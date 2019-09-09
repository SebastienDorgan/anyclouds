package azure_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type AZNetworkManagerTestSuite struct {
	tests.NetworkManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AZNetworkManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Mgr = p.GetNetworkManager()
}

func TestAZNetworkManagerTestSuite(t *testing.T) {
	suite.Run(t, new(AZNetworkManagerTestSuite))
}
