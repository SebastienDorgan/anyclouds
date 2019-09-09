package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/pkg/errors"
)

type ServerManager struct {
	Provider *Provider
}

func (mgr *ServerManager) resourceGroup() string {
	return mgr.Provider.Configuration.ResourceGroupName
}
func (mgr *ServerManager) Create(options *api.CreateServerOptions) (*api.Server, error) {
	publisher, offer, sku, version := ParseId(options.ImageID)
	future, err := mgr.Provider.VirtualMachinesClient.CreateOrUpdate(
		context.Background(),
		mgr.resourceGroup(),
		options.Name,
		compute.VirtualMachine{
			Location: to.StringPtr(mgr.Provider.Configuration.Location),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.VirtualMachineSizeTypesBasicA0,
				},
				StorageProfile: &compute.StorageProfile{
					ImageReference: &compute.ImageReference{
						Publisher: to.StringPtr(publisher),
						Offer:     to.StringPtr(offer),
						Sku:       to.StringPtr(sku),
						Version:   to.StringPtr(version),
					},
				},
				OsProfile: &compute.OSProfile{
					ComputerName: to.StringPtr(options.Name),
					//AdminUsername: to.StringPtr(username),
					//AdminPassword: to.StringPtr(password),
					LinuxConfiguration: &compute.LinuxConfiguration{
						SSH: &compute.SSHConfiguration{
							PublicKeys: &[]compute.SSHPublicKey{
								{
									Path: to.StringPtr(
										fmt.Sprintf("/home/%s/.ssh/authorized_keys",
											"dsjjkldkldgskjlgdjklgdskjljdgjgdjklgdjklgjgjgd")),
									KeyData: to.StringPtr(string(options.KeyPair.PublicKey)),
								},
							},
						},
					},
				},
				NetworkProfile: &compute.NetworkProfile{
					NetworkInterfaces: &[]compute.NetworkInterfaceReference{
						{
							ID: to.StringPtr("nic.ID"),
							NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
								Primary: to.BoolPtr(true),
							},
						},
					},
				},
			},
		},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating server %s", options.Name)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.VirtualMachinesClient.Client)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating server %s", options.Name)
	}
	return nil, nil
}

func (mgr *ServerManager) Delete(id string) error {
	panic("implement me")
}

func (mgr *ServerManager) List() ([]api.Server, error) {
	panic("implement me")
}

func (mgr *ServerManager) get(id string) (*compute.VirtualMachine, error) {
	res, err := mgr.Provider.VirtualMachinesClient.Get(context.Background(), mgr.resourceGroup(), id, "")
	return &res, err
}

func (mgr *ServerManager) Get(id string) (*api.Server, error) {
	panic("implement me")
}

func (mgr *ServerManager) Start(id string) error {
	panic("implement me")
}

func (mgr *ServerManager) Stop(id string) error {
	panic("implement me")
}

func (mgr *ServerManager) Resize(id string, templateID string) error {
	panic("implement me")
}
