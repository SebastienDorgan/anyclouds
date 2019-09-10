package openstack

import (
	"fmt"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"strings"

	"github.com/SebastienDorgan/anyclouds/api"
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
func (mgr *SecurityGroupManager) Create(options api.SecurityGroupOptions) (*api.SecurityGroup, error) {
	err := checkGroupName(options.Name)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "error creating security group")
	}
	createOpts := groups.CreateOpts{
		Name:        fmt.Sprintf("%s/%s", options.NetworkID, options.Name),
		Description: options.Description,
	}

	g, err := groups.Create(mgr.OpenStack.Network, createOpts).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "error creating security group")
	}
	return group(g), nil

}

//Delete deletes the Openstack security group identified by id
func (mgr *SecurityGroupManager) Delete(id string) error {
	err := groups.Delete(mgr.OpenStack.Network, id).ExtractErr()
	if err != nil {
		return errors.Wrap(ProviderError(err), "error listing security group")
	}
	return nil
}

func (mgr *SecurityGroupManager) list(opts groups.ListOpts) ([]api.SecurityGroup, error) {
	securityGroups, err := mgr.groups(opts)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "error listing security group")
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
func (mgr *SecurityGroupManager) List() ([]api.SecurityGroup, error) {
	listOpts := groups.ListOpts{}

	return mgr.list(listOpts)
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
func (mgr *SecurityGroupManager) get(id string, withRules bool) (*api.SecurityGroup, error) {
	listOpts := groups.ListOpts{
		ID: id,
	}

	securityGroups, err := mgr.list(listOpts)
	if err != nil {
		return nil, err
	}
	if len(securityGroups) == 0 {
		return nil, fmt.Errorf("Error getting  security group: group does not exists")
	} else if len(securityGroups) > 1 {
		return nil, fmt.Errorf("Error getting  security group: Provider error: multiple security groups exists with the same identifier")
	}
	group := &securityGroups[0]
	return group, nil
}

//Get returns the  security group identified by id fetching rules
func (mgr *SecurityGroupManager) Get(id string) (*api.SecurityGroup, error) {
	return mgr.get(id, true)
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
func (mgr *SecurityGroupManager) Attach(options api.SecurityGroupAttachmentOptions) error {
	//srv, err := servers.Get(mgr.OpenStack.Compute, id).Extract()
	sn, err := mgr.OpenStack.NetworkManager.GetSubnet(options.NetworkID, options.SubnetID)
	if err != nil {
		return errors.Wrapf(ProviderError(err), "error attaching security group %s to server %s on subnet %s of network %s",
			options.SecurityGroupID,
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
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
	return errors.Wrapf(ProviderError(err), "error attaching security group %s to server %s on subnet %s of network %s",
		options.SecurityGroupID,
		options.ServerID,
		options.SubnetID,
		options.NetworkID,
	)
}

//AddRule adds a security rule to an OpenStack security group
func (mgr *SecurityGroupManager) AddRule(options api.SecurityRuleOptions) (*api.SecurityRule, error) {
	opts := ruleOptions(&options)
	opts.SecGroupID = options.SecurityGroupID
	r, err := rules.Create(mgr.OpenStack.Network, opts).Extract()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error adding security rule")
	}
	return rule(r), nil
}

//DeleteRule deletes a security rule from an OpenStack security group
func (mgr *SecurityGroupManager) DeleteRule(groupID, ruleID string) error {
	err := rules.Delete(mgr.OpenStack.Network, ruleID).ExtractErr()
	return errors.Wrap(ProviderError(err), "Error deleting security rule")

}
