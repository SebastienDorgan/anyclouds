package providers

//Volume defines volume properties
type Volume struct {
	ID   string
	Name string
	Size int
	IOPS int
}

//VolumeOptions defines options to use when creating a volume
type VolumeOptions struct {
	Name    string
	Size    int
	MinIOPS int
}

//VolumeAttachment attachment between an instace and a volume
type VolumeAttachment struct {
	ID         string
	VolumeID   string
	InstanceID string
	Name       string
}

//VolumeManager defines volume management functions an anyclouds provider must provide
type VolumeManager interface {
	Create(options *VolumeOptions) (*Volume, error)
	Delete(id string) error
	List(filter *ResourceFilter) ([]Volume, error)
	Get(id string) (*Volume, error)
	Attach(volumeID string, instanceID string) (*VolumeAttachment, error)
	Detach(attachmentID string) error
	Attachement(volumeID string, instanceID string) (*VolumeAttachment, error)
	Attachments(instanceID string, filter *ResourceFilter) ([]VolumeAttachment, error)
}
