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
	Provider *Provider
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

func (mgr *ServerManager) networks(subnets []api.Subnet) []servers.Network {
	var nets []servers.Network
	for _, asn := range subnets {
		nets = append(nets, servers.Network{
			UUID: asn.NetworkID,
		})
	}
	return nets
}
func (mgr *ServerManager) createServer(options *api.CreateServerOptions) (*api.Server, error) {
	opts := servers.CreateOpts{
		FlavorRef:      options.TemplateID,
		ImageRef:       options.ImageID,
		Name:           options.Name,
		SecurityGroups: []string{options.DefaultSecurityGroup},
		Networks:       mgr.networks(options.Subnets),
	}
	keyID := uuid.New().String()
	err := mgr.Provider.KeyPairManager.Import(keyID, options.KeyPair.PublicKey)
	defer func() { _ = mgr.Provider.KeyPairManager.Delete(keyID) }()
	if err != nil {
		return nil, UnwrapOpenStackError(err)
	}
	srv, err := servers.Create(mgr.Provider.BaseServices.Compute, createServerOpts{
		CreateOpts: opts,
		KeyName:    "key",
	}).Extract()
	if err != nil {
		return nil, UnwrapOpenStackError(err)
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
func (mgr *ServerManager) Create(options api.CreateServerOptions) (*api.Server, api.CreateServerError) {
	srv, err := mgr.createServer(&options)
	if err != nil {
		return nil, api.NewCreateServerError(err, options)
	}
	srv, err = providers.WaitUntilServerReachStableState(mgr, srv.ID)
	return nil, api.NewCreateServerError(err, options)
}

func (mgr *ServerManager) findIP(srvID string) *floatingips.FloatingIP {
	page, err := floatingips.List(mgr.Provider.BaseServices.Compute).AllPages()
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
func (mgr *ServerManager) Delete(id string) api.DeleteServerError {
	ip := mgr.findIP(id)
	err := servers.Delete(mgr.Provider.BaseServices.Compute, id).ExtractErr()
	if err != nil {
		return api.NewDeleteServerError(UnwrapOpenStackError(err), id)
	}
	if ip != nil {
		err = floatingips.Delete(mgr.Provider.BaseServices.Compute, ip.ID).ExtractErr()
	}
	return api.NewDeleteServerError(UnwrapOpenStackError(err), id)
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

func (mgr *ServerManager) server(srv *servers.Server) *api.Server {
	if srv == nil {
		return nil
	}
	flavorID, _ := flavors.IDFromName(mgr.Provider.BaseServices.Compute, srv.Flavor["original_name"].(string))
	return &api.Server{
		ID:         srv.ID,
		ImageID:    srv.Image["id"].(string),
		TemplateID: flavorID,
		State:      state(srv.Status),
		Name:       srv.Name,
		CreatedAt:  srv.Created,
	}
}

//List list Servers
func (mgr *ServerManager) List() ([]api.Server, api.ListServersError) {
	page, err := servers.List(mgr.Provider.BaseServices.Compute, servers.ListOpts{}).AllPages()
	if err != nil {
		return nil, api.NewListServersError(UnwrapOpenStackError(err))
	}
	l, err := servers.ExtractServers(page)
	if err != nil {
		return nil, api.NewListServersError(UnwrapOpenStackError(err))
	}

	var res []api.Server
	for _, srv := range l {

		res = append(res, *mgr.server(&srv))
	}
	return res, nil
}

//Get get Servers
func (mgr *ServerManager) Get(id string) (*api.Server, api.GetServerError) {
	srv, err := servers.Get(mgr.Provider.BaseServices.Compute, id).Extract()
	return mgr.server(srv), api.NewGetServerError(UnwrapOpenStackError(err), id)
}

//Start starts an Server
func (mgr *ServerManager) Start(id string) api.StartServerError {
	err := startstop.Start(mgr.Provider.BaseServices.Compute, id).ExtractErr()
	return api.NewStartServerError(UnwrapOpenStackError(err), id)
}

//Stop stops an Server
func (mgr *ServerManager) Stop(id string) api.StopServerError {
	err := startstop.Stop(mgr.Provider.BaseServices.Compute, id).ExtractErr()
	return api.NewStopServerError(UnwrapOpenStackError(err), id)
}

func (mgr *ServerManager) waitResize(id string) *retry.Result {
	get := func() (interface{}, error) {
		return servers.Get(mgr.Provider.BaseServices.Compute, id).Extract()
	}
	finished := func(v interface{}, e error) bool {
		state := v.(*servers.Server).Status
		return state == "RESIZE"
	}
	return retry.With(get).For(3 * time.Minute).Every(time.Second).Until(finished).Go()
}

//Resize resize a server
func (mgr *ServerManager) Resize(id string, templateID string) api.ResizeServerError {
	err := servers.Resize(mgr.Provider.BaseServices.Compute, id, servers.ResizeOpts{
		FlavorRef: templateID,
	}).ExtractErr()
	if err != nil {
		servers.RevertResize(mgr.Provider.BaseServices.Compute, id)
		return api.NewResizeServerError(err, id, templateID)
	}
	res := mgr.waitResize(id)
	if res.LastError != nil {
		return api.NewResizeServerError(err, id, templateID)
	}
	srv := res.LastValue.(*servers.Server)
	if srv == nil {
		servers.RevertResize(mgr.Provider.BaseServices.Compute, id)
		err := fmt.Errorf("unable to retrive server state")
		return api.NewResizeServerError(err, id, templateID)
	}
	if srv.Status != "RESIZE" {
		servers.RevertResize(mgr.Provider.BaseServices.Compute, id)
		err := fmt.Errorf("unexpected server state: %s", srv.Status)
		return api.NewResizeServerError(err, id, templateID)
	}
	err = servers.ConfirmResize(mgr.Provider.BaseServices.Compute, id).ExtractErr()
	if err != nil {
		servers.RevertResize(mgr.Provider.BaseServices.Compute, id)
		return api.NewResizeServerError(err, id, templateID)
	}
	return nil
}
