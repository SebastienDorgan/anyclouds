package aws_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type AWSKeyPairManagerTestSuite struct {
	tests.KeyPairManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AWSKeyPairManagerTestSuite) SetupSuite() {
	suite.Mgr = GetProvider().GetKeyPairManager()
}

func TestAWSKeyPairManagerTestSuite(t *testing.T) {
	suite.Run(t, new(AWSKeyPairManagerTestSuite))
}
