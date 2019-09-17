package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/SebastienDorgan/anyclouds/api"
)

type NetworkManager struct {
	Provider *Provider
}

func (mgr *NetworkManager) CreateNetwork(options api.CreateNetworkOptions) (*api.Network, *api.CreateNetworkError) {
	n, err := mgr.createNetwork(options)
	return n, api.NewCreateNetworkError(err, options)
}

func (mgr *NetworkManager) createNetwork(options api.CreateNetworkOptions) (*api.Network, error) {
	future, err := mgr.Provider.BaseServices.VirtualNetworksClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), options.Name, network.VirtualNetwork{
		Location: &mgr.Provider.Configuration.Location,
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{
				AddressPrefixes: &[]string{options.CIDR},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.BaseServices.VirtualNetworksClient.Client)
	if err != nil {
		return nil, err
	}
	n, err := future.Result(mgr.Provider.BaseServices.VirtualNetworksClient)
	if err != nil {
		return nil, err
	}

	return &api.Network{
		ID:   *n.Name,
		Name: *n.Name,
		CIDR: (*n.VirtualNetworkPropertiesFormat.AddressSpace.AddressPrefixes)[0],
	}, nil
}

func (mgr *NetworkManager) DeleteNetwork(id string) *api.DeleteNetworkError {
	future, err := mgr.Provider.BaseServices.VirtualNetworksClient.Delete(context.Background(), mgr.resourceGroup(), id)
	if err != nil {
		return api.NewDeleteNetworkError(err, id)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.BaseServices.VirtualNetworksClient.Client)
	return api.NewDeleteNetworkError(err, id)
}

func (mgr *NetworkManager) resourceGroup() string {
	return mgr.Provider.Configuration.ResourceGroupName
}

func (mgr *NetworkManager) ListNetworks() ([]api.Network, *api.ListNetworksError) {
	list, err := mgr.Provider.BaseServices.VirtualNetworksClient.List(context.Background(), mgr.resourceGroup())
	if err != nil {
		return nil, api.NewListNetworksError(err)
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

func (mgr *NetworkManager) GetNetwork(id string) (*api.Network, *api.GetNetworkError) {
	n, err := mgr.Provider.BaseServices.VirtualNetworksClient.Get(context.Background(), mgr.resourceGroup(), id, "")
	if err != nil {
		return nil, api.NewGetNetworkError(err, id)
	}
	return &api.Network{
		ID:   *n.Name,
		Name: *n.Name,
		CIDR: *n.Tags["cidr"],
	}, nil
}

func (mgr *NetworkManager) CreateSubnet(options api.CreateSubnetOptions) (*api.Subnet, *api.CreateSubnetError) {
	future, err := mgr.Provider.BaseServices.SubnetsClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), options.NetworkID, options.Name, network.Subnet{
		SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
			AddressPrefix: &options.CIDR,
		},
	})
	if err != nil {
		return nil, api.NewCreateSubnetError(err, options)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.BaseServices.SubnetsClient.Client)
	if err != nil {
		return nil, api.NewCreateSubnetError(err, options)
	}
	sn, err := future.Result(mgr.Provider.BaseServices.SubnetsClient)
	if err != nil {
		return nil, api.NewCreateSubnetError(err, options)
	}
	return &api.Subnet{
		ID:        *sn.Name,
		NetworkID: options.NetworkID,
		Name:      *sn.Name,
		CIDR:      *sn.SubnetPropertiesFormat.AddressPrefix,
		IPVersion: api.IPVersion4,
	}, nil
}

func (mgr *NetworkManager) DeleteSubnet(networkID, subnetID string) *api.DeleteSubnetError {
	future, err := mgr.Provider.BaseServices.SubnetsClient.Delete(context.Background(), mgr.resourceGroup(), networkID, subnetID)
	if err != nil {
		return api.NewDeleteSubnetError(err, networkID, subnetID)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.BaseServices.SubnetsClient.Client)
	return api.NewDeleteSubnetError(err, networkID, subnetID)
}

func (mgr *NetworkManager) ListSubnets(networkID string) ([]api.Subnet, *api.ListSubnetsError) {
	n, err := mgr.Provider.BaseServices.VirtualNetworksClient.Get(context.Background(), mgr.resourceGroup(), networkID, "")
	if err != nil {
		return nil, api.NewListSubnetsError(err, networkID)
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

func (mgr *NetworkManager) GetSubnet(networkID, subnetID string) (*api.Subnet, *api.GetSubnetError) {
	sn, err := mgr.Provider.BaseServices.SubnetsClient.Get(context.Background(), mgr.resourceGroup(), networkID, subnetID, "")
	if err != nil {
		return nil, api.NewGetSubnetError(err, networkID, subnetID)
	}
	return &api.Subnet{
		ID:        *sn.Name,
		NetworkID: networkID,
		Name:      *sn.Name,
		CIDR:      *sn.SubnetPropertiesFormat.AddressPrefix,
		IPVersion: api.IPVersion4,
	}, nil
}
