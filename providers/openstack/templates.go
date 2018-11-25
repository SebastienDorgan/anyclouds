package openstack

import (
	"github.com/SebastienDorgan/anyclouds/providers"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/pkg/errors"
)

//InstanceTemplateManager defines instance template management functions a anyclouds provider must provide
type InstanceTemplateManager struct {
	OpenStack *Provider
}

//List returns available VM templates
func (mgr *InstanceTemplateManager) List(filter *providers.ResourceFilter) ([]providers.InstanceTemplate, error) {
	page, err := flavors.ListDetail(mgr.OpenStack.Compute, flavors.ListOpts{}).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing templates")
	}
	l, err := flavors.ExtractFlavors(page)
	tpls := []providers.InstanceTemplate{}

	for _, f := range l {
		if len(filter.Name) > 0 {
			if f.Name != filter.Name {
				continue
			}
		}
		if len(filter.ID) > 0 {
			if f.ID != filter.ID {
				continue
			}
		}
		t := providers.InstanceTemplate{
			ID:                f.ID,
			Name:              f.Name,
			CPUArch:           providers.ArchX86_64,
			NumberOfCPUCore:   f.VCPUs,
			CPUCoreFrequency:  1.0,
			RAMSize:           float64(f.RAM) / 1024.0,
			SystemDiskSize:    f.Disk,
			EphemeralDiskSize: f.Ephemeral,
			NumberOfGPU:       0,
			NumberOfGPUCore:   0,
			GPURAMSize:        0,
			GPUCoreFrequency:  0,
		}
		tpls = append(tpls, t)
	}
	return tpls, nil
}

//Get returns the template identified by ids
func (mgr *InstanceTemplateManager) Get(id string) (*providers.InstanceTemplate, error) {
	f, err := flavors.Get(mgr.OpenStack.Compute, id).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting template")
	}
	return &providers.InstanceTemplate{
		ID:                f.ID,
		Name:              f.Name,
		CPUArch:           providers.ArchX86_64,
		NumberOfCPUCore:   f.VCPUs,
		CPUCoreFrequency:  1.0,
		RAMSize:           float64(f.RAM) / 1024.0,
		SystemDiskSize:    f.Disk,
		EphemeralDiskSize: f.Ephemeral,
		NumberOfGPU:       0,
		NumberOfGPUCore:   0,
		GPURAMSize:        0,
		GPUCoreFrequency:  0,
	}, nil
}
