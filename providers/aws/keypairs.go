package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

//KeyPairManager openstack implementation of api.KeyPairManager
type KeyPairManager struct {
	AWS *Provider
}

//Import load a public key
func (mgr *KeyPairManager) Import(name string, publicKey []byte) error {
	_, err := mgr.AWS.EC2Client.ImportKeyPair(&ec2.ImportKeyPairInput{
		DryRun:            aws.Bool(false),
		KeyName:           aws.String(name),
		PublicKeyMaterial: publicKey,
	})
	return errors.Wrap(err, "Error loading key pair")
}

//Delete a key pair
func (mgr *KeyPairManager) Delete(name string) error {
	_, err := mgr.AWS.EC2Client.DeleteKeyPair(&ec2.DeleteKeyPairInput{
		DryRun:  aws.Bool(false),
		KeyName: aws.String(name),
	})
	return errors.Wrap(err, "Error deleting key pair")
}
