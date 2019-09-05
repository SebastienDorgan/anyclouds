package openstack_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type OSImageManagerTestSuite struct {
	tests.ImageManagerTestSuite
}

//SetupSuite set up image manager
func (suite *OSImageManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Mgr = p.GetImageManager()
}

func TestOSImageManagerTestSuite(t *testing.T) {
	suite.Run(t, new(OSImageManagerTestSuite))
}
