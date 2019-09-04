package tests

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/anyclouds/sshutils"
	"github.com/SebastienDorgan/talgo"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
)

type VolumeManagerTestSuite struct {
	suite.Suite
	Prov api.Provider
}

func (s *VolumeManagerTestSuite) ImportKeyPair(name string) (*sshutils.KeyPair, error) {
	kp, err := sshutils.CreateKeyPair(4096)
	if err != nil {
		return nil, err
	}
	_ = ioutil.WriteFile("private_key", kp.PrivateKey, 0640)
	_ = ioutil.WriteFile("public_key", kp.PublicKey, 0640)
	return kp, s.Prov.GetKeyPairManager().Import(name, kp.PublicKey)
}

func (s *VolumeManagerTestSuite) SelectTemplate() (*api.ServerTemplate, error) {
	cpu := 4
	ram := 15000
	tplMgr := s.Prov.GetTemplateManager()
	tpls, err := tplMgr.List()
	if err != nil {
		return nil, err
	}
	indexes := talgo.FindAll(len(tpls), func(i int) bool {
		if tpls[i].NumberOfCPUCore == cpu && tpls[i].RAMSize >= ram && tpls[i].GPUInfo == nil && tpls[i].Arch == api.ArchAmd64 {
			return true
		}
		return false
	})
	if len(indexes) == 0 {
		return nil, errors.Errorf("No matching template find")
	}
	selected := talgo.Select(len(indexes), func(i, j int) int {
		if tpls[indexes[i]].OneDemandPrice < tpls[indexes[j]].OneDemandPrice {
			return i
		} else {
			return j
		}
	})
	return &tpls[indexes[selected]], nil
}

func (s *VolumeManagerTestSuite) FindImage(tpl *api.ServerTemplate) (*api.Image, error) {
	imgs, err := s.Prov.GetImageManager().List()
	if err != nil {
		return nil, err
	}
	os := "UBUNTU"
	version := "18.04"
	for _, img := range imgs {
		if img.MinDisk < tpl.SystemDiskSize && img.MinRAM < tpl.RAMSize && CheckImageName(&img, os, version) {
			return &img, err
		}
	}
	return nil, errors.Errorf("Enable to fin image fitting template %s", tpl.Name)
}
func (s *VolumeManagerTestSuite) getDefaultNetwork() (*api.Network, error) {
	nets, err := s.Prov.GetNetworkManager().ListNetworks()
	if err != nil {
		return nil, err
	}
	for _, n := range nets {
		if n.Name == "" {
			return &n, err
		}
	}
	return nil, errors.Errorf("default network not found")
}

func (s *VolumeManagerTestSuite) getDefaultSecurityGroup() (*api.SecurityGroup, error) {
	sgs, err := s.Prov.GetSecurityGroupManager().List()
	if err != nil {
		return nil, err
	}
	for _, n := range sgs {
		if n.Name == "default" {
			return &n, err
		}
	}
	return nil, errors.Errorf("default security group not found")
}

func (s *VolumeManagerTestSuite) TestVolumeManager() {
	tpl, err := s.SelectTemplate()
	assert.NoError(s.T(), err)
	img, err := s.FindImage(tpl)
	assert.NoError(s.T(), err)
	_, err = s.ImportKeyPair("kp_test")
	assert.NoError(s.T(), err)
	n, err := s.getDefaultNetwork()
	assert.NoError(s.T(), err)
	sn, err := s.Prov.GetNetworkManager().CreateSubnet(&api.SubnetOptions{
		NetworkID: n.ID,
		Name:      "subnet",
		CIDR:      n.CIDR,
		IPVersion: api.IPVersion4,
	})
	assert.NoError(s.T(), err)
	sg, err := s.getDefaultSecurityGroup()
	assert.NoError(s.T(), err)
	rule, err := s.Prov.GetSecurityGroupManager().AddRule(&api.SecurityRuleOptions{
		SecurityGroupID: sg.ID,
		Direction:       api.RuleDirectionIngress,
		PortRange:       api.PortRange{From: 22, To: 22},
		Protocol:        api.ProtocolTCP,
		CIDR:            "0.0.0.0/0",
		Description:     "grant ssh acces",
	})
	assert.NoError(s.T(), err)
	server, err := s.Prov.GetServerManager().Create(&api.CreateServerOptions{
		Name:       "instance_with_volume",
		TemplateID: tpl.ID,
		ImageID:    img.ID,
		//SecurityGroups:  []string{sg.ID},
		Subnets:         []string{sn.ID},
		PublicIP:        true,
		BootstrapScript: nil,
		KeyPairName:     "kp_test",
		LeasingType:     api.LeasingTypeOnDemand,
		LeaseDuration:   0,
	})
	assert.NoError(s.T(), err)

	assert.NoError(s.T(), err)
	v, err := s.Prov.GetVolumeManager().Create(&api.VolumeOptions{
		Name:        "my volume",
		Size:        5,
		MinIOPS:     250,
		MinDataRate: 250,
	})
	assert.NoError(s.T(), err)
	att, err := s.Prov.GetVolumeManager().Attach(v.ID, server.ID, "/dev/sdh")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "/dev/sdh", att.Device)
	err = s.Prov.GetVolumeManager().Detach(v.ID, server.ID, true)
	assert.NoError(s.T(), err)
	err = s.Prov.GetVolumeManager().Delete(v.ID)
	assert.NoError(s.T(), err)
	err = s.Prov.GetSecurityGroupManager().DeleteRule(rule.ID)
	assert.NoError(s.T(), err)
	err = s.Prov.GetServerManager().Delete(server.ID)
	assert.NoError(s.T(), err)
	err = s.Prov.GetKeyPairManager().Delete("kp_test")
	assert.NoError(s.T(), err)
	err = WilfulDelete(s.Prov.GetNetworkManager().DeleteSubnet, sn.ID)
	assert.NoError(s.T(), err)

}
