package openstack_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type OSNetworkManagerTestSuite struct {
	tests.NetworkManagerTestSuite
}

//SetupSuite set up image manager
func (suite *OSNetworkManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Mgr = p.GetNetworkManager()
}

func TestOSNetworkManagerTestSuite(t *testing.T) {
	suite.Run(t, new(OSNetworkManagerTestSuite))
}
