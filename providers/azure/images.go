package azure

import (
	"context"
	"fmt"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type ImageManager struct {
	Provider *Provider
}

func createId(publisher, offer, sku, version string) string {
	return fmt.Sprintf("%s##%s##%s##%s", publisher, offer, sku, version)
}

func parseId(id string) (publisher, offer, sku, version string) {
	tokens := strings.Split(id, "##")
	publisher = tokens[0]
	offer = tokens[1]
	sku = tokens[2]
	version = tokens[3]
	return
}

func (mgr *ImageManager) List() ([]api.Image, error) {
	cfg := mgr.Provider.Configuration

	var images []api.Image
	for _, publisher := range cfg.VirtualMachineImagePublishers {
		offers, err := mgr.Provider.VirtualMachineImagesClient.ListOffers(context.Background(), cfg.Location, publisher)
		if err != nil {
			return nil, errors.Wrap(err, "error listing images")
		}
		for _, offer := range *offers.Value {
			skus, err := mgr.Provider.VirtualMachineImagesClient.ListSkus(context.Background(), cfg.Location, publisher, *offer.Name)
			if err != nil {
				return nil, errors.Wrap(err, "error listing images")
			}
			for _, sku := range *skus.Value {
				maxResult := int32(100)
				versions, err := mgr.Provider.VirtualMachineImagesClient.List(context.Background(), cfg.Location, publisher, *offer.Name, *sku.Name, "", &maxResult, "")
				if err != nil {
					return nil, errors.Wrap(err, "error listing images")
				}
				for _, version := range *versions.Value {
					//img, err := mgr.Provider.VirtualMachineImagesClient.Get(context.Background(), cfg.Location, publisher, *offer.Name, *sku.Name, *version.Name)
					//if err != nil {
					//	return nil, errors.Wrap(err, "error listing images")
					//}
					//
					id := createId(publisher, *offer.Name, *sku.Name, *version.Name)
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

func (mgr *ImageManager) Get(id string) (*api.Image, error) {
	cfg := mgr.Provider.Configuration
	publisher, offer, sku, version := parseId(id)
	_, err := mgr.Provider.VirtualMachineImagesClient.Get(context.Background(), cfg.Location, publisher, offer, sku, version)
	if err != nil {
		return nil, errors.Wrap(err, "error getting images")
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
