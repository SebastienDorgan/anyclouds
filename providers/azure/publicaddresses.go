package azure

import (
	"context"
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/anyclouds/iputils"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type PublicIPManager struct {
	Provider *Provider
}

//PoolProperty property of azure public address pool
type PoolProperty struct {
	ChangeNumber    int32    `json:"changeNumber,omitempty"`
	Region          string   `json:"region,omitempty"`
	Platform        string   `json:"platform,omitempty"`
	SystemService   string   `json:"systemService,omitempty"`
	AddressPrefixes []string `json:"addressPrefixes,omitempty"`
}

//AddressPool azure public address pool
type AddressPool struct {
	Name       string       `json:"name,omitempty"`
	ID         string       `json:"id,omitempty"`
	Properties PoolProperty `json:"properties,omitempty"`
}

//PublicAddressPools azure public address pools
type PublicAddressPools struct {
	ChangeNumber int32         `json:"changeNumber,omitempty"`
	Cloud        string        `json:"cloud,omitempty"`
	Values       []AddressPool `json:"values,omitempty"`
}

func (mgr *PublicIPManager) getPublicAddressPools(url string) ([]AddressPool, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Body.Close() }()
	var pools PublicAddressPools
	err = json.NewDecoder(r.Body).Decode(&pools)
	var selection []AddressPool
	for _, pls := range pools.Values {
		if pls.Properties.Region == mgr.Provider.Configuration.Location && pls.Properties.SystemService == "AzureAppService" {
			selection = append(selection, pls)
		}
	}
	return selection, err
}

