package openstack

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/pkg/errors"
)

type NetworkInterfacesManager struct {
	OpenStack *Provider
}

func convert(port *ports.Port) *api.NetworkInterface {
	return &api.NetworkInterface{
		ID:         port.ID,
		Name:       port.Name,
		MacAddress: port.MACAddress,
		NetworkID:  port.NetworkID,
		SubnetID:   port.FixedIPs[0].SubnetID,
		ServerID:   port.DeviceID,
		IPAddress:  port.FixedIPs[0].IPAddress,
		Primary:    false,
	}
}

func (mgr *NetworkInterfacesManager) Create(options *api.NetworkInterfaceOptions) (*api.NetworkInterface, error) {
	up := true
	ip := ports.IP{
		IPAddress: "",
		SubnetID:  options.SubnetID,
	}
	if options.IPAddress != nil {
		ip.IPAddress = *options.IPAddress
	}
	p, err := ports.Create(mgr.OpenStack.Network, &ports.CreateOpts{
		NetworkID:      options.NetworkID,
		Name:           options.Name,
		Description:    options.Name,
		AdminStateUp:   &up,
		FixedIPs:       []ports.IP{ip},
		DeviceID:       options.ServerID,
		SecurityGroups: &[]string{options.SecurityGroupID},
	}).Extract()
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	return convert(p), nil
}

func (mgr *NetworkInterfacesManager) Delete(id string) error {
	err := ports.Delete(mgr.OpenStack.Network, id).ExtractErr()
	return errors.Wrapf(err, "error deleting network interface %s", id)
}

func (mgr *NetworkInterfacesManager) Get(id string) (*api.NetworkInterface, error) {
	p, err := ports.Get(mgr.OpenStack.Network, id).Extract()
	if err != nil {
		return nil, errors.Wrapf(err, "error getting network interface %s", id)
	}
	return convert(p), nil
}

func (mgr *NetworkInterfacesManager) List() ([]api.NetworkInterface, error) {
	pages, err := ports.List(mgr.OpenStack.Network, ports.ListOpts{}).AllPages()
	if err != nil {
		return nil, errors.Wrap(err, "error listing network interfaces")
	}
	pts, err := ports.ExtractPorts(pages)
	if err != nil {
		return nil, errors.Wrap(err, "error listing network interfaces")
	}
	var list []api.NetworkInterface
	for _, p := range pts {
		list = append(list, *convert(&p))
	}
	return list, nil
}

func (mgr *NetworkInterfacesManager) UpdateSecurityGroup(options *api.NetworkInterfacesUpdateOptions) (*api.NetworkInterface, error) {
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
	return convert(port), nil
}
