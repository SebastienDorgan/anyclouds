package aws

import (
	"github.com/SebastienDorgan/anyclouds/api"
)

//NetworkManager defines networking functions a anyclouds provider must provide
type NetworkManager struct {
	AWS *Provider
}

//CreateNetwork creates a network
func (mgr *NetworkManager) CreateNetwork(options *api.NetworkOptions) (*api.Network, error) {
	return nil, nil
}

//DeleteNetwork deletes the netwok identified by id
func (mgr *NetworkManager) DeleteNetwork(id string) error {
	return nil
}

//ListNetworks lists networks
func (mgr *NetworkManager) ListNetworks() ([]api.Network, error) {
	return nil, nil
}

//GetNetwork returns the configuration of the network identified by id
func (mgr *NetworkManager) GetNetwork(id string) (*api.Network, error) {
	return nil, nil

}

//CreateSubnet creates a subnet
func (mgr *NetworkManager) CreateSubnet(options *api.SubnetOptions) (*api.Subnet, error) {
	return nil, nil
}

//DeleteSubnet deletes the subnet identified by id
func (mgr *NetworkManager) DeleteSubnet(id string) error {
	return nil
}

//ListSubnets lists the subnet
func (mgr *NetworkManager) ListSubnets(networkID string) ([]api.Subnet, error) {
	return nil, nil
}

//GetSubnet returns the configuration of the subnet identified by id
func (mgr *NetworkManager) GetSubnet(id string) (*api.Subnet, error) {
	return nil, nil
}
