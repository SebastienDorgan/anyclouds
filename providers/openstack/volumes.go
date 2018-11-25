package openstack

import (
	"github.com/SebastienDorgan/anyclouds/providers"
)

//VolumeManager defines volume management functions an anyclouds provider must provide
type VolumeManager struct {
	OpenStack *Provider
}

//Create creates a volume with options
func (mgr *VolumeManager) Create(options *providers.VolumeOptions) (*providers.Volume, error) {
	return nil, nil
}

//Delete deletes volume identified by id
func (mgr *VolumeManager) Delete(id string) error {
	return nil
}

//List lists volumes along filter
func (mgr *VolumeManager) List(filter *providers.ResourceFilter) ([]providers.Volume, error) {
	return nil, nil
}

//Get returns volume details
func (mgr *VolumeManager) Get(id string) (*providers.Volume, error) {
	return nil, nil
}

//Attach attaches a volume to an instance
func (mgr *VolumeManager) Attach(volumeID string, instanceID string) (*providers.VolumeAttachment, error) {
	return nil, nil
}

//Detach detach a volume from an instance
func (mgr *VolumeManager) Detach(attachmentID string) error {
	return nil
}

//Attachement returns the attachement between a volume and an instance
func (mgr *VolumeManager) Attachement(volumeID string, instanceID string) (*providers.VolumeAttachment, error) {
	return nil, nil
}

//Attachments returns all the attachements of an instance
func (mgr *VolumeManager) Attachments(instanceID string, filter *providers.ResourceFilter) ([]providers.VolumeAttachment, error) {
	return nil, nil
}
