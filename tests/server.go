package tests

import (
	"fmt"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/anyclouds/sshutils"
	"github.com/SebastienDorgan/talgo"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ssh"
	"strings"
)

//ServerManagerTestSuite test suite of api.ServerManager
type ServerManagerTestSuite struct {
	suite.Suite
	Prov api.Provider
}

func weight(tpl api.ServerTemplate) float64 {
	return float64(tpl.RAMSize)
}

func (s *ServerManagerTestSuite) CreateNetwork(mgr api.NetworkManager) (network *api.Network, subnet *api.Subnet, err error) {
	network, err = mgr.CreateNetwork(api.CreateNetworkOptions{
		CIDR: "10.0.0.0/16",
		Name: "Test Network",
	})
	if err != nil {
		return nil, nil, err
	}

	subnet, err = mgr.CreateSubnet(api.CreateSubnetOptions{
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
	sg, err := sgm.Create(api.SecurityGroupOptions{
		Name:        "TestSG",
		Description: "Test security group",
		NetworkID:   network.ID,
	})
	if err != nil {
		return nil, err
	}
	_, _ = sgm.AddSecurityRule(api.AddSecurityRuleOptions{
		SecurityGroupID: sg.ID,
		Direction:       api.RuleDirectionIngress,
		PortRange:       api.PortRange{From: 22, To: 22},
		Protocol:        api.ProtocolTCP,
		CIDR:            "0.0.0.0/0",
		Description:     "Grant ingress ssh traffic",
	})
	_, _ = sgm.AddSecurityRule(api.AddSecurityRuleOptions{
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
		if weight(templates[indexes[i]]) < weight(templates[indexes[j]]) {
			return i
		}
		return j
	})
	return &templates[indexes[selected]], nil
}

func CheckImageName(img *api.Image, os, version string) bool {
	name := strings.ToUpper(img.Name)
	return strings.Contains(name, os) && strings.Contains(name, version)
}

func (s *ServerManagerTestSuite) FindImage(imm api.ImageManager, tpl *api.ServerTemplate) (*api.Image, error) {
	images, err := imm.List()
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

//TestServerManagerOnDemandInstance Canonical test for ServerTemplateManager implementation
func (s *ServerManagerTestSuite) TestServerManagerOnDemandInstance() {
	kp, err := sshutils.CreateKeyPair(4096)
	assert.NoError(s.T(), err)

	mgr := s.Prov.GetNetworkManager()
	net, subnet, err := s.CreateNetwork(mgr)
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
	server, err := srvMgr.Create(api.CreateServerOptions{
		Name:                 "test_server",
		TemplateID:           tpl.ID,
		ImageID:              img.ID,
		DefaultSecurityGroup: sg.ID,
		Subnets: []api.Subnet{
			*subnet,
		},
		BootstrapScript: nil,
		KeyPair:         *kp,
	})
	assert.NoError(s.T(), err)
	publicIP, err := s.Prov.GetPublicIPAddressManager().Create(api.CreatePublicIPOptions{
		Name: "ip",
	})
	assert.NoError(s.T(), err)
	err = s.Prov.GetPublicIPAddressManager().Associate(api.AssociatePublicIPOptions{
		PublicIPId: publicIP.ID,
		ServerID:   server.ID,
		SubnetID:   subnet.ID,
	})
	assert.NoError(s.T(), err)
	nis, err := s.Prov.GetNetworkInterfaceManager().List(&api.ListNetworkInterfacesOptions{
		ServerID: &server.ID,
	})
	assert.Equal(s.T(), server.Name, "test_server")
	assert.Equal(s.T(), server.ImageID, img.ID)

	assert.Equal(s.T(), server.State, api.ServerReady)
	assert.Equal(s.T(), server.LeasingType, api.LeasingTypeOnDemand)
	assert.NotEmpty(s.T(), nis[0].PublicIPAddress)
	assert.Equal(s.T(), server.TemplateID, tpl.ID)
	err = srvMgr.Delete(server.ID)
	assert.NoError(s.T(), err)
	err = s.Prov.GetPublicIPAddressManager().Delete(publicIP.ID)
	assert.NoError(s.T(), err)
	err = sgm.Delete(sg.ID)
	assert.NoError(s.T(), err)
	err = mgr.DeleteSubnet(net.ID, subnet.ID)
	assert.NoError(s.T(), err)
	err = mgr.DeleteNetwork(net.ID)
	assert.NoError(s.T(), err)
}

//TestServerManagerSpotInstance Canonical test for ServerTemplateManager implementation
func (s *ServerManagerTestSuite) TestServerManagerSpotInstance() {
	kp, err := sshutils.CreateKeyPair(4096)
	assert.NoError(s.T(), err)

	mgr := s.Prov.GetNetworkManager()
	net, subnet, err := s.CreateNetwork(mgr)
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
	server, err := srvMgr.Create(api.CreateServerOptions{
		Name:                 "test_server",
		TemplateID:           tpl.ID,
		ImageID:              img.ID,
		DefaultSecurityGroup: sg.ID,
		Subnets: []api.Subnet{
			*subnet,
		},
		BootstrapScript: nil,
		KeyPair:         *kp,
		LowPriorityServerOptions: &api.LowPriorityServerOptions{
			HourlyPrice: tpl.OneDemandPrice / 4,
			Duration:    0,
		},
	})
	assert.NoError(s.T(), err)
	publicIP, err := s.Prov.GetPublicIPAddressManager().Create(api.CreatePublicIPOptions{
		Name: "ip",
	})
	assert.NoError(s.T(), err)
	err = s.Prov.GetPublicIPAddressManager().Associate(api.AssociatePublicIPOptions{
		PublicIPId: publicIP.ID,
		ServerID:   server.ID,
		SubnetID:   subnet.ID,
	})
	assert.NoError(s.T(), err)
	nis, err := s.Prov.GetNetworkInterfaceManager().List(&api.ListNetworkInterfacesOptions{
		ServerID: &server.ID,
	})
	assert.Equal(s.T(), server.Name, "test_server")
	assert.Equal(s.T(), server.ImageID, img.ID)

	auth, err := kp.AuthMethod()
	clt, err := sshutils.CreateClient(&sshutils.SSHConfig{
		Addr: fmt.Sprintf("%s:%d", nis[0].PublicIPAddress, 22),
		ClientConfig: &ssh.ClientConfig{
			Config:          ssh.Config{},
			User:            "ubuntu",
			Auth:            []ssh.AuthMethod{auth},
			Timeout:         0,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
		Proxy: nil,
	})
	println(fmt.Sprintf("%s:%d", nis[0].PublicIPAddress, 22))
	assert.NoError(s.T(), err)
	if clt != nil {
		session, err := clt.NewSession()
		assert.NoError(s.T(), err)
		resp, err := session.Output("hostname")
		assert.NoError(s.T(), err)
		assert.NotEmpty(s.T(), resp)
		fmt.Println("hostname", string(resp))
		_ = session.Close()
	}

	err = srvMgr.Delete(server.ID)
	assert.NoError(s.T(), err)
	err = sgm.Delete(sg.ID)
	assert.NoError(s.T(), err)
	err = s.Prov.GetPublicIPAddressManager().Delete(publicIP.ID)
	assert.NoError(s.T(), err)
	err = mgr.DeleteSubnet(net.ID, subnet.ID)
	assert.NoError(s.T(), err)
	err = mgr.DeleteNetwork(net.ID)
	assert.NoError(s.T(), err)
}
