package azure

import (
	"context"
	"fmt"
	"github.com/SebastienDorgan/anyclouds/api"
	"strings"
	"time"
)

type ImageManager struct {
	Provider *Provider
}

func createImageID(publisher, offer, sku, version string) string {
	return fmt.Sprintf("%s##%s##%s##%s", publisher, offer, sku, version)
}

func parseImageID(id string) (publisher, offer, sku, version string) {
	tokens := strings.Split(id, "##")
	publisher = tokens[0]
	offer = tokens[1]
	sku = tokens[2]
	version = tokens[3]
	return
}

func (mgr *ImageManager) list() ([]api.Image, error) {
	cfg := mgr.Provider.Configuration

	var images []api.Image
	for _, publisher := range cfg.VirtualMachineImagePublishers {
		offers, err := mgr.Provider.BaseServices.VirtualMachineImagesClient.ListOffers(context.Background(), cfg.Location, publisher)
		if err != nil {
			return nil, err
		}
		for _, offer := range *offers.Value {
			skus, err := mgr.Provider.BaseServices.VirtualMachineImagesClient.ListSkus(context.Background(), cfg.Location, publisher, *offer.Name)
			if err != nil {
				return nil, err
			}
			for _, sku := range *skus.Value {
				maxResult := int32(100)
				versions, err := mgr.Provider.BaseServices.VirtualMachineImagesClient.List(context.Background(), cfg.Location, publisher, *offer.Name, *sku.Name, "", &maxResult, "")
				if err != nil {
					return nil, err
				}
				for _, version := range *versions.Value {
					id := createImageID(publisher, *offer.Name, *sku.Name, *version.Name)
					images = append(images, api.Image{
						ID:        id,
						Name:      id,
						MinDisk:   0,
						MinRAM:    0,
						CreatedAt: time.Date(2016, 1, 1, 0, 0, 0, 0, nil),
						UpdatedAt: time.Date(2016, 1, 1, 0, 0, 0, 0, nil),
					})
				}
			}
		}
	}
	return images, nil

}

func (mgr *ImageManager) List() ([]api.Image, *api.ListImageError) {
	l, err := mgr.list()
	return l, api.NewListImageError(err)
}
func (mgr *ImageManager) get(id string) (*api.Image, error) {
	cfg := mgr.Provider.Configuration
	publisher, offer, sku, version := parseImageID(id)
	_, err := mgr.Provider.BaseServices.VirtualMachineImagesClient.Get(context.Background(), cfg.Location, publisher, offer, sku, version)
	if err != nil {
		return nil, err
	}

	return &api.Image{
		ID:        id,
		Name:      id,
		MinDisk:   0,
		MinRAM:    0,
		CreatedAt: time.Date(2016, 1, 1, 0, 0, 0, 0, nil),
		UpdatedAt: time.Date(2016, 1, 1, 0, 0, 0, 0, nil),
	}, nil
}

func (mgr *ImageManager) Get(id string) (*api.Image, *api.GetImageError) {
	i, err := mgr.get(id)
	return i, api.NewGetImageError(err, id)
}
