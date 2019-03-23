package api

import "time"

//CPUArch enum defining CPU Architectures
type CPUArch string

const (
	//ArchX86 x86 32 bits architecture
	ArchX86 CPUArch = "x86"
	//ArchX86_64 x86 64 bits architecture
	ArchX86_64 CPUArch = "x86_64"
	//ArchARM architecture
	ArchARM CPUArch = "ARM"
	//ArchUnknown unknown architecture
	ArchUnknown CPUArch = "UNKNOWN"
)

//GPU defines a GPU
type GPU struct {
	Number int
	Type   string
}

//ServerTemplate defines instace template type
type ServerTemplate struct {
	ID              string
	Name            string
	NumberOfCPUCore int
	//in MB
	RAMSize int
	//in GB
	SystemDiskSize int
	//in GB
	EphemeralDiskSize int
	CreatedAt         time.Time
	Arch              CPUArch
	CPUSpeed          float32
	GPU               *GPU
}

//ServerTemplateManager defines Server template management functions a anyclouds provider must provide
type ServerTemplateManager interface {
	List() ([]ServerTemplate, error)
	Get(id string) (*ServerTemplate, error)
}
