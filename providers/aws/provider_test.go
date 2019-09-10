package aws_test

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/SebastienDorgan/anyclouds/providers/aws"
	"github.com/stretchr/testify/assert"
)

func GetProvider() *aws.Provider {
	var provider aws.Provider
	usr, _ := user.Current()
	path := filepath.Join(usr.HomeDir, ".anyclouds/aws_test.json")
	file, err := os.Open(path)
	err = provider.Init(file, "json")
	if err != nil {
		return nil
	}
	return &provider
}

//TestCreate create Provider provider
func TestCreate(t *testing.T) {
	var provider aws.Provider
	usr, _ := user.Current()
	path := filepath.Join(usr.HomeDir, ".anyclouds/aws_test.json")
	file, err := os.Open(path)
	assert.NoError(t, err)
	err = provider.Init(file, "json")
	assert.NoError(t, err)
	images, err := provider.GetImageManager().List()
	assert.NoError(t, err)
	assert.True(t, len(images) > 0)
}
