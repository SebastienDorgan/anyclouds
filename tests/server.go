package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ssh"
	"sort"
)

//TemplateManagerTestSuite test suite off api.Temp
type ServerManagerTestSuite struct {
	suite.Suite
	Prov api.Provider
}

func weight(tpl api.ServerTemplate) float64 {
	return float64(tpl.NumberOfCPUCore) + float64(tpl.RAMSize)/4. + float64(tpl.SystemDiskSize)/10.
}

func findImg(tpl api.ServerTemplate, imgs []api.Image) *api.Image {
	for _, img := range imgs {
		if img.MinDisk < tpl.SystemDiskSize && img.MinRAM < tpl.RAMSize {
			return &img
		}
	}
	return nil
}

//TestServerTemplateManager Canonical test for ServerTemplateManager implementation
func (s *ServerManagerTestSuite) TestServerManager() {
	kpMgr := s.Prov.GetKeyPairManager()
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := privateKey.PublicKey
	pub, _ := ssh.NewPublicKey(&publicKey)
	pubBytes := ssh.MarshalAuthorizedKey(pub)
	err := kpMgr.Import("kptest", pubBytes)
	assert.NoError(s.T(), err)
	defer kpMgr.Delete("kptest")

	srvMgr := s.Prov.GetServerManager()
	tplMgr := s.Prov.GetTemplateManager()
	tpls, err := tplMgr.List()
	assert.NoError(s.T(), err)
	sort.Slice(tpls, func(i, j int) bool {
		return weight(tpls[i]) < weight(tpls[j])
	})
	tpl := tpls[0]
	imgMgr := s.Prov.GetImageManager()
	imgs, err := imgMgr.List()
	assert.NoError(s.T(), err)
	img := findImg(tpl, imgs)
	netMgr := s.Prov.GetNetworkManager()
	net, err := netMgr.CreateNetwork(&api.NetworkOptions{
		CIDR: "10.0.0.0/24",
	})
	assert.NoError(s.T(), err)
	defer RunSilent(netMgr.DeleteNetwork, net.ID)
	subNet, err := netMgr.CreateSubnet(&api.SubnetOptions{
		NetworkID: net.ID,
		Name:      "Test subnet",
		CIDR:      "10.0.0.0/24",
		IPVersion: api.IPVersion4,
	})
	assert.NoError(s.T(), err)
	sgMgr := s.Prov.GetSecurityGroupManager()
	sg, err := sgMgr.Create(&api.SecurityGroupOptions{
		Name:        "TestSG",
		Description: "Test security group",
		NetworkID:   net.ID,
	})
	assert.NoError(s.T(), err)
	defer RunSilent(sgMgr.Delete, sg.ID)

	server, err := srvMgr.Create(&api.CreateServerOptions{
		Name:       "test_server",
		TemplateID: tpls[0].ID,
		ImageID:    img.ID,
		SecurityGroups: []string{
			sg.ID,
		},
		Subnets: []string{
			subNet.ID,
		},
		PublicIP:        true,
		BootstrapScript: nil,
		KeyPairName:     "kptest",
		LeasingType:     api.LeasingTypeOnDemand,
	})
	assert.NoError(s.T(), err)
	defer RunSilent(srvMgr.Delete, server.ID)
	assert.Equal(s.T(), server.Name, "test_server")
	assert.Equal(s.T(), server.ImageID, img.ID)
	assert.Equal(s.T(), server.KeyPairName, "kptest")
	assert.Equal(s.T(), len(server.PrivateIPs[api.IPVersion4]), 1)
	assert.Equal(s.T(), server.State, api.ServerReady)
	assert.Equal(s.T(), server.LeasingType, api.LeasingTypeOnDemand)
	assert.NotEmpty(s.T(), server.PublicIPv4)
	assert.Equal(s.T(), server.SecurityGroups[0], sg.ID)
	assert.Equal(s.T(), server.TemplateID, tpl.ID)

}
