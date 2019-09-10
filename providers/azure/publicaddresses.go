package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/pkg/errors"
)

type PublicIPAddressManager struct {
	Provider *Provider
}

func (mgr *PublicIPAddressManager) ListAvailablePools() ([]api.PublicIPPool, error) {
	panic("implement me")
}

func checkAddress(addresses []string, address *string) bool {
	if address == nil {
		return false
	}
	if addresses == nil {
		return true
	}
	for _, addr := range addresses {
		if addr == *address {
			return true
		}
	}
	return false
}

func (mgr *PublicIPAddressManager) List(options *api.ListPublicIPAddressOptions) ([]api.PublicIP, error) {
	var addresses []string
	if options != nil && options.ServerID != nil {
		nis, err := mgr.Provider.NetworkInterfacesManager.List(&api.ListNetworkInterfacesOptions{
			ServerID: options.ServerID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error listing public ip address")
		}
		for _, ni := range nis {
			if len(ni.PublicIPAddress) > 0 {
				addresses = append(addresses, ni.PublicIPAddress)
			}
		}

	}
	ips, err := mgr.Provider.PublicIPAddressesClient.List(context.Background(), mgr.Provider.Configuration.ResourceGroupName)
	if err != nil {
		return nil, errors.Wrap(err, "error listing public ip address")
	}
	var list []api.PublicIP
	for ips.NotDone() {
		for _, ip := range ips.Values() {
			if checkAddress(addresses, ip.IPAddress) {
				list = append(list, *convertAddress(&ip))
			}
		}
		err := ips.NextWithContext(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "error listing public ip address")
		}
	}
	return list, nil
}

func convertAddress(address *network.PublicIPAddress) *api.PublicIP {
	return &api.PublicIP{
		ID:             *address.Name,
		Name:           *address.Name,
		Address:        *address.IPAddress,
		PrivateAddress: *address.IPConfiguration.PrivateIPAddress,
	}
}

func (mgr *PublicIPAddressManager) Allocate(options api.PublicIPAllocationOptions) (*api.PublicIP, error) {
	future, err := mgr.Provider.PublicIPAddressesClient.CreateOrUpdate(
		context.Background(),
		mgr.Provider.Configuration.ResourceGroupName,
		options.Name,
		network.PublicIPAddress{
			Response: autorest.Response{},
			Sku:      nil,
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAddressVersion:   network.IPv4,
				PublicIPAllocationMethod: network.Static,
			},
			Name: to.StringPtr(options.Name),

			Location: to.StringPtr(mgr.Provider.Configuration.Location),
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "error allocating public ip %s", options.Name)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.PublicIPAddressesClient.Client)
	if err != nil {
		return nil, errors.Wrapf(err, "error allocating public ip %s", options.Name)
	}
	ip, err := future.Result(mgr.Provider.PublicIPAddressesClient)
	if err != nil {
		return nil, errors.Wrapf(err, "error allocating public ip %s", options.Name)
	}

	return convertAddress(&ip), nil
}

func (mgr *PublicIPAddressManager) Associate(options api.PublicIPAssociationOptions) error {
	nis, err := mgr.Provider.NetworkInterfacesManager.list(&api.ListNetworkInterfacesOptions{
		SubnetID: &options.SubnetID,
		ServerID: &options.ServerID,
	})
	if err != nil {
		return errors.Wrapf(err, "error associating public ip %s with server %s", options.PublicIPId, options.ServerID)
	}
	var ipConf *network.InterfaceIPConfiguration
	var niToUpdate *network.Interface
	for _, ni := range nis {
		for _, ipc := range *ni.IPConfigurations {
			if *ipc.PrivateIPAddress == options.PrivateIP {
				ipConf = &ipc
				niToUpdate = &ni
			}
		}
	}
	if ipConf == nil || niToUpdate == nil {
		err = errors.Errorf("unable to find network inferface of server %s using private address %s", options.ServerID, options.PrivateIP)
		return errors.Wrapf(err, "error associating public ip %s with server %s", options.PublicIPId, options.ServerID)
	}
	addr, err := mgr.get(options.PublicIPId)
	if err != nil {
		return errors.Wrapf(err, "error associating public ip %s with server %s", options.PublicIPId, options.ServerID)
	}
	ipConf.PublicIPAddress = addr
	future, err := mgr.Provider.InterfacesClient.CreateOrUpdate(context.Background(), mgr.Provider.Configuration.ResourceGroupName, *niToUpdate.Name, *niToUpdate)
	if err != nil {
		return errors.Wrapf(err, "error associating public ip %s with server %s", options.PublicIPId, options.ServerID)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.InterfacesClient.Client)

	return errors.Wrapf(err, "error associating public ip %s with server %s", options.PublicIPId, options.ServerID)

}

func (mgr *PublicIPAddressManager) Dissociate(publicIPId string) error {
	panic("implement me")
}

func (mgr *PublicIPAddressManager) Release(publicIPId string) error {
	panic("implement me")
}

func (mgr *PublicIPAddressManager) get(publicIPId string) (*network.PublicIPAddress, error) {
	addr, err := mgr.Provider.PublicIPAddressesClient.Get(context.Background(), mgr.Provider.Configuration.ResourceGroupName, publicIPId, "")
	return &addr, err
}

func (mgr *PublicIPAddressManager) Get(publicIPId string) (*api.PublicIP, error) {
	panic("implement me")
}
