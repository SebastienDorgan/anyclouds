package providers

//ResourceFilter resource filter
type ResourceFilter struct {
	ID   string
	Name string
}

//Provider define a cloud provider
type Provider interface {
	Init(config interface{}) error
	GetNetworkManager() NetworkManager
	GetImageManager() ImageManager
	GetTemplateManager() InstanceTemplateManager
	GetSecurityGroupManager() SecurityGroupManager
	GetInstanceManager() InstanceManager
	GetVolumeManager() VolumeManager
}
