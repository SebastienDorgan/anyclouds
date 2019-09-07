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
	mgr := aws.PublicIPAddressManager{AWS: prov}
	pools, err := mgr.ListAvailablePools()
	assert.NoError(t, err)
	for _, pool := range pools {
		fmt.Printf("%v", pool)
	}
	ip, err := mgr.Allocate(&api.PublicIPAllocationOptions{
		Name: "test_ip",
	})
	assert.NoError(t, err)
	assert.Equal(t, "test_ip", ip.Name)
	assert.NotEmpty(t, ip.Address)
	assert.NotEmpty(t, ip.ID)
	ips, err := mgr.ListAllocated()
	assert.NoError(t, err)
	assert.True(t, len(ips) == 1)
	assert.Equal(t, ip.ID, ips[0].ID)
	assert.Equal(t, ip.Name, ips[0].Name)

	kp, err := sshutils.CreateKeyPair(4096)
	assert.NoError(t, err)
	err = prov.GetKeyPairManager().Import("test_kp", kp.PublicKey)
	assert.NoError(t, err)
	net, err := prov.GetNetworkManager().CreateNetwork(&api.NetworkOptions{
		CIDR: "10.0.0.0/16",
		Name: "test_network",
	})
	assert.NoError(t, err)
	sg, err := prov.GetSecurityGroupManager().Create(&api.SecurityGroupOptions{
		Name:        "sg",
		Description: "security group",
		NetworkID:   net.ID,
	})
	assert.NoError(t, err)
	_, err = prov.GetSecurityGroupManager().AddRule(&api.SecurityRuleOptions{
		SecurityGroupID: sg.ID,
		Direction:       api.RuleDirectionIngress,
		PortRange:       api.PortRange{From: 22, To: 22},
		Protocol:        api.ProtocolTCP,
		Description:     "grant ssh access",
	})
	assert.NoError(t, err)
	_, err = prov.GetSecurityGroupManager().AddRule(&api.SecurityRuleOptions{
		SecurityGroupID: sg.ID,
		Direction:       api.RuleDirectionIngress,
		Protocol:        api.ProtocolICMP,
		Description:     "grant icmp access",
	})
	assert.NoError(t, err)
	snet, err := prov.GetNetworkManager().CreateSubnet(&api.SubnetOptions{
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

	srv, err := prov.GetServerManager().Create(&api.CreateServerOptions{
		Name:            "test_server",
		TemplateID:      tpls[0].ID,
		ImageID:         images[selection[n]].ID,
		SecurityGroups:  []string{sg.ID},
		Subnets:         []string{snet.ID},
		PublicIP:        false,
		BootstrapScript: nil,
		KeyPairName:     "test_kp",
	})
	assert.NotNil(t, srv)
	if srv != nil {
		assert.Equal(t, "", srv.PublicIPv4)
		assert.Equal(t, "", srv.AccessIPv6)
		err = mgr.Associate(&api.PublicIPAssociationOptions{
			PublicIPId: ip.ID,
			ServerID:   srv.ID,
		})
		assert.NoError(t, err)
		srv, err = prov.GetServerManager().Get(srv.ID)
		assert.NoError(t, err)
		assert.Equal(t, ip.Address, srv.PublicIPv4)
		err = mgr.Dissociate(ip.ID)
		assert.NoError(t, err)
		err = prov.GetServerManager().Delete(srv.ID)
		assert.NoError(t, err)
	}

	err = prov.GetNetworkManager().DeleteSubnet(snet.ID)
	assert.NoError(t, err)
	err = prov.GetKeyPairManager().Delete("test_kp")
	assert.NoError(t, err)
	err = prov.GetSecurityGroupManager().Delete(sg.ID)
	assert.NoError(t, err)
	err = prov.GetNetworkManager().DeleteNetwork(net.ID)
	assert.NoError(t, err)
	err = mgr.Release(ip.ID)
	assert.NoError(t, err)
}