package openstack

import (
	"fmt"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/anyclouds/providers"
	"github.com/SebastienDorgan/retry"
	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/startstop"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/pkg/errors"
	"time"
)

//ServerManager defines Server management functions an anyclouds provider must provide
type ServerManager struct {
	OpenStack *Provider
}

//key_name

type createServerOpts struct {
	servers.CreateOpts
	KeyName string
}

func (o createServerOpts) ToServerCreateMap() (map[string]interface{}, error) {
	m, err := o.CreateOpts.ToServerCreateMap()
	if err != nil {
		return nil, err
	}
	m["key_name"] = o.KeyName
	return m, nil

}

func (mgr *ServerManager) networks(subnets []string) ([]servers.Network, error) {
	allSubnets, err := mgr.OpenStack.NetworkManager.listAllSubnets()
	var nets []servers.Network
	if err != nil {
		return nil, err
	}
	for _, asn := range allSubnets {
		for _, sn := range subnets {
			if asn.ID == sn {
				nets = append(nets, servers.Network{
					UUID: asn.NetworkID,
				})
			}
		}
	}
	return nets, nil
}
func (mgr *ServerManager) createServer(options *api.CreateServerOptions) (*api.Server, error) {
	nets, err := mgr.networks(options.Subnets)
	if err != nil {
		return nil, err
	}
	opts := servers.CreateOpts{
		FlavorRef:      options.TemplateID,
		ImageRef:       options.ImageID,
		Name:           options.Name,
		SecurityGroups: options.SecurityGroups,
		Networks:       nets,
	}
	keyId := uuid.New().String()
	err = mgr.OpenStack.KeyPairManager.Import(keyId, options.KeyPair.PublicKey)
	defer func() { _ = mgr.OpenStack.KeyPairManager.Delete(keyId) }()
	if err != nil {
		return nil, ProviderError(err)
	}
	srv, err := servers.Create(mgr.OpenStack.Compute, createServerOpts{
		CreateOpts: opts,
		KeyName:    "key",
	}).Extract()
	if err != nil {
		return nil, ProviderError(err)
	}
	s, err := providers.WaitUntilServerReachStableState(mgr, srv.ID)
	if s != nil && s.State != api.ServerReady {
		_ = mgr.Delete(s.ID)
		return nil, errors.Errorf("server in unexpected state: %s", s.State)
	}
	if s == nil {
		return nil, errors.Errorf("server in unexpected state: %s", "nil")
	}

	return s, nil
}

//Create creates an Server with options
func (mgr *ServerManager) Create(options *api.CreateServerOptions) (*api.Server, error) {
	srv, err := mgr.createServer(options)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating server %s", options.Name)
	}
	//if no public IP is requested the server is ready to be used
	if !options.PublicIP {
		return srv, nil
	}
	fip, err := floatingips.Create(mgr.OpenStack.Compute, &floatingips.CreateOpts{}).Extract()
	if err != nil {
		_ = mgr.Delete(srv.ID)
		return nil, errors.Wrap(err, "Error creating server")
	}
	err = floatingips.AssociateInstance(mgr.OpenStack.Compute, srv.ID, &floatingips.AssociateOpts{
		FloatingIP: fip.IP,
	}).ExtractErr()
	if err != nil {
		return nil, errors.Wrap(err, "Error creating server")
	}
	return providers.WaitUntilServerReachStableState(mgr, srv.ID)

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
		return api.ServerUnknownState
	case "SHELVED_OFFLOADED":
		return api.ServerUnknownState
	case "PAUSED":
		return api.ServerPaused
	case "SHUTOFF":
		return api.ServerShutoff
	case "UNKNOWN":
		return api.ServerUnknownState
	default:
		return api.ServerPending
	}

}

// convertAddresses converts addresses returned by the OpenStack driver arrange them by version in a map
func addresses(addresses map[string]interface{}) map[api.IPVersion][]string {
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

func ids(m []map[string]interface{}) []string {
	kl := make([]string, len(m))
	for i, m := range m {
		kl[i] = m["id"].(string)
	}
	return kl
}

func (mgr *ServerManager) server(srv *servers.Server) *api.Server {
	if srv == nil {
		return nil
	}
	flavorID, _ := flavors.IDFromName(mgr.OpenStack.Compute, srv.Flavor["original_name"].(string))
	return &api.Server{
		ID:             srv.ID,
		ImageID:        srv.Image["id"].(string),
		TemplateID:     flavorID,
		State:          state(srv.Status),
		PublicIPv4:     srv.AccessIPv4,
		AccessIPv6:     srv.AccessIPv6,
		PrivateIPs:     addresses(srv.Addresses),
		Name:           srv.Name,
		CreatedAt:      srv.Created,
		SecurityGroups: ids(srv.SecurityGroups),
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

	var res []api.Server
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
	return errors.Wrap(ProviderError(err), "Error stopping server")
}

func (mgr *ServerManager) waitResize(id string) *retry.Result {
	get := func() (interface{}, error) {
		return servers.Get(mgr.OpenStack.Compute, id).Extract()
	}
	finished := func(v interface{}, e error) bool {
		state := v.(*servers.Server).Status
		return state == "RESIZE"
	}
	return retry.With(get).For(3 * time.Minute).Every(time.Second).Until(finished).Go()
}

//Resize resize a server
func (mgr *ServerManager) Resize(id string, templateID string) error {
	err := servers.Resize(mgr.OpenStack.Compute, id, servers.ResizeOpts{
		FlavorRef: templateID,
	}).ExtractErr()
	if err != nil {
		servers.RevertResize(mgr.OpenStack.Compute, id)
		return errors.Wrap(ProviderError(err), "error resizing server")
	}
	res := mgr.waitResize(id)
	if res.LastError != nil {
		return errors.Wrap(ProviderError(res.LastError), "error resizing server")
	}
	srv := res.LastValue.(*servers.Server)
	if srv == nil {
		servers.RevertResize(mgr.OpenStack.Compute, id)
		err := fmt.Errorf("unable to retrive server state")
		return errors.Wrap(err, "error resizing server")
	}
	if srv.Status != "RESIZE" {
		servers.RevertResize(mgr.OpenStack.Compute, id)
		err := fmt.Errorf("unexpected server state: %s", srv.Status)
		return errors.Wrap(err, "error resizing server")
	}
	err = servers.ConfirmResize(mgr.OpenStack.Compute, id).ExtractErr()
	if err != nil {
		servers.RevertResize(mgr.OpenStack.Compute, id)
		return errors.Wrap(ProviderError(err), "error resizing server")
	}
	return nil
}
