package openstack

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
)

//ServerTemplateManager defines Server template management functions a anyclouds provider must provide
type ServerTemplateManager struct {
	OpenStack *Provider
}

//List returns available VM templates
func (mgr *ServerTemplateManager) List() ([]api.ServerTemplate, *api.ListServerTemplatesError) {
	page, err := flavors.ListDetail(mgr.OpenStack.Compute, flavors.ListOpts{}).AllPages()
	if err != nil {
		return nil, api.NewListServerTemplatesError(err)
	}
	l, err := flavors.ExtractFlavors(page)
	var templates []api.ServerTemplate

	for _, f := range l {
		t := api.ServerTemplate{
			ID:                f.ID,
			Name:              f.Name,
			NumberOfCPUCore:   f.VCPUs,
			RAMSize:           f.RAM,
			SystemDiskSize:    f.Disk,
			EphemeralDiskSize: f.Ephemeral,
			Arch:              api.ArchUnknown,
		}
		templates = append(templates, t)
	}
	return templates, nil
}

//Get returns the template identified by ids
func (mgr *ServerTemplateManager) Get(id string) (*api.ServerTemplate, *api.GetServerTemplateError) {
	f, err := flavors.Get(mgr.OpenStack.Compute, id).Extract()
	if err != nil {
		return nil, api.NewGetServerTemplateError(err, id)
	}
	return &api.ServerTemplate{
		ID:                f.ID,
		Name:              f.Name,
		NumberOfCPUCore:   f.VCPUs,
		RAMSize:           f.RAM,
		SystemDiskSize:    f.Disk,
		EphemeralDiskSize: f.Ephemeral,
		Arch:              api.ArchUnknown,
	}, nil
}
