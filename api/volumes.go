package api

//Volume defines volume properties
type Volume struct {
	ID       string
	Name     string
	Size     int64
	IOPS     int64
	DataRate int64
}

//VolumeOptions defines options to use when creating a volume
type VolumeOptions struct {
	Name        string
	Size        int64
	MinIOPS     int64
	MinDataRate int64
}

type ModifyVolumeOptions struct {
	ID          string
	Size        int64
	MinIOPS     int64
	MinDataRate int64
}

//VolumeAttachment attachment between an instance and a volume
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
	Modify(options *ModifyVolumeOptions) (*Volume, error)
	Attach(volumeID string, serverID string, device string) (*VolumeAttachment, error)
	Detach(volumeID string, serverID string, force bool) error
	Attachments(serverID string) ([]VolumeAttachment, error)
}
