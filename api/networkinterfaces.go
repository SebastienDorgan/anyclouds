package api

//NetworkInterface represents an network interface card
type NetworkInterface struct {
	ID         string
	Name       string
	MacAddress string
	NetworkID  string
	SubnetID   string
	ServerID   string
	IPAddress  string
	Primary    bool
}

//NetworkInterfaceOptions options that can be used to create a network interface card
type NetworkInterfaceOptions struct {
	Name            string
	NetworkID       string
	SubnetID        string
	ServerID        string
	SecurityGroupID string
	Primary         bool
	IPAddress       *string
}

//NetworkInterfacesUpdateOptions options can be used to update a network interface card
type NetworkInterfacesUpdateOptions struct {
	ID              string
	ServerID        *string
	SecurityGroupID *string
}

//NetworkInterfaceManager an interface providing an abastraction to manipulate network interface cards
type NetworkInterfaceManager interface {
	Create(options *NetworkInterfaceOptions) (*NetworkInterface, error)
	Delete(id string) error
	Get(id string) (*NetworkInterface, error)
	List() ([]NetworkInterface, error)
	Update(options *NetworkInterfacesUpdateOptions) (*NetworkInterface, error)
}
