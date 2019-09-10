package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/pkg/errors"
)

type NetworkManager struct {
	Provider *Provider
}

func (mgr *NetworkManager) CreateNetwork(options api.NetworkOptions) (*api.Network, error) {
	future, err := mgr.Provider.VirtualNetworksClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), options.Name, network.VirtualNetwork{
		Location: &mgr.Provider.Configuration.Location,
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{
				AddressPrefixes: &[]string{options.CIDR},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network %s", options.Name)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.VirtualNetworksClient.Client)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network %s", options.Name)
	}
	n, err := future.Result(mgr.Provider.VirtualNetworksClient)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network %s", options.Name)
	}

	return &api.Network{
		ID:   *n.Name,
		Name: *n.Name,
		CIDR: (*n.VirtualNetworkPropertiesFormat.AddressSpace.AddressPrefixes)[0],
	}, nil
}

func (mgr *NetworkManager) DeleteNetwork(id string) error {
	future, err := mgr.Provider.VirtualNetworksClient.Delete(context.Background(), mgr.resourceGroup(), id)
	if err != nil {
		return errors.Wrapf(err, "error deleting network %s", id)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.VirtualNetworksClient.Client)
	return errors.Wrapf(err, "error deleting network %s", id)
}

func (mgr *NetworkManager) resourceGroup() string {
	return mgr.Provider.Configuration.ResourceGroupName
}

func (mgr *NetworkManager) ListNetworks() ([]api.Network, error) {
	list, err := mgr.Provider.VirtualNetworksClient.List(context.Background(), mgr.resourceGroup())
	if err != nil {
		return nil, errors.Wrap(err, "error listing networks")
	}
	var nets []api.Network
	for _, n := range list.Values() {
		nets = append(nets, api.Network{
			ID:   *n.Name,
			Name: *n.Name,
			CIDR: (*n.VirtualNetworkPropertiesFormat.AddressSpace.AddressPrefixes)[0],
		})
	}
	return nets, nil
}

func (mgr *NetworkManager) GetNetwork(id string) (*api.Network, error) {
	n, err := mgr.Provider.VirtualNetworksClient.Get(context.Background(), mgr.resourceGroup(), id, "")
	if err != nil {
		return nil, errors.Wrapf(err, "error getting networks %s", id)
	}
	return &api.Network{
		ID:   *n.Name,
		Name: *n.Name,
		CIDR: *n.Tags["cidr"],
	}, nil
}

func (mgr *NetworkManager) CreateSubnet(options api.SubnetOptions) (*api.Subnet, error) {
	future, err := mgr.Provider.SubnetsClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), options.NetworkID, options.Name, network.Subnet{
		SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
			AddressPrefix: &options.CIDR,
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error creating subnet %s", options.Name)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.SubnetsClient.Client)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating subnet %s", options.Name)
	}
	sn, err := future.Result(mgr.Provider.SubnetsClient)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating subnet %s", options.Name)
	}
	return &api.Subnet{
		ID:        *sn.Name,
		NetworkID: options.NetworkID,
		Name:      *sn.Name,
		CIDR:      *sn.SubnetPropertiesFormat.AddressPrefix,
		IPVersion: api.IPVersion4,
	}, nil
}

func (mgr *NetworkManager) DeleteSubnet(networkID, subnetID string) error {
	future, err := mgr.Provider.SubnetsClient.Delete(context.Background(), mgr.resourceGroup(), networkID, subnetID)
	if err != nil {
		return errors.Wrapf(err, "error deleting subnet %s", subnetID)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.SubnetsClient.Client)
	return errors.Wrapf(err, "error deleting subnet %s", subnetID)
}

func (mgr *NetworkManager) ListSubnets(networkID string) ([]api.Subnet, error) {
	n, err := mgr.Provider.VirtualNetworksClient.Get(context.Background(), mgr.resourceGroup(), networkID, "")
	if err != nil {
		return nil, errors.Wrapf(err, "error getting networks %s", networkID)
	}
	var subnets []api.Subnet
	for _, sn := range *n.Subnets {
		subnets = append(subnets, api.Subnet{
			ID:        *n.Name,
			NetworkID: networkID,
			Name:      *sn.Name,
			CIDR:      *sn.SubnetPropertiesFormat.AddressPrefix,
			IPVersion: api.IPVersion4,
		})
	}
	return subnets, nil
}

func (mgr *NetworkManager) GetSubnet(networkID, subnetID string) (*api.Subnet, error) {
	sn, err := mgr.Provider.SubnetsClient.Get(context.Background(), mgr.resourceGroup(), networkID, subnetID, "")
	if err != nil {
		return nil, errors.Wrapf(err, "error getting subnet %s", subnetID)
	}
	return &api.Subnet{
		ID:        *sn.Name,
		NetworkID: networkID,
		Name:      *sn.Name,
		CIDR:      *sn.SubnetPropertiesFormat.AddressPrefix,
		IPVersion: api.IPVersion4,
	}, nil
}
