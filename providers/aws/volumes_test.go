package aws_test

import (
	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AWSVolumeManagerTestSuite struct {
	tests.VolumeManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AWSVolumeManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Prov = p
}

func TestAWSVolumeManagerTestSuite(t *testing.T) {
	suite.Run(t, new(AWSVolumeManagerTestSuite))
}
