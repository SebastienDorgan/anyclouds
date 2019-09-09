package api

//PublicIP represent a public ip address
type PublicIP struct {
	ID        string
	Name      string
	Address   string
	ServerID  string
	SubnetID  string
	PrivateIP string
}

//PublicIPAllocationOptions options that can be used to allocate a public ip address
type PublicIPAllocationOptions struct {
	Name        string
	Address     string
	AddressPool string
}

//PublicIPAssociationOptions options that can be used to associate a public ip adress to a server
type PublicIPAssociationOptions struct {
	PublicIPId string
	ServerID   string
	//If the server is attached to more than one network this field must be provided
	SubnetID string
	//If the server has more than one private IP by subnet this field can be provided to control the private IP
	// correlated with the public ip
	PrivateIP string
}

//AddressRange a contigues range of address
type AddressRange struct {
	FirstAddress string
	LastAddress  string
}

//PublicIPPool a pool of public IP
type PublicIPPool struct {
	ID     string
	Ranges []AddressRange
}

//PublicIPAddressManager an interface providing an abastraction to manipulate public ip addresses
type PublicIPAddressManager interface {
	ListAvailablePools() ([]PublicIPPool, error)
	ListAllocated() ([]PublicIP, error)
	Allocate(options *PublicIPAllocationOptions) (*PublicIP, error)
	Associate(options *PublicIPAssociationOptions) error
	Dissociate(publicIPId string) error
	Release(publicIPId string) error
	Get(publicIPId string) (*PublicIP, error)
}
