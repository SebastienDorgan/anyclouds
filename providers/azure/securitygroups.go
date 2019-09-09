package azure

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"math"
	"strconv"
	"strings"
)

type SecurityGroupManager struct {
	Provider *Provider
}

func (mgr *SecurityGroupManager) resourceGroup() string {
	return mgr.Provider.Configuration.ResourceGroupName
}

func (mgr *SecurityGroupManager) Create(options *api.SecurityGroupOptions) (*api.SecurityGroup, error) {
	tags := make(map[string]*string, 1)
	tags["networkID"] = &options.NetworkID
	future, err := mgr.Provider.SecurityGroupsClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), options.Name, network.SecurityGroup{
		Location: &mgr.Provider.Configuration.Location,
		Tags:     tags,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error creating security group %s", options.Name)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.SecurityGroupsClient.Client)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating security group %s", options.Name)
	}
	sg, err := future.Result(mgr.Provider.SecurityGroupsClient)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating security group %s", options.Name)
	}
	return &api.SecurityGroup{
		ID:        *sg.Name,
		Name:      *sg.Name,
		NetworkID: *sg.Tags["networkID"],
		Rules:     nil,
	}, nil
}

func (mgr *SecurityGroupManager) Delete(id string) error {
	future, err := mgr.Provider.SecurityGroupsClient.Delete(context.Background(), mgr.resourceGroup(), id)
	if err != nil {
		return errors.Wrapf(err, "error deleting security group %s", id)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.SecurityGroupsClient.Client)
	return errors.Wrapf(err, "error deleting security group %s", id)

}

func convertDirection(direction network.SecurityRuleDirection) api.RuleDirection {
	if direction == network.SecurityRuleDirectionInbound {
		return api.RuleDirectionIngress
	}
	return api.RuleDirectionEgress
}

func convertProtocol(protocol network.SecurityRuleProtocol) api.Protocol {
	if protocol == network.SecurityRuleProtocolTCP {
		return api.ProtocolTCP
	}
	if protocol == network.SecurityRuleProtocolUDP {
		return api.ProtocolTCP
	}
	if protocol == network.SecurityRuleProtocolIcmp {
		return api.ProtocolICMP
	}
	return api.ProtocolAny
}

func convertPortRange(sourcePortRange *string) *api.PortRange {
	if sourcePortRange == nil {
		return &api.PortRange{
			From: 0,
			To:   0,
		}
	}
	if *sourcePortRange == "*" || *sourcePortRange == "" {
		return &api.PortRange{
			From: 0,
			To:   math.MaxUint16,
		}
	}
	tokens := strings.Split(*sourcePortRange, "-")
	if len(tokens) == 1 {
		from, _ := strconv.Atoi(tokens[0])
		return &api.PortRange{
			From: from,
			To:   math.MaxUint16,
		}
	}
	if len(tokens) == 2 {
		rfrom, _ := strconv.Atoi(tokens[0])
		rto, _ := strconv.Atoi(tokens[1])
		return &api.PortRange{
			From: rfrom,
			To:   rto,
		}
	}

	return &api.PortRange{
		From: 0,
		To:   0,
	}
}

func convertRule(r *network.SecurityRule, sgIg string) *api.SecurityRule {
	if r == nil {
		return nil
	}

	return &api.SecurityRule{
		ID:              *r.ID,
		SecurityGroupID: sgIg,
		Direction:       convertDirection(r.Direction),
		PortRange:       *convertPortRange(r.DestinationPortRange),
		Protocol:        convertProtocol(r.Protocol),
		CIDR:            *r.DestinationAddressPrefix,
		Description:     *r.Description,
	}
}

func extractRules(sg *network.SecurityGroup) []api.SecurityRule {
	if sg.SecurityRules == nil {
		return nil
	}
	var rules []api.SecurityRule
	for _, r := range *sg.SecurityRules {
		rules = append(rules, *convertRule(&r, *sg.Name))
	}
	return rules
}

