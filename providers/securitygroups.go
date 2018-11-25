package providers

//Protocol TCP/UDP enum
type Protocol string

const (
	//ProtocolTCP const
	ProtocolTCP Protocol = "TCP"
	//ProtocolUDP const
	ProtocolUDP Protocol = "UDP"
)

//PortRange defines a port range
type PortRange struct {
	From int
	To   int
}

//SecurityRule define a security rule
type SecurityRule struct {
	ID              string
	SecurityGroupID string
	CIDR            string
	PortRange       PortRange
	Protocol        Protocol
}

//SecurityRuleOptions define security rule options when adding rule to a security group
type SecurityRuleOptions struct {
	SecurityGroupID string
	CIDR            string
	PortRange       PortRange
	Protocol        Protocol
}

//SecurityGroup defines security groups properties
type SecurityGroup struct {
	ID          string
	Name        string
	Description string
	Rules       []SecurityRule
}

//SecurityGroupOptions defines security groups properties
type SecurityGroupOptions struct {
	Name        string
	Description string
}

//SecurityGroupManager defines security group management functions a anyclouds provider must provide
type SecurityGroupManager interface {
	Create(options *SecurityGroupOptions) (*SecurityGroup, error)
	Delete(id string) error
	List(filter *ResourceFilter) ([]SecurityGroup, error)
	Get(id string) (*SecurityGroup, error)
	AddRule(rule *SecurityRuleOptions) (*SecurityRule, error)
	DeleteRule(ruleID string) error
}
