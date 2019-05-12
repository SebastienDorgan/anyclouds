package aws_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type AWSSecurityGroupManagerTestSuite struct {
	tests.SecurityGroupManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AWSSecurityGroupManagerTestSuite) SetupSuite() {
	suite.Prov, _ = GetProvider()
}

func TestAWSSecurityGroupManagerTestSuite(t *testing.T) {
	suite.Run(t, new(AWSSecurityGroupManagerTestSuite))
}
