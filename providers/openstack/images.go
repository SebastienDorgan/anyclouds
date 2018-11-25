package openstack

import (
	"github.com/SebastienDorgan/anyclouds/providers"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/pkg/errors"
)

//ImageManager defines image management functions a anyclouds provider must provide
type ImageManager struct {
	OpenStack *Provider
}

//List returns available image list
func (mgr *ImageManager) List(filter *providers.ResourceFilter) ([]providers.Image, error) {
	opts := images.ListOpts{
		ID:   filter.ID,
		Name: filter.Name,
	}

	// Retrieve a pager (i.e. a paginated collection)
	page, err := images.List(mgr.OpenStack.Compute, opts).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing images")
	}
	imageList, err := images.ExtractImages(page)
	var imgList []providers.Image
	for _, img := range imageList {
		imgList = append(imgList, providers.Image{
			ID:      img.ID,
			Name:    img.Name,
			MinDisk: 0,
			MinRAM:  0,
		})
	}
	return imgList, nil
}

//Get retuns the image identified by id
func (mgr *ImageManager) Get(id string) (*providers.Image, error) {
	img, err := images.Get(mgr.OpenStack.Compute, id).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting image: %s")
	}
	return &providers.Image{
		ID:      img.ID,
		Name:    img.Name,
		MinDisk: 0,
		MinRAM:  0,
	}, nil
}
