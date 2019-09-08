package api

import "io"

//Provider define a cloud provider
type Provider interface {
	Init(config io.Reader, format string) error
	GetNetworkManager() NetworkManager
	GetImageManager() ImageManager
	GetTemplateManager() ServerTemplateManager
	GetSecurityGroupManager() SecurityGroupManager
	GetServerManager() ServerManager
	GetVolumeManager() VolumeManager
	GetPublicIpAddressManager() PublicIPAddressManager
}
