package aws_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type AWSTemplateManagerTestSuite struct {
	tests.TemplateManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AWSTemplateManagerTestSuite) SetupSuite() {
	p, _ := GetProvider()
	suite.Mgr =
		p.GetTemplateManager()
}

func TestAWSTemplateTestSuite(t *testing.T) {
	suite.Run(t, new(AWSTemplateManagerTestSuite))
}
