package tests

import (
	"fmt"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/anyclouds/sshutils"
	"github.com/SebastienDorgan/retry"
	"github.com/SebastienDorgan/talgo"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ssh"
	"strings"
	"time"
)

//TemplateManagerTestSuite test suite off api.Temp
type ServerManagerTestSuite struct {
	suite.Suite
	Prov api.Provider
}

func weight(tpl api.ServerTemplate) float64 {
	return float64(tpl.RAMSize)
}

func noError() retry.Condition {
	return func(v interface{}, e error) bool {
		return e == nil
	}
}

func DeleteAction(f func(v string) error, id string) retry.Action {
	return func() (v interface{}, e error) {
		err := f(id)
		return nil, err
	}
}

func WilfulDelete(f func(v string) error, id string) error {
	return retry.With(DeleteAction(f, id)).Every(20 * time.Second).For(2 * time.Minute).Until(noError()).Go().LastError
}

func (s *ServerManagerTestSuite) CreateNetwork(netm api.NetworkManager) (network *api.Network, subnet *api.Subnet, err error) {
	network, err = netm.CreateNetwork(&api.NetworkOptions{
		CIDR: "10.0.0.0/16",
		Name: "Test Network",
	})
	if err != nil {
		return nil, nil, err
	}

	subnet, err = netm.CreateSubnet(&api.SubnetOptions{
		NetworkID: network.ID,
		Name:      "Test subnet",
		CIDR:      "10.0.0.0/24",
		IPVersion: api.IPVersion4,
	})
	if err != nil {
		return nil, nil, err
	}
	return
}

func (s *ServerManagerTestSuite) CreateSecurityGroup(sgm api.SecurityGroupManager, network *api.Network) (*api.SecurityGroup, error) {
	sg, err := sgm.Create(&api.SecurityGroupOptions{
		Name:        "TestSG",
		Description: "Test security group",
		NetworkID:   network.ID,
	})
	if err != nil {
		return nil, err
	}
	_, _ = sgm.AddRule(&api.SecurityRuleOptions{
		SecurityGroupID: sg.ID,
		Direction:       api.RuleDirectionIngress,
		PortRange:       api.PortRange{From: 22, To: 22},
		Protocol:        api.ProtocolTCP,
		CIDR:            "0.0.0.0/0",
		Description:     "Grant ingress ssh trafic",
	})
	_, _ = sgm.AddRule(&api.SecurityRuleOptions{
		SecurityGroupID: sg.ID,
		Direction:       api.RuleDirectionIngress,
		PortRange:       api.PortRange{From: 1, To: 65535},
		Protocol:        api.ProtocolICMP,
		CIDR:            "0.0.0.0/0",
		Description:     "Grant ICMP",
	})

	return sg, nil

}

func (s *ServerManagerTestSuite) SelectTemplate(tpm api.ServerTemplateManager) (*api.ServerTemplate, error) {
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
		if weight(tpls[indexes[i]]) < weight(tpls[indexes[j]]) {
			return i
		} else {
			return j
		}
	})
	return &tpls[indexes[selected]], nil
}

func CheckImageName(img *api.Image, os, version string) bool {
	name := strings.ToUpper(img.Name)
	return strings.Contains(name, os) && strings.Contains(name, version)
}

func (s *ServerManagerTestSuite) FindImage(imm api.ImageManager, tpl *api.ServerTemplate) (*api.Image, error) {
	imgs, err := imm.List()
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

//TestServerTemplateManager Canonical test for ServerTemplateManager implementation
func (s *ServerManagerTestSuite) TestServerManagerOnDemandInstance() {
	kp, err := sshutils.CreateKeyPair(4096)
	assert.NoError(s.T(), err)

	netm := s.Prov.GetNetworkManager()
	net, subnet, err := s.CreateNetwork(netm)
	assert.NoError(s.T(), err)

	sgm := s.Prov.GetSecurityGroupManager()
	sg, err := s.CreateSecurityGroup(sgm, net)
	assert.NoError(s.T(), err)

	tpm := s.Prov.GetTemplateManager()
	tpl, err := s.SelectTemplate(tpm)
	assert.NoError(s.T(), err)

	imm := s.Prov.GetImageManager()
	img, err := s.FindImage(imm, tpl)
	assert.NoError(s.T(), err)

	srvMgr := s.Prov.GetServerManager()
	server, err := srvMgr.Create(&api.CreateServerOptions{
		Name:       "test_server",
		TemplateID: tpl.ID,
		ImageID:    img.ID,
		SecurityGroups: []string{
			sg.ID,
		},
		Subnets: []string{
			subnet.ID,
		},
		PublicIP:        true,
		BootstrapScript: nil,
		KeyPair:         kp,
	})
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), server.Name, "test_server")
	assert.Equal(s.T(), server.ImageID, img.ID)
	assert.Equal(s.T(), server.KeyPairName, "kptest")
	assert.Equal(s.T(), len(server.PrivateIPs[api.IPVersion4]), 1)
	assert.Equal(s.T(), server.State, api.ServerReady)
	assert.Equal(s.T(), server.LeasingType, api.LeasingTypeOnDemand)
	assert.NotEmpty(s.T(), server.PublicIPv4)
	assert.Equal(s.T(), server.SecurityGroups[0], sg.ID)
	assert.Equal(s.T(), server.TemplateID, tpl.ID)
	err = srvMgr.Delete(server.ID)
	assert.NoError(s.T(), err)
	err = WilfulDelete(sgm.Delete, sg.ID)
	assert.NoError(s.T(), err)
	err = WilfulDelete(netm.DeleteSubnet, subnet.ID)
	assert.NoError(s.T(), err)
	err = WilfulDelete(netm.DeleteNetwork, net.ID)
	assert.NoError(s.T(), err)
}

