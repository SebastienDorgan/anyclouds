package aws

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

//KeyPairManager openstack implementation of api.KeyPairManager
type KeyPairManager struct {
	AWS *Provider
}

//Load load a public key
func (mgr *KeyPairManager) Load(name string, publicKey []byte) error {
	_, err := mgr.AWS.EC2Client.ImportKeyPair(&ec2.ImportKeyPairInput{
		DryRun:            aws.Bool(false),
		KeyName:           aws.String(name),
		PublicKeyMaterial: publicKey,
	})
	return errors.Wrap(err, "Error loading key pair")
}

//List available keys
func (mgr *KeyPairManager) List() ([]api.KeyPair, error) {
	out, err := mgr.AWS.EC2Client.DescribeKeyPairs(nil)
	if err != nil {
		return nil, errors.Wrap(err, "Error listing key pair")
	}
	var result []api.KeyPair
	for _, kp := range out.KeyPairs {
		result = append(result, api.KeyPair{
			Name:        *kp.KeyName,
			Fingerprint: *kp.KeyFingerprint,
		})
	}
	return result, nil
}

//Delete a key pair
func (mgr *KeyPairManager) Delete(name string) error {
	_, err := mgr.AWS.EC2Client.DeleteKeyPair(&ec2.DeleteKeyPairInput{
		DryRun:  aws.Bool(false),
		KeyName: aws.String(name),
	})
	return errors.Wrap(err, "Error deleting key pair")
}
