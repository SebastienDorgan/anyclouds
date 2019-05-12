package aws

import (
	"fmt"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

//NetworkManager defines networking functions a anyclouds provider must provide
type NetworkManager struct {
	AWS *Provider
}

//CreateNetwork creates a network
func (mgr *NetworkManager) CreateNetwork(options *api.NetworkOptions) (*api.Network, error) {
	out, err := mgr.AWS.EC2Client.CreateVpc(&ec2.CreateVpcInput{
		AmazonProvidedIpv6CidrBlock: aws.Bool(true),
		CidrBlock:                   &options.CIDR,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error creating network")
	}

	return &api.Network{
		ID:   *out.Vpc.VpcId,
		CIDR: *out.Vpc.CidrBlock,
	}, nil
}

//DeleteNetwork deletes the network identified by id
func (mgr *NetworkManager) DeleteNetwork(id string) error {
	_, err := mgr.AWS.EC2Client.DeleteVpc(&ec2.DeleteVpcInput{
		VpcId: &id,
	})
	return errors.Wrap(err, "Error deleting network")

}

//ListNetworks lists networks
func (mgr *NetworkManager) ListNetworks() ([]api.Network, error) {
	out, err := mgr.AWS.EC2Client.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, errors.Wrap(err, "Error listing network")
	}
	var result []api.Network
	for _, vpc := range out.Vpcs {
		result = append(result, api.Network{
			ID:   *vpc.VpcId,
			CIDR: *vpc.CidrBlock,
		})
	}
	return result, nil
}

//GetNetwork returns the configuration of the network identified by id
func (mgr *NetworkManager) GetNetwork(id string) (*api.Network, error) {
	out, err := mgr.AWS.EC2Client.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: []*string{&id},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error listing network")
	}

	if len(out.Vpcs) > 0 {
		return &api.Network{
			ID:   *out.Vpcs[0].VpcId,
			CIDR: *out.Vpcs[0].CidrBlock,
		}, nil
	}
	return nil, errors.Wrap(err, "Network not found")

}

func subnet(s *ec2.Subnet) *api.Subnet {
	sn := api.Subnet{
		ID:        *s.SubnetId,
		NetworkID: *s.VpcId,
	}
	if len(s.Ipv6CidrBlockAssociationSet) > 0 {
		sn.IPVersion = api.IPVersion6
		sn.CIDR = *s.Ipv6CidrBlockAssociationSet[0].Ipv6CidrBlock
	} else if s.CidrBlock != nil {
		sn.IPVersion = api.IPVersion4
		sn.CIDR = *s.CidrBlock
	}
	return &sn
}

//CreateSubnet creates a subnet
func (mgr *NetworkManager) CreateSubnet(options *api.SubnetOptions) (*api.Subnet, error) {

	input := ec2.CreateSubnetInput{
		VpcId: &options.NetworkID,
	}
	if options.IPVersion == api.IPVersion4 {
		input.CidrBlock = &options.CIDR
	} else if options.IPVersion == api.IPVersion6 {
		input.Ipv6CidrBlock = &options.CIDR
	}
	out, err := mgr.AWS.EC2Client.CreateSubnet(&input)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating subnet")
	}

	return subnet(out.Subnet), nil
}

//DeleteSubnet deletes the subnet identified by id
func (mgr *NetworkManager) DeleteSubnet(id string) error {
	_, err := mgr.AWS.EC2Client.DeleteSubnet(&ec2.DeleteSubnetInput{
		SubnetId: &id,
	})
	return errors.Wrap(err, "Error creating subnet")
}

//ListSubnets lists the subnet
func (mgr *NetworkManager) ListSubnets(networkID string) ([]api.Subnet, error) {
	out, err := mgr.AWS.EC2Client.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					&networkID,
				},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error listing subnets")
	}
	var result []api.Subnet
	for _, sn := range out.Subnets {
		result = append(result, *subnet(sn))
	}
	return result, nil
}

//GetSubnet returns the configuration of the subnet identified by id
func (mgr *NetworkManager) GetSubnet(id string) (*api.Subnet, error) {
	out, err := mgr.AWS.EC2Client.DescribeSubnets(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{&id},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error listing subnets")
	}
	for _, sn := range out.Subnets {
		if id == *sn.SubnetId {
			return subnet(sn), nil
		}
	}
	return nil, fmt.Errorf("subnet not found")
}
