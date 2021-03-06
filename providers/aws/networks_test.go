package aws_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type AWSNetworkManagerTestSuite struct {
	tests.NetworkManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AWSNetworkManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Mgr = p.GetNetworkManager()
}

func TestAWSNetworkManagerTestSuite(t *testing.T) {
	suite.Run(t, new(AWSNetworkManagerTestSuite))
}
