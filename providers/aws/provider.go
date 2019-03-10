package aws

import (
	"io"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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

//Retrieve adapts Config to AWS Providder interface
func (cfg *Config) Retrieve() (credentials.Value, error) {
	return credentials.Value{
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		ProviderName:    "anycloud",
	}, nil
}

//IsExpired adapts Config to AWS Providder interface
func (cfg *Config) IsExpired() bool {
	return false
}

//Provider AWS provider
type Provider struct {
	EC2Client            *ec2.EC2
	PricingClient        *pricing.Pricing
	Name                 string
	KeyPairManager       KeyPairManager
	ImagesManager        ImageManager
	NetworkManager       NetworkManager
	TemplateManager      ServerTemplateManager
	ServerManager        ServerManager
	SecurityGroupManager SecurityGroupManager
	VolumeManager        VolumeManager
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
func (p *Provider) Init(config io.Reader) error {
	v := viper.New()
	v.ReadConfig(config)
	cfg := Config{}
	err := v.Unmarshal(&cfg)
	if err != nil {
		return errors.Wrap(err, "Error reading provider configuration")
	}
	ec2session, err := session.NewSession(getEC2Config(&cfg))
	if err != nil {
		return errors.Wrap(err, "Error creation provider session")
	}
	p.EC2Client = ec2.New(ec2session)

	pricingSession, err := session.NewSession(getPricingConfig(&cfg))
	if err != nil {
		return errors.Wrap(err, "Error creation provider session")
	}
	p.PricingClient = pricing.New(pricingSession)
	p.ImagesManager.AWS = p
	p.NetworkManager.AWS = p
	p.ServerManager.AWS = p
	p.TemplateManager.AWS = p
	p.VolumeManager.AWS = p
	p.SecurityGroupManager.AWS = p
	p.KeyPairManager.AWS = p
	return nil
}

//GetKeyPairManager returns aws KeyPairManager
func (p *Provider) GetKeyPairManager() api.KeyPairManager {
	return &p.KeyPairManager
}

//GetNetworkManager returns aws NetworkManager
func (p *Provider) GetNetworkManager() api.NetworkManager {
	return &p.NetworkManager
}

//GetImageManager returns aws ImageManager
func (p *Provider) GetImageManager() api.ImageManager {
	return &p.ImagesManager
}

//GetTemplateManager returns aws IntanceTemplateManager
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
