package openstack_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type OSKeyPairManagerTestSuite struct {
	tests.KeyPairManagerTestSuite
}

//SetupSuite set up image manager
func (suite *OSKeyPairManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Mgr = p.GetKeyPairManager()
}

func TestOSKeyPairManagerTestSuite(t *testing.T) {
	suite.Run(t, new(OSKeyPairManagerTestSuite))
}
