package openstack

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/volumeactions"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/volumeattach"
)

//VolumeManager defines volume management functions an anyclouds provider must provide
type VolumeManager struct {
	Provider *Provider
}

//Create creates a volume with options
func (mgr *VolumeManager) Create(options api.CreateVolumeOptions) (*api.Volume, *api.CreateVolumeError) {
	v, err := volumes.Create(mgr.Provider.BaseServices.Volume, volumes.CreateOpts{
		Size:        int(options.Size),
		Metadata:    nil,
		Name:        options.Name,
		Multiattach: false,
	}).Extract()
	if err != nil {
		return nil, api.NewCreateVolumeError(err, options)
	}
	return &api.Volume{
		Name: v.Name,
		ID:   v.ID,
		Size: int64(v.Size),
	}, nil
}

//Delete deletes volume identified by id
func (mgr *VolumeManager) Delete(id string) *api.DeleteVolumeError {
	err := volumes.Delete(mgr.Provider.BaseServices.Volume, id, volumes.DeleteOpts{Cascade: true}).ExtractErr()
	return api.NewDeleteVolumeError(UnwrapOpenStackError(err), id)
}

//List lists volumes along filter
func (mgr *VolumeManager) List() ([]api.Volume, *api.ListVolumesError) {
	page, err := volumes.List(mgr.Provider.BaseServices.Volume, volumes.ListOpts{}).AllPages()
	if err != nil {
		return nil, api.NewListVolumesError(UnwrapOpenStackError(err))
	}
	l, err := volumes.ExtractVolumes(page)
	var res []api.Volume
	for _, v := range l {
		res = append(res, api.Volume{
			Name: v.Name,
			ID:   v.ID,
			Size: int64(v.Size),
		})
	}
	return res, nil
}

//Get returns volume details
func (mgr *VolumeManager) Get(id string) (*api.Volume, *api.GetVolumeError) {
	v, err := volumes.Get(mgr.Provider.BaseServices.Volume, id).Extract()
	if err != nil {
		return nil, api.NewGetVolumeError(UnwrapOpenStackError(err), id)
	}
	return &api.Volume{
		Name: v.Name,
		ID:   v.ID,
		Size: int64(v.Size),
	}, nil
}

//Attach attaches a volume to an Server
func (mgr *VolumeManager) Attach(options api.AttachVolumeOptions) (*api.VolumeAttachment, *api.AttachVolumeError) {
	va, err := volumeattach.Create(mgr.Provider.BaseServices.Compute, options.ServerID, volumeattach.CreateOpts{
		Device:   options.DevicePath,
		VolumeID: options.VolumeID,
	}).Extract()
	if err != nil {
		return nil, api.NewAttachVolumeError(UnwrapOpenStackError(err), options)
	}
	return &api.VolumeAttachment{
		ID:       va.ID,
		VolumeID: va.VolumeID,
		ServerID: va.ServerID,
		Device:   va.Device,
	}, nil
}

//Detach detach a volume from an Server
func (mgr *VolumeManager) Detach(options api.DetachVolumeOptions) *api.DetachVolumeError {
	att, err := mgr.attachment(options.VolumeID, options.ServerID)
	if err != nil {
		return api.NewDetachVolumeError(UnwrapOpenStackError(err), options)
	}
	err = volumeattach.Delete(mgr.Provider.BaseServices.Compute, options.ServerID, att.ID).Err
	return api.NewDetachVolumeError(UnwrapOpenStackError(err), options)
}

//attachment returns the attachment between a volume and an Server
func (mgr *VolumeManager) attachment(volumeID string, serverID string) (*api.VolumeAttachment, error) {
	attachments, err := mgr.ListAttachments(serverID)
	if err != nil {
		return nil, err
	}
	for _, va := range attachments {
		if va.VolumeID == volumeID && va.ServerID == serverID {
			return &va, nil
		}
	}
	return nil, nil
}

//ListAttachments returns all the attachments of an Server
func (mgr *VolumeManager) ListAttachments(serverID string) ([]api.VolumeAttachment, *api.ListVolumeAttachmentsError) {
	page, err := volumeattach.List(mgr.Provider.BaseServices.Compute, serverID).AllPages()
	if err != nil {
		return nil, api.NewListVolumeAttachmentsError(UnwrapOpenStackError(err), serverID)
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

func (mgr *VolumeManager) Resize(options api.ResizeVolumeOptions) (*api.Volume, *api.ResizeVolumeError) {
	err := volumeactions.ExtendSize(mgr.Provider.BaseServices.Volume, options.ID, volumeactions.ExtendSizeOpts{
		NewSize: int(options.Size)}).Err
	if err != nil {
		return nil, api.NewResizeVolumeError(UnwrapOpenStackError(err), options)
	}

	v, err := mgr.Get(options.ID)
	return v, api.NewResizeVolumeError(UnwrapOpenStackError(err), options)
}
