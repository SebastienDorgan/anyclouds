package aws

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
)

//ServerManager defines Server management functions an anyclouds provider must provide
type ServerManager struct {
	AWS *Provider
}

func createSpotInstance(options *api.CreateServerOptions) (*api.Server, error) {
	return nil, nil
}

//Create creates an Server with options
func (mgr *ServerManager) Create(options *api.CreateServerOptions) (*api.Server, error) {
	if options.Spot {
		return createSpotInstance(options)
	}
	return nil, nil
}

func (mgr *ServerManager) findIP(srvID string) *floatingips.FloatingIP {
	return nil
}

//Delete delete Server identified by id
func (mgr *ServerManager) Delete(id string) error {
	return nil
}

//List list Servers
func (mgr *ServerManager) List() ([]api.Server, error) {
	return nil, nil
}

//Get get Servers
func (mgr *ServerManager) Get(id string) (*api.Server, error) {
	return nil, nil
}

//Start starts an Server
func (mgr *ServerManager) Start(id string) error {
	return nil
}

//Stop stops an Server
func (mgr *ServerManager) Stop(id string) error {
	return nil
}

//Resize resize a server
func (mgr *ServerManager) Resize(id string, templateID string) error {
	return nil
}
