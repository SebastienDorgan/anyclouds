package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"golang.org/x/crypto/ssh"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/stretchr/testify/assert"
)

//KeyPairManager an api.KeyPairManager implementation
var KeyPairManager api.KeyPairManager

//TestKeyPairManager Canonical test for KeyPairManager implementation
func TestKeyPairManager(t *testing.T) {
	KeyPairManager.Delete("pktest")

	keypairs, err := KeyPairManager.List()
	assert.NoError(t, err)
	nkeys := len(keypairs)

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := privateKey.PublicKey
	pub, _ := ssh.NewPublicKey(&publicKey)
	pubBytes := ssh.MarshalAuthorizedKey(pub)

	err = KeyPairManager.Load("pktest", pubBytes)
	assert.NoError(t, err)

	keypairs, err = KeyPairManager.List()
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

	err = KeyPairManager.Delete("pktest")
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
