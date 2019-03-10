package aws

import (
	"fmt"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

//SecurityGroupManager defines security group management functions a anyclouds provider must provide
type SecurityGroupManager struct {
	AWS *Provider
}

func group(g *ec2.SecurityGroup) *api.SecurityGroup {
	return &api.SecurityGroup{
		Description: *g.Description,
		ID:          *g.GroupId,
		Name:        *g.GroupName,
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
	return nil, nil
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
	return nil
}

//RemoveServer remove a server from a security group
func (mgr *SecurityGroupManager) RemoveServer(id string, serverID string) error {
	return nil
}

//AddRule adds a security rule to an security group
func (mgr *SecurityGroupManager) AddRule(id string, options *api.SecurityRuleOptions) (*api.SecurityRule, error) {
	return nil, nil
}

//DeleteRule deletes a secuity rule from an security group
func (mgr *SecurityGroupManager) DeleteRule(ruleID string) error {
	return nil
}
