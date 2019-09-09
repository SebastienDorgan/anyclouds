package azure_test

import (
	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AZServerManagerTestSuite struct {
	tests.ServerManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AZServerManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Prov = p
}

func TestAZServerManagerTestSuite(t *testing.T) {
	suite.Run(t, new(AZServerManagerTestSuite))
}
