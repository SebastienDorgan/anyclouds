package openstack

import (
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/pkg/errors"
)

//KeyPairManager openstack implementation of api.KeyPairManager
type KeyPairManager struct {
	OpenStack *Provider
}

//Load load a public key
func (mgr *KeyPairManager) Load(name string, publicKey []byte) error {
	_, err := keypairs.Create(mgr.OpenStack.Compute, keypairs.CreateOpts{
		Name:      name,
		PublicKey: string(publicKey),
	}).Extract()
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error listing images")
	}
	return nil
}

//List available keys
func (mgr *KeyPairManager) List() ([]string, error) {
	pages, err := keypairs.List(mgr.OpenStack.Compute).AllPages()
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing images")
	}
	kps, err := keypairs.ExtractKeyPairs(pages)
	if err != nil {
		return nil, errors.Wrap(ProviderError(err), "Error listing images")
	}
	result := []string{}
	for _, kp := range kps {
		result = append(result, kp.Name)
	}
	return result, nil
}

//Delete a key pair
func (mgr *KeyPairManager) Delete(name string) error {
	err := keypairs.Delete(mgr.OpenStack.Compute, name).ExtractErr()
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error listing images")
	}
	return nil
}
