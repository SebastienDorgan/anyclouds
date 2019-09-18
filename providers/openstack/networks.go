package openstack

import (
	"fmt"
	"github.com/SebastienDorgan/anyclouds/api"
	gc "github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"
)

//NetworkManager defines networking functions a anyclouds provider must provide
type NetworkManager struct {
	Refactor *Provider
}

//CreateNetwork creates a network
func (mgr *NetworkManager) CreateNetwork(options api.CreateNetworkOptions) (*api.Network, api.CreateNetworkError) {
	up := true
	opts := networks.CreateOpts{
		AdminStateUp: &up,
		Name:         options.Name,
		Description:  options.CIDR,
	}
	network, err := networks.Create(mgr.Refactor.BaseServices.Network, opts).Extract()
	if err != nil {
		return nil, api.NewCreateNetworkError(UnwrapOpenStackError(err), options)
	}
	_, err = mgr.createRouter(network.ID)
	if err != nil {
		err2 := mgr.DeleteNetwork(network.ID)
		err = api.NewErrorStackFromError(err, err2)
		return nil, api.NewCreateNetworkError(err, options)
	}
	return &api.Network{
		ID:   network.ID,
		Name: network.Name,
		CIDR: network.Description,
	}, nil
}

//DeleteNetwork deletes the network identified by id
func (mgr *NetworkManager) DeleteNetwork(id string) api.DeleteNetworkError {
	r, err := mgr.findRouter(id)
	if err != nil {
		return api.NewDeleteNetworkError(err, id)
	}
	if r != nil && len(r.ID) > 0 {
		err = mgr.deleteRouter(r.ID)
		if err != nil {
			return api.NewDeleteNetworkError(err, id)
		}
	}

	err = networks.Delete(mgr.Refactor.BaseServices.Network, id).ExtractErr()
	if err != nil {
		return api.NewDeleteNetworkError(err, id)
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
func (mgr *NetworkManager) ListNetworks() ([]api.Network, api.ListNetworksError) {
	opts := networks.ListOpts{}
	page, err := networks.List(mgr.Refactor.BaseServices.Network, opts).AllPages()
	if err != nil {
		return nil, api.NewListNetworksError(err)
	}
	l, err := networks.ExtractNetworks(page)
	if err != nil {
		return nil, api.NewListNetworksError(err)
	}
	var nets []api.Network
	for _, n := range l {
		nets = append(nets, *network(&n))
	}
	return nets, nil
}

//GetNetwork returns the configuration of the network identified by id
func (mgr *NetworkManager) GetNetwork(id string) (*api.Network, api.GetNetworkError) {
	n, err := networks.Get(mgr.Refactor.BaseServices.Network, id).Extract()
	if err != nil {
		return nil, api.NewGetNetworkError(err, id)
	}
	return network(n), nil
}

func (mgr *NetworkManager) createRouter(networkID string) (*routers.Router, error) {
	up := true
	netID, err := networks.IDFromName(mgr.Refactor.BaseServices.Network, mgr.Refactor.Config.ExternalNetworkName)
	if err != nil {
		return nil, UnwrapOpenStackError(err)
	}
	gwi := routers.GatewayInfo{
		NetworkID: netID,
	}

	createOpts := routers.CreateOpts{
		Name:         networkID,
		AdminStateUp: &up,
		GatewayInfo:  &gwi,
	}

	router, err := routers.Create(mgr.Refactor.BaseServices.Network, createOpts).Extract()
	if err != nil {
		return nil, UnwrapOpenStackError(err)
	}
	return router, nil
}
func (mgr *NetworkManager) attachSubnetToRouter(routerID, subnetID string) error {
	_, err := routers.AddInterface(mgr.Refactor.BaseServices.Network, routerID, routers.AddInterfaceOpts{
		SubnetID: subnetID,
	}).Extract()
	if err != nil {
		return UnwrapOpenStackError(err)
	}
	return nil
}
func (mgr *NetworkManager) deleteRouter(id string) error {
	err := routers.Delete(mgr.Refactor.BaseServices.Network, id).ExtractErr()
	return UnwrapOpenStackError(err)
}

func (mgr *NetworkManager) findRouter(name string) (*routers.Router, error) {
	page, err := routers.List(mgr.Refactor.BaseServices.Network, routers.ListOpts{Name: name}).AllPages()
	if err != nil {
		return nil, UnwrapOpenStackError(err)
	}
	l, err := routers.ExtractRouters(page)
	if err != nil {
		return nil, UnwrapOpenStackError(err)
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
func (mgr *NetworkManager) CreateSubnet(options api.CreateSubnetOptions) (*api.Subnet, api.CreateSubnetError) {
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
		return nil, api.NewCreateSubnetError(err, options)
	}
	// Execute the operation and get back a subnets.Subnet struct
	subnet, err := subnets.Create(mgr.Refactor.BaseServices.Network, opts).Extract()
	if err != nil {
		return nil, api.NewCreateSubnetError(err, options)
	}

	err = mgr.attachSubnetToRouter(router.ID, subnet.ID)
	if err != nil {
		err2 := mgr.DeleteSubnet(options.NetworkID, subnet.ID)
		err = api.NewErrorStackFromError(err, err2)
		return nil, api.NewCreateSubnetError(err, options)
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
func (mgr *NetworkManager) DeleteSubnet(networkID, subnetID string) api.DeleteSubnetError {
	err := subnets.Delete(mgr.Refactor.BaseServices.Network, subnetID).ExtractErr()
	return api.NewDeleteSubnetError(UnwrapOpenStackError(err), networkID, subnetID)
}

//ListSubnets lists the subnet
func (mgr *NetworkManager) ListSubnets(networkID string) ([]api.Subnet, api.ListSubnetsError) {
	page, err := subnets.List(mgr.Refactor.BaseServices.Network, subnets.ListOpts{
		NetworkID: networkID,
	}).AllPages()
	if err != nil {
		return nil, api.NewListSubnetsError(UnwrapOpenStackError(err), networkID)
	}
	l, err := subnets.ExtractSubnets(page)
	if err != nil {
		return nil, api.NewListSubnetsError(UnwrapOpenStackError(err), networkID)
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
	page, err := subnets.List(mgr.Refactor.BaseServices.Network, subnets.ListOpts{}).AllPages()
	if err != nil {
		return nil, UnwrapOpenStackError(err)
	}
	l, err := subnets.ExtractSubnets(page)
	if err != nil {
		return nil, UnwrapOpenStackError(err)
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
func (mgr *NetworkManager) GetSubnet(networkID, subnetID string) (*api.Subnet, api.GetSubnetError) {
	sn, err := subnets.Get(mgr.Refactor.BaseServices.Network, subnetID).Extract()
	if err != nil {
		return nil, api.NewGetSubnetError(UnwrapOpenStackError(err), networkID, subnetID)
	}
	return &api.Subnet{
		CIDR:      sn.CIDR,
		ID:        sn.ID,
		IPVersion: api.IPVersion(sn.IPVersion),
		Name:      sn.Name,
		NetworkID: sn.NetworkID,
	}, nil
}
