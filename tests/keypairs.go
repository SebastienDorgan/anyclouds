package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"golang.org/x/crypto/ssh"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

//KeyPairManagerTestSuite test suite off api.KeyPairManagers
type KeyPairManagerTestSuite struct {
	suite.Suite
	Mgr api.KeyPairManager
}

//TestKeyPairManager Canonical test for KeyPairManager implementation
func (s *KeyPairManagerTestSuite) TestKeyPairManager(t *testing.T) {
	s.Mgr.Delete("pktest")

	keypairs, err := s.Mgr.List()
	assert.NoError(t, err)
	nkeys := len(keypairs)

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := privateKey.PublicKey
	pub, _ := ssh.NewPublicKey(&publicKey)
	pubBytes := ssh.MarshalAuthorizedKey(pub)

	err = s.Mgr.Load("pktest", pubBytes)
	assert.NoError(t, err)

	keypairs, err = s.Mgr.List()
	assert.NoError(t, err)
	assert.Equal(t, nkeys+1, len(keypairs))

	found := false
	for _, kp := range keypairs {
		if kp == "pktest" {
			found = true
			break
		}
	}
	assert.True(t, found)

	err = s.Mgr.Delete("pktest")
	assert.NoError(t, err)
	assert.Equal(t, nkeys, len(keypairs))

	found = false
	for _, kp := range keypairs {
		if kp == "pktest" {
			found = true
			break
		}
	}
	assert.False(t, found)
}
