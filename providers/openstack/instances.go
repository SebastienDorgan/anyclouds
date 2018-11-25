package openstack

import (
	"github.com/SebastienDorgan/anyclouds/providers"
)

//InstanceManager defines instance management functions an anyclouds provider must provide
type InstanceManager struct {
	OpenStack *Provider
}

//Create creates an instance with options
func (mgr *InstanceManager) Create(options *providers.InstanceOptions) (*providers.Instance, error) {
	return nil, nil
}

//Delete delete instance identified by id
func (mgr *InstanceManager) Delete(id string) error {
	return nil
}

//List list instances
func (mgr *InstanceManager) List(filter *providers.ResourceFilter) ([]providers.Instance, error) {
	return nil, nil
}

//Get get instances
func (mgr *InstanceManager) Get(id string) (*providers.Instance, error) {
	return nil, nil
}

//Start starts an instance
func (mgr *InstanceManager) Start(id string) error {
	return nil
}

//Stop stops an instance
func (mgr *InstanceManager) Stop(id string) error {
	return nil
}
