package azure

import (
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
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

	ImageManager api.ImageManager
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
}

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
	p.Authorizer, err = getAuthorizerForResource(&cfg)
	if err != nil {
		return errors.Wrap(err, "error reading provider configuration")
	}

	p.VirtualMachineImagesClient = compute.NewVirtualMachineImagesClient(cfg.SubscriptionID)
	p.VirtualMachineImagesClient.Authorizer = p.Authorizer
	err = p.VirtualMachineImagesClient.AddToUserAgent(cfg.UserAgent)
	if err != nil {
		return errors.Wrap(err, "error reading provider configuration")
	}

	p.VirtualMachineSizesClient = compute.NewVirtualMachineSizesClient(cfg.SubscriptionID)
	p.VirtualMachineSizesClient.Authorizer = p.Authorizer
	err = p.VirtualMachineSizesClient.AddToUserAgent(cfg.UserAgent)
	if err != nil {
		return errors.Wrap(err, "error reading provider configuration")
	}
	p.ImageManager = &ImageManager{Provider: p}
	return nil
}

func getAuthorizerForResource(config *Config) (autorest.Authorizer, error) {
	if config.UseDeviceFlow {
		deviceconfig := auth.NewDeviceFlowConfig(config.ClientID, config.TenantID)
		deviceconfig.Resource = config.ResourceManagerEndpoint
		return deviceconfig.Authorizer()

	} else {
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
}

func (p *Provider) GetNetworkManager() api.NetworkManager {
	panic("implement me")
}

func (p *Provider) GetImageManager() api.ImageManager {
	return p.ImageManager
}

func (p *Provider) GetTemplateManager() api.ServerTemplateManager {
	panic("implement me")
}

func (p *Provider) GetSecurityGroupManager() api.SecurityGroupManager {
	panic("implement me")
}

func (p *Provider) GetServerManager() api.ServerManager {
	panic("implement me")
}

func (p *Provider) GetVolumeManager() api.VolumeManager {
	panic("implement me")
}

func (p *Provider) GetPublicIpAddressManager() api.PublicIPAddressManager {
	panic("implement me")
}