func (mgr *SecurityGroupManager) List() ([]api.SecurityGroup, error) {
	res, err := mgr.Provider.SecurityGroupsClient.List(context.Background(), mgr.resourceGroup())
	if err != nil {
		return nil, errors.Wrap(err, "error listing security groups")
	}
	var sgs []api.SecurityGroup
	for _, sg := range res.Values() {

		sgs = append(sgs, api.SecurityGroup{
			ID:        *sg.Name,
			Name:      *sg.Name,
			NetworkID: *sg.Tags["networkID"],
			Rules:     extractRules(&sg),
		})
	}
	return sgs, nil
}

func (mgr *SecurityGroupManager) Get(id string) (*api.SecurityGroup, error) {
	sg, err := mgr.Provider.SecurityGroupsClient.Get(context.Background(), mgr.resourceGroup(), id, "")
	if err != nil {
		return nil, errors.Wrapf(err, "error getting security group %s", id)
	}

	return &api.SecurityGroup{
		ID:        *sg.Name,
		Name:      *sg.Name,
		NetworkID: *sg.Tags["networkID"],
		Rules:     extractRules(&sg),
	}, nil
}

func (mgr *SecurityGroupManager) Attach(options *api.SecurityGroupAttachmentOptions) error {
	sg, err := mgr.Provider.SecurityGroupsClient.Get(context.Background(), mgr.resourceGroup(), options.SecurityGroupID, "")
	if err != nil {
		return errors.Wrapf(err, "error attaching security group %s to server %s on subnet %s of network %s",
			options.SecurityGroupID,
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	srv, err := mgr.Provider.ServerManager.get(options.ServerID)
	if err != nil {
		return errors.Wrapf(err, "error attaching security group %s to server %s on subnet %s of network %s",
			options.SecurityGroupID,
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	if srv.NetworkProfile == nil || srv.NetworkProfile.NetworkInterfaces == nil || len(*srv.NetworkProfile.NetworkInterfaces) == 0 {
		err = errors.Errorf("no network interface found for subnet %s of network %s on server %s",
			options.SubnetID,
			options.NetworkID,
			options.ServerID,
		)
		return errors.Wrapf(err, "error attaching security group %s to server %s on subnet %s of network %s",
			options.SecurityGroupID,
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	done := false
	for _, nir := range *srv.NetworkProfile.NetworkInterfaces {
		ni, err := mgr.Provider.InterfacesClient.Get(context.Background(), mgr.resourceGroup(), *nir.ID, "")
		if err != nil {
			return errors.Wrapf(err, "error attaching security group %s to server %s on subnet %s of network %s",
				options.SecurityGroupID,
				options.ServerID,
				options.SubnetID,
				options.NetworkID,
			)
		}
		impacted := false
		for _, ipc := range *ni.IPConfigurations {
			if ipc.Subnet != nil && *ipc.Subnet.Name == options.SubnetID &&
				options.IPAddress == nil || (*ipc.PrivateIPAddress == *options.IPAddress) {
				impacted = true
				break
			}
		}
		if !impacted {
			continue
		}
		done = true
		ni.NetworkSecurityGroup = &sg
		future, err := mgr.Provider.InterfacesClient.CreateOrUpdate(context.Background(), *ni.Name, mgr.resourceGroup(), ni)
		if err != nil {
			return errors.Wrapf(err, "error attaching security group %s to server %s on subnet %s of network %s",
				options.SecurityGroupID,
				options.ServerID,
				options.SubnetID,
				options.NetworkID,
			)
		}
		err = future.WaitForCompletionRef(context.Background(), mgr.Provider.InterfacesClient.Client)
		if err != nil {
			return errors.Wrapf(err, "error attaching security group %s to server %s on subnet %s of network %s",
				options.SecurityGroupID,
				options.ServerID,
				options.SubnetID,
				options.NetworkID,
			)
		}
	}
	if !done {
		err = errors.Errorf("no network interface found for subnet %s of network %s on server %s",
			options.SubnetID,
			options.NetworkID,
			options.ServerID,
		)
		return errors.Wrapf(err, "error attaching security group %s to server %s on subnet %s of network %s",
			options.SecurityGroupID,
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	return nil
}

func convertAzProtocol(protocol api.Protocol) network.SecurityRuleProtocol {
	if protocol == api.ProtocolTCP {
		return network.SecurityRuleProtocolTCP
	}
	if protocol == api.ProtocolUDP {
		return network.SecurityRuleProtocolUDP
	}
	if protocol == api.ProtocolICMP {
		return network.SecurityRuleProtocolIcmp
	}

	return network.SecurityRuleProtocolAsterisk

}
func convertAzDirection(direction api.RuleDirection) network.SecurityRuleDirection {
	if direction == api.RuleDirectionEgress {
		return network.SecurityRuleDirectionOutbound
	}
	return network.SecurityRuleDirectionInbound
}
func convertAzPortRange(portRange api.PortRange) *string {
	return to.StringPtr(fmt.Sprintf("%d-%d", portRange.From, portRange.To))
}

func azSeurityRule(rule *api.SecurityRuleOptions) *network.SecurityRule {
	id := uuid.New()
	format := network.SecurityRulePropertiesFormat{
		Description:              &rule.Description,
		Protocol:                 convertAzProtocol(rule.Protocol),
		DestinationPortRange:     convertAzPortRange(rule.PortRange),
		DestinationAddressPrefix: &rule.CIDR,
		Direction:                convertAzDirection(rule.Direction),
	}
	return &network.SecurityRule{
		SecurityRulePropertiesFormat: &format,
		Name:                         to.StringPtr(id.String()),
	}
}

func (mgr *SecurityGroupManager) AddRule(options *api.SecurityRuleOptions) (*api.SecurityRule, error) {
	sg, err := mgr.Provider.SecurityGroupsClient.Get(context.Background(), mgr.resourceGroup(), options.SecurityGroupID, "")
	if err != nil {
		return nil, errors.Wrapf(err, "error adding security rule -%s- to security group %s ", options.Description, options.SecurityGroupID)
	}
	var rules []network.SecurityRule
	if sg.SecurityRules != nil {
		rules = append(rules, *sg.SecurityRules...)
	}
	rule := azSeurityRule(options)
	rules = append(rules, *rule)
	sg.SecurityRules = &rules
	future, err := mgr.Provider.SecurityGroupsClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), options.SecurityGroupID, sg)
	if err != nil {
		return nil, errors.Wrapf(err, "error adding security rule -%s- to security group %s ", options.Description, options.SecurityGroupID)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.SecurityGroupsClient.Client)
	if err != nil {
		return nil, errors.Wrapf(err, "error adding security rule -%s- to security group %s ", options.Description, options.SecurityGroupID)
	}
	sg, err = future.Result(mgr.Provider.SecurityGroupsClient)
	if err != nil {
		return nil, errors.Wrapf(err, "error adding security rule -%s- to security group %s ", options.Description, options.SecurityGroupID)
	}
	for _, r := range *sg.SecurityRules {
		if r.Name == rule.Name {
			return convertRule(rule, *sg.ID), nil
		}
	}
	return nil, errors.Errorf("error adding security rule -%s- to security group %s ", options.Description, options.SecurityGroupID)
}

func (mgr *SecurityGroupManager) DeleteRule(groupID, ruleID string) error {
	sg, err := mgr.Provider.SecurityGroupsClient.Get(context.Background(), mgr.resourceGroup(), groupID, "")
	if err != nil {
		return errors.Wrapf(err, "error deleting security rule -%s- to security group %s ", ruleID, groupID)
	}
	var rules []network.SecurityRule
	done := false
	if sg.SecurityRules != nil {
		for _, r := range *sg.SecurityRules {
			if *r.Name != ruleID {
				rules = append(rules, r)
			} else {
				done = true
			}
		}
	}
	if !done {
		err = errors.Errorf("rule %s does not exist in security group %s", ruleID, groupID)
		return errors.Wrapf(err, "error deleting security rule -%s- to security group %s ", ruleID, groupID)
	}
	sg.SecurityRules = &rules
	future, err := mgr.Provider.SecurityGroupsClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), groupID, sg)
	if err != nil {
		return errors.Wrapf(err, "error deleting security rule -%s- to security group %s ", ruleID, groupID)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.SecurityGroupsClient.Client)
	return errors.Wrapf(err, "error deleting security rule -%s- to security group %s ", ruleID, groupID)

}
