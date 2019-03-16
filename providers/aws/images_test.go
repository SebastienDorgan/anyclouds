package aws_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type AWSImageManagerTestSuite struct {
	tests.ImageManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AWSImageManagerTestSuite) SetupSuite() {
	suite.Mgr = GetProvider().GetImageManager()
}

func TestAWSImageManagerTestSuite(t *testing.T) {
	suite.Run(t, new(AWSImageManagerTestSuite))
}
