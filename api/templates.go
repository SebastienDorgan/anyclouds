package api

import "time"

//CPUArch enum defining CPU Architectures
type CPUArch string

//noinspection ALL
const (
	//Arch386 x86 32 bits architecture
	Arch386 CPUArch = "386"
	//ArchAmd64 x86 64 bits architecture
	ArchAmd64 CPUArch = "amd64"
	//ArchARM ARM 32 bits architecture
	ArchARM CPUArch = "arm"
	//ArchARM64 ARM 64 bits architecture
	ArchARM64 CPUArch = "arm"
	//ArchUnknown unknown architecture
	ArchUnknown CPUArch = "unknown"
)

//GPUInfo defines a GPUInfo
type GPUInfo struct {
	Number       int
	NumberOfCore int
	MemorySize   int
	Type         string
}

//ServerTemplate defines instance template type
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
	CPUFrequency      float32
	NetworkSpeed      int
	GPUInfo           *GPUInfo
	OneDemandPrice    float32
}

//ServerTemplateManager defines Server template management functions a anyclouds provider must provide
type ServerTemplateManager interface {
	List() ([]ServerTemplate, *ListServerTemplatesError)
	Get(id string) (*ServerTemplate, *GetServerTemplateError)
}

//ListServerTemplatesError list server templates error type
type ListServerTemplatesError struct {
	ErrorStack
}

//NewListServerTemplatesError  creates a new ListServerTemplatesError
func NewListServerTemplatesError(cause error) *ListServerTemplatesError {
	return &ListServerTemplatesError{ErrorStack: *NewErrorStack(cause, "error listing server templates")}
}

//GetServerTemplateError get server template error type
type GetServerTemplateError struct {
	ErrorStack
}

//NewGetServerTemplateError  creates a new GetServerTemplateError
func NewGetServerTemplateError(cause error, id string) *GetServerTemplateError {
	return &GetServerTemplateError{ErrorStack: *NewErrorStack(cause, "error get server templates", id)}
}
