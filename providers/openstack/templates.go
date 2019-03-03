package openstack

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/pkg/errors"
)

//ServerTemplateManager defines Server template management functions a anyclouds provider must provide
type ServerTemplateManager struct {
	OpenStack *Provider
}

//List returns available VM templates
func (mgr *ServerTemplateManager) List() ([]api.ServerTemplate, error) {
	page, err := flavors.ListDetail(mgr.OpenStack.Compute, flavors.ListOpts{}).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing templates")
	}
	l, err := flavors.ExtractFlavors(page)
	tpls := []api.ServerTemplate{}

	for _, f := range l {
		t := api.ServerTemplate{
			ID:                f.ID,
			Name:              f.Name,
			NumberOfCPUCore:   f.VCPUs,
			RAMSize:           float64(f.RAM) / 1024.0,
			SystemDiskSize:    f.Disk,
			EphemeralDiskSize: f.Ephemeral,
		}
		tpls = append(tpls, t)
	}
	return tpls, nil
}

//Get returns the template identified by ids
func (mgr *ServerTemplateManager) Get(id string) (*api.ServerTemplate, error) {
	f, err := flavors.Get(mgr.OpenStack.Compute, id).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting template")
	}
	return &api.ServerTemplate{
		ID:                f.ID,
		Name:              f.Name,
		NumberOfCPUCore:   f.VCPUs,
		RAMSize:           float64(f.RAM) / 1024.0,
		SystemDiskSize:    f.Disk,
		EphemeralDiskSize: f.Ephemeral,
	}, nil
}
