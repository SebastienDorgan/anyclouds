package openstack

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/anyclouds/iputils"
	"github.com/SebastienDorgan/talgo"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/pkg/errors"
)

//PublicIPAddressManager openstack implementation of api.PublicIPAddressManager
type PublicIPAddressManager struct {
	OpenStack *Provider
}

func (mgr *PublicIPAddressManager) ListAvailablePools() ([]api.PublicIPPool, error) {
	snets, err := mgr.OpenStack.GetNetworkManager().ListSubnets(mgr.OpenStack.ExternalNetworkID)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "error listing available public ip pools")
	}
	var pools []api.PublicIPPool
	for _, sn := range snets {
		if sn.IPVersion == api.IPVersion6 {
			continue
		}
		r, err := iputils.GetRange(sn.CIDR)
		if err != nil {
			return nil, errors.Wrap(ProviderError(err), "error listing available public ip pools")
		}
		pools = append(pools, api.PublicIPPool{
			ID: sn.ID,
			Ranges: []api.AddressRange{
				{
					FirstAddress: r.FirstIP.String(),
					LastAddress:  r.LastIP.String(),
				},
			},
		})
	}
	return pools, nil
}

func (mgr *PublicIPAddressManager) List(options *api.ListPublicIPAddressOptions) ([]api.PublicIP, error) {
	pages, err := floatingips.List(mgr.OpenStack.Network, floatingips.ListOpts{
		Status: "ACTIVE",
	}).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "error listing available public ip address")
	}
	fips, err := floatingips.ExtractFloatingIPs(pages)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "error listing available public ip address")
	}

	publicIPs := make([]api.PublicIP, len(fips))
	for i, fip := range fips {
		ip, err := toPublicIP(&fip)
		if err != nil {
			return nil, errors.Wrap(ProviderError(err), "error listing available public ip address")
		}
		publicIPs[i] = *ip
	}
	return publicIPs, nil
}
func (mgr *PublicIPAddressManager) Allocate(options *api.PublicIPAllocationOptions) (*api.PublicIP, error) {
	fip, err := floatingips.Create(mgr.OpenStack.Network, &floatingips.CreateOpts{
		Description:       options.Name,
		FloatingNetworkID: mgr.OpenStack.ExternalNetworkID,
		FloatingIP:        options.Address,
		SubnetID:          options.AddressPool,
	}).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "error allocating public ip address")
	}
	return &api.PublicIP{
		ID:      fip.ID,
		Name:    fip.Description,
		Address: fip.FloatingIP,
	}, nil
}

func (mgr *PublicIPAddressManager) Associate(options *api.PublicIPAssociationOptions) error {
	fip, err := floatingips.Get(mgr.OpenStack.Network, options.PublicIPId).Extract()
	if err != nil {
		return errors.Wrapf(ProviderError(err), "error associating public ip address %s to server %s", options.PublicIPId, options.ServerID)
	}
	portList, err := mgr.listPorts(options.ServerID)
	if err != nil {
		return errors.Wrapf(ProviderError(err), "error associating public ip address %s to server %s", options.PublicIPId, options.ServerID)
	}
	selectedPorts := make([]int, len(portList))
	for i := 0; i < len(portList); i++ {
		selectedPorts[i] = i
	}
	//Filter ports by subnet
	if len(options.SubnetID) > 0 {
		selectedPorts = talgo.FindAll(len(portList), func(i int) bool {
			return talgo.Any(len(portList[i].FixedIPs), func(j int) bool {
				return portList[i].FixedIPs[j].SubnetID == options.SubnetID
			})
		})
	}
	//Filter ports by IP
	if len(options.PrivateIP) > 0 {
		selection := talgo.FindAll(len(selectedPorts), func(i int) bool {
			return talgo.Any(len(portList[selectedPorts[i]].FixedIPs), func(j int) bool {
				return portList[selectedPorts[i]].FixedIPs[j].IPAddress == options.PrivateIP
			})
		})
		// get index of the orignal portList
		for i := 0; i < len(selection); i++ {
			selection[i] = selectedPorts[selection[i]]
		}
		selectedPorts = selection
	}
	_, err = floatingips.Update(mgr.OpenStack.Network, options.PublicIPId, floatingips.UpdateOpts{
		Description: &fip.Description,
		PortID:      &portList[selectedPorts[0]].ID,
		FixedIP:     options.PrivateIP,
	}).Extract()
	return errors.Wrapf(ProviderError(err), "error associating public ip address %s to server %s", options.PublicIPId, options.ServerID)

}

func (mgr *PublicIPAddressManager) listPorts(serverID string) ([]ports.Port, error) {
	up := true
	pages, err := ports.List(mgr.OpenStack.Network, ports.ListOpts{
		Status:       "ACTIVE",
		AdminStateUp: &up,
		DeviceID:     serverID,
	}).AllPages()
	if err != nil {
		return nil, err
	}
	portList, err := ports.ExtractPorts(pages)
	return portList, err
}

func (mgr *PublicIPAddressManager) Dissociate(publicIPID string) error {
	pip, err := mgr.Get(publicIPID)
	if err != nil {
		return errors.Wrapf(err, "error dissociating public ip address %s", publicIPID)
	}
	_, err = floatingips.Update(mgr.OpenStack.Network, publicIPID, floatingips.UpdateOpts{
		Description: &pip.Name,
		PortID:      nil,
		FixedIP:     "",
	}).Extract()
	return errors.Wrapf(err, "error dissociating public ip address %s", publicIPID)

}
func (mgr *PublicIPAddressManager) Release(publicIPId string) error {
	err := floatingips.Delete(mgr.OpenStack.Network, publicIPId).ExtractErr()
	return errors.Wrapf(err, "error releasing public ip address %s", publicIPId)
}

func (mgr *PublicIPAddressManager) Get(publicIPID string) (*api.PublicIP, error) {
	fip, err := floatingips.Get(mgr.OpenStack.Network, publicIPID).Extract()
	if err != nil {
		return nil, errors.Wrapf(err, "error dissociating public ip address %s", publicIPID)
	}
	return toPublicIP(fip)
}

func toPublicIP(fip *floatingips.FloatingIP) (*api.PublicIP, error) {
	return &api.PublicIP{
		ID:                 fip.ID,
		Name:               fip.Description,
		Address:            fip.FloatingIP,
		NetworkInterfaceID: fip.PortID,
		PrivateAddress:     fip.FixedIP,
	}, nil
}
