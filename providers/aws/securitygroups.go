package aws

import (
	"fmt"
	"strconv"
	"strings"

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
	return fmt.Sprintf("%s/%s/%s/%d/%d", r.SecurityGroupID, r.Direction, r.Protocol, r.PortRange.From, r.PortRange.To)
}

func idr(s string) *api.SecurityRule {
	tokens := strings.Split(s, "/")
	from, err := strconv.ParseInt(tokens[3], 10, 32)
	if err != nil {
		from = 0
	}
	to, err := strconv.ParseInt(tokens[4], 10, 32)
	if err != nil {
		to = 0
	}
	return &api.SecurityRule{
		SecurityGroupID: tokens[0],
		Direction:       api.RuleDirection(tokens[1]),
		Protocol:        api.Protocol(tokens[2]),
		PortRange: api.PortRange{
			From: int(from),
			To:   int(to),
		},
	}

}

func group(g *ec2.SecurityGroup) *api.SecurityGroup {
	rules := []api.SecurityRule{}
	for _, pi := range g.IpPermissions {
		r := api.SecurityRule{
			Direction: api.RuleDirectionIngress,
			PortRange: api.PortRange{
				From: i64toi(pi.FromPort),
				To:   i64toi(pi.ToPort),
			},
			Protocol: api.Protocol(*pi.IpProtocol),
		}
		r.ID = rid(r)
		rules = append(rules, r)
	}
	return &api.SecurityGroup{
		ID:    *g.GroupId,
		Name:  *g.GroupName,
		Rules: rules,
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

//AddRule adds a security rule to an security group
func (mgr *SecurityGroupManager) AddRule(id string, options *api.SecurityRuleOptions) (*api.SecurityRule, error) {
	out, err := mgr.AWS.EC2Client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{
			aws.String(id),
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error adding rule to security groups")
	}
	if len(out.SecurityGroups) == 0 {
		return nil, errors.Wrap(fmt.Errorf("Security group with ID %s not found", id), "Error adding rule to security groups")
	}
	if len(out.SecurityGroups) > 1 {
		return nil, errors.Wrap(fmt.Errorf("Multiple groups with ID %s not found", id), "Error adding rule to security groups")
	}
	group := out.SecurityGroups[0]

	if options.Direction == api.RuleDirectionIngress {

		inRules := append(group.IpPermissions, &ec2.IpPermission{
			IpProtocol: aws.String(string(options.Protocol)),
			FromPort:   aws.Int64(int64(options.PortRange.From)),
			ToPort:     aws.Int64(int64(options.PortRange.To)),
		})
		_, err := mgr.AWS.EC2Client.UpdateSecurityGroupRuleDescriptionsIngress(&ec2.UpdateSecurityGroupRuleDescriptionsIngressInput{
			GroupId:       &id,
			IpPermissions: inRules,
		})
		if err != nil {
			return nil, errors.Wrap(err, "Error adding rule to security group")
		}

	} else if options.Direction == api.RuleDirectionEgress {
		inRules := append(group.IpPermissions, &ec2.IpPermission{
			IpProtocol: aws.String(string(options.Protocol)),
			FromPort:   aws.Int64(int64(options.PortRange.From)),
			ToPort:     aws.Int64(int64(options.PortRange.To)),
		})
		_, err := mgr.AWS.EC2Client.UpdateSecurityGroupRuleDescriptionsEgress(&ec2.UpdateSecurityGroupRuleDescriptionsEgressInput{
			GroupId:       &id,
			IpPermissions: inRules,
		})
		if err != nil {
			return nil, errors.Wrap(err, "Error adding rule to security group")
		}
	}
	rule := api.SecurityRule{
		SecurityGroupID: id,
		Direction:       options.Direction,
		PortRange:       options.PortRange,
		Protocol:        options.Protocol,
	}
	rule.ID = rid(rule)
	return &rule, nil
}

//DeleteRule deletes a secuity rule from an security group
func (mgr *SecurityGroupManager) DeleteRule(ruleID string) error {
	rule := idr(ruleID)
	if rule.Direction == api.RuleDirectionIngress {
		_, err := mgr.AWS.EC2Client.RevokeSecurityGroupIngress(&ec2.RevokeSecurityGroupIngressInput{
			GroupId: &rule.SecurityGroupID,
			IpPermissions: []*ec2.IpPermission{
				&ec2.IpPermission{
					IpProtocol: aws.String(string(rule.Protocol)),
					FromPort:   aws.Int64(int64(rule.PortRange.From)),
					ToPort:     aws.Int64(int64(rule.PortRange.To)),
				},
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
				&ec2.IpPermission{
					IpProtocol: aws.String(string(rule.Protocol)),
					FromPort:   aws.Int64(int64(rule.PortRange.From)),
					ToPort:     aws.Int64(int64(rule.PortRange.To)),
				},
			},
		})
		if err != nil {
			return errors.Wrap(err, "Error deleting rule from security group")
		}
	}
	return nil

}
