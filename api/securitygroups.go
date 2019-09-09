package api

//Protocol valid options are empty string (any protocol), tcp or upd
type Protocol string

const (
	//ProtocolAny const
	ProtocolAny Protocol = ""
	//ProtocolTCP const
	ProtocolTCP Protocol = "tcp"
	//ProtocolUDP const
	ProtocolUDP Protocol = "udp"
	//ProtocolICMP const
	ProtocolICMP Protocol = "icmp"
)

//RuleDirection defines rule direction. Valid option are ingress or egress
type RuleDirection string

const (
	//RuleDirectionIngress Ingress direction
	//From outside the network to inside
	RuleDirectionIngress RuleDirection = "ingress"
	//RuleDirectionEgress Egress direction
	//From inside the network to outside
	RuleDirectionEgress RuleDirection = "egress"
)

//PortRange defines a port range
type PortRange struct {
	// The minimum port number in the range that is matched by the security group rule.
	// If the protocol is TCP, UDP this value must be less than or equal to the port_range_max attribute value.
	// If the protocol is ICMP, this value must be an ICMP type
	From int

	// The maximum port number in the range that is matched by the security group rule.
	// If the protocol is TCP, UDP this value must be greater than or equal to the port_range_min attribute value.
	// If the protocol is ICMP, this value must be an ICMP code.
	To int
}

//SecurityRule define a security rule
type SecurityRule struct {
	ID              string
	SecurityGroupID string
	Direction       RuleDirection
	PortRange       PortRange
	Protocol        Protocol
	CIDR            string
	Description     string
}

//SecurityRuleOptions define security rule options when adding rule to a security group
type SecurityRuleOptions struct {
	SecurityGroupID string
	Direction       RuleDirection
	PortRange       PortRange
	Protocol        Protocol
	CIDR            string
	Description     string
}

//SecurityGroup defines security groups properties
type SecurityGroup struct {
	ID        string
	Name      string
	NetworkID string
	Rules     []SecurityRule
}

//SecurityGroupOptions defines security groups properties
type SecurityGroupOptions struct {
	Name        string
	Description string
	NetworkID   string
}

//SecurityGroupAttachmentOptions options that can be used to attach a security group to a network inteface
type SecurityGroupAttachmentOptions struct {
	SecurityGroupID string
	ServerID        string
	NetworkID       string
	SubnetID        string
	IPAddress       *string
}

//SecurityGroupManager defines security group management functions a anyclouds provider must provide
type SecurityGroupManager interface {
	//Create a security group
	Create(options *SecurityGroupOptions) (*SecurityGroup, error)
	//Delete a security group
	Delete(id string) error
	//List security groups
	List() ([]SecurityGroup, error)
	//Get security group
	Get(id string) (*SecurityGroup, error)
	//Attach a security group to a server
	Attach(options *SecurityGroupAttachmentOptions) error
	//Add a rule to a security group
	AddRule(options *SecurityRuleOptions) (*SecurityRule, error)
	//Delete a rule
	DeleteRule(groupID, ruleID string) error
}
