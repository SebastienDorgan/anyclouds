package aws_test

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/anyclouds/providers/aws"
)

func GetProvider() (api.Provider, error) {
	usr, _ := user.Current()
	path := filepath.Join(usr.HomeDir, ".anyclouds/aws_test.json")
	file, err := os.Open(path)
	if err != nil {
		return nil, nil
	}
	provider := &aws.Provider{}
	err = provider.Init(file, "json")
	return provider, err
}
