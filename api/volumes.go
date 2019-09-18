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

//ResizeVolumeOptions options that can be used to modify a volume
type ResizeVolumeOptions struct {
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

//ListAttachmentsOptions options used list volume attachments
type ListAttachmentsOptions struct {
	VolumeID *string
	ServerID *string
}

//VolumeManager defines volume management functions an anyclouds provider must provide
type VolumeManager interface {
	Create(options CreateVolumeOptions) (*Volume, CreateVolumeError)
	Delete(id string) DeleteVolumeError
	List() ([]Volume, ListVolumesError)
	Get(id string) (*Volume, GetVolumeError)
	Resize(options ResizeVolumeOptions) (*Volume, ResizeVolumeError)
	Attach(options AttachVolumeOptions) (*VolumeAttachment, AttachVolumeError)
	Detach(options DetachVolumeOptions) DetachVolumeError
	ListAttachments(options *ListAttachmentsOptions) ([]VolumeAttachment, ListVolumeAttachmentsError)
}

//CreateVolumeError create volume error type
type CreateVolumeError interface {
	Error() string
}

//NewCreateVolumeError creates a new CreateVolumeError
func NewCreateVolumeError(cause error, options CreateVolumeOptions) CreateVolumeError {
	return NewErrorStack(cause, "error creating volume", options)
}

//DeleteVolumeError delete volume error type
type DeleteVolumeError interface {
	Error() string
}

//NewDeleteVolumeError creates a new DeleteVolumeError
func NewDeleteVolumeError(cause error, id string) DeleteVolumeError {
	return NewErrorStack(cause, "error deleting volume", id)
}

//ListVolumesError list volume error type
type ListVolumesError interface {
	Error() string
}

//NewListVolumesError creates a new ListVolumesError
func NewListVolumesError(cause error) ListVolumesError {
	return NewErrorStack(cause, "error listing volume")
}

//GetVolumeError get volume error type
type GetVolumeError interface {
	Error() string
}

//NewGetVolumeError creates a new GetVolumeError
func NewGetVolumeError(cause error, id string) GetVolumeError {
	return NewErrorStack(cause, "error getting volume", id)
}

//ResizeVolumeError resize volume error type
type ResizeVolumeError interface {
	Error() string
}

//NewResizeVolumeError creates a new ResizeVolumeError
func NewResizeVolumeError(cause error, options ResizeVolumeOptions) ResizeVolumeError {
	return NewErrorStack(cause, "error resizing volume", options)
}

//AttachVolumeError resize volume error type
type AttachVolumeError interface {
	Error() string
}

//NewAttachVolumeError creates a new AttachVolumeError
func NewAttachVolumeError(cause error, options AttachVolumeOptions) AttachVolumeError {
	return NewErrorStack(cause, "error attaching volume", options)
}

//DetachVolumeError resize volume error type
type DetachVolumeError interface {
	Error() string
}

//NewDetachVolumeError creates a new DetachVolumeError
func NewDetachVolumeError(cause error, options DetachVolumeOptions) DetachVolumeError {
	return NewErrorStack(cause, "error detaching volume", options)
}

//ListVolumeAttachmentsError list volume attachments error type
type ListVolumeAttachmentsError interface {
	Error() string
}

//NewListVolumeAttachmentsError creates a new ListVolumeAttachmentsError
func NewListVolumeAttachmentsError(cause error, options *ListAttachmentsOptions) ListVolumeAttachmentsError {
	return NewErrorStack(cause, "error listing attachments volume", options)
}
