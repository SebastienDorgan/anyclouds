package api

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

type NetworkInterfaceOptions struct {
	Name            string
	NetworkID       string
	SubnetID        string
	ServerID        string
	SecurityGroupID string
	Primary         bool
	IPAddress       *string
}

type NetworkInterfacesUpdateOptions struct {
	ID              string
	ServerID        *string
	SecurityGroupID *string
}

type NetworkInterfaceManager interface {
	Create(options *NetworkInterfaceOptions) (*NetworkInterface, error)
	Delete(id string) error
	Get(id string) (*NetworkInterface, error)
	List() ([]NetworkInterface, error)
	Update(options *NetworkInterfacesUpdateOptions) (*NetworkInterface, error)
}
