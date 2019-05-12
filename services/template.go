package services

import (
	"math"
	"sort"

	"github.com/SebastienDorgan/anyclouds/api"
)

func fakeExtra(srv *api.ServerTemplate) *TemplateSelector {
	return &TemplateSelector{
		CPUArch:           api.ArchX8664,
		NumberOfCPUCore:   srv.NumberOfCPUCore,
		CPUCoreFrequency:  0.0,
		RAMSize:           srv.RAMSize,
		SystemDiskSize:    srv.SystemDiskSize,
		EphemeralDiskSize: srv.EphemeralDiskSize,
		NumberOfGPU:       0,
		NumberOfGPUCore:   0,
		GPURAMSize:        0,
		GPUCoreFrequency:  0,
		HourlyPrice:       0,
	}
}

func mergeExtra(srv *api.ServerTemplate, ext *TemplateSelector) *TemplateSelector {
	ext2 := *ext
	ext2.NumberOfCPUCore = srv.NumberOfCPUCore
	ext2.RAMSize = srv.RAMSize
	ext2.SystemDiskSize = srv.SystemDiskSize
	ext2.EphemeralDiskSize = srv.EphemeralDiskSize
	return &ext2
}

func checkWithExtra(ext *TemplateSelector, filter *TemplateSelector) bool {
	if ext.CPUArch != filter.CPUArch {
		return false
	}
	if ext.NumberOfCPUCore < filter.NumberOfCPUCore {
		return false
	}
	if ext.CPUCoreFrequency < filter.CPUCoreFrequency {
		return false
	}
	if ext.RAMSize < filter.RAMSize {
		return false
	}
	if ext.SystemDiskSize < filter.SystemDiskSize {
		return false
	}
	if ext.EphemeralDiskSize < filter.EphemeralDiskSize {
		return false
	}
	if ext.NumberOfGPU < filter.NumberOfGPU {
		return false
	}
	if ext.NumberOfGPUCore < filter.NumberOfGPUCore {
		return false
	}
	if ext.GPURAMSize < filter.GPURAMSize {
		return false
	}
	if ext.GPUCoreFrequency < filter.GPUCoreFrequency {
		return false
	}
	return true
}

//DRFTemplateResultsWeights weight to be considered by the DRFScorer
type DRFTemplateResultsWeights struct {
	NumberOfCPUCore   float64
	CPUCoreFrequency  float64
	RAMSize           float64
	SystemDiskSize    float64
	EphemeralDiskSize float64
	NumberOfGPU       float64
	NumberOfGPUCore   float64
	GPURAMSize        float64
	GPUCoreFrequency  float64
}

//DefaultDRFTemplateResultsWeights default DRF scorer configuration
var DefaultDRFTemplateResultsWeights = DRFTemplateResultsWeights{
	NumberOfCPUCore:   1.0,
	CPUCoreFrequency:  1.0,
	RAMSize:           1.0 / 4.0,
	SystemDiskSize:    1.0 / 10,
	EphemeralDiskSize: 1.0 / 40,
	NumberOfGPU:       2.0,
	NumberOfGPUCore:   1.0 / 200.0,
	GPURAMSize:        1.0 / 4.0,
	GPUCoreFrequency:  1.0,
}

//Scorer scores a template result
type Scorer func(tpl *TemplateResult) float64

//DRFScorer scores result againts filter using Dominant Resource Fairness algorithm
func DRFScorer(weights *DRFTemplateResultsWeights, filter *TemplateSelector) Scorer {
	w := weights
	if w == nil {
		w = &DefaultDRFTemplateResultsWeights
	}
	return func(tpl *TemplateResult) float64 {
		score := float64(tpl.NumberOfCPUCore-filter.NumberOfCPUCore) * w.NumberOfCPUCore
		score += float64(tpl.CPUCoreFrequency-filter.CPUCoreFrequency) * w.CPUCoreFrequency
		score += float64(tpl.RAMSize-filter.RAMSize) * w.RAMSize
		score += float64(tpl.SystemDiskSize-filter.SystemDiskSize) * w.SystemDiskSize
		score += float64(tpl.EphemeralDiskSize-filter.EphemeralDiskSize) * w.EphemeralDiskSize
		score += float64(tpl.NumberOfGPU-filter.NumberOfGPU) * w.NumberOfGPU
		score += float64(tpl.NumberOfGPUCore-filter.NumberOfGPUCore) * w.NumberOfGPUCore
		score += float64(tpl.GPURAMSize-filter.GPURAMSize) * w.GPURAMSize
		score += float64(tpl.GPUCoreFrequency-filter.GPUCoreFrequency) * w.GPUCoreFrequency
		return score
	}
}

//HourlyPriceScorer score results by houly price
func HourlyPriceScorer(tpl *TemplateResult) float64 {
	return tpl.HourlyPrice
}

//MonthlyPriceScorer score results by taking account of the minimum renting duration in month
func MonthlyPriceScorer(tpl *TemplateResult, numberOfMonthes int) float64 {
	if tpl.MonthlyPrices == nil {
		return 365.25 / 12.0 * 24 * tpl.HourlyPrice
	} else if p, ok := tpl.MonthlyPrices[numberOfMonthes]; ok {
		return p
	} else {
		minPrice := math.MaxFloat64
		for k, p := range tpl.MonthlyPrices {
			minPrice = math.Min(p/float64(k)*float64(numberOfMonthes), minPrice)
		}
		return minPrice
	}
}

type results struct {
	results []TemplateResult
	scorer  *Scorer
}

func (res *results) Len() int {
	return len(res.results)
}
func (res *results) Swap(i, j int) {
	res.results[i], res.results[j] = res.results[j], res.results[i]
}
func (res *results) Score(i int) float64 {

	if res.scorer == nil {
		return 0.0
	}
	scorer := *res.scorer
	return scorer(&res.results[i])

}
func (res *results) Less(i, j int) bool {
	return res.Score(i) < res.Score(j)
}

//SearchTemplates finds template matching filter
func (sv *ProviderService) SearchTemplates(tpls []api.ServerTemplate, filter *TemplateSelector) ([]TemplateResult, error) {
	var res []TemplateResult
	for _, tpl := range tpls {
		var ext *TemplateSelector
		if tmp, ok := sv.extra[tpl.Name]; ok {
			ext = mergeExtra(&tpl, &tmp)
		} else {
			ext = fakeExtra(&tpl)
		}
		if checkWithExtra(ext, filter) {
			res = append(res, TemplateResult{
				ID:               tpl.ID,
				TemplateSelector: *ext,
			})
		}
	}

	s := DRFScorer(&DefaultDRFTemplateResultsWeights, filter)
	results := results{
		results: res,
		scorer:  &s,
	}
	sort.Sort(&results)
	return res, nil
}

//SortTemplateResultList sort template results in ascending order using scorer
func SortTemplateResultList(templates []TemplateResult, scorer Scorer) {
	results := results{
		results: templates,
		scorer:  &scorer,
	}
	sort.Sort(&results)
}
