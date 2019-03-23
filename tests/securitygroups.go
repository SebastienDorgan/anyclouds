package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/SebastienDorgan/anyclouds/api"
)

//SecurityGroupManagerTestSuite test suite for api.SecurityGroupManager
type SecurityGroupManagerTestSuite struct {
	suite.Suite
	Prov api.Provider
}

//TestSecurityGroupManager Canonical test for SecurityGroupManager implementation
func (s *SecurityGroupManagerTestSuite) TestSecurityGroupManager() {

	n, err := s.Prov.GetNetworkManager().CreateNetwork(&api.NetworkOptions{
		CIDR: "10.0.0.0/16",
	})
	assert.NoError(s.T(), err)

	Mgr := s.Prov.GetSecurityGroupManager()
	sgl, err := Mgr.List()
	assert.NoError(s.T(), err)

	l0 := len(sgl)
	sg, err := Mgr.Create(&api.SecurityGroupOptions{
		Description: "test security group",
		Name:        "test_sg",
		NetworkID:   n.ID,
	})

	Mgr.AddRule(sg.ID, api.SecurityRuleOptions{
		Description: "rule",
		Direction:   api.RuleDirectionIngress,
	})
	assert.NoError(s.T(), err)
	sgl, err = Mgr.List()
	assert.Equal(s.T(), l0+1, len(sgl))

	Mgr.Delete(sg.ID)
	assert.NoError(s.T(), err)
	sgl, err = Mgr.List()
	assert.Equal(s.T(), l0, len(sgl))
}
