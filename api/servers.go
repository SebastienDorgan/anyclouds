package api

import "io"

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
	//ServerUnknwonState State of an Server in error
	ServerUnknwonState ServerState = "UNKNOWN"
)

//Server defines Server properties
type Server struct {
	ID         string
	Name       string
	TemplateID string
	ImageID    string
	PrivateIPs map[IPVersion][]string
	PublicIPv4 string
	PublicIPv6 string
	State      ServerState
}

//ServerOptions defines options to use when creating an Server
type ServerOptions struct {
	Name            string
	TemplateID      string
	ImageID         string
	SecurityGroupID string
	Networks        []string
	PublicIP        bool
	BootstrapScript io.Reader
}

//ServerManager defines Server management functions an anyclouds provider must provide
type ServerManager interface {
	Create(options *ServerOptions) (*Server, error)
	Delete(id string) error
	List() ([]Server, error)
	Get(id string) (*Server, error)
	Start(id string) error
	Stop(id string) error
}
