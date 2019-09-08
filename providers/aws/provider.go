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

//Config AWS session configuration
type Config struct {
	// AWS Region
	Region string
	// AWS Access key ID
	AccessKeyID string

	// AWS Secret Access Key
	SecretAccessKey string

	// AWS Session Token
	SessionToken string

	// Provider used to get credentials
	ProviderName string
}

//Retrieve adapts Config to AWS Provider interface
func (cfg *Config) Retrieve() (credentials.Value, error) {
	return credentials.Value{
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		ProviderName:    "anyclouds",
	}, nil
}

//IsExpired adapts Config to AWS Provider interface
func (cfg *Config) IsExpired() bool {
	return false
}

//Provider AWS provider
type Provider struct {
	EC2Client              *ec2.EC2
	OpsWorksClient         *opsworks.OpsWorks
	PricingClient          *pricing.Pricing
	KeyPairManager         *KeyPairManager
	ImagesManager          *ImageManager
	NetworkManager         *NetworkManager
	TemplateManager        *ServerTemplateManager
	ServerManager          *ServerManager
	SecurityGroupManager   *SecurityGroupManager
	VolumeManager          *VolumeManager
	PublicIPAddressManager *PublicIPAddressManager
	Region                 string
	RegionName             string
	AvailabilityZone       string
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

//Init initialize AWS Provider
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
	p.EC2Client = ec2.New(ec2session)
	p.OpsWorksClient = opsworks.New(ec2session)

	pricingSession, err := session.NewSession(getPricingConfig(&cfg))
	if err != nil {
		return errors.Wrap(err, "Error creation provider session")
	}
	p.PricingClient = pricing.New(pricingSession)
	p.ImagesManager = &ImageManager{AWS: p}
	p.NetworkManager = &NetworkManager{AWS: p}
	p.ServerManager = &ServerManager{AWS: p}
	p.TemplateManager = &ServerTemplateManager{AWS: p}
	p.VolumeManager = &VolumeManager{AWS: p}
	p.SecurityGroupManager = &SecurityGroupManager{AWS: p}
	p.KeyPairManager = &KeyPairManager{AWS: p}
	p.PublicIPAddressManager = &PublicIPAddressManager{AWS: p}
	p.Region = cfg.Region
	p.RegionName = v.GetString("RegionName")
	p.AvailabilityZone = v.GetString("AvailabilityZone")
	return nil

}

//Name name of the provider
func (p *Provider) Name() string {
	return "AWS"
}

//GetNetworkManager returns aws NetworkManager
func (p *Provider) GetNetworkManager() api.NetworkManager {
	return p.NetworkManager
}

//GetImageManager returns aws ImageManager
func (p *Provider) GetImageManager() api.ImageManager {
	return p.ImagesManager
}

//GetTemplateManager returns aws ServerTemplateManager
func (p *Provider) GetTemplateManager() api.ServerTemplateManager {
	return p.TemplateManager
}

//GetSecurityGroupManager returns aws SecurityGroupManager
func (p *Provider) GetSecurityGroupManager() api.SecurityGroupManager {
	return p.SecurityGroupManager
}

//GetServerManager returns aws ServerManager
func (p *Provider) GetServerManager() api.ServerManager {
	return p.ServerManager
}

//GetVolumeManager returns aws OpenStack VolumeManager
func (p *Provider) GetVolumeManager() api.VolumeManager {
	return p.VolumeManager
}

func (p *Provider) GetPublicIpAddressManager() api.PublicIPAddressManager {
	return p.PublicIPAddressManager
}
