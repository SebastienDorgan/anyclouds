package openstack

import (
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"io"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/spf13/viper"

	gc "github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/pkg/errors"
)

/*Config fields are the union of those recognized by each Provider identity implementation and
provider.
*/
type Config struct {
	// IdentityEndpoint specifies the HTTP endpoint that is required to work with
	// the Identity API of the appropriate version. While it's ultimately needed by
	// all of the identity services, it will often be populated by a provider-level
	// function.
	IdentityEndpoint string

	// Username is required if using Identity V2 API. Consult with your provider's
	// control panel to discover your account's username. In Identity V3, either
	// UserID or a combination of Username and DomainID or DomainName are needed.
	Username string
	UserID   string

	// Exactly one of Password or APIKey is required for the Identity V2 and V3
	// APIs. Consult with your provider's control panel to discover your account's
	// preferred method of authentication.
	Password string
	APIKey   string

	// At most one of DomainID and DomainName must be provided if using Username
	// with Identity V3. Otherwise, either are optional.
	DomainID   string
	DomainName string

	// The TenantID and TenantName fields are optional for the Identity V2 API.
	// Some api allow you to specify a TenantName instead of the TenantId.
	// Some require both. Your provider's authentication policies will determine
	// how these fields influence authentication.
	TenantID   string
	TenantName string

	// AllowReauth should be set to true if you grant permission for anyclouds to
	// cache your credentials in memory, and to allow anyclouds to attempt to
	// re-authenticate automatically if/when your token expires.  If you set it to
	// false, it will not cache these settings, but re-authentication will not be
	// possible.  This setting defaults to false.
	//
	// NOTE: The reauth function will try to re-authenticate endlessly if left unchecked.
	// The way to limit the number of attempts is to provide a custom HTTP client to the provider client
	// and provide a transport that implements the RoundTripper interface and stores the number of failed retries.
	// For an example of this, see here: https://github.com/gophercloud/rack/blob/1.0.0/auth/clients.go#L311
	AllowReauth bool

	// TokenID allows users to authenticate (possibly as another user) with an
	// authentication token ID.
	TokenID string

	//Openstack region (data center) where the infrastructure will be created
	Region string

	//PublicIPPool name of the floating IP pool
	//Necessary only if UseFloatingIP is true
	FloatingIPPool string

	//Name of the external network used to get public ip addresses
	ExternalNetworkName string
}

//UnwrapOpenStackError creates an error string from openstack api error
func UnwrapOpenStackError(err error) error {
	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case gc.ErrDefault401:
		return errors.Errorf("code: 401, reason: %s", string(e.Body[:]))
	case *gc.ErrDefault401:
		return errors.Errorf("code: 401, reason: %s", string(e.Body[:]))
	case gc.ErrDefault404:
		return errors.Errorf("code: 404, reason: %s", string(e.Body[:]))
	case *gc.ErrDefault404:
		return errors.Errorf("code: 404, reason: %s", string(e.Body[:]))
	case gc.ErrDefault500:
		return errors.Errorf("code: 500, reason: %s", string(e.Body[:]))
	case *gc.ErrDefault500:
		return errors.Errorf("code: 500, reason: %s", string(e.Body[:]))
	case gc.ErrUnexpectedResponseCode:
		return errors.Errorf("code: %d, reason: %s", e.Actual, string(e.Body[:]))
	case *gc.ErrUnexpectedResponseCode:
		return errors.Errorf("code: %d, reason: %s", e.Actual, string(e.Body[:]))
	default:
		return e
	}
}

type BaseServices struct {
	client  *gc.ProviderClient
	Compute *gc.ServiceClient
	Network *gc.ServiceClient
	Volume  *gc.ServiceClient
}

type Configuration struct {
	ExternalNetworkName string
	ExternalNetworkID   string
}

//Provider Provider provider
type Provider struct {
	Config                   Configuration
	BaseServices             BaseServices
	KeyPairManager           KeyPairManager
	ImagesManager            ImageManager
	NetworkManager           NetworkManager
	NetworkInterfacesManager NetworkInterfacesManager
	TemplateManager          ServerTemplateManager
	ServerManager            ServerManager
	SecurityGroupManager     SecurityGroupManager
	VolumeManager            VolumeManager
	PublicIPAddressManager   PublicIPManager
}

