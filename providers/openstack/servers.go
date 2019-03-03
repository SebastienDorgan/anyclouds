package openstack

import (
	"time"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/retry"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/pkg/errors"
)

//ServerManager defines Server management functions an anyclouds provider must provide
type ServerManager struct {
	OpenStack *Provider
}

func (mgr *ServerManager) createServer(options *api.ServerOptions) (*servers.Server, error) {
	opts := servers.CreateOpts{
		FlavorRef: options.TemplateID,
		ImageRef:  options.ImageID,
		Name:      options.Name,
	}
	for _, n := range options.Networks {
		opts.Networks = append(opts.Networks, servers.Network{
			UUID: n,
		})
	}
	return servers.Create(mgr.OpenStack.Compute, opts).Extract()
}

func (mgr *ServerManager) waitNotTransientState(srv *servers.Server) *retry.Result {
	get := func() (interface{}, error) {
		return mgr.Get(srv.ID)
	}
	finished := func(v interface{}, e error) bool {
		state := v.(*api.Server).State
		return state != api.ServerTransientState
	}
	return retry.With(get).For(3 * time.Minute).Every(time.Second).Until(finished).Go()

}

//Create creates an Server with options
func (mgr *ServerManager) Create(options *api.ServerOptions) (*api.Server, error) {
	srv, err := mgr.createServer(options)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error creating server")
	}
	res := mgr.waitNotTransientState(srv)
	if res.Timeout || res.LastError != nil {
		if res.LastError != nil {
			return nil, errors.Wrap(res.LastError, "Error creating server")
		}
		return nil, errors.Wrap(errors.Errorf("Timeout"), "Error creating server")
	}
	s := res.LastValue.(*api.Server)
	if s.State != api.ServerReady {
		return nil, errors.Wrap(errors.Errorf("Server in unexpected state: %s", s.State), "Error creating server")
	}
	//if no public IP is requested the server is ready to be used
	if !options.PublicIP {
		return s, nil
	}
	fip, err := floatingips.Create(mgr.OpenStack.Compute, &floatingips.CreateOpts{}).Extract()
	if err != nil {
		mgr.Delete(srv.ID)
		return nil, errors.Wrap(err, "Error creating server")
	}
	err = floatingips.AssociateInstance(mgr.OpenStack.Compute, srv.ID, &floatingips.AssociateOpts{
		FloatingIP: fip.IP,
	}).ExtractErr()
	if err != nil {
		return nil, errors.Wrap(err, "Error creating server")
	}
	//One more get to update PublicIPv4 and PublicIPv6
	return mgr.Get(srv.ID)
}

func (mgr *ServerManager) findIP(srvID string) *floatingips.FloatingIP {
	page, err := floatingips.List(mgr.OpenStack.Compute).AllPages()
	if err != nil {
		return nil
	}
	l, err := floatingips.ExtractFloatingIPs(page)
	if err != nil {
		return nil
	}
	for _, ip := range l {
		if ip.InstanceID == srvID {
			return &ip
		}
	}
	return nil
}

//Delete delete Server identified by id
func (mgr *ServerManager) Delete(id string) error {
	ip := mgr.findIP(id)
	err := servers.Delete(mgr.OpenStack.Compute, id).ExtractErr()
	if err != nil {
		return errors.Wrap(err, "Error deleting server")
	}
	if ip != nil {
		err = floatingips.Delete(mgr.OpenStack.Compute, ip.ID).ExtractErr()
	}
	return errors.Wrap(err, "Error deleting server")
}

func state(status string) api.ServerState {
	switch status {
	case "ACTIVE":
		return api.ServerReady
	case "DELETED":
		return api.ServerDeleted
	case "SOFT_DELETED":
		return api.ServerDeleted
	case "ERROR":
		return api.ServerInError
	case "SHELVED":
		return api.ServerUnknwonState
	case "SHELVED_OFFLOADED":
		return api.ServerUnknwonState
	case "PAUSED":
		return api.ServerPaused
	case "SHUTOFF":
		return api.ServerShutoff
	case "UNKNOWN":
		return api.ServerUnknwonState
	default:
		return api.ServerTransientState
	}

}

// convertAdresses converts adresses returned by the OpenStack driver arrange them by version in a map
func adresses(addresses map[string]interface{}) map[api.IPVersion][]string {
	addrs := make(map[api.IPVersion][]string)
	for _, obj := range addresses {
		for _, networkAddresses := range obj.([]interface{}) {
			address := networkAddresses.(map[string]interface{})
			version := address["version"].(float64)
			fixedIP := address["addr"].(string)
			switch version {
			case 4:
				addrs[api.IPVersion4] = append(addrs[api.IPVersion4], fixedIP)
			case 6:
				addrs[api.IPVersion6] = append(addrs[api.IPVersion6], fixedIP)
			}
		}
	}
	return addrs
}

func (mgr *ServerManager) server(srv *servers.Server) *api.Server {
	if srv == nil {
		return nil
	}
	flavorID, _ := flavors.IDFromName(mgr.OpenStack.Compute, srv.Flavor["original_name"].(string))
	return &api.Server{
		ID:         srv.ID,
		ImageID:    srv.Image["id"].(string),
		TemplateID: flavorID,
		State:      state(srv.Status),
		PublicIPv4: srv.AccessIPv4,
		PublicIPv6: srv.AccessIPv6,
		PrivateIPs: adresses(srv.Addresses),
		Name:       srv.Name,
	}
}

//List list Servers
func (mgr *ServerManager) List() ([]api.Server, error) {
	page, err := servers.List(mgr.OpenStack.Compute, servers.ListOpts{}).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing servers")
	}
	l, err := servers.ExtractServers(page)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing servers")
	}

	res := []api.Server{}
	for _, srv := range l {

		res = append(res, *mgr.server(&srv))
	}
	return res, nil
}

//Get get Servers
func (mgr *ServerManager) Get(id string) (*api.Server, error) {
	srv, err := servers.Get(mgr.OpenStack.Compute, id).Extract()
	return mgr.server(srv), errors.Wrap(ProviderError(err), "Error getting server")
}

//Start starts an Server
func (mgr *ServerManager) Start(id string) error {
	err := startstop.Start(mgr.OpenStack.Compute, id).ExtractErr()
	return errors.Wrap(ProviderError(err), "Error starting server")
}

//Stop stops an Server
func (mgr *ServerManager) Stop(id string) error {
	err := startstop.Stop(mgr.OpenStack.Compute, id).ExtractErr()
	return errors.Wrap(ProviderError(err), "Error stoping server")
}
