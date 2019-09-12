package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/profiles/preview/preview/commerce/mgmt/commerce"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/pkg/errors"
	"time"
)

type ServerTemplateManager struct {
	Provider *Provider
}

func (mgr *ServerTemplateManager) GetVMMeters() ([]commerce.MeterInfo, error) {
	filter := fmt.Sprintf("OfferDurableId eq ’%s’ and Currency eq ’%s’ and Locale eq ’en-US’ and RegionInfo eq ’%s’",
		mgr.Provider.Configuration.OfferNumber,
		mgr.Provider.Configuration.Currency,
		mgr.Provider.Configuration.RegionInfo)

	result, err := mgr.Provider.RateCardClient.Get(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	if result.Meters == nil {
		return nil, errors.Errorf("unable to retrieve meter info for offer number %s", mgr.Provider.Configuration.OfferNumber)
	}
	var vmMeters []commerce.MeterInfo
	for _, mi := range *result.Meters {
		if mi.MeterRegion != nil && *mi.MeterRegion == mgr.Provider.Configuration.Location &&
			mi.MeterCategory != nil && *mi.MeterCategory == "Virtual Machines" &&
			mi.MeterName != nil && *mi.MeterName == "Compute Hours" {
			vmMeters = append(vmMeters, mi)
		}
	}
	return vmMeters, nil
}

func GetMeter(vmMeters []commerce.MeterInfo, sizeName string) *commerce.MeterInfo {
	for _, mi := range vmMeters {
		if *mi.MeterSubCategory == sizeName {
			return &mi
		}
	}
	return nil
}

func (mgr *ServerTemplateManager) List() ([]api.ServerTemplate, *api.ListServerTemplatesError) {
	list, err := mgr.Provider.VirtualMachineSizesClient.List(context.Background(), mgr.Provider.Configuration.Location)
	if err != nil {
		return nil, api.NewListServerTemplatesError(err)
	}
	var templates []api.ServerTemplate
	vmMeters, err := mgr.GetVMMeters()
	for _, size := range *list.Value {
		meterInfo := GetMeter(vmMeters, *size.Name)
		templates = append(templates, api.ServerTemplate{
			ID:                *size.Name,
			Name:              *size.Name,
			NumberOfCPUCore:   int(*size.NumberOfCores),
			RAMSize:           int(*size.MemoryInMB),
			SystemDiskSize:    int(*size.OsDiskSizeInMB / 1000),
			EphemeralDiskSize: int(*size.ResourceDiskSizeInMB / 1000),
			CreatedAt:         time.Date(2016, 1, 1, 0, 0, 0, 0, nil),
			Arch:              api.ArchAmd64,
			CPUFrequency:      0,
			NetworkSpeed:      0,
			GPUInfo:           nil,
			OneDemandPrice:    float32(*meterInfo.MeterRates["0"]),
		})
	}

	return templates, nil
}

func (mgr *ServerTemplateManager) Get(id string) (*api.ServerTemplate, *api.GetServerTemplateError) {
	list, err := mgr.List()
	if err != nil {
		return nil, api.NewGetServerTemplateError(err, id)
	}
	for _, tpl := range list {
		if tpl.ID == id {
			return &tpl, nil
		}
	}
	return nil, api.NewGetServerTemplateError(err, id)
}
