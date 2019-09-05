package openstack_test

import (
	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
	"testing"
)

type OSServerManagerTestSuite struct {
	tests.ServerManagerTestSuite
}

//SetupSuite set up image manager
func (suite *OSServerManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Prov = p
}

func TestOSServerManagerTestSuite(t *testing.T) {
	suite.Run(t, new(OSServerManagerTestSuite))
}