//TestServerTemplateManager Canonical test for ServerTemplateManager implementation
func (s *ServerManagerTestSuite) TestServerManagerSpotInstance() {
	kp, err := sshutils.CreateKeyPair(4096)
	assert.NoError(s.T(), err)

	netm := s.Prov.GetNetworkManager()
	net, subnet, err := s.CreateNetwork(netm)
	assert.NoError(s.T(), err)

	sgm := s.Prov.GetSecurityGroupManager()
	sg, err := s.CreateSecurityGroup(sgm, net)
	assert.NoError(s.T(), err)

	tpm := s.Prov.GetTemplateManager()
	tpl, err := s.SelectTemplate(tpm)
	assert.NoError(s.T(), err)

	imm := s.Prov.GetImageManager()
	img, err := s.FindImage(imm, tpl)
	assert.NoError(s.T(), err)

	srvMgr := s.Prov.GetServerManager()
	server, err := srvMgr.Create(&api.CreateServerOptions{
		Name:       "test_server",
		TemplateID: tpl.ID,
		ImageID:    img.ID,
		SecurityGroups: []string{
			sg.ID,
		},
		Subnets: []string{
			subnet.ID,
		},
		PublicIP:        true,
		BootstrapScript: nil,
		KeyPair:         kp,
		SpotServerOptions: &api.SpotServerOptions{
			HourlyPrice: tpl.OneDemandPrice / 4,
			Duration:    0,
		},
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), server.Name, "test_server")
	assert.Equal(s.T(), server.ImageID, img.ID)
	assert.Equal(s.T(), server.KeyPairName, "kptest")
	assert.Equal(s.T(), len(server.PrivateIPs[api.IPVersion4]), 1)
	assert.Equal(s.T(), server.State, api.ServerReady)
	assert.Equal(s.T(), server.LeasingType, api.LeasingTypeOnDemand)
	assert.NotEmpty(s.T(), server.PublicIPv4)
	assert.Equal(s.T(), server.SecurityGroups[0], sg.ID)
	assert.Equal(s.T(), server.TemplateID, tpl.ID)

	auth, err := kp.AuthMethod()
	clt, err := sshutils.CreateClient(&sshutils.SSHConfig{
		Addr: fmt.Sprintf("%s:%d", server.PublicIPv4, 22),
		ClientConfig: &ssh.ClientConfig{
			Config:          ssh.Config{},
			User:            "ubuntu",
			Auth:            []ssh.AuthMethod{auth},
			Timeout:         0,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
		Proxy: nil,
	})
	println(fmt.Sprintf("%s:%d", server.PublicIPv4, 22))
	assert.NoError(s.T(), err)
	session, err := clt.NewSession()
	assert.NoError(s.T(), err)
	resp, err := session.Output("hostname")
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), resp)
	fmt.Println("hostname", string(resp))
	_ = session.Close()
	err = srvMgr.Delete(server.ID)
	assert.NoError(s.T(), err)
	err = WilfulDelete(sgm.Delete, sg.ID)
	assert.NoError(s.T(), err)
	err = WilfulDelete(netm.DeleteSubnet, subnet.ID)
	assert.NoError(s.T(), err)
	err = WilfulDelete(netm.DeleteNetwork, net.ID)
	assert.NoError(s.T(), err)
}
