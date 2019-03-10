package services

import (
	"github.com/SebastienDorgan/anyclouds/api"
)

//TemplateExtra template additionnal information
type TemplateSelector struct {
	Name            string      `json:"name,omitempty"`
	CPUArch         api.CPUArch `json:"cpu_arch,omitempty"`
	NumberOfCPUCore int         `json:"number_of_cpu_core,omitempty"`
	//in GHz
	CPUCoreFrequency float64 `json:"cpu_core_frequency,omitempty"`
	//in MB
	RAMSize int `json:"ram_size,omitempty"`
	//in GB
	SystemDiskSize int `json:"system_disk_size,omitempty"`
	//in I/O per seconds
	SystemDiskIOPS int `json:"system_disk_iops,omitempty"`
	//in GB
	EphemeralDiskSize int `json:"ephemeral_disk_size,omitempty"`
	//in I/O per seconds
	EphemeralDiskIOPS int     `json:"ephemeral_disk_iops,omitempty"`
	NumberOfGPU       int     `json:"number_of_gpu,omitempty"`
	NumberOfGPUCore   int     `json:"number_of_gpu_core,omitempty"`
	GPURAMSize        float64 `json:"gpuram_size,omitempty"`
	//in GHz
	GPUCoreFrequency float64 `json:"gpu_core_frequency,omitempty"`
	//choose your currency
	HourlyPrice   float64         `json:"hourly_price,omitempty"`
	MonthlyPrices map[int]float64 `json:"monthly_prices,omitempty"`
}

//TemplatesExtra templates additionnal information
type TemplatesExtra map[string]TemplateSelector

//TemplateResult result of a template selection
type TemplateResult struct {
	TemplateSelector
	Name string
	ID   string
}
