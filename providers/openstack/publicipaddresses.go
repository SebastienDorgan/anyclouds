package openstack

import (
	"encoding/binary"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/talgo"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/pkg/errors"
	"net"
)

//NetworkManager defines networking functions a anyclouds provider must provide
type PublicIPAddressManager struct {
	OpenStack *Provider
}

func ipToUin32(ip *net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

func uint32toIP(ui uint32) net.IP {
	ip := net.IPv4(0xffff, 0xffff, 0xffff, 0xffff)
	binary.BigEndian.PutUint32(ip, ui)
	return ip
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
		ip, ipNet, err := net.ParseCIDR(sn.CIDR)
		if err != nil {
			return nil, errors.Wrap(ProviderError(err), "error listing available public ip pools")
		}
		uip := ipToUin32(&ip) + 1 //Firts ip of the subnet
		firstIP := uint32toIP(uip)
		//increment until uip+1 not in subnet
		for ; ipNet.Contains(uint32toIP(uip + 1)); uip++ {
		}
		lastIP := uint32toIP(uip)
		pools = append(pools, api.PublicIPPool{
			ID: sn.ID,
			Ranges: []api.AddressRange{
				{
					FirstAddress: firstIP.String(),
					LastAddress:  lastIP.String(),
				},
			},
		})
	}
	return pools, nil
}

func (mgr *PublicIPAddressManager) ListAllocated() ([]api.PublicIP, error) {
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
		ip, err := mgr.toPublicIP(&fip)
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

func (mgr *PublicIPAddressManager) Dissociate(publicIpID string) error {
	pip, err := mgr.Get(publicIpID)
	if err != nil {
		return errors.Wrapf(err, "error dissociating public ip address %s", publicIpID)
	}
	_, err = floatingips.Update(mgr.OpenStack.Network, publicIpID, floatingips.UpdateOpts{
		Description: &pip.Name,
		PortID:      nil,
		FixedIP:     "",
	}).Extract()
	return errors.Wrapf(err, "error dissociating public ip address %s", publicIpID)

}
func (mgr *PublicIPAddressManager) Release(publicIPId string) error {
	err := floatingips.Delete(mgr.OpenStack.Network, publicIPId).ExtractErr()
	return errors.Wrapf(err, "error releasing public ip address %s", publicIPId)
}

func (mgr *PublicIPAddressManager) Get(publicIpId string) (*api.PublicIP, error) {
	fip, err := floatingips.Get(mgr.OpenStack.Network, publicIpId).Extract()
	if err != nil {
		return nil, errors.Wrapf(err, "error dissociating public ip address %s", publicIpId)
	}
	return mgr.toPublicIP(fip)
}

func (mgr *PublicIPAddressManager) toPublicIP(fip *floatingips.FloatingIP) (*api.PublicIP, error) {
	var ServerID, SubnetID string
	var ps *ports.Port

	if len(fip.PortID) > 0 {
		var err error
		ps, err = ports.Get(mgr.OpenStack.Network, fip.PortID).Extract()
		if err != nil {
			return nil, err
		}
	}

	if ps != nil {
		ServerID = ps.DeviceID
		//take the subnet matching fip fixed ip
		for _, ip := range ps.FixedIPs {
			if ip.IPAddress == fip.FixedIP {
				SubnetID = ip.SubnetID
				break
			}
		}
	}
	return &api.PublicIP{
		ID:        fip.ID,
		Name:      fip.Description,
		Address:   fip.FloatingIP,
		ServerID:  ServerID,
		SubnetID:  SubnetID,
		PrivateIP: fip.FixedIP,
	}, nil
}
