package openstack

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/anyclouds/iputils"
	"github.com/SebastienDorgan/talgo"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

//PublicIPManager openstack implementation of api.PublicIPManager
type PublicIPManager struct {
	OpenStack *Provider
}

func (mgr *PublicIPManager) ListAvailablePools() ([]api.PublicIPPool, *api.ListAvailablePublicIPPoolsError) {
	subnets, err := mgr.OpenStack.GetNetworkManager().ListSubnets(mgr.OpenStack.ExternalNetworkID)
	if err != nil {
		return nil, api.NewListAvailablePublicIPPoolsError(UnwrapOpenStackError(err))
	}
	var pools []api.PublicIPPool
	for _, sn := range subnets {
		if sn.IPVersion == api.IPVersion6 {
			continue
		}
		r, err := iputils.GetRange(sn.CIDR)
		if err != nil {
			return nil, api.NewListAvailablePublicIPPoolsError(UnwrapOpenStackError(err))
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

func checkPublicIP(nis []api.NetworkInterface, fip *floatingips.FloatingIP) bool {
	if nis == nil {
		return true
	}
	for _, ni := range nis {
		if ni.PublicIPAddress == fip.FloatingIP {
			return true
		}
	}
	return false
}

func (mgr *PublicIPManager) List(options *api.ListPublicIPsOptions) ([]api.PublicIP, *api.ListPublicIPsError) {
	pages, err := floatingips.List(mgr.OpenStack.Network, floatingips.ListOpts{}).AllPages()
	if err != nil {
		return nil, api.NewListPublicIPsError(UnwrapOpenStackError(err), options)
	}
	fips, err := floatingips.ExtractFloatingIPs(pages)
	if err != nil {
		return nil, api.NewListPublicIPsError(UnwrapOpenStackError(err), options)
	}
	var nis []api.NetworkInterface
	if options != nil {
		nis, err = mgr.OpenStack.NetworkInterfacesManager.List(&api.ListNetworkInterfacesOptions{
			ServerID: options.ServerID,
		})
		if err != nil {
			return nil, api.NewListPublicIPsError(UnwrapOpenStackError(err), options)
		}
	}
	var publicIPs []api.PublicIP
	for _, fip := range fips {
		if checkPublicIP(nis, &fip) {
			publicIPs = append(publicIPs, *toPublicIP(&fip))
		}
	}
	return publicIPs, nil
}
func (mgr *PublicIPManager) Create(options api.CreatePublicIPOptions) (*api.PublicIP, *api.CreatePublicIPError) {
	var ipAddress string
	if options.IPAddress != nil {
		ipAddress = *options.IPAddress
	}
	var ipAddressPool string
	if options.IPAddressPoolID != nil {
		ipAddressPool = *options.IPAddressPoolID
	}
	fip, err := floatingips.Create(mgr.OpenStack.Network, &floatingips.CreateOpts{
		Description:       options.Name,
		FloatingNetworkID: mgr.OpenStack.ExternalNetworkID,
		FloatingIP:        ipAddress,
		SubnetID:          ipAddressPool,
	}).Extract()
	if err != nil {
		return nil, api.NewCreatePublicIPError(UnwrapOpenStackError(err), options)
	}
	return &api.PublicIP{
		ID:      fip.ID,
		Name:    fip.Description,
		Address: fip.FloatingIP,
	}, nil
}

func (mgr *PublicIPManager) Associate(options api.AssociatePublicIPOptions) *api.AssociatePublicIPError {
	fip, err := floatingips.Get(mgr.OpenStack.Network, options.PublicIPId).Extract()
	if err != nil {
		return api.NewAssociatePublicIPError(UnwrapOpenStackError(err), options)
	}
	portList, err := mgr.listPorts(options.ServerID)
	if err != nil {
		return api.NewAssociatePublicIPError(UnwrapOpenStackError(err), options)
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
		// get index of the original portList
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
	return api.NewAssociatePublicIPError(UnwrapOpenStackError(err), options)

}

func (mgr *PublicIPManager) listPorts(serverID string) ([]ports.Port, error) {
	up := true
	pages, err := ports.List(mgr.OpenStack.Network, ports.ListOpts{
		AdminStateUp: &up,
		DeviceID:     serverID,
	}).AllPages()
	if err != nil {
		return nil, err
	}
	portList, err := ports.ExtractPorts(pages)
	return portList, err
}

func (mgr *PublicIPManager) Dissociate(publicIPID string) *api.DissociatePublicIPError {
	var err error
	pip, err := mgr.Get(publicIPID)
	if err != nil {
		return api.NewDissociatePublicIPError(err, publicIPID)
	}
	_, err = floatingips.Update(mgr.OpenStack.Network, publicIPID, floatingips.UpdateOpts{
		Description: &pip.Name,
		PortID:      nil,
		FixedIP:     "",
	}).Extract()
	return api.NewDissociatePublicIPError(UnwrapOpenStackError(err), publicIPID)

}
func (mgr *PublicIPManager) Delete(publicIPId string) *api.DeletePublicIPError {
	err := floatingips.Delete(mgr.OpenStack.Network, publicIPId).ExtractErr()
	return api.NewDeletePublicIPError(UnwrapOpenStackError(err), publicIPId)
}

func (mgr *PublicIPManager) Get(publicIPID string) (*api.PublicIP, *api.GetPublicIPError) {
	fip, err := floatingips.Get(mgr.OpenStack.Network, publicIPID).Extract()
	if err != nil {
		return nil, api.NewGetPublicIPError(UnwrapOpenStackError(err), publicIPID)
	}
	return toPublicIP(fip), nil
}

func toPublicIP(fip *floatingips.FloatingIP) *api.PublicIP {
	return &api.PublicIP{
		ID:                 fip.ID,
		Name:               fip.Description,
		Address:            fip.FloatingIP,
		NetworkInterfaceID: fip.PortID,
		PrivateAddress:     fip.FixedIP,
	}
}
