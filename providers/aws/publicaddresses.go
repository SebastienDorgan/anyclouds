package aws

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

type PublicIPAddressManager struct {
	AWS *Provider
}

func addressRanges(pool *ec2.PublicIpv4Pool) []api.AddressRange {
	var ranges []api.AddressRange
	for _, r := range pool.PoolAddressRanges {
		ranges = append(ranges, api.AddressRange{
			FirstAddress: *r.FirstAddress,
			LastAddress:  *r.LastAddress,
		})
	}
	return ranges
}

func (mgr *PublicIPAddressManager) ListAvailablePools() ([]api.PublicIPPool, error) {
	var pools []api.PublicIPPool
	out, err := mgr.AWS.EC2Client.DescribePublicIpv4Pools(&ec2.DescribePublicIpv4PoolsInput{})
	if err != nil {
		return nil, errors.Wrap(err, "error listing public ip pools")
	}
	for _, pool := range out.PublicIpv4Pools {
		pools = append(pools, api.PublicIPPool{
			ID:     *pool.PoolId,
			Ranges: addressRanges(pool),
		})
	}
	for out.NextToken != nil {
		out, err = mgr.AWS.EC2Client.DescribePublicIpv4Pools(&ec2.DescribePublicIpv4PoolsInput{
			NextToken: out.NextToken,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error listing public ip pools")
		}
		for _, pool := range out.PublicIpv4Pools {
			pools = append(pools, api.PublicIPPool{
				ID:     *pool.PoolId,
				Ranges: addressRanges(pool),
			})
		}
	}
	return pools, nil
}
func (mgr *PublicIPAddressManager) ListAllocated() ([]api.PublicIP, error) {
	out, err := mgr.AWS.EC2Client.DescribeAddresses(&ec2.DescribeAddressesInput{})
	if err != nil {
		return nil, errors.Wrap(err, "error listing public ips")
	}
	var addresses []api.PublicIP
	for _, addr := range out.Addresses {
		a, err := mgr.toPublicIP(addr)
		if err != nil {
			return nil, errors.Wrap(err, "error listing public ips")
		}
		addresses = append(addresses, *a)
	}
	return addresses, nil
}

func toAllocateAddressInput(options *api.PublicIPAllocationOptions) *ec2.AllocateAddressInput {
	var address *string
	if len(options.Address) > 0 {
		address = aws.String(options.Address)
	}
	var addressPool *string
	if len(options.AddressPool) > 0 {
		addressPool = aws.String(options.AddressPool)
	}
	return &ec2.AllocateAddressInput{
		Address:        address,
		Domain:         aws.String("vpc"),
		PublicIpv4Pool: addressPool,
	}

}
func (mgr *PublicIPAddressManager) Allocate(options *api.PublicIPAllocationOptions) (*api.PublicIP, error) {
	out, err := mgr.AWS.EC2Client.AllocateAddress(toAllocateAddressInput(options))
	if err != nil {
		return nil, errors.Wrapf(err, "error allocating public ip %s", options.Name)
	}
	err = mgr.AWS.AddTags(*out.AllocationId, map[string]string{"name": options.Name})
	if err != nil {
		_ = mgr.Release(*out.AllocationId)
		return nil, errors.Wrapf(err, "error allocating public ip %s", options.Name)
	}
	return &api.PublicIP{
		Name:    options.Name,
		ID:      *out.AllocationId,
		Address: *out.PublicIp,
	}, nil
}
func (mgr *PublicIPAddressManager) Associate(options *api.PublicIPAssociationOptions) error {
	input, err := mgr.toAssociatedAddressInput(options)
	if err != nil {
		return errors.Wrapf(err, "error associating public ip %s with server %s", options.PublicIPId, options.ServerID)
	}
	_, err = mgr.AWS.EC2Client.AssociateAddress(input)
	return errors.Wrapf(err, "error associating public ip %s with server %s", options.PublicIPId, options.ServerID)
}

func (mgr *PublicIPAddressManager) toAssociatedAddressInput(options *api.PublicIPAssociationOptions) (*ec2.AssociateAddressInput, error) {
	var privateIp *string
	if len(options.PrivateIP) > 0 {
		privateIp = aws.String(options.PrivateIP)
	}
	var networkInterface *string
	if len(options.SubnetID) > 0 {
		ni, err := mgr.getNetworkInterface(options, privateIp)
		if err != nil {
			return nil, err
		}
		networkInterface = ni.NetworkInterfaceId
	}
	return &ec2.AssociateAddressInput{
		AllocationId:       aws.String(options.PublicIPId),
		InstanceId:         aws.String(options.ServerID),
		NetworkInterfaceId: networkInterface,
		PrivateIpAddress:   privateIp,
	}, nil
}

func (mgr *PublicIPAddressManager) getNetworkInterface(options *api.PublicIPAssociationOptions, privateIp *string) (*ec2.NetworkInterface, error) {
	out, err := mgr.AWS.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		DryRun: nil,
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.instance-id"),
				Values: []*string{aws.String(options.ServerID)},
			},
			{
				Name:   aws.String("subnet-id"),
				Values: []*string{aws.String(options.SubnetID)},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	networkInterface := out.NetworkInterfaces[0]
	if privateIp != nil {
		for _, ni := range out.NetworkInterfaces {
			if *ni.PrivateIpAddress == *privateIp {
				networkInterface = ni
				break
			}
		}
	}
	return networkInterface, nil
}
func (mgr *PublicIPAddressManager) Dissociate(publicIpID string) error {
	addr, err := mgr.getAddress(publicIpID)
	if err != nil {
		return errors.Wrapf(err, "error disassociating public ip %s", publicIpID)
	}
	_, err = mgr.AWS.EC2Client.DisassociateAddress(&ec2.DisassociateAddressInput{
		AssociationId: addr.AssociationId,
	})
	return errors.Wrapf(err, "error disassociating public ip %s", publicIpID)
}
func (mgr *PublicIPAddressManager) Release(publicIPId string) error {
	_, err := mgr.AWS.EC2Client.ReleaseAddress(&ec2.ReleaseAddressInput{
		AllocationId: aws.String(publicIPId),
	})
	return errors.Wrapf(err, "error releasing public ip %s", publicIPId)
}

func (mgr *PublicIPAddressManager) Get(publicIPId string) (*api.PublicIP, error) {
	addr, err := mgr.getAddress(publicIPId)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting public ip %s", publicIPId)
	}
	ip, err := mgr.toPublicIP(addr)
	return ip, errors.Wrapf(err, "error getting public ip %s", publicIPId)

}

