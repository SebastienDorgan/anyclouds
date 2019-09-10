package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/pkg/errors"
)

type NetworkInterfacesManager struct {
	Provider *Provider
}

func (mgr *NetworkInterfacesManager) resourceGroup() string {
	return mgr.Provider.Configuration.ResourceGroupName
}
func (mgr *NetworkInterfacesManager) create(options *api.CreateNetworkInterfaceOptions) (*network.Interface, error) {
	var subResource *network.SubResource
	if options.ServerID != nil {
		subResource = &network.SubResource{ID: options.ServerID}
	}
	tags := make(map[string]*string)
	tags["network-id"] = &options.NetworkID
	if options.ServerID != nil {
		tags["server-id"] = options.ServerID
	}
	future, err := mgr.Provider.InterfacesClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), options.Name, network.Interface{
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
			VirtualMachine: subResource,
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				{
					InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAddress:          options.PrivateIPAddress,
						PrivateIPAllocationMethod: network.Dynamic,
						Subnet: &network.Subnet{
							Name: &options.Name,
						},
						Primary: to.BoolPtr(options.Primary),
					},
					Name: &options.Name,
				},
			},
			Primary:                     to.BoolPtr(options.Primary),
			EnableAcceleratedNetworking: to.BoolPtr(true),
			EnableIPForwarding:          to.BoolPtr(options.Primary),
			NetworkSecurityGroup: &network.SecurityGroup{
				Name: &options.SecurityGroupID,
			},
		},
		Name:     &options.Name,
		Location: &mgr.Provider.Configuration.Location,
		Tags:     tags,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			*options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.InterfacesClient.Client)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			*options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	ni, err := future.Result(mgr.Provider.InterfacesClient)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			*options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	if ni.IPConfigurations == nil {
		err = errors.Errorf("no ip configuration for interface %s", *ni.ID)
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			*options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	return &ni, nil
}
func (mgr *NetworkInterfacesManager) Create(options api.CreateNetworkInterfaceOptions) (*api.NetworkInterface, error) {
	ni, err := mgr.create(&options)
	return convertNetworkInterface(ni), err
}

func convertNetworkInterface(ni *network.Interface) *api.NetworkInterface {
	ipConf := *ni.IPConfigurations
	var srvID string
	if srvName, ok := ni.Tags["server-id"]; ok {
		srvID = *srvName
	}
	var netID string
	if netName, ok := ni.Tags["network-id"]; ok {
		netID = *netName
	}
	var PrivateIPAddress string
	if ipConf[0].PrivateIPAddress != nil {
		PrivateIPAddress = *ipConf[0].PrivateIPAddress
	}
	var publicIPAddress string
	if ipConf[0].PublicIPAddress != nil {
		publicIPAddress = *ipConf[0].PublicIPAddress.IPAddress
	}

	return &api.NetworkInterface{
		ID:               *ni.Name,
		Name:             *ni.Name,
		MacAddress:       *ni.MacAddress,
		NetworkID:        netID,
		SubnetID:         *ipConf[0].Subnet.Name,
		ServerID:         srvID,
		PrivateIPAddress: PrivateIPAddress,
		PublicIPAddress:  publicIPAddress,
		SecurityGroupID:  *ni.NetworkSecurityGroup.Name,
	}
}

func (mgr *NetworkInterfacesManager) Delete(id string) error {
	future, err := mgr.Provider.InterfacesClient.Delete(context.Background(), mgr.resourceGroup(), id)
	if err != nil {
		return errors.Wrapf(err, "error deleting network interface %s", id)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.InterfacesClient.Client)
	return errors.Wrapf(err, "error deleting network interface %s", id)
}

func (mgr *NetworkInterfacesManager) Get(id string) (*api.NetworkInterface, error) {
	res, err := mgr.Provider.InterfacesClient.Get(context.Background(), mgr.resourceGroup(), id, "")
	if err != nil {
		return nil, errors.Wrapf(err, "error getting network interface %s", id)
	}

	return convertNetworkInterface(&res), nil

}

func checkNI(ni *api.NetworkInterface, options *api.ListNetworkInterfacesOptions) bool {
	if options == nil {
		return true
	}
	if options.ServerID != nil && *options.ServerID != ni.ServerID {
		return false
	}
	if options.NetworkID != nil && *options.NetworkID != ni.NetworkID {
		return false
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

func (mgr *NetworkInterfacesManager) list(options *api.ListNetworkInterfacesOptions) ([]network.Interface, error) {
	res, err := mgr.Provider.InterfacesClient.List(context.Background(), mgr.resourceGroup())
	if err != nil {
		return nil, err
	}
	var list []network.Interface
	for res.NotDone() {
		for _, ni := range res.Values() {
			n := convertNetworkInterface(&ni)
			if checkNI(n, options) {
				list = append(list, ni)
			}
		}
		err := res.NextWithContext(context.Background())
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (mgr *NetworkInterfacesManager) List(options *api.ListNetworkInterfacesOptions) ([]api.NetworkInterface, error) {
	nis, err := mgr.list(options)
	if err != nil {
		return nil, errors.Wrap(err, "error listing network interfaces")
	}
	var list []api.NetworkInterface
	for _, ni := range nis {
		list = append(list, *convertNetworkInterface(&ni))
	}
	return list, nil
}

func (mgr *NetworkInterfacesManager) Update(options api.UpdateNetworkInterfacesOptions) (*api.NetworkInterface, error) {
	res, err := mgr.Provider.InterfacesClient.Get(context.Background(), mgr.resourceGroup(), options.ID, "")
	if err != nil {
		return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
	}
	if options.SecurityGroupID != nil {
		res.NetworkSecurityGroup = &network.SecurityGroup{
			Name: options.SecurityGroupID,
		}
	}
	if options.ServerID != nil {
		res.VirtualMachine = &network.SubResource{ID: options.ServerID}
	}

	future, err := mgr.Provider.InterfacesClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), *res.Name, res)
	if err != nil {
		return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.InterfacesClient.Client)
	ni, err := mgr.Get(options.ID)
	return ni, errors.Wrapf(err, "error updating network interface %s", options.ID)
}
