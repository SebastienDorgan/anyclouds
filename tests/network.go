package tests

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

//NetworkManagerTestSuite test suite off api.NetworkManager
type NetworkManagerTestSuite struct {
	suite.Suite
	Mgr api.NetworkManager
}

//TestNetworks Canonical test for NetworkManager implementation
//func (s *NetworkManagerTestSuite) TestNetworks() {
//	Mgr := s.Mgr
//	nets, err := Mgr.ListNetworks()
//	l0 := len(nets)
//	n, err := Mgr.CreateNetwork(&api.CreateNetworkOptions{
//		CIDR: "10.0.0.0/16",
//		DeviceName: "test_net",
//	})
//	assert.NoError(s.T(), err)
//	nets, err = Mgr.ListNetworks()
//	assert.NoError(s.T(), err)
//	assert.Equal(s.T(), l0+1, len(nets))
//	found := false
//	for _, nn := range nets {
//		if nn.ID == n.ID {
//			found = true
//			assert.Equal(s.T(), n.CIDR, nn.CIDR)
//			assert.Equal(s.T(), n.DeviceName, nn.DeviceName)
//			break
//		}
//	}
//	ng, err := Mgr.GetNetwork(n.ID)
//	assert.NoError(s.T(), err)
//	assert.Equal(s.T(), ng.ID, n.ID)
//	assert.Equal(s.T(), ng.CIDR, n.CIDR)
//	assert.Equal(s.T(), ng.DeviceName, n.DeviceName)
//	assert.True(s.T(), found)
//	err = Mgr.DeleteNetwork(n.ID)
//	assert.NoError(s.T(), err)
//	nets, err = Mgr.ListNetworks()
//	assert.Equal(s.T(), l0, len(nets))
//	found = false
//	for _, nn := range nets {
//		if nn.ID == n.ID {
//			found = true
//			break
//		}
//	}
//	assert.False(s.T(), found)
//}

//TestSubnets canonical tests for subnets
func (s *NetworkManagerTestSuite) TestSubnets() {
	n, err := s.Mgr.CreateNetwork(api.CreateNetworkOptions{
		Name: "test_network",
		CIDR: "10.0.0.0/16",
	})
	assert.NoError(s.T(), err)
	sns, err := s.Mgr.ListSubnets(n.ID)
	assert.NoError(s.T(), err)
	l0 := len(sns)

	sn, err := s.Mgr.CreateSubnet(api.SubnetOptions{
		CIDR:      "10.0.1.0/24",
		IPVersion: api.IPVersion4,
		NetworkID: n.ID,
		Name:      "test_subnet",
	})
	assert.NoError(s.T(), err)

	tmp, err := s.Mgr.GetSubnet(n.ID, sn.ID)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), tmp.ID, sn.ID)
	assert.Equal(s.T(), tmp.CIDR, sn.CIDR)
	assert.Equal(s.T(), tmp.IPVersion, sn.IPVersion)
	assert.Equal(s.T(), tmp.NetworkID, sn.NetworkID)
	assert.Equal(s.T(), tmp.Name, sn.Name)

	sns, err = s.Mgr.ListSubnets(n.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), l0+1, len(sns))
	found := false
	for _, st := range sns {
		if st.ID == sn.ID {
			assert.Equal(s.T(), st.CIDR, sn.CIDR)
			assert.Equal(s.T(), st.IPVersion, sn.IPVersion)
			assert.Equal(s.T(), st.NetworkID, sn.NetworkID)
			assert.Equal(s.T(), st.Name, sn.Name)
			found = true
			break
		}
	}
	assert.True(s.T(), found)

	err = s.Mgr.DeleteSubnet(n.ID, sn.ID)
	assert.NoError(s.T(), err)

	sns, err = s.Mgr.ListSubnets(n.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), l0, len(sns))

	found = false
	for _, st := range sns {
		if st.ID == sn.ID {
			found = true
			break
		}
	}

	assert.False(s.T(), found)

	err = s.Mgr.DeleteNetwork(n.ID)
	assert.NoError(s.T(), err)

}
