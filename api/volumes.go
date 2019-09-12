package api

//Volume defines volume properties
type Volume struct {
	ID       string
	Name     string
	Size     int64
	IOPS     int64
	DataRate int64
}

//CreateVolumeOptions defines options to use when creating a volume
type CreateVolumeOptions struct {
	Name        string
	Size        int64
	MinIOPS     int64
	MinDataRate int64
}

//ModifyVolumeOptions options that can be used to modify a volume
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

//AttachVolumeOptions options used to attach a volume
type AttachVolumeOptions struct {
	VolumeID   string
	ServerID   string
	DevicePath string
}

//DetachVolumeOptions options used to detach a volume
type DetachVolumeOptions struct {
	VolumeID string
	ServerID string
	Force    bool
}

//VolumeManager defines volume management functions an anyclouds provider must provide
type VolumeManager interface {
	Create(options CreateVolumeOptions) (*Volume, error)
	Delete(id string) error
	List() ([]Volume, error)
	Get(id string) (*Volume, error)
	Modify(options *ModifyVolumeOptions) (*Volume, error)
	Attach(options AttachVolumeOptions) (*VolumeAttachment, error)
	Detach(options DetachVolumeOptions) error
	Attachments(serverID string) ([]VolumeAttachment, error)
}

//CreateVolumeError create volume error type
type CreateVolumeError struct {
	ErrorStack
}

//NewCreateVolumeError creates a new CreateVolumeError
func NewCreateVolumeError(cause error, options CreateVolumeOptions) *CreateVolumeError {
	return &CreateVolumeError{ErrorStack: *NewErrorStack(cause, "error creating volume", options)}
}
