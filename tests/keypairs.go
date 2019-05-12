package tests

import (
	"crypto/rand"
	"crypto/rsa"

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
func (s *KeyPairManagerTestSuite) TestKeyPairManager() {
	_ = s.Mgr.Delete("pktest")

	keypairs, err := s.Mgr.List()
	assert.NoError(s.T(), err)
	keysLen := len(keypairs)

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := privateKey.PublicKey
	pub, _ := ssh.NewPublicKey(&publicKey)
	pubBytes := ssh.MarshalAuthorizedKey(pub)

	err = s.Mgr.Load("pktest", pubBytes)
	assert.NoError(s.T(), err)

	keypairs, err = s.Mgr.List()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), keysLen+1, len(keypairs))

	found := false
	for _, kp := range keypairs {
		if kp.Name == "pktest" {
			found = true
			break
		}
	}
	assert.True(s.T(), found)

	err = s.Mgr.Delete("pktest")
	assert.NoError(s.T(), err)

	keypairs, err = s.Mgr.List()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), keysLen, len(keypairs))

	found = false
	for _, kp := range keypairs {
		if kp.Name == "pktest" {
			found = true
			break
		}
	}
	assert.False(s.T(), found)
}
