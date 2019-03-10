package tests

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/stretchr/testify/assert"
)

//NetworkManager an api.NetworkManager implementation
var NetworkManager api.NetworkManager

//TestNetworks Canonical test for NetworkManager implementation
func TestNetworks(t *testing.T) {
	networks, err := NetworkManager.ListNetworks()
	assert.NoError(t, err)
	nnetworks := len(networks)

	opts := api.NetworkOptions{
		Name: "nettest",
	}
	net, err := NetworkManager.CreateNetwork(&opts)
	assert.NoError(t, err)

	networks, err = NetworkManager.ListNetworks()
	assert.Equal(t, nnetworks+1, len(networks))

	found := false
	for _, n := range networks {
		if n.ID == net.ID && n.Name == net.Name {
			found = true
			break
		}
	}
	assert.True(t, found)

	err = NetworkManager.DeleteNetwork(net.ID)
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
