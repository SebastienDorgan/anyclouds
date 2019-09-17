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

//AddSecurityRuleOptions define security rule options when adding rule to a security group
type AddSecurityRuleOptions struct {
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

//AttachSecurityGroupOptions options that can be used to attach a security group to a network interface
type AttachSecurityGroupOptions struct {
	SecurityGroupID string
	ServerID        string
	NetworkID       string
	SubnetID        string
	IPAddress       *string
}

//SecurityGroupManager defines security group management functions a anyclouds provider must provide
type SecurityGroupManager interface {
	//Create a security group
	Create(options SecurityGroupOptions) (*SecurityGroup, *CreateSecurityGroupError)
	//Delete a security group
	Delete(id string) *DeleteSecurityGroupError
	//List security groups
	List() ([]SecurityGroup, *ListSecurityGroupsError)
	//Get security group
	Get(id string) (*SecurityGroup, *GetSecurityGroupError)
	//Attach a security group to a server
	Attach(options AttachSecurityGroupOptions) *AttachSecurityGroupError
	//Add a rule to a security group
	AddSecurityRule(options AddSecurityRuleOptions) (*SecurityRule, *AddSecurityRuleError)
	//Delete a rule
	RemoveSecurityRule(groupID, ruleID string) *RemoveSecurityRuleError
}

//CreateSecurityGroupError create security group error type
type CreateSecurityGroupError struct {
	ErrorStack
}

//NewCreateSecurityGroupError creates a new CreateSecurityGroupError
func NewCreateSecurityGroupError(cause error, options SecurityGroupOptions) *CreateSecurityGroupError {
	return &CreateSecurityGroupError{ErrorStack: *NewErrorStack(cause, "error creating security group", options)}
}

//DeleteSecurityGroupError delete security group error type
type DeleteSecurityGroupError struct {
	ErrorStack
}

//NewDeleteSecurityGroupError creates a new DeleteSecurityGroupError
func NewDeleteSecurityGroupError(cause error, id string) *DeleteSecurityGroupError {
	return &DeleteSecurityGroupError{ErrorStack: *NewErrorStack(cause, "error deleting security group", id)}
}

//ListSecurityGroupsError list security group error type
type ListSecurityGroupsError struct {
	ErrorStack
}

//NewListSecurityGroupsError creates a new ListSecurityGroupsError
func NewListSecurityGroupsError(cause error) *ListSecurityGroupsError {
	return &ListSecurityGroupsError{ErrorStack: *NewErrorStack(cause, "error listing security groups")}
}

//GetSecurityGroupError get security group error type
type GetSecurityGroupError struct {
	ErrorStack
}

//NewGetSecurityGroupError creates a new GetSecurityGroupError
func NewGetSecurityGroupError(cause error, id string) *GetSecurityGroupError {
	return &GetSecurityGroupError{ErrorStack: *NewErrorStack(cause, "error getting security group", id)}
}

//AttachSecurityGroupError attach security group error type
type AttachSecurityGroupError struct {
	ErrorStack
}

//NewAttachSecurityGroupError creates a new AttachSecurityGroupError
func NewAttachSecurityGroupError(cause error, options AttachSecurityGroupOptions) *AttachSecurityGroupError {
	return &AttachSecurityGroupError{ErrorStack: *NewErrorStack(cause, "error attaching security group", options)}
}

////DetachSecurityGroupError detach security group error type
//type DetachSecurityGroupError struct {
//	ErrorStack
//}
//
////NewDetachSecurityGroupError creates a new DetachSecurityGroupError
//func NewDetachSecurityGroupError(cause error, id string) *DetachSecurityGroupError {
//	return &DetachSecurityGroupError{ErrorStack: *NewErrorStack(cause, "error detaching security group", id)}
//}

//AddSecurityRuleError add security rule error type
type AddSecurityRuleError struct {
	ErrorStack
}

//NewAddSecurityRuleError creates a new AddSecurityRuleError
func NewAddSecurityRuleError(cause error, options AddSecurityRuleOptions) *AddSecurityRuleError {
	return &AddSecurityRuleError{ErrorStack: *NewErrorStack(cause, "error adding security rule", options)}
}

//RemoveSecurityRuleError remove security rule error type
type RemoveSecurityRuleError struct {
	ErrorStack
}

//NewRemoveSecurityRuleError creates a new RemoveSecurityRuleError
func NewRemoveSecurityRuleError(cause error, id string, ruleID string) *RemoveSecurityRuleError {
	return &RemoveSecurityRuleError{ErrorStack: *NewErrorStack(cause, "error removing security rule", id, ruleID)}
}
