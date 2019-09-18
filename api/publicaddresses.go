package api

//PublicIP represent a public ip address
type PublicIP struct {
	ID                 string
	Name               string
	Address            string
	NetworkInterfaceID string
	PrivateAddress     string
}

//CreatePublicIPOptions options that can be used to allocate a public ip address
type CreatePublicIPOptions struct {
	Name            string
	IPAddress       *string
	IPAddressPoolID *string
}

//AssociatePublicIPOptions options that can be used to associate a public ip address to a server
type AssociatePublicIPOptions struct {
	PublicIPId string
	ServerID   string
	//If the server is attached to more than one network this field must be provided
	SubnetID string
	//If the server has more than one private IP by subnet this field can be provided to control the private IP
	// correlated with the public ip
	PrivateIP string
}

//AddressRange a contiguous range of address
type AddressRange struct {
	FirstAddress string
	LastAddress  string
}

//PublicIPPool a pool of public IP
type PublicIPPool struct {
	ID     string
	Ranges []AddressRange
}

//ListPublicIPsOptions options to be used to list public ip address
type ListPublicIPsOptions struct {
	ServerID *string
}

//PublicIPManager an interface providing an abstraction to manipulate public ip addresses
type PublicIPManager interface {
	ListAvailablePools() ([]PublicIPPool, ListAvailablePublicIPPoolsError)
	List(options *ListPublicIPsOptions) ([]PublicIP, ListPublicIPsError)
	Create(options CreatePublicIPOptions) (*PublicIP, CreatePublicIPError)
	Associate(options AssociatePublicIPOptions) AssociatePublicIPError
	Dissociate(id string) DissociatePublicIPError
	Delete(id string) DeletePublicIPError
	Get(id string) (*PublicIP, GetPublicIPError)
}

//ListAvailablePublicIPPoolsError list available public ip pools error type
type ListAvailablePublicIPPoolsError interface {
	Error() string
}

//NewListAvailablePublicIPPoolsError create a new ListAvailablePublicIPPoolsError
func NewListAvailablePublicIPPoolsError(cause error) ListAvailablePublicIPPoolsError {
	if cause == nil {
		return nil
	}
	return NewErrorStack(cause, "error listing available public ip pools")
}

//ListPublicIPsError list created public ips error type
type ListPublicIPsError interface {
	Error() string
}

//NewListPublicIPsError create a new ListPublicIPsError
func NewListPublicIPsError(cause error, options *ListPublicIPsOptions) ListPublicIPsError {
	if cause == nil {
		return nil
	}
	return NewErrorStack(cause, "error listing created public ips", options)
}

//CreatePublicIPError create public ip error type
type CreatePublicIPError interface {
	Error() string
}

//NewCreatePublicIPError create a new CreatePublicIPError
func NewCreatePublicIPError(cause error, options CreatePublicIPOptions) CreatePublicIPError {
	if cause == nil {
		return nil
	}
	return NewErrorStack(cause, "error creating public ip", options)
}

//AssociatePublicIPError associate public ip error type
type AssociatePublicIPError interface {
	Error() string
}

//NewAssociatePublicIPError create a new AssociatePublicIPError
func NewAssociatePublicIPError(cause error, options AssociatePublicIPOptions) AssociatePublicIPError {
	if cause == nil {
		return nil
	}
	return NewErrorStack(cause, "error associating public ip", options)
}

//DissociatePublicIPError dissociate public ip error type
type DissociatePublicIPError interface {
	Error() string
}

//NewDissociatePublicIPError create a new AssociatePublicIPError
func NewDissociatePublicIPError(cause error, id string) DissociatePublicIPError {
	if cause == nil {
		return nil
	}
	return NewErrorStack(cause, "error dissociating public ip", id)
}

//DeletePublicIPError delete public ip error type
type DeletePublicIPError interface {
	Error() string
}

//NewDeletePublicIPError create a new DeletePublicIPError
func NewDeletePublicIPError(cause error, id string) DeletePublicIPError {
	if cause == nil {
		return nil
	}
	return NewErrorStack(cause, "error deleting public ip", id)
}

//GetPublicIPError get public ip error type
type GetPublicIPError interface {
	Error() string
}

//NewGetPublicIPError create a new GetPublicIPError
func NewGetPublicIPError(cause error, id string) GetPublicIPError {
	if cause == nil {
		return nil
	}
	return NewErrorStack(cause, "error getting public ip", id)
}
