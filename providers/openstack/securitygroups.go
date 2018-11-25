package openstack

import (
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/secgroups"

	"github.com/SebastienDorgan/anyclouds/providers"

	"github.com/pkg/errors"
)

//SecurityGroupManager defines security group management functions a anyclouds provider must provide
type SecurityGroupManager struct {
	OpenStack *Provider
}

func convertRule(r *secgroups.Rule) *providers.SecurityRule {
	return &providers.SecurityRule{
		ID:       r.ID,
		Protocol: providers.Protocol(r.IPProtocol),
		CIDR:     r.IPRange.CIDR,
		PortRange: providers.PortRange{
			From: r.FromPort,
			To:   r.ToPort,
		},
	}
}

func convert(group *secgroups.SecurityGroup) *providers.SecurityGroup {
	res := providers.SecurityGroup{
		Name:        group.Name,
		Description: group.Description,
		ID:          group.ID,
	}
	for _, r := range group.Rules {
		rule := convertRule(&r)
		res.Rules = append(res.Rules, *rule)
	}
	return &res
}

//Create creates an openstack security group
func (sec *SecurityGroupManager) Create(options *providers.SecurityGroupOptions) (*providers.SecurityGroup, error) {

	createOpts := secgroups.CreateOpts{
		Name:        options.Name,
		Description: options.Description,
	}

	group, err := secgroups.Create(sec.OpenStack.Compute, createOpts).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error creating security group")
	}

	return convert(group), nil
}

//Delete deletes the Openstack security group identified by id
func (sec *SecurityGroupManager) Delete(id string) error {
	err := secgroups.Delete(sec.OpenStack.Compute, id).ExtractErr()
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error deleting security group")
	}
	return nil
}

//List list all Openstack security groups defined in the tenant
func (sec *SecurityGroupManager) List(filter *providers.ResourceFilter) ([]providers.SecurityGroup, error) {
	page, err := secgroups.List(sec.OpenStack.Compute).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing security group")
	}
	groups := []providers.SecurityGroup{}
	l, err := secgroups.ExtractSecurityGroups(page)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing security group")
	}
	for _, g := range l {
		if len(filter.Name) > 0 {
			if g.Name != filter.Name {
				continue
			}
		}
		if len(filter.ID) > 0 {
			if g.ID != filter.ID {
				continue
			}
		}
		group := convert(&g)
		groups = append(groups, *group)
	}
	return groups, nil
}

//Get returns the Openstack security group identified by id
func (sec *SecurityGroupManager) Get(id string) (*providers.SecurityGroup, error) {
	sg, err := secgroups.Get(sec.OpenStack.Compute, id).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting security group")
	}
	return convert(sg), nil
}

//AddRule adds a security rule to an OpenStack security group
func (sec *SecurityGroupManager) AddRule(rule *providers.SecurityRuleOptions) (*providers.SecurityRule, error) {
	r, err := secgroups.CreateRule(sec.OpenStack.Compute, secgroups.CreateRuleOpts{
		CIDR:          rule.CIDR,
		ParentGroupID: rule.SecurityGroupID,
		FromPort:      rule.PortRange.From,
		ToPort:        rule.PortRange.To,
		IPProtocol:    string(rule.Protocol),
	}).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error adding rule to security group")
	}

	return convertRule(r), nil
}

//DeleteRule deletes a secuity rule from an OpenStack security group
func (sec *SecurityGroupManager) DeleteRule(ruleID string) error {
	err := secgroups.DeleteRule(sec.OpenStack.Compute, ruleID).ExtractErr()
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error deleting security group")
	}
	return nil
}
