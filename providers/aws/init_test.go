package aws_test

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/anyclouds/providers/aws"
)

func GetProvider() api.Provider {
	user, _ := user.Current()
	path := filepath.Join(user.HomeDir, ".anyclouds/aws_test.json")
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	provider := &aws.Provider{}
	provider.Init(file, "json")
	return provider
}
