package openstack_test

import (
	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
	"testing"
)

type OSVolumeManagerTestSuite struct {
	tests.VolumeManagerTestSuite
}

//SetupSuite set up image manager
func (suite *OSVolumeManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Prov = p
}

func TestOSVolumeManagerTestSuite(t *testing.T) {
	suite.Run(t, new(OSVolumeManagerTestSuite))
}
