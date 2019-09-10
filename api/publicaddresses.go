package api

//PublicIP represent a public ip address
type PublicIP struct {
	ID                 string
	Name               string
	Address            string
	NetworkInterfaceID string
	PrivateAddress     string
}

//AllocatePublicIPAddressOptions options that can be used to allocate a public ip address
type AllocatePublicIPAddressOptions struct {
	Name        string
	Address     string
	AddressPool string
}

//AssociatePublicIPOptions options that can be used to associate a public ip adress to a server
type AssociatePublicIPOptions struct {
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

//ListPublicIPAddressOptions options to be used to list public ip address
type ListPublicIPAddressOptions struct {
	ServerID *string
}

//PublicIPAddressManager an interface providing an abastraction to manipulate public ip addresses
type PublicIPAddressManager interface {
	ListAvailablePools() ([]PublicIPPool, error)
	List(options *ListPublicIPAddressOptions) ([]PublicIP, error)
	Create(options AllocatePublicIPAddressOptions) (*PublicIP, error)
	Associate(options AssociatePublicIPOptions) error
	Dissociate(publicIPId string) error
	Delete(publicIPId string) error
	Get(publicIPId string) (*PublicIP, error)
}
