package aws_test

import (
	"fmt"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/anyclouds/providers/aws"
	"github.com/SebastienDorgan/anyclouds/sshutils"
	"github.com/SebastienDorgan/talgo"
	"github.com/stretchr/testify/assert"
	"sort"
	"strings"
	"testing"
)

func TestPublicAddresses(t *testing.T) {
	prov := GetProvider()
	mgr := aws.PublicIPAddressManager{Provider: prov}
	pools, err := mgr.ListAvailablePools()
	assert.NoError(t, err)
	for _, pool := range pools {
		fmt.Printf("%v", pool)
	}
	ip, err := mgr.Create(api.AllocatePublicIPAddressOptions{
		Name: "test_ip",
	})
	assert.NoError(t, err)
	assert.Equal(t, "test_ip", ip.Name)
	assert.NotEmpty(t, ip.Address)
	assert.NotEmpty(t, ip.ID)
	ips, err := mgr.List(&api.ListPublicIPAddressOptions{})
	assert.NoError(t, err)
	assert.True(t, len(ips) == 1)
	assert.Equal(t, ip.ID, ips[0].ID)
	assert.Equal(t, ip.Name, ips[0].Name)

	kp, err := sshutils.CreateKeyPair(4096)
	assert.NoError(t, err)
	kp, err = sshutils.CreateKeyPair(4096)
	assert.NoError(t, err)
	net, err := prov.GetNetworkManager().CreateNetwork(api.CreateNetworkOptions{
		CIDR: "10.0.0.0/16",
		Name: "test_network",
	})
	assert.NoError(t, err)
	sg, err := prov.GetSecurityGroupManager().Create(api.SecurityGroupOptions{
		Name:        "sg",
		Description: "security group",
		NetworkID:   net.ID,
	})
	assert.NoError(t, err)
	_, err = prov.GetSecurityGroupManager().AddSecurityRule(api.AddSecurityRuleOptions{
		SecurityGroupID: sg.ID,
		Direction:       api.RuleDirectionIngress,
		PortRange:       api.PortRange{From: 22, To: 22},
		Protocol:        api.ProtocolTCP,
		Description:     "grant ssh access",
	})
	assert.NoError(t, err)
	_, err = prov.GetSecurityGroupManager().AddSecurityRule(api.AddSecurityRuleOptions{
		SecurityGroupID: sg.ID,
		Direction:       api.RuleDirectionIngress,
		Protocol:        api.ProtocolICMP,
		Description:     "grant icmp access",
	})
	assert.NoError(t, err)
	snet, err := prov.GetNetworkManager().CreateSubnet(api.SubnetOptions{
		NetworkID: net.ID,
		Name:      "test_subnet",
		CIDR:      "10.0.0.0/24",
		IPVersion: api.IPVersion4,
	})
	assert.NoError(t, err)
	tpls, err := prov.GetTemplateManager().List()
	assert.NoError(t, err)
	sort.Slice(tpls, func(i, j int) bool {
		return tpls[i].OneDemandPrice < tpls[j].OneDemandPrice
	})
	images, err := prov.GetImageManager().List()
	assert.NoError(t, err)
	selection := talgo.FindAll(len(images), func(i int) bool {
		name := strings.ToUpper(images[i].Name)
		return strings.Contains(name, "UBUNTU") && strings.Contains(name, "18.04")
	})
	n := talgo.Select(len(selection), func(i, j int) int {
		if images[selection[i]].UpdatedAt.After(images[selection[j]].UpdatedAt) {
			return i
		}
		return j
	})
	assert.True(t, n < len(selection))

	srv, err := prov.GetServerManager().Create(api.CreateServerOptions{
		Name:                 "test_server",
		TemplateID:           tpls[0].ID,
		ImageID:              images[selection[n]].ID,
		DefaultSecurityGroup: sg.ID,
		Subnets:              []api.Subnet{*snet},
		BootstrapScript:      nil,
		KeyPair:              *kp,
	})
	assert.NotNil(t, srv)
	if srv != nil {
		err = mgr.Associate(api.AssociatePublicIPOptions{
			PublicIPId: ip.ID,
			ServerID:   srv.ID,
		})
		assert.NoError(t, err)
		nis, err := prov.GetNetworkInterfaceManager().List(&api.ListNetworkInterfacesOptions{ServerID: &srv.ID})
		assert.NoError(t, err)
		assert.Equal(t, ip.Address, nis[0].PublicIPAddress)
		err = mgr.Dissociate(ip.ID)
		assert.NoError(t, err)
		err = prov.GetServerManager().Delete(srv.ID)
		assert.NoError(t, err)
	}

	err = prov.GetNetworkManager().DeleteSubnet(net.ID, snet.ID)
	assert.NoError(t, err)
	err = prov.GetSecurityGroupManager().Delete(sg.ID)
	assert.NoError(t, err)
	err = prov.GetNetworkManager().DeleteNetwork(net.ID)
	assert.NoError(t, err)
	err = mgr.Delete(ip.ID)
	assert.NoError(t, err)
}
