package openstack

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/pkg/errors"
)

type NetworkInterfacesManager struct {
	OpenStack *Provider
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

func (mgr *NetworkInterfacesManager) Create(options *api.CreateNetworkInterfaceOptions) (*api.NetworkInterface, error) {
	up := true
	ip := ports.IP{
		IPAddress: "",
		SubnetID:  options.SubnetID,
	}
	if options.IPAddress != nil {
		ip.IPAddress = *options.IPAddress
	}
	var serverID string
	if options.ServerID != nil {
		serverID = *options.ServerID
	}
	p, err := ports.Create(mgr.OpenStack.Network, &ports.CreateOpts{
		NetworkID:      options.NetworkID,
		Name:           options.Name,
		Description:    options.Name,
		AdminStateUp:   &up,
		FixedIPs:       []ports.IP{ip},
		DeviceID:       serverID,
		SecurityGroups: &[]string{options.SecurityGroupID},
	}).Extract()
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			*options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	return convert(p, nil), nil
}

func (mgr *NetworkInterfacesManager) Delete(id string) error {
	err := ports.Delete(mgr.OpenStack.Network, id).ExtractErr()
	return errors.Wrapf(err, "error deleting network interface %s", id)
}

func (mgr *NetworkInterfacesManager) Get(id string) (*api.NetworkInterface, error) {
	publicIPs, _ := mgr.OpenStack.PublicIPAddressManager.List(&api.ListPublicIPAddressOptions{})
	p, err := ports.Get(mgr.OpenStack.Network, id).Extract()
	if err != nil {
		return nil, errors.Wrapf(err, "error getting network interface %s", id)
	}
	return convert(p, publicIPs), nil
}
func checkNI(ni *api.NetworkInterface, options *api.ListNetworkInterfacesOptions) bool {
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

func (mgr *NetworkInterfacesManager) List(options *api.ListNetworkInterfacesOptions) ([]api.NetworkInterface, error) {
	var netID string
	if options.NetworkID != nil {
		netID = *options.NetworkID
	}
	var srvID string
	if options.ServerID != nil {
		srvID = *options.ServerID
	}
	publicIPs, _ := mgr.OpenStack.PublicIPAddressManager.List(&api.ListPublicIPAddressOptions{})
	pages, err := ports.List(mgr.OpenStack.Network, ports.ListOpts{
		NetworkID: netID,
		DeviceID:  srvID,
	}).AllPages()
	if err != nil {
		return nil, errors.Wrap(err, "error listing network interfaces")
	}
	pts, err := ports.ExtractPorts(pages)
	if err != nil {
		return nil, errors.Wrap(err, "error listing network interfaces")
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

func (mgr *NetworkInterfacesManager) Update(options *api.UpdateNetworkInterfacesOptions) (*api.NetworkInterface, error) {
	opts := ports.UpdateOpts{
		DeviceID: options.ServerID,
	}
	if options.SecurityGroupID != nil {
		opts.SecurityGroups = &[]string{*options.SecurityGroupID}
	}
	port, err := ports.Update(mgr.OpenStack.Network, options.ID, opts).Extract()
	if err != nil {
		return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
	}
	publicIPs, _ := mgr.OpenStack.PublicIPAddressManager.List(&api.ListPublicIPAddressOptions{})
	return convert(port, publicIPs), nil
}
