package providers

import "io"

//InstanceState possible states of an instance
type InstanceState string

const (
	//InstanceBuilding State of an instance in building
	InstanceBuilding InstanceState = "BUILDING"
	//InstanceStarted State of a started instance
	InstanceStarted InstanceState = "STARTED"
	//InstanceStarting State of an instance starting
	InstanceStarting InstanceState = "STARTING"
	//InstanceStopped State of a stopped instance
	InstanceStopped InstanceState = "STOPPED"
	//InstanceStopping State of an instance stopping
	InstanceStopping InstanceState = "STOPPING"
	//InstanceInError State of an instance in error
	InstanceInError InstanceState = "ERROR"
)

//Instance defines instance properties
type Instance struct {
	ID            string
	Name          string
	TemplateID    string
	ImageID       string
	PrivateIPS    []string
	PublicIP      string
	InstanceState InstanceState
}

//InstanceOptions defines options to use when creating an instance
type InstanceOptions struct {
	Name            string
	TemplateID      string
	ImageID         string
	SecurityGroupID string
	Subnets         []string
	PublicIP        bool
	BootstrapScript io.Reader
}

//InstanceManager defines instance management functions an anyclouds provider must provide
type InstanceManager interface {
	Create(options *InstanceOptions) (*Instance, error)
	Delete(id string) error
	List(filter *ResourceFilter) ([]Instance, error)
	Get(id string) (*Instance, error)
	Start(id string) error
	Stop(id string) error
}
