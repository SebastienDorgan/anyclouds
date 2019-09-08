package azure_test

import (
	"testing"

	"github.com/SebastienDorgan/anyclouds/tests"
	"github.com/stretchr/testify/suite"
)

type AZImageManagerTestSuite struct {
	tests.ImageManagerTestSuite
}

//SetupSuite set up image manager
func (suite *AZImageManagerTestSuite) SetupSuite() {
	p := GetProvider()
	suite.Mgr = p.GetImageManager()
}

func TestAZImageManagerTestSuite(t *testing.T) {
	suite.Run(t, new(AZImageManagerTestSuite))
}
