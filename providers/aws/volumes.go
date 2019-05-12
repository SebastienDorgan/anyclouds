package aws

import (
	"github.com/SebastienDorgan/anyclouds/api"
)

//VolumeManager defines volume management functions an anyclouds provider must provide
type VolumeManager struct {
	AWS *Provider
}

//Create creates a volume with options
func (mgr *VolumeManager) Create(options *api.VolumeOptions) (*api.Volume, error) {
	return nil, nil
}

//Delete deletes volume identified by id
func (mgr *VolumeManager) Delete(id string) error {
	return nil
}

//List lists volumes along filter
func (mgr *VolumeManager) List() ([]api.Volume, error) {
	return nil, nil
}

//Get returns volume details
func (mgr *VolumeManager) Get(id string) (*api.Volume, error) {
	return nil, nil
}

//Attach attaches a volume to an Server
func (mgr *VolumeManager) Attach(volumeID string, serverID string) (*api.VolumeAttachment, error) {
	return nil, nil
}

//Detach detach a volume from an Server
func (mgr *VolumeManager) Detach(volumeID string, serverID string) error {
	return nil
}

//Attachment returns the attachment between a volume and an Server
func (mgr *VolumeManager) Attachment(volumeID string, serverID string) (*api.VolumeAttachment, error) {
	return nil, nil
}

//Attachments returns all the attachments of an Server
func (mgr *VolumeManager) Attachments(serverID string) ([]api.VolumeAttachment, error) {
	return nil, nil
}
