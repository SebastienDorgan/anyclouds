package openstack

import (
	"fmt"
	"strings"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/secgroups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
	"github.com/pkg/errors"
)

//SecurityGroupManager defines security group management functions a anyclouds provider must provide
type SecurityGroupManager struct {
	OpenStack *Provider
}

func group(g *groups.SecGroup) *api.SecurityGroup {
	tokens := strings.Split("/", g.Name)
	return &api.SecurityGroup{
		Name:        tokens[0],
		NetworkID:   tokens[1],
		ID:          g.ID,
		Description: g.Description,
	}
}

func chechGroupName(name string) error {
	if strings.Contains(name, "/") {
		return fmt.Errorf("Invalid '/' character in group name '%s'", name)
	}
	return nil
}

//Create creates an openstack security group
func (sec *SecurityGroupManager) Create(options *api.SecurityGroupOptions) (*api.SecurityGroup, error) {
	err := chechGroupName(options.Name)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error creating security group")
	}
	createOpts := groups.CreateOpts{
		Name:        fmt.Sprintf("%s/%s", options.NetworkID, options.Name),
		Description: options.Description,
	}

	g, err := groups.Create(sec.OpenStack.Network, createOpts).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error creating security group")
	}
	return group(g), nil

}

//Delete deletes the Openstack security group identified by id
func (sec *SecurityGroupManager) Delete(id string) error {
	err := groups.Delete(sec.OpenStack.Network, id).ExtractErr()
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error listing security group")
	}
	return nil
}

func (sec *SecurityGroupManager) list(opts groups.ListOpts) ([]api.SecurityGroup, error) {
	allPages, err := groups.List(sec.OpenStack.Network, opts).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing security group")
	}

	allGroups, err := groups.ExtractGroups(allPages)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing security group")
	}
	result := []api.SecurityGroup{}
	for _, g := range allGroups {
		group, err := sec.get(g.ID, false)
		if err != nil {
			return nil, errors.Wrap(ProviderError(err), "Error listing security group")
		}
		result = append(result, *group)
	}
	return result, nil
}

//List list all Openstack security groups defined in the tenant
func (sec *SecurityGroupManager) List() ([]api.SecurityGroup, error) {
	listOpts := groups.ListOpts{}

	return sec.list(listOpts)
}

func rule(r *rules.SecGroupRule) *api.SecurityRule {
	return &api.SecurityRule{
		ID:              r.ID,
		SecurityGroupID: r.SecGroupID,
		Direction:       api.RuleDirection(r.Direction),
		PortRange: api.PortRange{
			From: r.PortRangeMin,
			To:   r.PortRangeMax,
		},
		Protocol:    api.Protocol(r.Protocol),
		Description: r.Description,
	}
}

func ruleOptions(rule *api.SecurityRuleOptions) *rules.CreateOpts {
	return &rules.CreateOpts{
		Description:  rule.Description,
		Direction:    rules.RuleDirection(rule.Direction),
		PortRangeMax: rule.PortRange.To,
		PortRangeMin: rule.PortRange.From,
		Protocol:     rules.RuleProtocol(rule.Protocol),
	}
}

//Get returns the Openstack security group identified by id
func (sec *SecurityGroupManager) get(id string, withRules bool) (*api.SecurityGroup, error) {
	listOpts := groups.ListOpts{
		ID: id,
	}

	groups, err := sec.list(listOpts)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("Error getting  security group: group does not exists")
	} else if len(groups) > 1 {
		return nil, fmt.Errorf("Error getting  security group: Provider error: multiple security groups exists with the same identifier")
	}
	group := &groups[0]

	if !withRules {
		return group, nil
	}
	ruleOpts := rules.ListOpts{
		SecGroupID: group.ID,
	}

	allPages, err := rules.List(sec.OpenStack.Network, ruleOpts).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting  security group")
	}

	allRules, err := rules.ExtractRules(allPages)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error getting  security group")
	}
	rules := []api.SecurityRule{}
	for _, r := range allRules {
		rules = append(rules, *rule(&r))
	}
	group.Rules = rules
	return group, nil
}

//Get returns the  security group identified by id fetching rules
func (sec *SecurityGroupManager) Get(id string) (*api.SecurityGroup, error) {
	return sec.get(id, true)
}

//AddServer a server to a security group
func (sec *SecurityGroupManager) AddServer(id string, serverID string) error {
	err := secgroups.AddServer(sec.OpenStack.Compute, serverID, id).ExtractErr()
	return errors.Wrap(ProviderError(err), "Error deleting security rule")
}

//RemoveServer remove a server from a security group
func (sec *SecurityGroupManager) RemoveServer(id string, serverID string) error {
	err := secgroups.RemoveServer(sec.OpenStack.Compute, serverID, id).ExtractErr()
	return errors.Wrap(ProviderError(err), "Error deleting security rule")
}

//ListByServer list security groups by server
func (sec *SecurityGroupManager) ListByServer(serverID string) ([]api.SecurityGroup, error) {
	page, err := secgroups.ListByServer(sec.OpenStack.Compute, serverID).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing security group by server")
	}
	groups, err := secgroups.ExtractSecurityGroups(page)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing security group by server")
	}
	result := []api.SecurityGroup{}
	for _, g := range groups {
		group := api.SecurityGroup{
			Description: g.Description,
			ID:          g.ID,
			Name:        g.Name,
		}
		result = append(result, group)
	}
	return result, nil
}

//AddRule adds a security rule to an OpenStack security group
func (sec *SecurityGroupManager) AddRule(id string, options *api.SecurityRuleOptions) (*api.SecurityRule, error) {
	opts := ruleOptions(options)
	opts.SecGroupID = id
	r, err := rules.Create(sec.OpenStack.Network, opts).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error adding security rule")
	}
	return rule(r), nil
}

//DeleteRule deletes a secuity rule from an OpenStack security group
func (sec *SecurityGroupManager) DeleteRule(ruleID string) error {
	err := rules.Delete(sec.OpenStack.Network, ruleID).ExtractErr()
	return errors.Wrap(ProviderError(err), "Error deleting security rule")

}