func (mgr *PublicIPManager) ListAvailablePools() ([]api.PublicIPPool, api.ListAvailablePublicIPPoolsError) {
	addressPools, err := mgr.getPublicAddressPools(mgr.Provider.Configuration.PublicAddressesURL)
	if err != nil {
		return nil, api.NewListAvailablePublicIPPoolsError(err)
	}
	var pools []api.PublicIPPool
	for _, p := range addressPools {
		pool := api.PublicIPPool{
			ID: p.Name,
		}
		for _, prefix := range p.Properties.AddressPrefixes {
			addressRange, err := iputils.GetRange(prefix)
			if err != nil {
				return nil, api.NewListAvailablePublicIPPoolsError(err)
			}
			pool.Ranges = append(pool.Ranges, api.AddressRange{
				FirstAddress: addressRange.FirstIP.String(),
				LastAddress:  addressRange.LastIP.String(),
			})
		}
		pools = append(pools, pool)
	}
	return pools, nil
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

func (mgr *PublicIPManager) List(options *api.ListPublicIPsOptions) ([]api.PublicIP, api.ListPublicIPsError) {
	var addresses []string
	if options != nil && options.ServerID != nil {
		nis, err := mgr.Provider.NetworkInterfacesManager.List(&api.ListNetworkInterfacesOptions{
			ServerID: options.ServerID,
		})
		if err != nil {
			return nil, api.NewListPublicIPsError(err, options)
		}
		for _, ni := range nis {
			if len(ni.PublicIPAddress) > 0 {
				addresses = append(addresses, ni.PublicIPAddress)
			}
		}

	}
	ips, err := mgr.Provider.BaseServices.PublicIPAddressesClient.List(context.Background(), mgr.Provider.Configuration.ResourceGroupName)
	if err != nil {
		return nil, api.NewListPublicIPsError(err, options)
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
			return nil, api.NewListPublicIPsError(err, options)
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

func (mgr *PublicIPManager) Create(options api.CreatePublicIPOptions) (*api.PublicIP, api.CreatePublicIPError) {
	future, err := mgr.Provider.BaseServices.PublicIPAddressesClient.CreateOrUpdate(
		context.Background(),
		mgr.Provider.Configuration.ResourceGroupName,
		options.Name,
		network.PublicIPAddress{
			Response: autorest.Response{},
			Sku: &network.PublicIPAddressSku{
				Name: network.PublicIPAddressSkuNameStandard,
			},
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: network.Static,
				PublicIPAddressVersion:   network.IPv4,
				IPAddress:                options.IPAddress,
				PublicIPPrefix:           &network.SubResource{ID: options.IPAddressPoolID},
			},
			Name:     to.StringPtr(options.Name),
			Location: to.StringPtr(mgr.Provider.Configuration.Location),
		},
	)

	if err != nil {
		return nil, api.NewCreatePublicIPError(err, options)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.BaseServices.PublicIPAddressesClient.Client)
	if err != nil {
		return nil, api.NewCreatePublicIPError(err, options)
	}
	ip, err := future.Result(mgr.Provider.BaseServices.PublicIPAddressesClient)
	if err != nil {
		return nil, api.NewCreatePublicIPError(err, options)
	}

	return convertAddress(&ip), nil
}

func (mgr *PublicIPManager) Associate(options api.AssociatePublicIPOptions) api.AssociatePublicIPError {
	nis, err := mgr.Provider.NetworkInterfacesManager.listAzure(&api.ListNetworkInterfacesOptions{
		SubnetID: &options.SubnetID,
		ServerID: &options.ServerID,
	})
	if err != nil {
		return api.NewAssociatePublicIPError(err, options)
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
		err = errors.Errorf("unable to find network interface of server %s using private address %s", options.ServerID, options.PrivateIP)
		return api.NewAssociatePublicIPError(err, options)
	}
	addr, err := mgr.get(options.PublicIPId)
	if err != nil {
		return api.NewAssociatePublicIPError(err, options)
	}
	ipConf.PublicIPAddress = addr
	future, err := mgr.Provider.BaseServices.InterfacesClient.CreateOrUpdate(context.Background(), mgr.Provider.Configuration.ResourceGroupName, *niToUpdate.Name, *niToUpdate)
	if err != nil {
		return api.NewAssociatePublicIPError(err, options)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.BaseServices.InterfacesClient.Client)

	return api.NewAssociatePublicIPError(err, options)

}

func (mgr *PublicIPManager) Dissociate(publicIPId string) api.DissociatePublicIPError {
	var err error
	ip, err := mgr.Get(publicIPId)
	if err != nil {
		return api.NewDissociatePublicIPError(err, publicIPId)
	}
	if len(ip.NetworkInterfaceID) == 0 {
		return nil
	}
	ni, err := mgr.Provider.NetworkInterfacesManager.get(ip.NetworkInterfaceID)
	if err != nil {
		return api.NewDissociatePublicIPError(err, publicIPId)
	}
	if ni.IPConfigurations == nil {
		return nil
	}
	for i := range *ni.IPConfigurations {
		ipConf := (*ni.IPConfigurations)[i]
		if ipConf.PublicIPAddress != nil && *ipConf.PublicIPAddress.Name == ip.Name {
			ipConf.PublicIPAddress = nil
		}
	}
	future, err := mgr.Provider.BaseServices.InterfacesClient.CreateOrUpdate(context.Background(), mgr.Provider.Configuration.ResourceGroupName, *ni.Name, *ni)
	if err != nil {
		return api.NewDissociatePublicIPError(err, publicIPId)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.BaseServices.InterfacesClient.Client)

	return api.NewDissociatePublicIPError(err, publicIPId)
}

func (mgr *PublicIPManager) Delete(publicIPId string) api.DeletePublicIPError {
	_, err := mgr.Provider.BaseServices.PublicIPAddressesClient.Delete(context.Background(), mgr.Provider.Configuration.ResourceGroupName, publicIPId)
	return api.NewDeletePublicIPError(err, publicIPId)
}

func (mgr *PublicIPManager) get(publicIPId string) (*network.PublicIPAddress, error) {
	addr, err := mgr.Provider.BaseServices.PublicIPAddressesClient.Get(context.Background(), mgr.Provider.Configuration.ResourceGroupName, publicIPId, "")
	return &addr, err
}

func (mgr *PublicIPManager) Get(publicIPId string) (*api.PublicIP, api.GetPublicIPError) {
	ip, err := mgr.get(publicIPId)
	return convertAddress(ip), api.NewGetPublicIPError(err, publicIPId)
}
