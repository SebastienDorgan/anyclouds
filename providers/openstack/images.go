package openstack

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/pkg/errors"
	"time"
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
		im, err := mgr.Get(img.ID)
		if err != nil {
			return nil, errors.Wrap(ProviderError(err), "Error listing images")
		}
		imgList = append(imgList, *im)
	}
	return imgList, nil
}

//Get returns the image identified by id
func (mgr *ImageManager) Get(id string) (*api.Image, error) {
	res := images.Get(mgr.OpenStack.Compute, id)
	img, err := res.Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting image: %s")
	}
	if len(img.ID) > 0 {
		return &api.Image{
			ID:        img.ID,
			Name:      img.Name,
			MinDisk:   img.MinDiskGigabytes,
			MinRAM:    img.MinRAMMegabytes,
			CreatedAt: img.CreatedAt,
			UpdatedAt: img.UpdatedAt,
		}, nil
	}
	//img, err := images.Get(mgr.OpenStack.Compute, id).Extract()
	type image map[string]interface{}
	type newImage struct {
		Image image `json:"image"`
	}
	var ni newImage
	err = res.ExtractInto(&ni)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting image: %s")
	}
	createdAt, _ := time.Parse(time.RFC3339, ni.Image["created"].(string))
	updatedAt, _ := time.Parse(time.RFC3339, ni.Image["updated"].(string))
	//minDisk, _ := strconv.Atoi(ni.Image["minDisk"].(string))
	//minRam, _ := strconv.Atoi(ni.Image["minRam"].(string))
	return &api.Image{
		ID:        ni.Image["id"].(string),
		Name:      ni.Image["name"].(string),
		MinDisk:   int(ni.Image["minDisk"].(float64)),
		MinRAM:    int(ni.Image["minRam"].(float64)),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil

}
