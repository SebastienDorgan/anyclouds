package azure

import (
	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/Azure/azure-sdk-for-go/profiles/preview/preview/commerce/mgmt/commerce"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"io"
)

type Provider struct {
	Configuration              Config
	Authorizer                 autorest.Authorizer
	VirtualMachineImagesClient compute.VirtualMachineImagesClient
	VirtualMachineSizesClient  compute.VirtualMachineSizesClient
	VirtualNetworksClient      network.VirtualNetworksClient
	SubnetsClient              network.SubnetsClient
	SecurityGroupsClient       network.SecurityGroupsClient
	VirtualMachinesClient      compute.VirtualMachinesClient
	InterfacesClient           network.InterfacesClient
	RateCardClient             commerce.RateCardClient
	PublicIPAddressesClient    network.PublicIPAddressesClient

	ImageManager             *ImageManager
	ServerTemplateManager    *ServerTemplateManager
	NetworkManager           *NetworkManager
	SecurityGroupManager     *SecurityGroupManager
	ServerManager            *ServerManager
	NetworkInterfacesManager *NetworkInterfacesManager
	PublicIPAddressManager   *PublicIPManager
}

type Config struct {
	TenantID                      string
	ClientID                      string
	ClientSecret                  string
	ActiveDirectoryEndpoint       string
	ResourceManagerEndpoint       string
	UseDeviceFlow                 bool
	SubscriptionID                string
	UserAgent                     string
	Location                      string
	VirtualMachineImagePublishers []string
	DefaultVMUserName             string
	ResourceGroupName             string
	OfferNumber                   string
	Currency                      string
	RegionInfo                    string
	PublicAddressesURL            string
}

func (p *Provider) Init(config io.Reader, format string) error {

	v := viper.New()
	v.SetConfigType(format)
	err := v.ReadConfig(config)
	if err != nil {
		return errors.Wrap(err, "error initializing azure provider")
	}
	cfg := Config{}
	err = v.Unmarshal(&cfg)
	if err != nil {
		return errors.Wrap(err, "error initializing azure provider")
	}
	p.Authorizer, err = getAuthorizerForResource(&cfg)
	if err != nil {
		return errors.Wrap(err, "error initializing azure provider")
	}

	p.VirtualMachineImagesClient = compute.NewVirtualMachineImagesClient(cfg.SubscriptionID)
	p.VirtualMachineImagesClient.Authorizer = p.Authorizer
	err = p.VirtualMachineImagesClient.AddToUserAgent(cfg.UserAgent)
	if err != nil {
		return errors.Wrap(err, "error initializing azure provider")
	}

	p.VirtualMachineSizesClient = compute.NewVirtualMachineSizesClient(cfg.SubscriptionID)
	p.VirtualMachineSizesClient.Authorizer = p.Authorizer
	err = p.VirtualMachineSizesClient.AddToUserAgent(cfg.UserAgent)
	if err != nil {
		return errors.Wrap(err, "error initializing azure provider")
	}

	p.VirtualNetworksClient = network.NewVirtualNetworksClient(cfg.SubscriptionID)
	p.VirtualNetworksClient.Authorizer = p.Authorizer
	err = p.VirtualNetworksClient.AddToUserAgent(cfg.UserAgent)
	if err != nil {
		return errors.Wrap(err, "error initializing azure provider")
	}
	p.SubnetsClient = network.NewSubnetsClient(cfg.SubscriptionID)
	p.SubnetsClient.Authorizer = p.Authorizer
	err = p.SubnetsClient.AddToUserAgent(cfg.UserAgent)

	p.SecurityGroupsClient = network.NewSecurityGroupsClient(cfg.SubscriptionID)
	p.SecurityGroupsClient.Authorizer = p.Authorizer
	err = p.SecurityGroupsClient.AddToUserAgent(cfg.UserAgent)

	p.VirtualMachinesClient = compute.NewVirtualMachinesClient(cfg.SubscriptionID)
	p.VirtualMachinesClient.Authorizer = p.Authorizer
	err = p.VirtualMachinesClient.AddToUserAgent(cfg.UserAgent)
	if err != nil {
		return errors.Wrap(err, "error initializing azure provider")
	}

	p.RateCardClient = commerce.NewRateCardClient(cfg.SubscriptionID)
	p.RateCardClient.Authorizer = p.Authorizer
	err = p.RateCardClient.AddToUserAgent(cfg.UserAgent)
	if err != nil {
		return errors.Wrap(err, "error initializing azure provider")
	}

	p.PublicIPAddressesClient = network.NewPublicIPAddressesClient(cfg.SubscriptionID)
	p.PublicIPAddressesClient.Authorizer = p.Authorizer
	err = p.PublicIPAddressesClient.AddToUserAgent(cfg.UserAgent)
	if err != nil {
		return errors.Wrap(err, "error initializing azure provider")
	}

	p.ImageManager = &ImageManager{Provider: p}
	p.ServerTemplateManager = &ServerTemplateManager{Provider: p}
	p.NetworkManager = &NetworkManager{Provider: p}
	p.SecurityGroupManager = &SecurityGroupManager{Provider: p}
	p.ServerManager = &ServerManager{Provider: p}
	p.NetworkInterfacesManager = &NetworkInterfacesManager{Provider: p}
	p.PublicIPAddressManager = &PublicIPManager{Provider: p}

	return nil
}

func getAuthorizerForResource(config *Config) (autorest.Authorizer, error) {
	if config.UseDeviceFlow {
		deviceFlowConfig := auth.NewDeviceFlowConfig(config.ClientID, config.TenantID)
		deviceFlowConfig.Resource = config.ResourceManagerEndpoint
		return deviceFlowConfig.Authorizer()

	}
	oauthConfig, err := adal.NewOAuthConfig(
		config.ActiveDirectoryEndpoint, config.TenantID)
	if err != nil {
		return nil, err
	}

	token, err := adal.NewServicePrincipalToken(
		*oauthConfig, config.ClientID, config.ClientSecret, config.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}
	return autorest.NewBearerAuthorizer(token), nil

}

func (p *Provider) GetNetworkManager() api.NetworkManager {
	return p.NetworkManager
}

func (p *Provider) GetImageManager() api.ImageManager {
	return p.ImageManager
}

func (p *Provider) GetTemplateManager() api.ServerTemplateManager {
	return p.ServerTemplateManager
}

func (p *Provider) GetSecurityGroupManager() api.SecurityGroupManager {
	return p.SecurityGroupManager
}

func (p *Provider) GetServerManager() api.ServerManager {
	return p.ServerManager
}

func (p *Provider) GetNetworkInterfaceManager() api.NetworkInterfaceManager {
	return p.NetworkInterfacesManager
}

func (p *Provider) GetVolumeManager() api.VolumeManager {
	panic("implement me")
}

func (p *Provider) GetPublicIPAddressManager() api.PublicIPManager {
	return p.PublicIPAddressManager
}
