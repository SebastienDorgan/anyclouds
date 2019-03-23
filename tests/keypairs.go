package tests

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"

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
	s.Mgr.Delete("pktest")

	keypairs, err := s.Mgr.List()
	assert.NoError(s.T(), err)
	nkeys := len(keypairs)

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := privateKey.PublicKey
	pub, _ := ssh.NewPublicKey(&publicKey)
	pubBytes := ssh.MarshalAuthorizedKey(pub)
	hash := md5.New()
	hash.Write(x509.MarshalPKCS1PrivateKey(privateKey))
	finger := hex.EncodeToString(hash.Sum(nil))

	err = s.Mgr.Load("pktest", pubBytes)
	assert.NoError(s.T(), err)

	keypairs, err = s.Mgr.List()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), nkeys+1, len(keypairs))

	found := false
	ftest := ""
	for _, kp := range keypairs {
		if kp.Name == "pktest" {
			found = true
			ftest = kp.Fingerprint
			break
		}
	}
	assert.True(s.T(), found)
	assert.Equal(s.T(), finger, ftest)

	err = s.Mgr.Delete("pktest")
	assert.NoError(s.T(), err)

	keypairs, err = s.Mgr.List()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), nkeys, len(keypairs))

	found = false
	for _, kp := range keypairs {
		if kp.Name == "pktest" {
			found = true
			break
		}
	}
	assert.False(s.T(), found)
}
