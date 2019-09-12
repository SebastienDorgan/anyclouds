package openstack

import (
	"fmt"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"strings"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
)

//SecurityGroupManager defines security group management functions a anyclouds provider must provide
type SecurityGroupManager struct {
	OpenStack *Provider
}

func group(g *groups.SecGroup) *api.SecurityGroup {
	tokens := strings.Split("/", g.Name)
	var rls []api.SecurityRule
	for _, r := range g.Rules {
		rls = append(rls, *rule(&r))
	}
	return &api.SecurityGroup{
		Name:      tokens[0],
		NetworkID: tokens[1],
		ID:        g.ID,
	}
}

func checkGroupName(name string) error {
	if strings.Contains(name, "/") {
		return fmt.Errorf("invalid '/' character in group name '%s'", name)
	}
	return nil
}

//Create creates an openstack security group
func (mgr *SecurityGroupManager) Create(options api.SecurityGroupOptions) (*api.SecurityGroup, *api.CreateSecurityGroupError) {
	err := checkGroupName(options.Name)
	if err != nil {
		return nil, api.NewCreateSecurityGroupError(UnwrapOpenStackError(err), options)
	}
	createOpts := groups.CreateOpts{
		Name:        fmt.Sprintf("%s/%s", options.NetworkID, options.Name),
		Description: options.Description,
	}

	g, err := groups.Create(mgr.OpenStack.Network, createOpts).Extract()
	if err != nil {
		return nil, api.NewCreateSecurityGroupError(UnwrapOpenStackError(err), options)
	}
	return group(g), nil

}

//Delete deletes the Openstack security group identified by id
func (mgr *SecurityGroupManager) Delete(id string) *api.DeleteSecurityGroupError {
	err := groups.Delete(mgr.OpenStack.Network, id).ExtractErr()
	if err != nil {
		return api.NewDeleteSecurityGroupError(UnwrapOpenStackError(err), id)
	}
	return nil
}

func (mgr *SecurityGroupManager) list(opts groups.ListOpts) ([]api.SecurityGroup, error) {
	securityGroups, err := mgr.groups(opts)
	if err != nil {
		return nil, UnwrapOpenStackError(err)
	}
	var result []api.SecurityGroup
	for _, g := range securityGroups {
		result = append(result, *group(&g))
	}
	return result, nil
}

func (mgr *SecurityGroupManager) groups(opts groups.ListOpts) ([]groups.SecGroup, error) {
	allPages, err := groups.List(mgr.OpenStack.Network, opts).AllPages()
	if err != nil {
		return nil, err
	}
	return groups.ExtractGroups(allPages)

}

//List list all Openstack security groups defined in the tenant
func (mgr *SecurityGroupManager) List() ([]api.SecurityGroup, *api.ListSecurityGroupsError) {
	listOpts := groups.ListOpts{}

	l, err := mgr.list(listOpts)
	return l, api.NewListSecurityGroupsError(err)
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

func ruleOptions(rule *api.AddSecurityRuleOptions) *rules.CreateOpts {
	return &rules.CreateOpts{
		Description:  rule.Description,
		Direction:    rules.RuleDirection(rule.Direction),
		PortRangeMax: rule.PortRange.To,
		PortRangeMin: rule.PortRange.From,
		Protocol:     rules.RuleProtocol(rule.Protocol),
	}
}

//Get returns the Openstack security group identified by id
func (mgr *SecurityGroupManager) get(id string, withRules bool) (*api.SecurityGroup, error) {
	listOpts := groups.ListOpts{
		ID: id,
	}

	securityGroups, err := mgr.list(listOpts)
	if err != nil {
		return nil, err
	}
	if len(securityGroups) == 0 {
		return nil, fmt.Errorf("security group does not exists")
	} else if len(securityGroups) > 1 {
		return nil, fmt.Errorf("multiple security groups exists with the same identifier")
	}
	group := &securityGroups[0]
	return group, nil
}

//Get returns the  security group identified by id fetching rules
func (mgr *SecurityGroupManager) Get(id string) (*api.SecurityGroup, *api.GetSecurityGroupError) {
	sg, err := mgr.get(id, true)
	return sg, api.NewGetSecurityGroupError(UnwrapOpenStackError(err), id)
}

func checkIPs(subnetID string, ip string, ips []ports.IP) bool {
	for _, addr := range ips {
		if subnetID == addr.SubnetID && addr.IPAddress == ip {
			return true
		}
	}
	return false
}

//Attach a server to a security group
func (mgr *SecurityGroupManager) Attach(options api.AttachSecurityGroupOptions) *api.AttachSecurityGroupError {
	//srv, err := servers.Get(mgr.OpenStack.Compute, id).Extract()
	var err error
	sn, err := mgr.OpenStack.NetworkManager.GetSubnet(options.NetworkID, options.SubnetID)
	if err != nil {
		return api.NewAttachSecurityGroupError(UnwrapOpenStackError(err), options)
	}
	pts, err := mgr.OpenStack.PublicIPAddressManager.listPorts(options.ServerID)
	for _, p := range pts {
		if p.NetworkID == sn.NetworkID && (options.IPAddress == nil || checkIPs(options.SubnetID, *options.IPAddress, p.FixedIPs)) {
			ports.Update(mgr.OpenStack.Network, p.ID, ports.UpdateOpts{
				Name:                &p.Name,
				Description:         &p.Description,
				AdminStateUp:        &p.AdminStateUp,
				FixedIPs:            &p.FixedIPs,
				DeviceID:            &p.DeviceID,
				DeviceOwner:         &p.DeviceOwner,
				SecurityGroups:      &[]string{options.SecurityGroupID},
				AllowedAddressPairs: &p.AllowedAddressPairs,
			})
		}
	}
	return api.NewAttachSecurityGroupError(UnwrapOpenStackError(err), options)
}

//AddSecurityRule adds a security rule to an OpenStack security group
func (mgr *SecurityGroupManager) AddSecurityRule(options api.AddSecurityRuleOptions) (*api.SecurityRule, *api.AddSecurityRuleError) {
	opts := ruleOptions(&options)
	opts.SecGroupID = options.SecurityGroupID
	r, err := rules.Create(mgr.OpenStack.Network, opts).Extract()
	if err != nil {
		return nil, api.NewAddSecurityRuleError(UnwrapOpenStackError(err), options)
	}
	return rule(r), nil
}

//RemoveSecurityRule deletes a security rule from an OpenStack security group
func (mgr *SecurityGroupManager) RemoveSecurityRule(groupID, ruleID string) *api.RemoveSecurityRuleError {
	err := rules.Delete(mgr.OpenStack.Network, ruleID).ExtractErr()
	return api.NewRemoveSecurityRuleError(UnwrapOpenStackError(err), groupID, ruleID)
}
