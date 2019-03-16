package tests

import (
	"reflect"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/SebastienDorgan/anyclouds/api"
)

//ImageManagerTestSuite test suite for api.ImageManager
type ImageManagerTestSuite struct {
	suite.Suite
	Mgr api.ImageManager
}

//TestImageManager Canonical test for ImageManager implementation
func (s *ImageManagerTestSuite) TestImageManager() {
	images, err := s.Mgr.List()
	assert.NoError(s.T(), err)
	for _, img := range images {
		image, err := s.Mgr.Get(img.ID)
		assert.NoError(s.T(), err)
		assert.True(s.T(), reflect.DeepEqual(img, *image))
	}
}
