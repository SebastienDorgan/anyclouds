package aws_test

import (
	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AWSServerManagerTestSuite struct {
	tests.ServerManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AWSServerManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Prov = p
}

func TestAWSServerManagerTestSuite(t *testing.T) {
	suite.Run(t, new(AWSServerManagerTestSuite))
}
