package tests

import (
	"reflect"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/SebastienDorgan/anyclouds/api"
)

//SecurityGroupManagerTestSuite test suite for api.SecurityGroupManager
type SecurityGroupManagerTestSuite struct {
	suite.Suite
	Prov api.Provider
}

func (s *SecurityGroupManagerTestSuite) deleteNetwork(id string) {
	err := s.Prov.GetNetworkManager().DeleteNetwork(id)
	assert.NoError(s.T(), err)
}

//TestSecurityGroupManager Canonical test for SecurityGroupManager implementation
func (s *SecurityGroupManagerTestSuite) TestSecurityGroupManager() {

	n, err := s.Prov.GetNetworkManager().CreateNetwork(api.NetworkOptions{
		CIDR: "10.0.0.0/16",
	})
	assert.NoError(s.T(), err)
	defer s.deleteNetwork(n.ID)
	Mgr := s.Prov.GetSecurityGroupManager()
	sgl, err := Mgr.List()
	assert.NoError(s.T(), err)

	l0 := len(sgl)
	sg, err := Mgr.Create(api.SecurityGroupOptions{
		Description: "test security group",
		Name:        "test_sg",
		NetworkID:   n.ID,
	})
	assert.NoError(s.T(), err)
	_, err = Mgr.AddRule(api.SecurityRuleOptions{
		SecurityGroupID: sg.ID,
		Description:     "rule",
		Direction:       api.RuleDirectionIngress,
		PortRange: api.PortRange{
			From: 0,
			To:   10000,
		},
		Protocol: api.ProtocolTCP,
		//CIDR:     "0.0.0.0/0",
	})
	assert.NoError(s.T(), err)
	sgl, err = Mgr.List()
	assert.Equal(s.T(), l0+1, len(sgl))

	sg, err = Mgr.Get(sg.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(sg.Rules))
	if len(sg.Rules) == 0 {
		assert.FailNow(s.T(), "No rule added")
	}
	r := sg.Rules[0]

	assert.Equal(s.T(), r.Description, "rule")
	assert.Equal(s.T(), r.Direction, api.RuleDirectionIngress)
	assert.Equal(s.T(), r.Protocol, api.ProtocolTCP)
	assert.True(s.T(), reflect.DeepEqual(r.PortRange, api.PortRange{
		From: 0,
		To:   10000,
	}))

	err = Mgr.DeleteRule(sg.ID, r.ID)
	sg, err = Mgr.Get(sg.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, len(sg.Rules))
	err = Mgr.Delete(sg.ID)
	assert.NoError(s.T(), err)
	sgl, err = Mgr.List()
	assert.Equal(s.T(), l0, len(sgl))

}
