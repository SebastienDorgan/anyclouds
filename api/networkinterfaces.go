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

//UpdateNetworkInterfacesOptions options can be used to update a network interface card
type UpdateNetworkInterfacesOptions struct {
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

//NetworkInterfaceManager an interface providing an abastraction to manipulate network interface cards
type NetworkInterfaceManager interface {
	Create(options CreateNetworkInterfaceOptions) (*NetworkInterface, error)
	Delete(id string) error
	Get(id string) (*NetworkInterface, error)
	List(options *ListNetworkInterfacesOptions) ([]NetworkInterface, error)
	Update(options UpdateNetworkInterfacesOptions) (*NetworkInterface, error)
}
