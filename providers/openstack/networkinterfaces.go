package openstack

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

type NetworkInterfacesManager struct {
	Provider *Provider
}

func convert(port *ports.Port, publicIPs []api.PublicIP) *api.NetworkInterface {
	var publicIPAddress string
	for _, ip := range publicIPs {
		if ip.NetworkInterfaceID == port.ID {
			publicIPAddress = ip.Address
		}
	}
	return &api.NetworkInterface{
		ID:               port.ID,
		Name:             port.Name,
		MacAddress:       port.MACAddress,
		NetworkID:        port.NetworkID,
		SubnetID:         port.FixedIPs[0].SubnetID,
		ServerID:         port.DeviceID,
		PrivateIPAddress: port.FixedIPs[0].IPAddress,
		PublicIPAddress:  publicIPAddress,
		SecurityGroupID:  port.SecurityGroups[0],
	}
}

func (mgr *NetworkInterfacesManager) Create(options api.CreateNetworkInterfaceOptions) (*api.NetworkInterface, api.CreateNetworkInterfaceError) {
	up := true
	ip := ports.IP{
		IPAddress: "",
		SubnetID:  options.SubnetID,
	}
	if options.PrivateIPAddress != nil {
		ip.IPAddress = *options.PrivateIPAddress
	}
	var serverID string
	if options.ServerID != nil {
		serverID = *options.ServerID
	}
	p, err := ports.Create(mgr.Provider.BaseServices.Network, &ports.CreateOpts{
		NetworkID:      options.NetworkID,
		Name:           options.Name,
		Description:    options.Name,
		AdminStateUp:   &up,
		FixedIPs:       []ports.IP{ip},
		DeviceID:       serverID,
		SecurityGroups: &[]string{options.SecurityGroupID},
	}).Extract()
	if err != nil {
		return nil, api.NewCreateNetworkInterfaceError(err, options)
	}
	return convert(p, nil), nil
}

func (mgr *NetworkInterfacesManager) Delete(id string) api.DeleteNetworkInterfaceError {
	err := ports.Delete(mgr.Provider.BaseServices.Network, id).ExtractErr()
	return api.NewDeleteNetworkInterfaceError(err, id)
}

func (mgr *NetworkInterfacesManager) Get(id string) (*api.NetworkInterface, api.GetNetworkInterfaceError) {
	publicIPs, _ := mgr.Provider.PublicIPAddressManager.List(&api.ListPublicIPsOptions{})
	p, err := ports.Get(mgr.Provider.BaseServices.Network, id).Extract()
	if err != nil {
		return nil, api.NewGetNetworkInterfaceError(err, id)
	}
	return convert(p, publicIPs), nil
}

func checkNI(ni *api.NetworkInterface, options *api.ListNetworkInterfacesOptions) bool {
	if options == nil {
		return true
	}
	if options.SubnetID != nil && *options.SubnetID != ni.SubnetID {
		return false
	}
	if options.SecurityGroupID != nil && *options.SecurityGroupID != ni.SecurityGroupID {
		return false
	}
	if options.PrivateIPAddress != nil && *options.PrivateIPAddress != ni.PrivateIPAddress {
		return false
	}
	return true
}

func (mgr *NetworkInterfacesManager) list(options *api.ListNetworkInterfacesOptions) ([]api.NetworkInterface, error) {
	var netID string
	if options != nil && options.NetworkID != nil {
		netID = *options.NetworkID
	}
	var srvID string
	if options != nil && options.ServerID != nil {
		srvID = *options.ServerID
	}
	publicIPs, _ := mgr.Provider.PublicIPAddressManager.List(&api.ListPublicIPsOptions{})
	pages, err := ports.List(mgr.Provider.BaseServices.Network, ports.ListOpts{
		NetworkID: netID,
		DeviceID:  srvID,
	}).AllPages()
	if err != nil {
		return nil, err
	}
	pts, err := ports.ExtractPorts(pages)
	if err != nil {
		return nil, err
	}

	var list []api.NetworkInterface
	for _, p := range pts {
		ni := convert(&p, publicIPs)
		if checkNI(ni, options) {
			list = append(list, *ni)
		}

	}
	return list, nil
}

func (mgr *NetworkInterfacesManager) List(options *api.ListNetworkInterfacesOptions) ([]api.NetworkInterface, api.ListNetworkInterfacesError) {
	l, err := mgr.list(options)
	return l, api.NewListNetworkInterfacesError(err, options)
}

func (mgr *NetworkInterfacesManager) update(options api.UpdateNetworkInterfaceOptions) (*api.NetworkInterface, error) {
	opts := ports.UpdateOpts{
		DeviceID: options.ServerID,
	}
	if options.SecurityGroupID != nil {
		opts.SecurityGroups = &[]string{*options.SecurityGroupID}
	}
	port, err := ports.Update(mgr.Provider.BaseServices.Network, options.ID, opts).Extract()
	if err != nil {
		return nil, err
	}
	publicIPs, _ := mgr.Provider.PublicIPAddressManager.List(&api.ListPublicIPsOptions{})
	return convert(port, publicIPs), nil
}

func (mgr *NetworkInterfacesManager) Update(options api.UpdateNetworkInterfaceOptions) (*api.NetworkInterface, api.UpdateNetworkInterfaceError) {
	ni, err := mgr.Update(options)
	return ni, api.NewUpdateNetworkInterfaceError(err, options)
}
