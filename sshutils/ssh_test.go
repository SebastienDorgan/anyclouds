package sshutils_test

import (
	"github.com/SebastienDorgan/anyclouds/sshutils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestCreateKeyPair(t *testing.T) {
	kp, err := sshutils.CreateKeyPair(4096)
	assert.NoError(t, err)
	_ = ioutil.WriteFile("public_key", kp.PublicKey, 0644)
	_ = ioutil.WriteFile("private_key", kp.PrivateKey, 0644)
}
