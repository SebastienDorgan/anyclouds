package azure

import (
	"context"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/pkg/errors"
	"time"
)

type ServerTemplateManager struct {
	Provider *Provider
}

func (mgr *ServerTemplateManager) List() ([]api.ServerTemplate, error) {
	list, err := mgr.Provider.VirtualMachineSizesClient.List(context.Background(), mgr.Provider.Configuration.Location)
	if err != nil {
		return nil, errors.Wrap(err, "error listing server images")
	}
	var templates []api.ServerTemplate
	for _, size := range *list.Value {
		templates = append(templates, api.ServerTemplate{
			ID:                *size.Name,
			Name:              *size.Name,
			NumberOfCPUCore:   int(*size.NumberOfCores),
			RAMSize:           int(*size.MemoryInMB),
			SystemDiskSize:    int(*size.OsDiskSizeInMB / 1000),
			EphemeralDiskSize: int(*size.ResourceDiskSizeInMB / 1000),
			CreatedAt:         time.Date(2016, 1, 1, 0, 0, 0, 0, nil),
			Arch:              api.ArchAmd64,
			CPUFrequency:      0,
			NetworkSpeed:      0,
			GPUInfo:           nil,
			OneDemandPrice:    0,
		})
	}
	return templates, nil
}

func (mgr *ServerTemplateManager) Get(id string) (*api.ServerTemplate, error) {
	list, err := mgr.List()
	if err != nil {
		return nil, errors.Wrapf(err, "error getting server image %s", id)
	}
	for _, tpl := range list {
		if tpl.ID == id {
			return &tpl, nil
		}
	}
	return nil, errors.Errorf("error getting server image %s: image not found", id)
}
