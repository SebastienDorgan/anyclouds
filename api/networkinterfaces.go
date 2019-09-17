package api

//NetworkInterface represents an network interface card
type NetworkInterface struct {
	ID               string
	Name             string
	MacAddress       string
	NetworkID        string
	SubnetID         string
	ServerID         string
	PublicIPAddress  string
	PrivateIPAddress string
	SecurityGroupID  string
}

//CreateNetworkInterfaceOptions options that can be used to create a network interface card
type CreateNetworkInterfaceOptions struct {
	Name             string
	NetworkID        string
	SubnetID         string
	ServerID         *string
	SecurityGroupID  string
	Primary          bool
	PrivateIPAddress *string
}

//UpdateNetworkInterfaceOptions options can be used to update a network interface card
type UpdateNetworkInterfaceOptions struct {
	ID              string
	ServerID        *string
	SecurityGroupID *string
}

//ListNetworkInterfacesOptions options can be used to list network interface cards
type ListNetworkInterfacesOptions struct {
	NetworkID        *string
	SubnetID         *string
	ServerID         *string
	SecurityGroupID  *string
	PrivateIPAddress *string
}

//NetworkInterfaceManager an interface providing an abstraction to manipulate network interface cards
type NetworkInterfaceManager interface {
	Create(options CreateNetworkInterfaceOptions) (*NetworkInterface, *CreateNetworkInterfaceError)
	Delete(id string) *DeleteNetworkInterfaceError
	Get(id string) (*NetworkInterface, *GetNetworkInterfaceError)
	List(options *ListNetworkInterfacesOptions) ([]NetworkInterface, *ListNetworkInterfacesError)
	Update(options UpdateNetworkInterfaceOptions) (*NetworkInterface, *UpdateNetworkInterfaceError)
}

//CreateNetworkInterfaceError create network interface error type
type CreateNetworkInterfaceError struct {
	ErrorStack
}

//NewCreateNetworkInterfaceError create a new CreateNetworkInterfaceError
func NewCreateNetworkInterfaceError(cause error, options CreateNetworkInterfaceOptions) *CreateNetworkInterfaceError {
	if cause == nil {
		return nil
	}
	return &CreateNetworkInterfaceError{*NewErrorStack(cause, "error creating network interface", options)}
}

//DeleteNetworkInterfaceError delete network interface error type
type DeleteNetworkInterfaceError struct {
	ErrorStack
}

//NewDeleteNetworkInterfaceError create a new DeleteNetworkInterfaceError
func NewDeleteNetworkInterfaceError(cause error, id string) *DeleteNetworkInterfaceError {
	if cause == nil {
		return nil
	}
	return &DeleteNetworkInterfaceError{*NewErrorStack(cause, "error deleting network interface", id)}
}

//GetNetworkInterfaceError get network interface error type
type GetNetworkInterfaceError struct {
	ErrorStack
}

//NewGetNetworkInterfaceError create a new GetNetworkInterfaceError
func NewGetNetworkInterfaceError(cause error, id string) *GetNetworkInterfaceError {
	if cause == nil {
		return nil
	}
	return &GetNetworkInterfaceError{*NewErrorStack(cause, "error getting network interface", id)}
}

//ListNetworkInterfacesError list network interface error type
type ListNetworkInterfacesError struct {
	ErrorStack
}

//NewListNetworkInterfacesError create a new ListNetworkInterfacesError
func NewListNetworkInterfacesError(cause error, options *ListNetworkInterfacesOptions) *ListNetworkInterfacesError {
	if cause == nil {
		return nil
	}
	return &ListNetworkInterfacesError{*NewErrorStack(cause, "error listing network interfaces", options)}
}

//UpdateNetworkInterfaceError update network interface error type
type UpdateNetworkInterfaceError struct {
	ErrorStack
}

//NewUpdateNetworkInterfaceError create a new UpdateNetworkInterfaceError
func NewUpdateNetworkInterfaceError(cause error, options UpdateNetworkInterfaceOptions) *UpdateNetworkInterfaceError {
	if cause == nil {
		return nil
	}
	return &UpdateNetworkInterfaceError{*NewErrorStack(cause, "error updating network interface", options)}
}
