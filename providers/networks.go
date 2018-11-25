package providers

//Network defines network properties
type Network struct {
	//unique identifier of the network
	ID string
	//name of the network
	Name string
}

//NetworkOptions defines options to use when creating a network
type NetworkOptions struct {
	//name of the network
	Name string
}

//IPVersion ip version enum
type IPVersion int

const (
	//IPVersion6 IPv6
	IPVersion6 IPVersion = 6
	//IPVersion4 IPv4
	IPVersion4 IPVersion = 4
)

//SubnetOptions defines options to use when creating a sub network
type SubnetOptions struct {
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
	//unnique identifier of the sub network
	ID string
	//identifier of the parent network (i.e. Network.ID)
	NetworkID string
	//name of the sub network
	Name string
	//CIDR of the sub network
	CIDR string
	//IP Version
	IPVersion IPVersion
}

//NetworkManager defines networking functions a anyclouds provider must provide
type NetworkManager interface {
	CreateNetwork(options *NetworkOptions) (*Network, error)
	DeleteNetwork(id string) error
	ListNetworks(filter *ResourceFilter) ([]Network, error)
	GetNetwork(id string) (*Network, error)

	CreateSubnet(options *SubnetOptions) (*Subnet, error)
	DeleteSubnet(id string) error
	ListSubnets(networkID string, filter *ResourceFilter) ([]Subnet, error)
	GetSubnet(id string) (*Subnet, error)
}
