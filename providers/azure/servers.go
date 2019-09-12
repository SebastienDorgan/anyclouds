package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/sethvargo/go-password/password"
	"time"
)

type ServerManager struct {
	Provider *Provider
}

func (mgr *ServerManager) resourceGroup() string {
	return mgr.Provider.Configuration.ResourceGroupName
}

func (mgr *ServerManager) createNetworkInterfaces(options *api.CreateServerOptions) ([]compute.NetworkInterfaceReference, error) {
	var nis []compute.NetworkInterfaceReference
	for _, sn := range options.Subnets {
		ni, err := mgr.Provider.NetworkInterfacesManager.Create(api.CreateNetworkInterfaceOptions{
			Name:             fmt.Sprintf("NI-%s", sn.Name),
			NetworkID:        sn.NetworkID,
			SubnetID:         sn.ID,
			ServerID:         nil,
			Primary:          false,
			PrivateIPAddress: nil,
			SecurityGroupID:  options.DefaultSecurityGroup,
		})
		if err != nil {
			return nil, err
		}
		nis = append(nis, compute.NetworkInterfaceReference{
			NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{Primary: to.BoolPtr(false)},
			ID:                                  &ni.ID,
		})
	}
	if nis != nil {
		nis[0].Primary = to.BoolPtr(true)
	}
	return nis, nil
}
func (mgr *ServerManager) Create(options api.CreateServerOptions) (*api.Server, *api.CreateServerError) {
	publisher, offer, sku, version := parseImageID(options.ImageID)
	nis, err := mgr.createNetworkInterfaces(&options)
	if err != nil {
		return nil, api.NewCreateServerError(err, options)
	}
	priority := compute.Regular
	if options.LowPriorityServerOptions != nil {
		priority = compute.Low
	}
	future, err := mgr.Provider.VirtualMachinesClient.CreateOrUpdate(
		context.Background(),
		mgr.resourceGroup(),
		options.Name,
		compute.VirtualMachine{
			Location: to.StringPtr(mgr.Provider.Configuration.Location),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.VirtualMachineSizeTypes(options.TemplateID),
				},
				Priority: priority,
				StorageProfile: &compute.StorageProfile{
					ImageReference: &compute.ImageReference{
						Publisher: to.StringPtr(publisher),
						Offer:     to.StringPtr(offer),
						Sku:       to.StringPtr(sku),
						Version:   to.StringPtr(version),
					},
				},
				OsProfile: &compute.OSProfile{
					ComputerName:  to.StringPtr(options.Name),
					AdminUsername: to.StringPtr(mgr.Provider.Configuration.DefaultVMUserName),
					AdminPassword: to.StringPtr(password.MustGenerate(16, 5, 5, false, false)),
					LinuxConfiguration: &compute.LinuxConfiguration{
						SSH: &compute.SSHConfiguration{
							PublicKeys: &[]compute.SSHPublicKey{
								{
									Path: to.StringPtr(
										fmt.Sprintf("/home/%s/.ssh/authorized_keys",
											mgr.Provider.Configuration.DefaultVMUserName)),
									KeyData: to.StringPtr(string(options.KeyPair.PublicKey)),
								},
							},
						},
					},
				},
				NetworkProfile: &compute.NetworkProfile{
					NetworkInterfaces: &nis,
				},
			},
		},
	)
	if err != nil {
		return nil, api.NewCreateServerError(err, options)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.VirtualMachinesClient.Client)
	if err != nil {
		return nil, api.NewCreateServerError(err, options)
	}
	vm, err := future.Result(mgr.Provider.VirtualMachinesClient)
	if err != nil {
		return nil, api.NewCreateServerError(err, options)
	}
	return mgr.server(&vm), nil
}

func imageID(reference *compute.ImageReference) string {
	if reference == nil {
		return ""
	}
	return createImageID(*reference.Publisher, *reference.Offer, *reference.Sku, *reference.Version)
}

func (mgr *ServerManager) server(vm *compute.VirtualMachine) *api.Server {
	leasingType := api.LeasingTypeOnDemand
	if vm.Priority == compute.Low {
		leasingType = api.LeasingTypeSpot
	}
	srv := &api.Server{
		ID:            *vm.Name,
		Name:          *vm.Name,
		TemplateID:    string(vm.HardwareProfile.VMSize),
		ImageID:       imageID(vm.StorageProfile.ImageReference),
		State:         "",
		CreatedAt:     time.Time{},
		LeasingType:   leasingType,
		LeaseDuration: 0,
	}
	return srv
}

func (mgr *ServerManager) Delete(id string) *api.DeleteServerError {
	future, err := mgr.Provider.VirtualMachinesClient.Delete(context.Background(), mgr.resourceGroup(), id)
	if err != nil {
		return api.NewDeleteServerError(err, id)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.VirtualMachinesClient.Client)
	return api.NewDeleteServerError(err, id)
}

func (mgr *ServerManager) List() ([]api.Server, *api.ListServersError) {
	it, err := mgr.Provider.VirtualMachinesClient.List(context.Background(), mgr.resourceGroup())
	if err != nil {
		return nil, api.NewListServersError(err)
	}
	var servers []api.Server
	for it.NotDone() {
		vms := it.Values()
		for _, vm := range vms {
			servers = append(servers, *mgr.server(&vm))
		}
		err = it.NextWithContext(context.Background())
		if err != nil {
			return nil, api.NewListServersError(err)
		}
	}
	return servers, nil
}

func (mgr *ServerManager) list() ([]compute.VirtualMachine, error) {
	it, err := mgr.Provider.VirtualMachinesClient.List(context.Background(), mgr.resourceGroup())
	if err != nil {
		return nil, err
	}
	var servers []compute.VirtualMachine
	for it.NotDone() {
		vms := it.Values()
		for _, vm := range vms {
			servers = append(servers, vm)
		}
		err = it.NextWithContext(context.Background())
		if err != nil {
			return nil, err
		}
	}
	return servers, nil
}

func (mgr *ServerManager) get(id string) (*compute.VirtualMachine, error) {
	res, err := mgr.Provider.VirtualMachinesClient.Get(context.Background(), mgr.resourceGroup(), id, "")
	return &res, err
}

func (mgr *ServerManager) Get(id string) (*api.Server, *api.GetServerError) {
	vm, err := mgr.get(id)
	return mgr.server(vm), api.NewGetServerError(err, id)
}

func (mgr *ServerManager) Start(id string) *api.StartServerError {
	future, err := mgr.Provider.VirtualMachinesClient.Start(context.Background(), mgr.resourceGroup(), id)
	if err != nil {
		return api.NewStartServerError(err, id)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.VirtualMachinesClient.Client)
	return api.NewStartServerError(err, id)
}

func (mgr *ServerManager) Stop(id string) *api.StopServerError {
	future, err := mgr.Provider.VirtualMachinesClient.PowerOff(context.Background(), mgr.resourceGroup(), id, to.BoolPtr(false))
	if err != nil {
		return api.NewStopServerError(err, id)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.VirtualMachinesClient.Client)
	return api.NewStopServerError(err, id)
}

func (mgr *ServerManager) Resize(id string, templateID string) *api.ResizeServerError {
	vm, err := mgr.get(id)
	if err != nil {
		return api.NewResizeServerError(err, id, templateID)
	}
	vm.HardwareProfile.VMSize = compute.VirtualMachineSizeTypes(templateID)
	future, err := mgr.Provider.VirtualMachinesClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), id, *vm)
	if err != nil {
		return api.NewResizeServerError(err, id, templateID)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.VirtualMachinesClient.Client)
	return api.NewResizeServerError(err, id, templateID)
}