//Init initialize Provider Provider
func (p *Provider) Init(config io.Reader, format string) error {
	v := viper.New()
	v.SetConfigType(format)
	err := v.ReadConfig(config)
	if err != nil {
		return errors.Wrap(err, "error reading provider configuration")
	}
	cfg := Config{}
	err = v.Unmarshal(&cfg)
	if err != nil {
		return errors.Wrap(err, "error reading provider configuration")
	}
	opts := gc.AuthOptions{
		IdentityEndpoint: cfg.IdentityEndpoint,
		Username:         cfg.Username,
		UserID:           cfg.UserID,
		Password:         cfg.Password,
		DomainID:         cfg.DomainID,
		DomainName:       cfg.DomainName,
		TenantID:         cfg.TenantID,
		TenantName:       cfg.TenantName,
		AllowReauth:      cfg.AllowReauth,
		TokenID:          cfg.TokenID,
	}

	// Openstack client
	p.BaseServices.client, err = openstack.AuthenticatedClient(opts)
	if err != nil {
		return errors.Wrap(UnwrapOpenStackError(err), "Error initializing openstack driver")
	}
	// Compute API
	p.BaseServices.Compute, err = openstack.NewComputeV2(p.BaseServices.client, gc.EndpointOpts{
		Region: cfg.Region,
	})
	if err != nil {
		return errors.Wrap(UnwrapOpenStackError(err), "Error initializing openstack driver")
	}
	//Network API
	p.BaseServices.Network, err = openstack.NewNetworkV2(p.BaseServices.client, gc.EndpointOpts{
		Name:   "neutron",
		Region: cfg.Region,
	})
	if err != nil {
		return errors.Wrap(UnwrapOpenStackError(err), "Error initializing openstack driver")
	}
	//Volume API
	p.BaseServices.Volume, err = openstack.NewBlockStorageV3(p.BaseServices.client, gc.EndpointOpts{
		Region: cfg.Region,
	})
	if err != nil {
		return errors.Wrap(UnwrapOpenStackError(err), "Error initializing openstack driver")
	}

	p.ImagesManager.Provider = p
	p.NetworkManager.Refactor = p
	p.NetworkInterfacesManager.Provider = p
	p.ServerManager.Provider = p
	p.TemplateManager.Provider = p
	p.VolumeManager.Provider = p
	p.SecurityGroupManager.Provider = p
	p.KeyPairManager.Provider = p
	p.PublicIPAddressManager.OpenStack = p

	p.Config.ExternalNetworkName = cfg.ExternalNetworkName
	extNetID, err := networks.IDFromName(p.BaseServices.Network, p.Config.ExternalNetworkName)
	p.Config.ExternalNetworkID = extNetID
	return errors.Wrap(UnwrapOpenStackError(err), "Error initializing openstack driver")
}

//GetNetworkManager returns an Provider NetworkManager
func (p *Provider) GetNetworkManager() api.NetworkManager {
	return &p.NetworkManager
}

func (p *Provider) GetNetworkInterfaceManager() api.NetworkInterfaceManager {
	return &p.NetworkInterfacesManager
}

//GetImageManager returns an Provider ImageManager
func (p *Provider) GetImageManager() api.ImageManager {
	return &p.ImagesManager
}

//GetTemplateManager returns an Provider ServerTemplateManager
func (p *Provider) GetTemplateManager() api.ServerTemplateManager {
	return &p.TemplateManager
}

//GetSecurityGroupManager returns an Provider SecurityGroupManager
func (p *Provider) GetSecurityGroupManager() api.SecurityGroupManager {
	return &p.SecurityGroupManager
}

//GetServerManager returns an Provider ServerManager
func (p *Provider) GetServerManager() api.ServerManager {
	return &p.ServerManager
}

//GetVolumeManager returns an Provider VolumeManager
func (p *Provider) GetVolumeManager() api.VolumeManager {
	return &p.VolumeManager
}
func (p *Provider) GetPublicIPAddressManager() api.PublicIPManager {
	return &p.PublicIPAddressManager
}
