package anyclouds

import (
	"time"

	"github.com/SebastienDorgan/anyclouds/api"
)

//HardwareResources server hardware resources
type HardwareResources struct {
	CPUArch         api.CPUArch
	NumberOfCPUCore int
	//in GHz
	CPUCoreFrequency float64
	//in GB
	RAMSize float64
	//in GB
	SystemDiskSize int
	//in I/O per seconds
	SystemDiskIOPS int
	//in GB
	EphemeralDiskSize int
	//in I/O per seconds
	EphemeralDiskIOPS int
	NumberOfGPU       int
	NumberOfGPUCore   int
	GPURAMSize        float64
	//in GHz
	GPUCoreFrequency float64
}

//Resource a resource is qulified by an indentifier, a name, and a creation date.
type Resource struct {
	ID           string
	Name         string
	CreationDate time.Time
}

//Server popoerties of a server
type Server struct {
	Resource
	HardwareResources
}

//Network properties of a network
type Network struct {
	Resource
	CIDR string
}

//NetworkManager defines functions offered by anyclouds to manage networks
type NetworkManager interface {
	Create(name string, CIDR string) (*Network, error)
	Delete(ref string) error
	Attach(ref string, serverRef string) error
	Detach(ref string, serverRef string) error
	Inspect(ref string) (*Network, error)
}

//ServerManager defines functions offered by anyclouds to manage serves
type ServerManager interface {
	Create(name string, rentalDuration int, networkRef string, resource *HardwareResources) (*Server, error)
	Delete(ref string) error
	Inspect(ref string) (*Server, error)
	Resize(ref string, resource *HardwareResources) (*Server, error)
	Start(ref string) error
	Stop(ref string) error
}

//VolumeManager defines functions offered by anyclouds to manage volumes
type VolumeManager interface {
}
