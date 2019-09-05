package openstack_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type OSTemplateManagerTestSuite struct {
	tests.TemplateManagerTestSuite
}

//SetupSuite set up image manager
func (suite *OSTemplateManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Mgr =
		p.GetTemplateManager()
}

func TestOSTemplateTestSuite(t *testing.T) {
	suite.Run(t, new(OSTemplateManagerTestSuite))
}
