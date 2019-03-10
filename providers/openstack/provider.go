package openstack

import (
	"io"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/spf13/viper"

	gc "github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/pkg/errors"
)

/*Config fields are the union of those recognized by each OpenStack identity implementation and
provider.
*/
type Config struct {
	// IdentityEndpoint specifies the HTTP endpoint that is required to work with
	// the Identity API of the appropriate version. While it's ultimately needed by
	// all of the identity services, it will often be populated by a provider-level
	// function.
	IdentityEndpoint string `json:"identity_endpoint,omitempty"`

	// Username is required if using Identity V2 API. Consult with your provider's
	// control panel to discover your account's username. In Identity V3, either
	// UserID or a combination of Username and DomainID or DomainName are needed.
	Username string `json:"username,omitempty"`
	UserID   string `json:"user_id,omitempty"`

	// Exactly one of Password or APIKey is required for the Identity V2 and V3
	// APIs. Consult with your provider's control panel to discover your account's
	// preferred method of authentication.
	Password string `json:"password,omitempty"`
	APIKey   string `json:"api_key,omitempty"`

	// At most one of DomainID and DomainName must be provided if using Username
	// with Identity V3. Otherwise, either are optional.
	DomainID   string `json:"domain_id,omitempty"`
	DomainName string `json:"domain_name,omitempty"`

	// The TenantID and TenantName fields are optional for the Identity V2 API.
	// Some api allow you to specify a TenantName instead of the TenantId.
	// Some require both. Your provider's authentication policies will determine
	// how these fields influence authentication.
	TenantID   string `json:"tenant_id,omitempty"`
	TenantName string `json:"tenant_name,omitempty"`

	// AllowReauth should be set to true if you grant permission for Gophercloud to
	// cache your credentials in memory, and to allow Gophercloud to attempt to
	// re-authenticate automatically if/when your token expires.  If you set it to
	// false, it will not cache these settings, but re-authentication will not be
	// possible.  This setting defaults to false.
	//
	// NOTE: The reauth function will try to re-authenticate endlessly if left unchecked.
	// The way to limit the number of attempts is to provide a custom HTTP client to the provider client
	// and provide a transport that implements the RoundTripper interface and stores the number of failed retries.
	// For an example of this, see here: https://github.com/gophercloud/rack/blob/1.0.0/auth/clients.go#L311
	AllowReauth bool `json:"allow_reauth,omitempty"`

	// TokenID allows users to authenticate (possibly as another user) with an
	// authentication token ID.
	TokenID string `json:"token_id,omitempty"`

	//Openstack region (data center) where the infrstructure will be created
	Region string `json:"region,omitempty"`

	//FloatingIPPool name of the floating IP pool
	//Necessary only if UseFloatingIP is true
	FloatingIPPool string `json:"floating_ip_pool,omitempty"`
}

// ProviderError creates an error string from openstack api error
func ProviderError(err error) error {
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
		// logrus.Debugf("Error code not yet handled specifically: ProviderErrorToString(%+v)\n", err)
		return e
	}
}

//Provider OpenStack provider
type Provider struct {
	client               *gc.ProviderClient
	Compute              *gc.ServiceClient
	Network              *gc.ServiceClient
	Volume               *gc.ServiceClient
	Name                 string
	KeyPairManager       KeyPairManager
	ImagesManager        ImageManager
	NetworkManager       NetworkManager
	TemplateManager      ServerTemplateManager
	ServerManager        ServerManager
	SecurityGroupManager SecurityGroupManager
	VolumeManager        VolumeManager
}

//Init initialize OpenStack Provider
func (p *Provider) Init(config io.Reader) error {
	v := viper.New()
	v.ReadConfig(config)
	cfg := Config{}
	err := v.Unmarshal(&cfg)
	if err != nil {
		return errors.Wrap(err, "Error reading provider configuration")
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
	p.client, err = openstack.AuthenticatedClient(opts)
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error initiliazing openstack driver")
	}
	// Compute API
	p.Compute, err = openstack.NewComputeV2(p.client, gc.EndpointOpts{
		Region: cfg.Region,
	})
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error initiliazing openstack driver")
	}
	//Network API
	p.Network, err = openstack.NewNetworkV2(p.client, gc.EndpointOpts{
		Name:   "neutron",
		Region: cfg.Region,
	})
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error initiliazing openstack driver")
	}
	//Volume API
	p.Volume, err = openstack.NewBlockStorageV1(p.client, gc.EndpointOpts{
		Region: cfg.Region,
	})
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error initiliazing openstack driver")
	}
	p.ImagesManager.OpenStack = p
	p.NetworkManager.OpenStack = p
	p.ServerManager.OpenStack = p
	p.TemplateManager.OpenStack = p
	p.VolumeManager.OpenStack = p
	p.SecurityGroupManager.OpenStack = p
	p.KeyPairManager.OpenStack = p
	return nil
}

//GetKeyPairManager returns aws KeyPairManager
func (p *Provider) GetKeyPairManager() api.KeyPairManager {
	return &p.KeyPairManager
}

//GetNetworkManager returns an OpenStack NetworkManager
func (p *Provider) GetNetworkManager() api.NetworkManager {
	return &p.NetworkManager
}

//GetImageManager returns an OpenStack ImageManager
func (p *Provider) GetImageManager() api.ImageManager {
	return &p.ImagesManager
}

//GetTemplateManager returns an OpenStack IntanceTemplateManager
func (p *Provider) GetTemplateManager() api.ServerTemplateManager {
	return &p.TemplateManager
}

//GetSecurityGroupManager returns an OpenStack SecurityGroupManager
func (p *Provider) GetSecurityGroupManager() api.SecurityGroupManager {
	return &p.SecurityGroupManager
}

//GetServerManager returns an OpenStack ServerManager
func (p *Provider) GetServerManager() api.ServerManager {
	return &p.ServerManager
}

//GetVolumeManager returns an OpenStack VolumeManager
func (p *Provider) GetVolumeManager() api.VolumeManager {
	return &p.VolumeManager
}
