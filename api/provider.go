package api

import "io"

//Provider implement api.Provider for Provider
type Provider interface {
	Init(config io.Reader, format string) error
	GetNetworkManager() NetworkManager
	GetImageManager() ImageManager
	GetTemplateManager() ServerTemplateManager
	GetSecurityGroupManager() SecurityGroupManager
	GetServerManager() ServerManager
	GetVolumeManager() VolumeManager
	GetPublicIPAddressManager() PublicIPAddressManager
	GetNetworkInterfaceManager() NetworkInterfaceManager
}
