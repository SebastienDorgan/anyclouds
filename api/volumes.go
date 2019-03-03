package api

//Volume defines volume properties
type Volume struct {
	ID   string
	Name string
	Size int
	Type string
}

//VolumeOptions defines options to use when creating a volume
type VolumeOptions struct {
	Name string
	Size int
	Type string
}

//VolumeAttachment attachment between an instace and a volume
type VolumeAttachment struct {
	ID       string
	VolumeID string
	ServerID string
	Device   string
}

//VolumeManager defines volume management functions an anyclouds provider must provide
type VolumeManager interface {
	Create(options *VolumeOptions) (*Volume, error)
	Delete(id string) error
	List() ([]Volume, error)
	Get(id string) (*Volume, error)
	Attach(volumeID string, serverID string) (*VolumeAttachment, error)
	Detach(volumeID string, serverID string) error
	Attachments(serverID string) ([]VolumeAttachment, error)
}
