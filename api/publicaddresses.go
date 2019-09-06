package api

type PublicIPAllocationOptions struct {
	Name        string
	Address     string
	AddressPool string
}

type PublicIPAssociationOptions struct {
	PublicIPId string
	ServerID   string
	//If the server is attached to more than one network this field must be provided
	SubnetID string
	//If the server has more than one private IP by subnet this field can be provided to control the private IP
	// correlated with the public ip
	PrivateIP string
}
type AddressRange struct {
	FirstAddress string
	LastAddress  string
}
type PublicIPPool struct {
	ID     string
	Ranges []AddressRange
}
type PublicIP struct {
	ID        string
	Name      string
	Address   string
	ServerID  string
	SubnetID  string
	PrivateIP string
}
type PublicIPAddressManager interface {
	ListAvailablePools() ([]PublicIPPool, error)
	ListAllocated() ([]PublicIP, error)
	Allocate(options *PublicIPAllocationOptions) (*PublicIP, error)
	Associate(options *PublicIPAssociationOptions) error
	Dissociate(publicIpID string) error
	Release(publicIPId string) error
	Get(publicIPId string) (*PublicIP, error)
}