func (mgr *PublicIPAddressManager) toPublicIP(addr *ec2.Address) (*api.PublicIP, error) {
	var serverID string
	var subnetID string
	if addr.InstanceId != nil {
		serverID = *addr.InstanceId
		out, err := mgr.AWS.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{addr.NetworkInterfaceId},
		})
		if err != nil {
			return nil, err
		}
		if len(out.NetworkInterfaces) == 0 {
			return nil, errors.Errorf("error getting public ip %s, no network interface assocaited to server %s", *addr.AllocationId, *addr.InstanceId)
		}
		subnetID = *out.NetworkInterfaces[0].SubnetId
	}
	var name string
	for _, t := range addr.Tags {
		if *t.Key == "name" {
			name = *t.Value
			break
		}
	}
	var privateAddress string
	if addr.PrivateIpAddress != nil {
		privateAddress = *addr.PrivateIpAddress
	}

	return &api.PublicIP{
		ID:        *addr.AllocationId,
		Name:      name,
		Address:   *addr.PublicIp,
		ServerID:  serverID,
		SubnetID:  subnetID,
		PrivateIP: privateAddress,
	}, nil
}

func (mgr *PublicIPAddressManager) getAddress(publicIPId string) (*ec2.Address, error) {
	out, err := mgr.AWS.EC2Client.DescribeAddresses(&ec2.DescribeAddressesInput{
		AllocationIds: []*string{aws.String(publicIPId)},
	})
	if err != nil {
		return nil, err
	}
	return out.Addresses[0], nil
}
