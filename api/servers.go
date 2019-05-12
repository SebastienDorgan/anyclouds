package api

import (
	"io"
	"time"
)

//ServerState possible states of an Server
type ServerState string

const (
	//ServerBuilding State of an Server in building
	ServerBuilding ServerState = "BUILDING"
	//ServerReady State of a started Server
	ServerReady ServerState = "READY"
	//ServerPaused State of a paused server
	ServerPaused ServerState = "PAUSED"
	//ServerDeleted State of a deleted server
	ServerDeleted ServerState = "DELETED"
	//ServerShutoff State of a shutoff server
	ServerShutoff ServerState = "SHUTOFF"
	//ServerInError State of an Server in error
	ServerInError ServerState = "ERROR"
	//ServerTransientState State of an Server in transient state
	ServerTransientState ServerState = "TRANSIENT"
	//ServerUnknownState State of an Server in error
	ServerUnknownState ServerState = "UNKNOWN"
)

//Server defines Server properties
type Server struct {
	ID             string
	Name           string
	TemplateID     string
	ImageID        string
	SecurityGroups []string
	PrivateIPs     map[IPVersion][]string
	PublicIPv4     string
	PublicIPv6     string
	State          ServerState
	CreatedAt      time.Time
	KeyPairName    string
	LeasingType    LeasingType
	LeaseDuration  time.Duration
}

//LeasingType type of leasing
type LeasingType int

const (
	//LeasingTypeOnDemand to lease on demand server instance
	LeasingTypeOnDemand LeasingType = 0
	//LeasingTypeSpot to lease on demand spot instance
	LeasingTypeSpot LeasingType = 1
	//LeasingTypeReserved to lease reserved instance
	LeasingTypeReserved LeasingType = 2
)

//CreateServerOptions defines options to use when creating an Server
type CreateServerOptions struct {
	Name            string
	TemplateID      string
	ImageID         string
	SecurityGroups  []string
	Subnets         []string
	PublicIP        bool
	BootstrapScript io.Reader
	KeyPairName     string
	LeasingType     LeasingType
	LeaseDuration   time.Duration
}

//ServerManager defines Server management functions an anyclouds provider must provide
type ServerManager interface {
	Create(options *CreateServerOptions) (*Server, error)
	Delete(id string) error
	List() ([]Server, error)
	Get(id string) (*Server, error)
	Start(id string) error
	Stop(id string) error
	Resize(id string, templateID string) error
}
