package aws

import (
	"fmt"
	"sort"

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
		return nil, errors.Wrapf(err, "error creating network %s", options.Name)
	}

	_, err = mgr.AWS.EC2Client.CreateTags(&ec2.CreateTagsInput{
		DryRun: aws.Bool(false),
		Resources: []*string{
			out.Vpc.VpcId,
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("name"),
				Value: aws.String(options.Name),
			},
		},
	})
	if err != nil {
		_ = mgr.DeleteNetwork(*out.Vpc.VpcId)
		return nil, errors.Wrapf(err, "Error creating network %s", options.Name)
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
	return errors.Wrapf(err, "error deleting network %s", id)

}

func network(v *ec2.Vpc) *api.Network {
	net := &api.Network{
		ID:   *v.VpcId,
		CIDR: *v.CidrBlock,
	}
	n := sort.Search(len(v.Tags), func(i int) bool {
		return *v.Tags[i].Key == "name"
	})
	if n < len(v.Tags) {
		net.Name = *v.Tags[n].Value
	}
	return net
}

//ListNetworks lists networks
func (mgr *NetworkManager) ListNetworks() ([]api.Network, error) {
	out, err := mgr.AWS.EC2Client.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, errors.Wrap(err, "Error listing network")
	}
	var result []api.Network
	for _, vpc := range out.Vpcs {
		result = append(result, *network(vpc))
	}
	return result, nil
}

//GetNetwork returns the configuration of the network identified by id
func (mgr *NetworkManager) GetNetwork(id string) (*api.Network, error) {
	out, err := mgr.AWS.EC2Client.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: []*string{&id},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error getting network %s", id)
	}
	if len(out.Vpcs) > 0 {
		return network(out.Vpcs[0]), nil
	}
	return nil, errors.Wrapf(err, "network %s not found", id)

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
	n := sort.Search(len(s.Tags), func(i int) bool {
		return *s.Tags[i].Key == "name"
	})
	if n < len(s.Tags) {
		sn.Name = *s.Tags[n].Value
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
	_, err = mgr.AWS.EC2Client.CreateTags(&ec2.CreateTagsInput{
		DryRun: aws.Bool(false),
		Resources: []*string{
			out.Subnet.SubnetId,
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("name"),
				Value: aws.String(options.Name),
			},
		},
	})
	if err != nil {
		_ = mgr.DeleteSubnet(*out.Subnet.SubnetId)
		return nil, errors.Wrapf(err, "Error creating subnet %s", options.Name)
	}
	return mgr.GetSubnet(*out.Subnet.SubnetId)
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
		return nil, errors.Wrap(err, "error listing subnets")
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
		return nil, errors.Wrapf(err, "error getting subnet %s", id)
	}
	for _, sn := range out.Subnets {
		if id == *sn.SubnetId {
			return subnet(sn), nil
		}
	}
	return nil, fmt.Errorf("subnet %s not found", id)
}
