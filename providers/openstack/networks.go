package openstack

import (
	"fmt"
	"github.com/SebastienDorgan/anyclouds/api"
	gc "github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
	"github.com/pkg/errors"
)

//NetworkManager defines networking functions a anyclouds provider must provide
type NetworkManager struct {
	OpenStack *Provider
}

//CreateNetwork creates a network
func (mgr *NetworkManager) CreateNetwork(options api.NetworkOptions) (*api.Network, error) {
	up := true
	opts := networks.CreateOpts{
		AdminStateUp: &up,
		Name:         options.Name,
		Description:  options.CIDR,
	}
	network, err := networks.Create(mgr.OpenStack.Network, opts).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error creating network")
	}
	_, err = mgr.createRouter(network.ID)
	if err != nil {
		_ = mgr.DeleteNetwork(network.ID)
		return nil, errors.Wrap(ProviderError(err), "Error creating network")
	}
	return &api.Network{
		ID:   network.ID,
		Name: network.Name,
		CIDR: network.Description,
	}, nil
}

//DeleteNetwork deletes the network identified by id
func (mgr *NetworkManager) DeleteNetwork(id string) error {
	r, err := mgr.findRouter(id)
	if err != nil {
		return errors.Wrap(ProviderError(err), "error deleting network")
	}
	if r != nil && len(r.ID) > 0 {
		err = mgr.deleteRouter(r.ID)
		if err != nil {
			return errors.Wrap(ProviderError(err), "error deleting network")
		}
	}

	err = networks.Delete(mgr.OpenStack.Network, id).ExtractErr()
	if err != nil {
		return errors.Wrap(ProviderError(err), "error deleting network")
	}
	return nil
}

func network(net *networks.Network) *api.Network {
	return &api.Network{
		ID:   net.ID,
		Name: net.Name,
		CIDR: net.Description,
	}
}

//ListNetworks lists networks
func (mgr *NetworkManager) ListNetworks() ([]api.Network, error) {
	opts := networks.ListOpts{}
	page, err := networks.List(mgr.OpenStack.Network, opts).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing network")
	}
	l, err := networks.ExtractNetworks(page)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing network")
	}
	var nets []api.Network
	for _, n := range l {
		nets = append(nets, *network(&n))
	}
	return nets, nil
}

//GetNetwork returns the configuration of the network identified by id
func (mgr *NetworkManager) GetNetwork(id string) (*api.Network, error) {
	n, err := networks.Get(mgr.OpenStack.Network, id).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting network")
	}
	return network(n), nil
}

func (mgr *NetworkManager) createRouter(networkID string) (*routers.Router, error) {
	up := true
	netID, err := networks.IDFromName(mgr.OpenStack.Network, mgr.OpenStack.ExternalNetworkName)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error creating subnet")
	}
	gwi := routers.GatewayInfo{
		NetworkID: netID,
	}

	createOpts := routers.CreateOpts{
		Name:         networkID,
		AdminStateUp: &up,
		GatewayInfo:  &gwi,
	}

	router, err := routers.Create(mgr.OpenStack.Network, createOpts).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error creating subnet")
	}
	return router, nil
}
func (mgr *NetworkManager) attachSubnetToRouter(routerID, subnetID string) error {
	_, err := routers.AddInterface(mgr.OpenStack.Network, routerID, routers.AddInterfaceOpts{
		SubnetID: subnetID,
	}).Extract()
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error creating subnet")
	}
	return nil
}
func (mgr *NetworkManager) deleteRouter(id string) error {
	return routers.Delete(mgr.OpenStack.Network, id).ExtractErr()
}

func (mgr *NetworkManager) findRouter(name string) (*routers.Router, error) {
	page, err := routers.List(mgr.OpenStack.Network, routers.ListOpts{Name: name}).AllPages()
	if err != nil {
		return nil, err
	}
	l, err := routers.ExtractRouters(page)
	if err != nil {
		return nil, err
	}
	if len(l) > 1 {
		return nil, fmt.Errorf("state error: subnet is associated to more than one router")
	}
	if len(l) == 0 {
		return nil, nil
	}
	return &l[0], nil
}

//CreateSubnet creates a subnet
func (mgr *NetworkManager) CreateSubnet(options api.SubnetOptions) (*api.Subnet, error) {
	dhcp := true
	opts := subnets.CreateOpts{
		NetworkID:  options.NetworkID,
		CIDR:       options.CIDR,
		IPVersion:  gc.IPVersion(options.IPVersion),
		Name:       options.Name,
		EnableDHCP: &dhcp,
	}
	router, err := mgr.findRouter(options.Name)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "error creating subnet")
	}
	// Execute the operation and get back a subnets.Subnet struct
	subnet, err := subnets.Create(mgr.OpenStack.Network, opts).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "error creating subnet")
	}

	err = mgr.attachSubnetToRouter(router.ID, subnet.ID)
	if err != nil {
		_ = mgr.DeleteSubnet(options.NetworkID, subnet.ID)
		return nil, errors.Wrap(ProviderError(err), "Error creating subnet")
	}

	return &api.Subnet{
		ID:        subnet.ID,
		Name:      subnet.Name,
		IPVersion: api.IPVersion(subnet.IPVersion),
		CIDR:      subnet.CIDR,
		NetworkID: subnet.NetworkID,
	}, nil
}

//DeleteSubnet deletes the subnet identified by id
func (mgr *NetworkManager) DeleteSubnet(networkID, subnetID string) error {
	err := subnets.Delete(mgr.OpenStack.Network, subnetID).ExtractErr()
	return errors.Wrap(ProviderError(err), "Error creating subnet: cannot delete router associated to subnet")
}

//ListSubnets lists the subnet
func (mgr *NetworkManager) ListSubnets(networkID string) ([]api.Subnet, error) {
	page, err := subnets.List(mgr.OpenStack.Network, subnets.ListOpts{
		NetworkID: networkID,
	}).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing subnets")
	}
	l, err := subnets.ExtractSubnets(page)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing subnets")
	}
	var res []api.Subnet
	for _, sn := range l {
		item := api.Subnet{
			CIDR:      sn.CIDR,
			ID:        sn.ID,
			IPVersion: api.IPVersion(sn.IPVersion),
			Name:      sn.Name,
			NetworkID: sn.NetworkID,
		}
		res = append(res, item)
	}
	return res, nil
}

func (mgr *NetworkManager) listAllSubnets() ([]api.Subnet, error) {
	page, err := subnets.List(mgr.OpenStack.Network, subnets.ListOpts{}).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing subnets")
	}
	l, err := subnets.ExtractSubnets(page)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing subnets")
	}
	var res []api.Subnet
	for _, sn := range l {
		item := api.Subnet{
			CIDR:      sn.CIDR,
			ID:        sn.ID,
			IPVersion: api.IPVersion(sn.IPVersion),
			Name:      sn.Name,
			NetworkID: sn.NetworkID,
		}
		res = append(res, item)
	}
	return res, nil
}

//GetSubnet returns the configuration of the subnet identified by id
func (mgr *NetworkManager) GetSubnet(networkID, subnetID string) (*api.Subnet, error) {
	sn, err := subnets.Get(mgr.OpenStack.Network, subnetID).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting subnet")
	}
	return &api.Subnet{
		CIDR:      sn.CIDR,
		ID:        sn.ID,
		IPVersion: api.IPVersion(sn.IPVersion),
		Name:      sn.Name,
		NetworkID: sn.NetworkID,
	}, nil
}
