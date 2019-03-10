package tests

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/SebastienDorgan/anyclouds/api"
)

//ImageManager an api.ImageManager implementation
var ImageManager api.ImageManager

//TestImageManager Canonical test for ImageManager implementation
func TestImageManager(t *testing.T) {
	images, err := ImageManager.List()
	assert.NoError(t, err)
	for _, img := range images {
		image, err := ImageManager.Get(img.ID)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(img, image))
	}
}
