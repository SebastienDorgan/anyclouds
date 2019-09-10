package sshutils_test

import (
	"fmt"
	"github.com/SebastienDorgan/anyclouds/sshutils"
	"github.com/sethvargo/go-password/password"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateKeyPair(t *testing.T) {
	kp, err := sshutils.CreateKeyPair(4096)
	assert.NoError(t, err)
	assert.NotNil(t, kp.PublicKey)
	assert.NotNil(t, kp.PrivateKey)
	fmt.Println(len(kp.PrivateKey), len(kp.PublicKey))
	fmt.Println(password.MustGenerate(16, 5, 5, false, false))
}
