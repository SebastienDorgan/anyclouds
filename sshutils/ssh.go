package sshutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
)

//SSHConfig defines a ssh configuration going throw a series of proxies
type SSHConfig struct {
	Addr         string
	ClientConfig *ssh.ClientConfig
	Proxy        *SSHConfig
}

//CreateClient creates ssh.Client using a SSHConfig
func CreateClient(cfg *SSHConfig) (*ssh.Client, error) {
	if cfg.Proxy == nil {
		return ssh.Dial("tcp", cfg.Addr, cfg.ClientConfig)
	}

	client, err := CreateClient(cfg.Proxy)
	if err != nil {
		return nil, err
	}
	conn, err := client.Dial("tcp", cfg.Addr)
	if err != nil {
		return nil, err
	}
	c, channels, reqs, err := ssh.NewClientConn(conn, cfg.Addr, cfg.ClientConfig)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, channels, reqs), nil

}

//KeyPair a key pair
type KeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

// CreateKeyPair creates a key pair using bitsize bits
func CreateKeyPair(bitsize int) (pair *KeyPair, err error) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, bitsize)
	publicKey := privateKey.PublicKey
	pub, err := ssh.NewPublicKey(&publicKey)
	if err != nil {
		return nil, err
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(pub)
	priBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: priBytes,
		},
	)
	return &KeyPair{
		PublicKey:  publicKeyBytes,
		PrivateKey: privateKeyBytes,
	}, nil
}

//AuthMethod returns the ssh.AuthMethod corresponding to this KeyPair
func (kp *KeyPair) AuthMethod() (ssh.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKey(kp.PrivateKey)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}
