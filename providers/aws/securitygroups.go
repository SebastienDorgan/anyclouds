package aws

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"github.com/SebastienDorgan/retry"
	"net"
	"time"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

//SecurityGroupManager defines security group management functions a anyclouds provider must provide
type SecurityGroupManager struct {
	Provider *Provider
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
	_ = enc.Encode(r)
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}

func idr(s string) *api.SecurityRule {
	bin, _ := base64.StdEncoding.DecodeString(s)
	buffer := bytes.NewBuffer(bin)
	dec := gob.NewDecoder(buffer)
	r := api.SecurityRule{}
	_ = dec.Decode(&r)
	return &r
}

func group(g *ec2.SecurityGroup) *api.SecurityGroup {
	var rules []api.SecurityRule
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

func (mgr *SecurityGroupManager) getGroup(id string) retry.Action {
	return func() (v interface{}, e error) {
		return mgr.Get(id)
	}
}

func noError() retry.Condition {
	return func(v interface{}, e error) bool {
		return e == nil
	}
}

//Create creates an security group
func (mgr *SecurityGroupManager) Create(options api.SecurityGroupOptions) (*api.SecurityGroup, api.CreateSecurityGroupError) {
	out, err := mgr.Provider.AWSServices.EC2Client.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		Description: aws.String(options.Description),
		GroupName:   aws.String(options.Name),
		VpcId:       aws.String(options.NetworkID),
	})
	if err != nil {
		return nil, api.NewCreateSecurityGroupError(err, options)
	}
	res := retry.With(mgr.getGroup(*out.GroupId)).Every(10 * time.Second).For(1 * time.Minute).Until(noError()).Go()
	if res.LastError != nil {
		return nil, api.NewCreateSecurityGroupError(err, options)
	}
	return res.LastValue.(*api.SecurityGroup), nil
}

//Delete deletes the security group identified by id
func (mgr *SecurityGroupManager) Delete(id string) api.DeleteSecurityGroupError {
	_, err := mgr.Provider.AWSServices.EC2Client.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(id),
	})
	return api.NewDeleteSecurityGroupError(err, id)
}

func groups(in []*ec2.SecurityGroup) []api.SecurityGroup {
	var result []api.SecurityGroup
	for _, g := range in {
		result = append(result, *group(g))
	}
	return result
}

//List list all security groups defined in the tenant
func (mgr *SecurityGroupManager) List() ([]api.SecurityGroup, api.ListSecurityGroupsError) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeSecurityGroups(nil)
	if err != nil {
		return nil, api.NewListSecurityGroupsError(err)
	}
	return groups(out.SecurityGroups), nil
}

//Get returns the security group identified by id
func (mgr *SecurityGroupManager) Get(id string) (*api.SecurityGroup, api.GetSecurityGroupError) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{
			aws.String(id),
		},
	})
	if err != nil {
		return nil, api.NewGetSecurityGroupError(err, id)
	}
	if len(out.SecurityGroups) == 0 {
		err = fmt.Errorf("security group %s not found", id)
		return nil, api.NewGetSecurityGroupError(err, id)
	}
	if len(out.SecurityGroups) > 1 {
		err = fmt.Errorf("multiple security groups identified by %s found", id)
		return nil, api.NewGetSecurityGroupError(err, id)
	}
	return group(out.SecurityGroups[0]), nil
}

//Attach a server to a security group
func (mgr *SecurityGroupManager) Attach(options api.AttachSecurityGroupOptions) api.AttachSecurityGroupError {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.instance-id"),
				Values: []*string{aws.String(options.ServerID)},
			},
			{
				Name:   aws.String("subnet-id"),
				Values: []*string{aws.String(options.SubnetID)},
			},
		},
	})
	if err != nil {
		return api.NewAttachSecurityGroupError(err, options)
	}
	if out.NetworkInterfaces == nil {
		err = errors.Errorf("no network interface found for subnet %s of network %s on server %s",
			options.SubnetID,
			options.NetworkID,
			options.ServerID,
		)
		return api.NewAttachSecurityGroupError(err, options)
	}
	for _, ni := range out.NetworkInterfaces {
		if options.IPAddress != nil && *ni.PrivateIpAddress != *options.IPAddress {
			continue
		}
		_, err = mgr.Provider.AWSServices.EC2Client.ModifyNetworkInterfaceAttribute(&ec2.ModifyNetworkInterfaceAttributeInput{
			Groups:             []*string{aws.String(options.SecurityGroupID)},
			NetworkInterfaceId: ni.NetworkInterfaceId,
		})
		if err != nil {
			return api.NewAttachSecurityGroupError(err, options)
		}
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

func ipRange(o *api.AddSecurityRuleOptions) *ec2.IpRange {
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

func ipv6Range(o *api.AddSecurityRuleOptions) *ec2.Ipv6Range {
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

func ipPermission(options *api.AddSecurityRuleOptions) (*ec2.IpPermission, error) {
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

//AddSecurityRule adds a security rule to an security group
func (mgr *SecurityGroupManager) AddSecurityRule(options api.AddSecurityRuleOptions) (*api.SecurityRule, api.AddSecurityRuleError) {
	p, err := ipPermission(&options)
	if err != nil {
		return nil, api.NewAddSecurityRuleError(err, options)
	}

	if options.Direction == api.RuleDirectionIngress {
		_, err = mgr.Provider.AWSServices.EC2Client.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: &options.SecurityGroupID,
			IpPermissions: []*ec2.IpPermission{
				p,
			},
		})
		if err != nil {
			return nil, api.NewAddSecurityRuleError(err, options)
		}

	} else if options.Direction == api.RuleDirectionEgress {
		_, err = mgr.Provider.AWSServices.EC2Client.AuthorizeSecurityGroupEgress(&ec2.AuthorizeSecurityGroupEgressInput{
			GroupId: &options.SecurityGroupID,
			IpPermissions: []*ec2.IpPermission{
				p,
			},
		})
		if err != nil {
			return nil, api.NewAddSecurityRuleError(err, options)
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

//RemoveSecurityRule deletes a security rule from an security group
func (mgr *SecurityGroupManager) RemoveSecurityRule(groupID, ruleID string) api.RemoveSecurityRuleError {
	rule := idr(ruleID)
	ipPerm, err := ipPermissionFromRule(rule)
	if err != nil {
		return api.NewRemoveSecurityRuleError(err, groupID, ruleID)
	}
	if rule.Direction == api.RuleDirectionIngress {
		_, err := mgr.Provider.AWSServices.EC2Client.RevokeSecurityGroupIngress(&ec2.RevokeSecurityGroupIngressInput{
			GroupId: &rule.SecurityGroupID,
			IpPermissions: []*ec2.IpPermission{
				ipPerm,
			},
		})
		if err != nil {
			return api.NewRemoveSecurityRuleError(err, groupID, ruleID)
		}
	}
	if rule.Direction == api.RuleDirectionEgress {
		_, err := mgr.Provider.AWSServices.EC2Client.RevokeSecurityGroupEgress(&ec2.RevokeSecurityGroupEgressInput{
			GroupId: &rule.SecurityGroupID,
			IpPermissions: []*ec2.IpPermission{
				ipPerm,
			},
		})
		if err != nil {
			return api.NewRemoveSecurityRuleError(err, groupID, ruleID)
		}
	}
	return nil

}
