package aws_test

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/SebastienDorgan/anyclouds/providers/aws"
	"github.com/stretchr/testify/assert"
)

//TestCreate create AWS provider
func TestCreate(t *testing.T) {
	var provider aws.Provider = aws.Provider{}
	user, _ := user.Current()
	path := filepath.Join(user.HomeDir, ".anyclouds/aws_test.json")
	file, err := os.Open(path)
	assert.NoError(t, err)
	provider.Init(file, "json")
	images, err := provider.GetImageManager().List()
	assert.NoError(t, err)
	for _, img := range images {
		fmt.Println(img.Name, img.CreatedAt)
	}
}