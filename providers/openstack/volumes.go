package openstack

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v1/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/pkg/errors"
)

//VolumeManager defines volume management functions an anyclouds provider must provide
type VolumeManager struct {
	OpenStack *Provider
}

//Create creates a volume with options
func (mgr *VolumeManager) Create(options *api.VolumeOptions) (*api.Volume, error) {
	v, err := volumes.Create(mgr.OpenStack.Volume, volumes.CreateOpts{
		Name:       options.Name,
		Size:       options.Size,
		VolumeType: options.Type,
	}).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error creating volume")
	}
	return &api.Volume{
		Name: v.Name,
		ID:   v.ID,
		Size: v.Size,
		Type: v.VolumeType,
	}, nil
}

//Delete deletes volume identified by id
func (mgr *VolumeManager) Delete(id string) error {
	err := volumes.Delete(mgr.OpenStack.Volume, id).ExtractErr()
	return errors.Wrap(ProviderError(err), "Error deleting volume")
}

//List lists volumes along filter
func (mgr *VolumeManager) List() ([]api.Volume, error) {
	page, err := volumes.List(mgr.OpenStack.Volume, volumes.ListOpts{}).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing volume")
	}
	l, err := volumes.ExtractVolumes(page)
	var res []api.Volume
	for _, v := range l {
		res = append(res, api.Volume{
			Name: v.Name,
			ID:   v.ID,
			Size: v.Size,
			Type: v.VolumeType,
		})
	}
	return res, nil
}

//Get returns volume details
func (mgr *VolumeManager) Get(id string) (*api.Volume, error) {
	v, err := volumes.Get(mgr.OpenStack.Volume, id).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting volume")
	}
	return &api.Volume{
		Name: v.Name,
		ID:   v.ID,
		Size: v.Size,
		Type: v.VolumeType,
	}, nil
}

//Attach attaches a volume to an Server
func (mgr *VolumeManager) Attach(volumeID string, serverID string) (*api.VolumeAttachment, error) {
	va, err := volumeattach.Create(mgr.OpenStack.Compute, serverID, volumeattach.CreateOpts{
		VolumeID: volumeID,
	}).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error attaching volume to server")
	}
	return &api.VolumeAttachment{
		ID:       va.ID,
		VolumeID: va.VolumeID,
		ServerID: va.ServerID,
		Device:   va.Device,
	}, nil
}

//Detach detach a volume from an Server
func (mgr *VolumeManager) Detach(volumeID string, serverID string) error {
	err := volumeattach.List(mgr.OpenStack.Compute, "").Err
	return errors.Wrap(ProviderError(err), "Error detaching volume from server")
}

//Attachment returns the attachment between a volume and an Server
func (mgr *VolumeManager) Attachment(volumeID string, serverID string) (*api.VolumeAttachment, error) {
	attachments, err := mgr.Attachments(serverID)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Retrieving  attachment")
	}
	for _, va := range attachments {
		if va.VolumeID == volumeID && va.ServerID == serverID {
			return &va, nil
		}
	}
	return nil, nil
}

//Attachments returns all the attachments of an Server
func (mgr *VolumeManager) Attachments(serverID string) ([]api.VolumeAttachment, error) {
	page, err := volumeattach.List(mgr.OpenStack.Compute, serverID).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Retrieving listing attachments")
	}
	var res []api.VolumeAttachment
	l, err := volumeattach.ExtractVolumeAttachments(page)
	for _, va := range l {
		res = append(res, api.VolumeAttachment{
			ID:       va.ID,
			VolumeID: va.VolumeID,
			ServerID: va.ServerID,
			Device:   va.Device,
		})
	}
	return res, nil
}
