package api

import "io"

//Provider define a cloud provider
type Provider interface {
	Init(config io.Reader) error
	Name() string
	GetNetworkManager() NetworkManager
	GetImageManager() ImageManager
	GetTemplateManager() ServerTemplateManager
	GetSecurityGroupManager() SecurityGroupManager
	GetServerManager() ServerManager
	GetVolumeManager() VolumeManager
}
