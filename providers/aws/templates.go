package aws

import (
	"github.com/SebastienDorgan/anyclouds/api"
)

//ServerTemplateManager defines Server template management functions a anyclouds provider must provide
type ServerTemplateManager struct {
	AWS *Provider
}

//List returns available VM templates
func (mgr *ServerTemplateManager) List() ([]api.ServerTemplate, error) {
	return nil, nil
}

//Get returns the template identified by ids
func (mgr *ServerTemplateManager) Get(id string) (*api.ServerTemplate, error) {
	return nil, nil
}
