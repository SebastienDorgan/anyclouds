package openstack

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/pkg/errors"
)

//ImageManager defines image management functions a anyclouds provider must provide
type ImageManager struct {
	OpenStack *Provider
}

//List returns available image list
func (mgr *ImageManager) List() ([]api.Image, error) {
	opts := images.ListOpts{}

	// Retrieve a pager (i.e. a paginated collection)
	page, err := images.List(mgr.OpenStack.Compute, opts).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing images")
	}
	imageList, err := images.ExtractImages(page)
	var imgList []api.Image
	for _, img := range imageList {
		imgList = append(imgList, api.Image{
			ID:        img.ID,
			Name:      img.Name,
			MinDisk:   img.MinDiskGigabytes,
			MinRAM:    img.MinRAMMegabytes,
			CreatedAt: img.CreatedAt,
			UpdatedAt: img.UpdatedAt,
		})
	}
	return imgList, nil
}

//Get returns the image identified by id
func (mgr *ImageManager) Get(id string) (*api.Image, error) {
	res := images.Get(mgr.OpenStack.Compute, id)
	//img, err := images.Get(mgr.OpenStack.Compute, id).Extract()
	str := res.PrettyPrintJSON()
	println(str)
	img, err := res.Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting image: %s")
	}
	return &api.Image{
		ID:        img.ID,
		Name:      img.Name,
		MinDisk:   img.MinDiskGigabytes,
		MinRAM:    img.MinRAMMegabytes,
		CreatedAt: img.CreatedAt,
		UpdatedAt: img.UpdatedAt,
	}, nil
}
