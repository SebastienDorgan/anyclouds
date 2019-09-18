package api

import (
	"github.com/SebastienDorgan/anyclouds/sshutils"
	"io"
	"time"
)

//ServerState possible states of an Server
type ServerState string

const (
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
	//ServerPending State of an Server in transient state
	ServerPending ServerState = "TRANSIENT"
	//ServerUnknownState State of an Server in error
	ServerUnknownState ServerState = "UNKNOWN"
)

//LowPriorityServerOptions options that can be used to create a spot instance
type LowPriorityServerOptions struct {
	HourlyPrice float32
	Duration    time.Duration
}

//ReservedServerOptions options that can be use to create a reserved instance
type ReservedServerOptions struct {
	Duration time.Duration
}

//Server defines Server properties
type Server struct {
	ID            string
	Name          string
	TemplateID    string
	ImageID       string
	State         ServerState
	CreatedAt     time.Time
	LeasingType   LeasingType
	LeaseDuration time.Duration
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
	Name                     string
	TemplateID               string
	ImageID                  string
	DefaultSecurityGroup     string
	Subnets                  []Subnet
	BootstrapScript          io.Reader
	KeyPair                  sshutils.KeyPair
	LowPriorityServerOptions *LowPriorityServerOptions
	ReservedServerOptions    *ReservedServerOptions
}

//ServerManager defines Server management functions an anyclouds provider must provide
type ServerManager interface {
	Create(options CreateServerOptions) (*Server, CreateServerError)
	Delete(id string) DeleteServerError
	List() ([]Server, ListServersError)
	Get(id string) (*Server, GetServerError)
	Start(id string) StartServerError
	Stop(id string) StopServerError
	Resize(id string, templateID string) ResizeServerError
}

//CreateServerError create server error type
type CreateServerError interface {
	Error() string
}

//NewCreateServerError creates a new CreateServerError
func NewCreateServerError(cause error, options CreateServerOptions) CreateServerError {
	return NewErrorStack(cause, "error creating server", options)
}

//DeleteServerError delete server error type
type DeleteServerError interface {
	Error() string
}

//NewDeleteServerError creates a new DeleteServerError
func NewDeleteServerError(cause error, id string) DeleteServerError {
	return NewErrorStack(cause, "error deleting server", id)
}

//ListServersError list servers error type
type ListServersError interface {
	Error() string
}

//NewListServersError creates a new ListServersError
func NewListServersError(cause error) ListServersError {
	return NewErrorStack(cause, "error listing servers")
}

//GetServerError get server error type
type GetServerError interface {
	Error() string
}

//NewGetServerError creates a new GetServerError
func NewGetServerError(cause error, id string) GetServerError {
	return NewErrorStack(cause, "error getting server", id)
}

//StartServerError start server error type
type StartServerError interface {
	Error() string
}

//NewStartServerError creates a new StartServerError
func NewStartServerError(cause error, id string) StartServerError {
	return NewErrorStack(cause, "error starting server", id)
}

//StopServerError start server error type
type StopServerError interface {
	Error() string
}

//NewStopServerError creates a new StopServerError
func NewStopServerError(cause error, id string) StopServerError {
	return NewErrorStack(cause, "error stopping server", id)
}

//ResizeServerError start server error type
type ResizeServerError interface {
	Error() string
}

//NewResizeServerError creates a new ResizeServerError
func NewResizeServerError(cause error, id string, templateID string) ResizeServerError {
	return NewErrorStack(cause, "error resizing server", id, templateID)
}
