package providers

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

//InstanceTemplate defines instace template type
type InstanceTemplate struct {
	ID              string
	Name            string
	CPUArch         CPUArch
	NumberOfCPUCore int
	//in GHz
	CPUCoreFrequency float64
	//in GB
	RAMSize float64
	//in GB
	SystemDiskSize int
	//in GB
	EphemeralDiskSize int
	NumberOfGPU       int
	NumberOfGPUCore   int
	GPURAMSize        int
	//in GHz
	GPUCoreFrequency int
}

//InstanceTemplateManager defines instance template management functions a anyclouds provider must provide
type InstanceTemplateManager interface {
	List(filter *ResourceFilter) ([]InstanceTemplate, error)
	Get(id string) (*InstanceTemplate, error)
}
