package azure_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type AZSecurityGroupManagerTestSuite struct {
	tests.SecurityGroupManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AZSecurityGroupManagerTestSuite) SetupSuite() {
	suite.Prov = GetProvider()
}

func TestAZSecurityGroupManagerTestSuite(t *testing.T) {
	suite.Run(t, new(AZSecurityGroupManagerTestSuite))
}
