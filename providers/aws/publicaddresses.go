package aws

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type PublicIPManager struct {
	Provider *Provider
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

func (mgr *PublicIPManager) ListAvailablePools() ([]api.PublicIPPool, api.ListAvailablePublicIPPoolsError) {
	var pools []api.PublicIPPool
	out, err := mgr.Provider.AWSServices.EC2Client.DescribePublicIpv4Pools(&ec2.DescribePublicIpv4PoolsInput{})
	if err != nil {
		return nil, api.NewListAvailablePublicIPPoolsError(err)
	}
	for _, pool := range out.PublicIpv4Pools {
		pools = append(pools, api.PublicIPPool{
			ID:     *pool.PoolId,
			Ranges: addressRanges(pool),
		})
	}
	for out.NextToken != nil {
		out, err = mgr.Provider.AWSServices.EC2Client.DescribePublicIpv4Pools(&ec2.DescribePublicIpv4PoolsInput{
			NextToken: out.NextToken,
		})
		if err != nil {
			return nil, api.NewListAvailablePublicIPPoolsError(err)
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
func createListPublicIPAddressFilters(options *api.ListPublicIPsOptions) []*ec2.Filter {
	if options == nil {
		return nil
	}
	if options.ServerID == nil {
		return nil
	}
	return []*ec2.Filter{
		{
			Name:   aws.String("instance-id"),
			Values: []*string{options.ServerID},
		},
	}
}

func (mgr *PublicIPManager) List(options *api.ListPublicIPsOptions) ([]api.PublicIP, api.ListPublicIPsError) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeAddresses(&ec2.DescribeAddressesInput{
		Filters: createListPublicIPAddressFilters(options),
	})
	if err != nil {
		return nil, api.NewListPublicIPsError(err, options)
	}
	var addresses []api.PublicIP
	for _, addr := range out.Addresses {
		a, err := toPublicIP(addr)
		if err != nil {
			return nil, api.NewListPublicIPsError(err, options)
		}
		addresses = append(addresses, *a)
	}
	return addresses, nil
}

func toAllocateAddressInput(options *api.CreatePublicIPOptions) *ec2.AllocateAddressInput {
	return &ec2.AllocateAddressInput{
		Address:        options.IPAddress,
		Domain:         aws.String("vpc"),
		PublicIpv4Pool: options.IPAddressPoolID,
	}

}
func (mgr *PublicIPManager) Create(options api.CreatePublicIPOptions) (*api.PublicIP, api.CreatePublicIPError) {
	out, err := mgr.Provider.AWSServices.EC2Client.AllocateAddress(toAllocateAddressInput(&options))
	if err != nil {
		return nil, api.NewCreatePublicIPError(err, options)
	}
	err = mgr.Provider.AddTags(*out.AllocationId, map[string]string{"name": options.Name})
	if err != nil {
		err2 := mgr.Delete(*out.AllocationId)
		err = api.NewErrorStackFromError(err, err2)
		return nil, api.NewCreatePublicIPError(err, options)
	}
	return &api.PublicIP{
		Name:    options.Name,
		ID:      *out.AllocationId,
		Address: *out.PublicIp,
	}, nil
}
func (mgr *PublicIPManager) Associate(options api.AssociatePublicIPOptions) api.AssociatePublicIPError {
	input, err := mgr.toAssociatedAddressInput(&options)
	if err != nil {
		return api.NewAssociatePublicIPError(err, options)
	}
	_, err = mgr.Provider.AWSServices.EC2Client.AssociateAddress(input)
	return api.NewAssociatePublicIPError(err, options)
}

func (mgr *PublicIPManager) toAssociatedAddressInput(options *api.AssociatePublicIPOptions) (*ec2.AssociateAddressInput, error) {
	var privateIP *string
	if len(options.PrivateIP) > 0 {
		privateIP = aws.String(options.PrivateIP)
	}
	var networkInterface *string
	if len(options.SubnetID) > 0 {
		ni, err := mgr.getNetworkInterface(options, privateIP)
		if err != nil {
			return nil, err
		}
		networkInterface = ni.NetworkInterfaceId
	}
	return &ec2.AssociateAddressInput{
		AllocationId:       aws.String(options.PublicIPId),
		InstanceId:         aws.String(options.ServerID),
		NetworkInterfaceId: networkInterface,
		PrivateIpAddress:   privateIP,
	}, nil
}

func (mgr *PublicIPManager) getNetworkInterface(options *api.AssociatePublicIPOptions, privateIP *string) (*ec2.NetworkInterface, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
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
	if privateIP != nil {
		for _, ni := range out.NetworkInterfaces {
			if *ni.PrivateIpAddress == *privateIP {
				networkInterface = ni
				break
			}
		}
	}
	return networkInterface, nil
}
func (mgr *PublicIPManager) Dissociate(publicIPID string) api.DissociatePublicIPError {
	addr, err := mgr.getAddress(publicIPID)
	if err != nil {
		return api.NewDissociatePublicIPError(err, publicIPID)
	}
	_, err = mgr.Provider.AWSServices.EC2Client.DisassociateAddress(&ec2.DisassociateAddressInput{
		AssociationId: addr.AssociationId,
	})
	return api.NewDissociatePublicIPError(err, publicIPID)
}
func (mgr *PublicIPManager) Delete(publicIPId string) api.DeletePublicIPError {
	_, err := mgr.Provider.AWSServices.EC2Client.ReleaseAddress(&ec2.ReleaseAddressInput{
		AllocationId: aws.String(publicIPId),
	})
	return api.NewDeletePublicIPError(err, publicIPId)
}

func (mgr *PublicIPManager) Get(publicIPId string) (*api.PublicIP, api.GetPublicIPError) {
	addr, err := mgr.getAddress(publicIPId)
	if err != nil {
		return nil, api.NewGetPublicIPError(err, publicIPId)
	}
	ip, err := toPublicIP(addr)
	return ip, api.NewGetPublicIPError(err, publicIPId)

}

func toPublicIP(addr *ec2.Address) (*api.PublicIP, error) {
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
		ID:                 *addr.AllocationId,
		Name:               name,
		Address:            *addr.PublicIp,
		NetworkInterfaceID: *addr.NetworkInterfaceId,
		PrivateAddress:     privateAddress,
	}, nil
}

func (mgr *PublicIPManager) getAddress(publicIPId string) (*ec2.Address, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeAddresses(&ec2.DescribeAddressesInput{
		AllocationIds: []*string{aws.String(publicIPId)},
	})
	if err != nil {
		return nil, err
	}
	return out.Addresses[0], nil
}
