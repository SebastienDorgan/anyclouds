package openstack_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type OSSecurityGroupManagerTestSuite struct {
	tests.SecurityGroupManagerTestSuite
}

//SetupSuite set up image manager
func (suite *OSSecurityGroupManagerTestSuite) SetupSuite() {
	suite.Prov = GetProvider()
}

func TestAWSSecurityGroupManagerTestSuite(t *testing.T) {
	suite.Run(t, new(OSSecurityGroupManagerTestSuite))
}
