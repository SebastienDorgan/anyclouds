package tests

import (
	"github.com/SebastienDorgan/anyclouds/sshutils"

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
	kp, err := sshutils.CreateKeyPair(2048)
	assert.NoError(s.T(), err)

	err = s.Mgr.Import("pktest", kp.PublicKey)
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
