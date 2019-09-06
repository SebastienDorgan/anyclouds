package openstack

import "github.com/SebastienDorgan/anyclouds/api"

//NetworkManager defines networking functions a anyclouds provider must provide
type PublicIPAddressManager struct {
	OpenStack *Provider
}

func (mgr *PublicIPAddressManager) ListAvailablePools() ([]api.PublicIPPool, error) {
	//TODO
	panic("implement me")
	return nil, nil
}
func (mgr *PublicIPAddressManager) ListAllocated() ([]api.PublicIP, error) {
	//TODO
	panic("implement me")
	return nil, nil
}
func (mgr *PublicIPAddressManager) Allocate(options *api.PublicIPAllocationOptions) (*api.PublicIP, error) {
	panic("implement me")
	//TODO
	return nil, nil
}
func (mgr *PublicIPAddressManager) Associate(options *api.PublicIPAssociationOptions) error {
	//TODO
	panic("implement me")
	return nil
}
func (mgr *PublicIPAddressManager) Dissociate(publicIpID string) error {
	panic("implement me")
	//TODO
	return nil
}
func (mgr *PublicIPAddressManager) Release(publicIPId string) error {
	panic("implement me")
	//TODO
	return nil
}
func (mgr *PublicIPAddressManager) Get(publicIPId string) (*api.PublicIP, error) {
	panic("implement me")
	//TODO
	return nil, nil
}
