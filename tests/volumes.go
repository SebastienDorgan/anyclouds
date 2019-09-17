package tests

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/anyclouds/sshutils"
	"github.com/SebastienDorgan/talgo"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VolumeManagerTestSuite struct {
	suite.Suite
	Prov api.Provider
}

func (s *VolumeManagerTestSuite) SelectTemplate() (*api.ServerTemplate, error) {
	cpu := 4
	ram := 15000
	tplMgr := s.Prov.GetTemplateManager()
	templates, err := tplMgr.List()
	if err != nil {
		return nil, err
	}
	indexes := talgo.FindAll(len(templates), func(i int) bool {
		if templates[i].NumberOfCPUCore == cpu && templates[i].RAMSize >= ram && templates[i].GPUInfo == nil && templates[i].Arch == api.ArchAmd64 {
			return true
		}
		return false
	})
	if len(indexes) == 0 {
		return nil, errors.Errorf("No matching template find")
	}
	selected := talgo.Select(len(indexes), func(i, j int) int {
		if templates[indexes[i]].OneDemandPrice < templates[indexes[j]].OneDemandPrice {
			return i
		}
		return j
	})
	return &templates[indexes[selected]], nil
}

func (s *VolumeManagerTestSuite) FindImage(tpl *api.ServerTemplate) (*api.Image, error) {
	images, err := s.Prov.GetImageManager().List()
	if err != nil {
		return nil, err
	}
	os := "UBUNTU"
	version := "18.04"
	for _, img := range images {
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
	kp, err := sshutils.CreateKeyPair(4096)
	assert.NoError(s.T(), err)
	n, err := s.getDefaultNetwork()
	assert.NoError(s.T(), err)
	sn, err := s.Prov.GetNetworkManager().CreateSubnet(api.CreateSubnetOptions{
		NetworkID: n.ID,
		Name:      "subnet",
		CIDR:      n.CIDR,
		IPVersion: api.IPVersion4,
	})
	assert.NoError(s.T(), err)
	server, err := s.Prov.GetServerManager().Create(api.CreateServerOptions{
		Name:            "instance_with_volume",
		TemplateID:      tpl.ID,
		ImageID:         img.ID,
		Subnets:         []api.Subnet{*sn},
		BootstrapScript: nil,
		KeyPair:         *kp,
	})
	assert.NoError(s.T(), err)
	ni, err := s.Prov.GetNetworkInterfaceManager().List(&api.ListNetworkInterfacesOptions{
		NetworkID: &n.ID,
		SubnetID:  &sn.ID,
		ServerID:  &server.ID,
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(ni))
	assert.Equal(s.T(), n.ID, ni[0].NetworkID)
	assert.Equal(s.T(), sn.ID, ni[0].SubnetID)
	assert.Equal(s.T(), server.ID, ni[0].ServerID)
	sgID := ni[0].SecurityGroupID
	rule, err := s.Prov.GetSecurityGroupManager().AddSecurityRule(api.AddSecurityRuleOptions{
		SecurityGroupID: sgID,
		Direction:       api.RuleDirectionIngress,
		PortRange:       api.PortRange{From: 22, To: 22},
		Protocol:        api.ProtocolTCP,
		CIDR:            "0.0.0.0/0",
		Description:     "grant ssh access",
	})
	assert.NoError(s.T(), err)

	v, err := s.Prov.GetVolumeManager().Create(api.CreateVolumeOptions{
		Name:        "my volume",
		Size:        5,
		MinIOPS:     250,
		MinDataRate: 250,
	})
	assert.NoError(s.T(), err)
	att, err := s.Prov.GetVolumeManager().Attach(api.AttachVolumeOptions{
		VolumeID:   v.ID,
		ServerID:   server.ID,
		DevicePath: "/dev/sdh",
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "/dev/sdh", att.Device)
	err = s.Prov.GetVolumeManager().Detach(api.DetachVolumeOptions{
		VolumeID: v.ID,
		ServerID: server.ID,
		Force:    true,
	})
	assert.NoError(s.T(), err)
	err = s.Prov.GetVolumeManager().Delete(v.ID)
	assert.NoError(s.T(), err)
	err = s.Prov.GetSecurityGroupManager().RemoveSecurityRule(sgID, rule.ID)
	assert.NoError(s.T(), err)
	err = s.Prov.GetServerManager().Delete(server.ID)
	assert.NoError(s.T(), err)
	err = s.Prov.GetNetworkManager().DeleteSubnet(n.ID, sn.ID)
	assert.NoError(s.T(), err)

}
