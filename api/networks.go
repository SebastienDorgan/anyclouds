package api

//Network defines network properties
type Network struct {
	//unique identifier of the network
	ID string
	//DeviceName of the network
	Name string
	//Cidr of the network
	CIDR string
}

//CreateNetworkOptions defines options to use when creating a network
type CreateNetworkOptions struct {
	//name of the network
	CIDR string
	//DeviceName of the network
	Name string
}

//IPVersion ip version enum
type IPVersion int

const (
	//IPVersion6 IPv6
	IPVersion6 IPVersion = 6
	//IPVersion4 IPv4
	IPVersion4 IPVersion = 4
	//IPVersion4And6 IPv4 and IPv6
	IPVersion4And6 IPVersion = 10
)

//CreateSubnetOptions defines options to use when creating a sub network
type CreateSubnetOptions struct {
	//NetworkID identifier of the parent network
	NetworkID string
	//name of the sub network
	Name string
	//CIDR of the sub network
	CIDR string
	//IP Version
	IPVersion IPVersion
}

//Subnet defines sub network properties
type Subnet struct {
	//unique identifier of the sub network
	ID string
	//identifier of the parent network (i.e. Network.ID)
	NetworkID string
	//name of the sub networkn
	Name string
	//CIDR of the sub network
	CIDR string
	//IP Version
	IPVersion IPVersion
}

//NetworkManager defines networking functions a anyclouds provider must provide
type NetworkManager interface {
	CreateNetwork(options CreateNetworkOptions) (*Network, *CreateNetworkError)
	DeleteNetwork(id string) *DeleteNetworkError
	ListNetworks() ([]Network, *ListNetworksError)
	GetNetwork(id string) (*Network, *GetNetworkError)

	CreateSubnet(options CreateSubnetOptions) (*Subnet, *CreateSubnetError)
	DeleteSubnet(networkID string, subnetID string) *DeleteSubnetError
	ListSubnets(networkID string) ([]Subnet, *ListSubnetsError)
	GetSubnet(networkID, subnetID string) (*Subnet, *GetSubnetError)
}

//CreateNetworkError create network error type
type CreateNetworkError struct {
	ErrorStack
}

//NewCreateNetworkError create a new CreateNetworkError
func NewCreateNetworkError(cause error, options CreateNetworkOptions) *CreateNetworkError {
	if cause == nil {
		return nil
	}
	return &CreateNetworkError{*NewErrorStack(cause, "error creating network", options)}
}

//DeleteNetworkError delete subnet error type
type DeleteNetworkError struct {
	ErrorStack
}

//NewDeleteNetworkError create a new DeleteNetworkError
func NewDeleteNetworkError(cause error, id string) *DeleteNetworkError {
	if cause == nil {
		return nil
	}
	return &DeleteNetworkError{*NewErrorStack(cause, "error deleting network", id)}
}

//GetNetworkError get network error type
type GetNetworkError struct {
	ErrorStack
}

//NewGetNetworkError create a new GetNetworkError
func NewGetNetworkError(cause error, id string) *GetNetworkError {
	if cause == nil {
		return nil
	}
	return &GetNetworkError{*NewErrorStack(cause, "error getting network", id)}
}

//ListNetworksError list network error type
type ListNetworksError struct {
	ErrorStack
}

//NewListNetworksError create a new ListNetworkInterfacesError
func NewListNetworksError(cause error) *ListNetworksError {
	if cause == nil {
		return nil
	}
	return &ListNetworksError{*NewErrorStack(cause, "error listing network")}
}

//CreateSubnetError create subnet error type
type CreateSubnetError struct {
	ErrorStack
}

//NewCreateSubnetError create a new CreateSubnetError
func NewCreateSubnetError(cause error, options CreateSubnetOptions) *CreateSubnetError {
	if cause == nil {
		return nil
	}
	return &CreateSubnetError{*NewErrorStack(cause, "error creating subnet", options)}
}

//GetSubnetError get subnet error type
type GetSubnetError struct {
	ErrorStack
}

//NewGetSubnetError create a new GetSubnetError
func NewGetSubnetError(cause error, networkID string, subnetID string) *GetSubnetError {
	if cause == nil {
		return nil
	}
	return &GetSubnetError{*NewErrorStack(cause, "error getting subnet", networkID, subnetID)}
}

//ListSubnetsError list subnet error type
type ListSubnetsError struct {
	ErrorStack
}

//NewListSubnetsError create a new ListSubnetsError
func NewListSubnetsError(cause error, networkID string) *ListSubnetsError {
	if cause == nil {
		return nil
	}
	return &ListSubnetsError{*NewErrorStack(cause, "error listing subnets", networkID)}
}

//DeleteSubnetError delete subnet error type
type DeleteSubnetError struct {
	ErrorStack
}

//NewDeleteSubnetError create a new DeleteSubnetError
func NewDeleteSubnetError(cause error, networkID string, subnetID string) *DeleteSubnetError {
	if cause == nil {
		return nil
	}
	return &DeleteSubnetError{*NewErrorStack(cause, "error deleting network", networkID, subnetID)}
}
