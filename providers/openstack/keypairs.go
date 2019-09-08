package openstack

import (
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/pkg/errors"
)

//KeyPairManager openstack implementation of api.KeyPairManager
type KeyPairManager struct {
	OpenStack *Provider
}

//Import load a public key
func (mgr *KeyPairManager) Import(name string, publicKey []byte) error {
	_, err := keypairs.Create(mgr.OpenStack.Compute, keypairs.CreateOpts{
		Name:      name,
		PublicKey: string(publicKey),
	}).Extract()
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error listing images")
	}
	return nil
}

//Delete a key pair
func (mgr *KeyPairManager) Delete(name string) error {
	err := keypairs.Delete(mgr.OpenStack.Compute, name).ExtractErr()
	if err != nil {
		return errors.Wrap(ProviderError(err), "Error listing images")
	}
	return nil
}
