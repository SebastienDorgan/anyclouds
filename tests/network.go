package tests

import (
	"testing"

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
func (s *NetworkManagerTestSuite) TestNetworks(t *testing.T) {
	networks, err := s.Mgr.ListNetworks()
	assert.NoError(t, err)
	nnetworks := len(networks)

	opts := api.NetworkOptions{
		Name: "nettest",
	}
	net, err := s.Mgr.CreateNetwork(&opts)
	assert.NoError(t, err)

	networks, err = s.Mgr.ListNetworks()
	assert.Equal(t, nnetworks+1, len(networks))

	found := false
	for _, n := range networks {
		if n.ID == net.ID && n.Name == net.Name {
			found = true
			break
		}
	}
	assert.True(t, found)

	err = s.Mgr.DeleteNetwork(net.ID)
	assert.NoError(t, err)
	assert.Equal(t, nnetworks, len(networks))

	found = false
	for _, n := range networks {
		if n.ID == net.ID && n.Name == net.Name {
			found = true
			break
		}
	}
	assert.False(t, found)

}
