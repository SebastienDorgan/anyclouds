package aws

import (
	"io"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/aws/aws-sdk-go/service/pricing"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

//Config Provider session configuration
type Config struct {
	// Provider Region
	Region string
	// Provider Access key ID
	AccessKeyID string

	// Provider Secret Access Key
	SecretAccessKey string

	// Provider Session Token
	SessionToken string

	// Provider used to get credentials
	ProviderName string
}

//Retrieve adapts Config to Provider Provider interface
func (cfg *Config) Retrieve() (credentials.Value, error) {
	return credentials.Value{
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		ProviderName:    "anyclouds",
	}, nil
}

//IsExpired adapts Config to Provider Provider interface
func (cfg *Config) IsExpired() bool {
	return false
}

//Configuration configuration of the Provider Provider
type Configuration struct {
	Region           string
	RegionName       string
	AvailabilityZone string
}

//RawServices aws raw services
type RawServices struct {
	EC2Client      *ec2.EC2
	OpsWorksClient *opsworks.OpsWorks
	PricingClient  *pricing.Pricing
}

//Provider Provider provider
type Provider struct {
	Configuration           Configuration
	AWSServices             RawServices
	KeyPairManager          KeyPairManager
	ImagesManager           ImageManager
	NetworkManager          NetworkManager
	NetworkInterfaceManager NetworkInterfaceManager
	TemplateManager         ServerTemplateManager
	ServerManager           ServerManager
	SecurityGroupManager    SecurityGroupManager
	VolumeManager           VolumeManager
	PublicIPAddressManager  PublicIPManager
}

func getEC2Config(cfg *Config) *aws.Config {
	return &aws.Config{
		Region:      aws.String(cfg.Region),
		Credentials: credentials.NewCredentials(cfg),
	}
}

func getPricingConfig(cfg *Config) *aws.Config {
	return &aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewCredentials(cfg),
	}
}

//Init initialize Provider Provider
func (p *Provider) Init(config io.Reader, format string) error {
	v := viper.New()
	v.SetConfigType(format)
	err := v.ReadConfig(config)
	if err != nil {
		return errors.Wrap(err, "Error creation provider session")
	}
	cfg := Config{
		AccessKeyID:     v.GetString("AccessKeyID"),
		Region:          v.GetString("Region"),
		SecretAccessKey: v.GetString("SecretAccessKey"),
	}
	ec2session, err := session.NewSession(getEC2Config(&cfg))
	if err != nil {
		return errors.Wrap(err, "Error creation provider session")
	}
	p.AWSServices.EC2Client = ec2.New(ec2session)
	p.AWSServices.OpsWorksClient = opsworks.New(ec2session)

	pricingSession, err := session.NewSession(getPricingConfig(&cfg))
	if err != nil {
		return errors.Wrap(err, "Error creation provider session")
	}
	p.AWSServices.PricingClient = pricing.New(pricingSession)
	p.ImagesManager.Provider = p
	p.NetworkManager.Provider = p
	p.NetworkInterfaceManager.Provider = p
	p.ServerManager.Provider = p
	p.TemplateManager.Provider = p
	p.VolumeManager.Provider = p
	p.SecurityGroupManager.Provider = p
	p.KeyPairManager.Provider = p
	p.PublicIPAddressManager.Provider = p
	p.Configuration.Region = cfg.Region
	p.Configuration.RegionName = v.GetString("RegionName")
	p.Configuration.AvailabilityZone = v.GetString("AvailabilityZone")
	return nil

}

//Name name of the provider
func (p *Provider) Name() string {
	return "Provider"
}

//GetNetworkManager returns aws NetworkManager
func (p *Provider) GetNetworkManager() api.NetworkManager {
	return &p.NetworkManager
}

//GetNetworkInterfaceManager returns aws NetworkInterfaceManager
func (p *Provider) GetNetworkInterfaceManager() api.NetworkInterfaceManager {
	return &p.NetworkInterfaceManager
}

//GetImageManager returns aws ImageManager
func (p *Provider) GetImageManager() api.ImageManager {
	return &p.ImagesManager
}

//GetTemplateManager returns aws ServerTemplateManager
func (p *Provider) GetTemplateManager() api.ServerTemplateManager {
	return &p.TemplateManager
}

//GetSecurityGroupManager returns aws SecurityGroupManager
func (p *Provider) GetSecurityGroupManager() api.SecurityGroupManager {
	return &p.SecurityGroupManager
}

//GetServerManager returns aws ServerManager
func (p *Provider) GetServerManager() api.ServerManager {
	return &p.ServerManager
}

//GetVolumeManager returns aws OpenStack VolumeManager
func (p *Provider) GetVolumeManager() api.VolumeManager {
	return &p.VolumeManager
}

func (p *Provider) GetPublicIPAddressManager() api.PublicIPManager {
	return &p.PublicIPAddressManager
}
