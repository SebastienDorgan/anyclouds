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

func (mgr *NetworkInterfacesManager) Create(options *api.NetworkInterfaceOptions) (*api.NetworkInterface, error) {
	future, err := mgr.Provider.InterfacesClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), options.Name, network.Interface{
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
			VirtualMachine: &network.SubResource{ID: &options.ServerID},
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				{
					InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAddress:          options.IPAddress,
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
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.InterfacesClient.Client)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	ni, err := future.Result(mgr.Provider.InterfacesClient)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	if ni.IPConfigurations == nil {
		err = errors.Errorf("no ip configuration for interface %s", *ni.ID)
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	return convert(&ni), nil
}

func convert(ni *network.Interface) *api.NetworkInterface {
	ipConf := *ni.IPConfigurations
	return &api.NetworkInterface{
		ID:         *ni.Name,
		Name:       *ni.Name,
		MacAddress: *ni.MacAddress,
		NetworkID:  *ipConf[0].Subnet.ID,
		SubnetID:   *ipConf[0].Subnet.ID,
		ServerID:   *ni.VirtualMachine.ID,
		IPAddress:  *ipConf[0].PrivateIPAddress,
		Primary:    *ni.Primary,
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
	return convert(&res), nil

}

func (mgr *NetworkInterfacesManager) List() ([]api.NetworkInterface, error) {
	res, err := mgr.Provider.InterfacesClient.List(context.Background(), mgr.resourceGroup())
	if err != nil {
		return nil, errors.Wrap(err, "error listing network interfaces")
	}
	var list []api.NetworkInterface
	for res.NotDone() {
		for _, ni := range res.Values() {
			list = append(list, *convert(&ni))
		}
		err := res.NextWithContext(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "error listing network interfaces")
		}
	}
	return list, nil
}

func (mgr *NetworkInterfacesManager) Update(options *api.NetworkInterfacesUpdateOptions) (*api.NetworkInterface, error) {
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
