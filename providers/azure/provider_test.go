package azure_test

import (
	"github.com/SebastienDorgan/anyclouds/providers/azure"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetProvider() *azure.Provider {
	var provider azure.Provider
	usr, _ := user.Current()
	path := filepath.Join(usr.HomeDir, ".anyclouds/azure.json")
	file, err := os.Open(path)
	err = provider.Init(file, "json")
	if err != nil {
		return nil
	}
	return &provider
}

//TestCreate create AWS provider
func TestCreate(t *testing.T) {
	var provider azure.Provider
	usr, _ := user.Current()
	path := filepath.Join(usr.HomeDir, ".anyclouds/azure.json")
	file, err := os.Open(path)
	assert.NoError(t, err)
	err = provider.Init(file, "json")
	assert.NoError(t, err)
	images, err := provider.GetImageManager().List()
	assert.NoError(t, err)
	assert.True(t, len(images) > 0)
}
