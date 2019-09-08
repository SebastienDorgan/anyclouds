package azure_test

import (
	"github.com/SebastienDorgan/anyclouds/providers/azure"
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type AZTemplateManagerTestSuite struct {
	tests.TemplateManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AZTemplateManagerTestSuite) SetupSuite() {
	p := azure.GetProvider()
	suite.Mgr =
		p.GetTemplateManager()
}

func TestAZTemplateTestSuite(t *testing.T) {
	suite.Run(t, new(AZTemplateManagerTestSuite))
}
