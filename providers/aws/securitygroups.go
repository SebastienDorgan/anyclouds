package aws

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

//SecurityGroupManager defines security group management functions a anyclouds provider must provide
type SecurityGroupManager struct {
	AWS *Provider
}

func i64toi(i64 *int64) int {
	if i64 == nil {
		return 0
	}
	return int(*i64)
}

func rid(r api.SecurityRule) string {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	enc.Encode(r)
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}

func idr(s string) *api.SecurityRule {
	bin, _ := base64.StdEncoding.DecodeString(s)
	buffer := bytes.NewBuffer(bin)
	dec := gob.NewDecoder(buffer)
	r := api.SecurityRule{}
	dec.Decode(&r)
	return &r
}

func group(g *ec2.SecurityGroup) *api.SecurityGroup {
	rules := []api.SecurityRule{}
	for _, pi := range g.IpPermissions {
		if len(pi.IpRanges) == 0 {
			continue
		}
		r := api.SecurityRule{
			SecurityGroupID: *g.GroupId,
			Direction:       api.RuleDirectionIngress,
			PortRange: api.PortRange{
				From: i64toi(pi.FromPort),
				To:   i64toi(pi.ToPort),
			},
			Protocol:    api.Protocol(*pi.IpProtocol),
			CIDR:        *pi.IpRanges[0].CidrIp,
			Description: *pi.IpRanges[0].Description,
		}
		r.ID = rid(r)
		rules = append(rules, r)
	}
	return &api.SecurityGroup{
		ID:        *g.GroupId,
		Name:      *g.GroupName,
		NetworkID: *g.VpcId,
		Rules:     rules,
	}
}

//Create creates an security group
func (mgr *SecurityGroupManager) Create(options *api.SecurityGroupOptions) (*api.SecurityGroup, error) {
	out, err := mgr.AWS.EC2Client.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		Description: aws.String(options.Description),
		GroupName:   aws.String(options.Name),
		VpcId:       aws.String(options.NetworkID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error creating security groups")
	}
	g, err := mgr.Get(*out.GroupId)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating security groups")
	}
	return g, err
}

//Delete deletes the security group identified by id
func (mgr *SecurityGroupManager) Delete(id string) error {
	_, err := mgr.AWS.EC2Client.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(id),
	})
	return errors.Wrap(err, "Error deleting security groups")
}

//List list all security groups defined in the tenant
func (mgr *SecurityGroupManager) List() ([]api.SecurityGroup, error) {
	out, err := mgr.AWS.EC2Client.DescribeSecurityGroups(nil)
	if err != nil {
		return nil, errors.Wrap(err, "Error listing security groups")
	}
	result := []api.SecurityGroup{}
	for _, g := range out.SecurityGroups {
		result = append(result, *group(g))
	}
	return result, nil
}

//ListByServer list security groups by server
func (mgr *SecurityGroupManager) ListByServer(serverID string) ([]api.SecurityGroup, error) {
	out, err := mgr.AWS.EC2Client.DescribeInstanceAttribute(&ec2.DescribeInstanceAttributeInput{
		InstanceId: &serverID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error listing security groups of server")
	}
	result := []api.SecurityGroup{}
	for _, g := range out.Groups {
		group, err := mgr.Get(*g.GroupId)
		if err != nil {
			continue
		}
		result = append(result, *group)
	}
	return result, nil
}

//Get returns the security group identified by id
func (mgr *SecurityGroupManager) Get(id string) (*api.SecurityGroup, error) {
	out, err := mgr.AWS.EC2Client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{
			aws.String(id),
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error listing security groups")
	}
	if len(out.SecurityGroups) == 0 {
		return nil, errors.Wrap(fmt.Errorf("Security group with ID %s not found", id), "Error listing security groups")
	}
	if len(out.SecurityGroups) > 1 {
		return nil, errors.Wrap(fmt.Errorf("Multiple groups with ID %s not found", id), "Error listing security groups")
	}
	return group(out.SecurityGroups[0]), nil
}

//AddServer a server to a security group
func (mgr *SecurityGroupManager) AddServer(id string, serverID string) error {
	groups, err := mgr.ListByServer(serverID)
	if err != nil {
		return errors.Wrap(err, "Error adding security groups to server")
	}
	groupIds := []*string{}
	for _, g := range groups {
		groupIds = append(groupIds, &g.ID)
	}
	groupIds = append(groupIds, &id)

	_, err = mgr.AWS.EC2Client.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
		InstanceId: &id,
		Groups:     groupIds,
	})
	if err != nil {
		return errors.Wrap(err, "Error adding security groups to server")
	}
	return nil
}

//RemoveServer remove a server from a security group
func (mgr *SecurityGroupManager) RemoveServer(id string, serverID string) error {
	groups, err := mgr.ListByServer(serverID)
	if err != nil {
		return errors.Wrap(err, "Error removing security groups from server")
	}
	groupIds := []*string{}
	for _, g := range groups {
		if serverID != g.ID {
			groupIds = append(groupIds, &g.ID)
		}
	}
	_, err = mgr.AWS.EC2Client.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
		InstanceId: &id,
		Groups:     groupIds,
	})
	if err != nil {
		return errors.Wrap(err, "Error removing security groups from server")
	}
	return nil
}

