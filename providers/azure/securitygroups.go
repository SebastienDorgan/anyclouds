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

func (mgr *SecurityGroupManager) Create(options api.SecurityGroupOptions) (*api.SecurityGroup, *api.CreateSecurityGroupError) {
	tags := make(map[string]*string, 1)
	tags["networkID"] = &options.NetworkID
	future, err := mgr.Provider.SecurityGroupsClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), options.Name, network.SecurityGroup{
		Location: &mgr.Provider.Configuration.Location,
		Tags:     tags,
	})
	if err != nil {
		return nil, api.NewCreateSecurityGroupError(err, options)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.SecurityGroupsClient.Client)
	if err != nil {
		return nil, api.NewCreateSecurityGroupError(err, options)
	}
	sg, err := future.Result(mgr.Provider.SecurityGroupsClient)
	if err != nil {
		return nil, api.NewCreateSecurityGroupError(err, options)
	}
	return &api.SecurityGroup{
		ID:        *sg.Name,
		Name:      *sg.Name,
		NetworkID: *sg.Tags["networkID"],
		Rules:     nil,
	}, nil
}

func (mgr *SecurityGroupManager) Delete(id string) *api.DeleteSecurityGroupError {
	future, err := mgr.Provider.SecurityGroupsClient.Delete(context.Background(), mgr.resourceGroup(), id)
	if err != nil {
		return api.NewDeleteSecurityGroupError(err, id)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.SecurityGroupsClient.Client)
	return api.NewDeleteSecurityGroupError(err, id)
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

func (mgr *SecurityGroupManager) List() ([]api.SecurityGroup, *api.ListSecurityGroupsError) {
	res, err := mgr.Provider.SecurityGroupsClient.List(context.Background(), mgr.resourceGroup())
	if err != nil {
		return nil, api.NewListSecurityGroupsError(err)
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

func (mgr *SecurityGroupManager) Get(id string) (*api.SecurityGroup, *api.GetSecurityGroupError) {
	sg, err := mgr.Provider.SecurityGroupsClient.Get(context.Background(), mgr.resourceGroup(), id, "")
	if err != nil {
		return nil, api.NewGetSecurityGroupError(err, id)
	}

	return &api.SecurityGroup{
		ID:        *sg.Name,
		Name:      *sg.Name,
		NetworkID: *sg.Tags["networkID"],
		Rules:     extractRules(&sg),
	}, nil
}

func (mgr *SecurityGroupManager) Attach(options api.AttachSecurityGroupOptions) *api.AttachSecurityGroupError {
	sg, err := mgr.Provider.SecurityGroupsClient.Get(context.Background(), mgr.resourceGroup(), options.SecurityGroupID, "")
	if err != nil {
		return api.NewAttachSecurityGroupError(err, options)
	}
	srv, err := mgr.Provider.ServerManager.get(options.ServerID)
	if err != nil {
		return api.NewAttachSecurityGroupError(err, options)
	}
	if srv.NetworkProfile == nil || srv.NetworkProfile.NetworkInterfaces == nil || len(*srv.NetworkProfile.NetworkInterfaces) == 0 {
		err = errors.Errorf("network interface not found")
		return api.NewAttachSecurityGroupError(err, options)
	}
	done := false
	for _, nir := range *srv.NetworkProfile.NetworkInterfaces {
		ni, err := mgr.Provider.InterfacesClient.Get(context.Background(), mgr.resourceGroup(), *nir.ID, "")
		if err != nil {
			return api.NewAttachSecurityGroupError(err, options)
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
			return api.NewAttachSecurityGroupError(err, options)
		}
		err = future.WaitForCompletionRef(context.Background(), mgr.Provider.InterfacesClient.Client)
		if err != nil {
			return api.NewAttachSecurityGroupError(err, options)
		}
	}
	if !done {
		err = errors.Errorf("network interface not found")
		return api.NewAttachSecurityGroupError(err, options)
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

func azSecurityRule(rule *api.AddSecurityRuleOptions) *network.SecurityRule {
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

func (mgr *SecurityGroupManager) AddSecurityRule(options api.AddSecurityRuleOptions) (*api.SecurityRule, *api.AddSecurityRuleError) {
	sg, err := mgr.Provider.SecurityGroupsClient.Get(context.Background(), mgr.resourceGroup(), options.SecurityGroupID, "")
	if err != nil {
		return nil, api.NewAddSecurityRuleError(err, options)
	}
	var rules []network.SecurityRule
	if sg.SecurityRules != nil {
		rules = append(rules, *sg.SecurityRules...)
	}
	rule := azSecurityRule(&options)
	rules = append(rules, *rule)
	sg.SecurityRules = &rules
	future, err := mgr.Provider.SecurityGroupsClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), options.SecurityGroupID, sg)
	if err != nil {
		return nil, api.NewAddSecurityRuleError(err, options)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.SecurityGroupsClient.Client)
	if err != nil {
		return nil, api.NewAddSecurityRuleError(err, options)
	}
	sg, err = future.Result(mgr.Provider.SecurityGroupsClient)
	if err != nil {
		return nil, api.NewAddSecurityRuleError(err, options)
	}
	for _, r := range *sg.SecurityRules {
		if r.Name == rule.Name {
			return convertRule(rule, *sg.ID), nil
		}
	}
	return nil, api.NewAddSecurityRuleError(err, options)
}

func (mgr *SecurityGroupManager) RemoveSecurityRule(groupID, ruleID string) *api.RemoveSecurityRuleError {
	sg, err := mgr.Provider.SecurityGroupsClient.Get(context.Background(), mgr.resourceGroup(), groupID, "")
	if err != nil {
		return api.NewRemoveSecurityRuleError(err, groupID, ruleID)
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
		err = errors.Errorf("rule does not exist")
		return api.NewRemoveSecurityRuleError(err, groupID, ruleID)
	}
	sg.SecurityRules = &rules
	future, err := mgr.Provider.SecurityGroupsClient.CreateOrUpdate(context.Background(), mgr.resourceGroup(), groupID, sg)
	if err != nil {
		return api.NewRemoveSecurityRuleError(err, groupID, ruleID)
	}
	err = future.WaitForCompletionRef(context.Background(), mgr.Provider.SecurityGroupsClient.Client)
	return api.NewRemoveSecurityRuleError(err, groupID, ruleID)

}