func isV6(CIDR string) (bool, error) {
	if len(CIDR) == 0 {
		return false, nil
	}
	_, n, e := net.ParseCIDR(CIDR)
	if e != nil {
		return false, e
	}
	if s, _ := n.Mask.Size(); s == 16 {
		return true, nil
	}
	return false, nil

}

func ipRange(o *api.SecurityRuleOptions) *ec2.IpRange {
	cidr := o.CIDR
	if len(cidr) == 0 {
		cidr = "0.0.0.0/0"
	}
	return &ec2.IpRange{
		CidrIp:      &cidr,
		Description: &o.Description,
	}
}

func ipRangeFromRule(o *api.SecurityRule) *ec2.IpRange {
	cidr := o.CIDR
	if len(cidr) == 0 {
		cidr = "0.0.0.0/0"
	}
	return &ec2.IpRange{
		CidrIp:      &cidr,
		Description: &o.Description,
	}
}

func ipv6Range(o *api.SecurityRuleOptions) *ec2.Ipv6Range {
	return &ec2.Ipv6Range{
		CidrIpv6:    &o.CIDR,
		Description: &o.Description,
	}
}

func ipv6RangeFromRule(o *api.SecurityRule) *ec2.Ipv6Range {
	return &ec2.Ipv6Range{
		CidrIpv6:    &o.CIDR,
		Description: &o.Description,
	}
}

func ipPermission(options *api.SecurityRuleOptions) (*ec2.IpPermission, error) {
	p := &ec2.IpPermission{
		IpProtocol: aws.String(string(options.Protocol)),
		FromPort:   aws.Int64(int64(options.PortRange.From)),
		ToPort:     aws.Int64(int64(options.PortRange.To)),
	}

	v6, err := isV6(options.CIDR)
	if err != nil {
		return nil, err
	}
	if v6 {
		p.Ipv6Ranges = []*ec2.Ipv6Range{
			ipv6Range(options),
		}
	} else {
		p.IpRanges = []*ec2.IpRange{
			ipRange(options),
		}
	}

	return p, nil
}

func ipPermissionFromRule(r *api.SecurityRule) (*ec2.IpPermission, error) {
	p := &ec2.IpPermission{
		IpProtocol: aws.String(string(r.Protocol)),
		FromPort:   aws.Int64(int64(r.PortRange.From)),
		ToPort:     aws.Int64(int64(r.PortRange.To)),
	}

	v6, err := isV6(r.CIDR)
	if err != nil {
		return nil, err
	}
	if v6 {
		p.Ipv6Ranges = []*ec2.Ipv6Range{
			ipv6RangeFromRule(r),
		}
	} else {
		p.IpRanges = []*ec2.IpRange{
			ipRangeFromRule(r),
		}
	}

	return p, nil
}

//AddRule adds a security rule to an security group
func (mgr *SecurityGroupManager) AddRule(options *api.SecurityRuleOptions) (*api.SecurityRule, error) {
	p, err := ipPermission(options)
	if err != nil {
		return nil, errors.Wrap(err, "Error adding rule to security groups")
	}

	if options.Direction == api.RuleDirectionIngress {
		_, err = mgr.AWS.EC2Client.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: &options.SecurityGroupID,
			IpPermissions: []*ec2.IpPermission{
				p,
			},
		})
		if err != nil {
			return nil, errors.Wrap(err, "Error adding rule to security group")
		}

	} else if options.Direction == api.RuleDirectionEgress {
		_, err = mgr.AWS.EC2Client.AuthorizeSecurityGroupEgress(&ec2.AuthorizeSecurityGroupEgressInput{
			GroupId: &options.SecurityGroupID,
			IpPermissions: []*ec2.IpPermission{
				p,
			},
		})
		if err != nil {
			return nil, errors.Wrap(err, "Error adding rule to security group")
		}
	}
	rule := api.SecurityRule{
		SecurityGroupID: options.SecurityGroupID,
		Direction:       options.Direction,
		PortRange:       options.PortRange,
		Protocol:        options.Protocol,
		Description:     options.Description,
	}
	rule.ID = rid(rule)
	return &rule, nil
}

//DeleteRule deletes a secuity rule from an security group
func (mgr *SecurityGroupManager) DeleteRule(ruleID string) error {
	rule := idr(ruleID)
	ipPerm, err := ipPermissionFromRule(rule)
	if err != nil {
		return errors.Wrap(err, "Error deleting rule from security group")
	}
	if rule.Direction == api.RuleDirectionIngress {
		_, err := mgr.AWS.EC2Client.RevokeSecurityGroupIngress(&ec2.RevokeSecurityGroupIngressInput{
			GroupId: &rule.SecurityGroupID,
			IpPermissions: []*ec2.IpPermission{
				ipPerm,
			},
		})
		if err != nil {
			return errors.Wrap(err, "Error deleting rule from security group")
		}
	}
	if rule.Direction == api.RuleDirectionEgress {
		_, err := mgr.AWS.EC2Client.RevokeSecurityGroupEgress(&ec2.RevokeSecurityGroupEgressInput{
			GroupId: &rule.SecurityGroupID,
			IpPermissions: []*ec2.IpPermission{
				ipPerm,
			},
		})
		if err != nil {
			return errors.Wrap(err, "Error deleting rule from security group")
		}
	}
	return nil

}
