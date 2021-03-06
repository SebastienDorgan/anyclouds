package aws

import (
	"fmt"
	"sort"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//NetworkManager defines networking functions a anyclouds provider must provide
type NetworkManager struct {
	Provider *Provider
}

//CreateNetwork creates a network
func (mgr *NetworkManager) CreateNetwork(options api.CreateNetworkOptions) (*api.Network, api.CreateNetworkError) {
	n, err := mgr.createNetwork(options)
	return n, api.NewCreateNetworkError(err, options)
}

func (mgr *NetworkManager) createNetwork(options api.CreateNetworkOptions) (*api.Network, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.CreateVpc(&ec2.CreateVpcInput{
		AmazonProvidedIpv6CidrBlock: aws.Bool(true),
		CidrBlock:                   &options.CIDR,
	})
	if err != nil {
		return nil, err
	}

	err = mgr.Provider.AddTags(*out.Vpc.VpcId, map[string]string{"name": options.Name})
	if err != nil {
		err2 := mgr.DeleteNetwork(*out.Vpc.VpcId)
		return nil, api.NewErrorStackFromError(err, err2)
	}

	gw, err := mgr.addInternetGateway(out)
	if err != nil {
		err2 := mgr.DeleteNetwork(*out.Vpc.VpcId)
		return nil, api.NewErrorStackFromError(err, err2)
	}
	_, err = mgr.populateRouteTable(*out.Vpc.VpcId, gw)
	if err != nil {
		err2 := mgr.DeleteNetwork(*out.Vpc.VpcId)
		return nil, api.NewErrorStackFromError(err, err2)
	}
	return mgr.GetNetwork(*out.Vpc.VpcId)
}

func (mgr *NetworkManager) addInternetGateway(out *ec2.CreateVpcOutput) (*ec2.InternetGateway, error) {
	outGW, err := mgr.Provider.AWSServices.EC2Client.CreateInternetGateway(&ec2.CreateInternetGatewayInput{})
	if err != nil {
		return nil, err
	}
	_, err = mgr.Provider.AWSServices.EC2Client.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		DryRun:            nil,
		InternetGatewayId: outGW.InternetGateway.InternetGatewayId,
		VpcId:             out.Vpc.VpcId,
	})
	return outGW.InternetGateway, err
}

func (mgr *NetworkManager) removeInternetGateway(id string) error {
	out, err := mgr.getInternetGateway(id)
	if err != nil {
		return err
	}
	for _, gw := range out.InternetGateways {
		_, err = mgr.Provider.AWSServices.EC2Client.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
			InternetGatewayId: gw.InternetGatewayId,
			VpcId:             aws.String(id),
		})
		if err != nil {
			return err
		}
		_, err = mgr.Provider.AWSServices.EC2Client.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: gw.InternetGatewayId,
		})
		if err != nil {
			return err
		}
	}
	return nil

}

func (mgr *NetworkManager) getInternetGateway(vpcID string) (*ec2.DescribeInternetGatewaysOutput, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.vpc-id"),
				Values: []*string{aws.String(vpcID)},
			},
		},
	})
	return out, err
}

//DeleteNetwork deletes the network identified by id
func (mgr *NetworkManager) DeleteNetwork(id string) api.DeleteNetworkError {
	return api.NewDeleteNetworkError(mgr.deleteNetwork(id), id)
}

func (mgr *NetworkManager) deleteNetwork(id string) error {
	err := mgr.removeInternetGateway(id)
	if err != nil {
		return err
	}

	_, err = mgr.Provider.AWSServices.EC2Client.DeleteVpc(&ec2.DeleteVpcInput{
		VpcId: &id,
	})
	return err
}

func (mgr *NetworkManager) getRouteTable(networkID string) (*ec2.RouteTable, error) {
	tables, err := mgr.Provider.AWSServices.EC2Client.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		DryRun: nil,
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(networkID)},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(tables.RouteTables) == 0 {
		return nil, fmt.Errorf("route table of network %s not found", networkID)
	}
	return tables.RouteTables[0], err
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
func (mgr *NetworkManager) ListNetworks() ([]api.Network, api.ListNetworksError) {
	ns, err := mgr.ListNetworks()
	return ns, api.NewListNetworksError(err)
}

func (mgr *NetworkManager) listNetworks() ([]api.Network, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}
	var result []api.Network
	for _, vpc := range out.Vpcs {
		result = append(result, *network(vpc))
	}
	return result, nil
}

//GetNetwork returns the configuration of the network identified by id
func (mgr *NetworkManager) GetNetwork(id string) (*api.Network, api.GetNetworkError) {
	n, err := mgr.getNetwork(id)
	return n, api.NewGetNetworkError(err, id)
}

func (mgr *NetworkManager) getNetwork(id string) (*api.Network, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: []*string{&id},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Vpcs) > 0 {
		return network(out.Vpcs[0]), nil
	}
	return nil, err

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
func (mgr *NetworkManager) CreateSubnet(options api.CreateSubnetOptions) (*api.Subnet, api.CreateSubnetError) {
	sn, err := mgr.createSubnet(options)
	return sn, api.NewCreateSubnetError(err, options)
}

func (mgr *NetworkManager) createSubnet(options api.CreateSubnetOptions) (*api.Subnet, error) {
	input := ec2.CreateSubnetInput{
		AvailabilityZone: aws.String(mgr.Provider.Configuration.AvailabilityZone),
		VpcId:            &options.NetworkID,
	}
	if options.IPVersion == api.IPVersion4 {
		input.CidrBlock = &options.CIDR
	} else if options.IPVersion == api.IPVersion6 {
		input.Ipv6CidrBlock = &options.CIDR
	}
	out, err := mgr.Provider.AWSServices.EC2Client.CreateSubnet(&input)
	if err != nil {
		return nil, err
	}

	err = mgr.Provider.AddTags(*out.Subnet.SubnetId, map[string]string{"name": options.Name})
	if err != nil {
		err2 := mgr.DeleteSubnet(options.NetworkID, *out.Subnet.SubnetId)
		return nil, api.NewErrorStackFromError(err, err2)
	}
	err = mgr.associateRouteTable(&options, out)
	if err != nil {
		err2 := mgr.DeleteSubnet(options.NetworkID, *out.Subnet.SubnetId)
		return nil, api.NewErrorStackFromError(err, err2)
	}
	return mgr.GetSubnet(options.NetworkID, *out.Subnet.SubnetId)
}

func (mgr *NetworkManager) associateRouteTable(options *api.CreateSubnetOptions, out *ec2.CreateSubnetOutput) error {
	rt, err := mgr.getRouteTable(options.NetworkID)
	if err != nil {
		return err
	}
	_, err = mgr.Provider.AWSServices.EC2Client.AssociateRouteTable(&ec2.AssociateRouteTableInput{
		RouteTableId: rt.RouteTableId,
		SubnetId:     out.Subnet.SubnetId,
	})
	return err
}

func (mgr *NetworkManager) populateRouteTable(networkID string, gw *ec2.InternetGateway) (*ec2.RouteTable, error) {
	rt, err := mgr.getRouteTable(networkID)
	if err != nil {
		return nil, err
	}
	_, err = mgr.Provider.AWSServices.EC2Client.CreateRoute(&ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            gw.InternetGatewayId,
		RouteTableId:         rt.RouteTableId,
	})
	if err != nil {
		_, _ = mgr.Provider.AWSServices.EC2Client.DeleteRouteTable(&ec2.DeleteRouteTableInput{
			RouteTableId: rt.RouteTableId,
		})
		return nil, err
	}
	_, err = mgr.Provider.AWSServices.EC2Client.CreateRoute(&ec2.CreateRouteInput{
		DestinationIpv6CidrBlock: aws.String("::/0"),
		GatewayId:                gw.InternetGatewayId,
		RouteTableId:             rt.RouteTableId,
	})
	if err != nil {
		_, _ = mgr.Provider.AWSServices.EC2Client.DeleteRoute(&ec2.DeleteRouteInput{
			DestinationCidrBlock: aws.String("0.0.0.0/0"),
			RouteTableId:         rt.RouteTableId,
		})
		_, _ = mgr.Provider.AWSServices.EC2Client.DeleteRouteTable(&ec2.DeleteRouteTableInput{
			RouteTableId: rt.RouteTableId,
		})
		return nil, err
	}
	return rt, err
}

//DeleteSubnet deletes the subnet identified by id
func (mgr *NetworkManager) DeleteSubnet(networkID, subnetID string) api.DeleteSubnetError {
	_, err := mgr.Provider.AWSServices.EC2Client.DeleteSubnet(&ec2.DeleteSubnetInput{
		SubnetId: &subnetID,
	})
	return api.NewDeleteSubnetError(err, networkID, subnetID)
}

//ListSubnets lists the subnet
func (mgr *NetworkManager) ListSubnets(networkID string) ([]api.Subnet, api.ListSubnetsError) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeSubnets(&ec2.DescribeSubnetsInput{
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
		return nil, api.NewListSubnetsError(err, networkID)
	}
	var result []api.Subnet
	for _, sn := range out.Subnets {
		result = append(result, *subnet(sn))
	}
	return result, nil
}

//GetSubnet returns the configuration of the subnet identified by id
func (mgr *NetworkManager) GetSubnet(networkID, subnetID string) (*api.Subnet, api.GetSubnetError) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeSubnets(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{&subnetID},
	})
	if err != nil {
		return nil, api.NewGetSubnetError(err, networkID, subnetID)
	}
	for _, sn := range out.Subnets {
		if subnetID == *sn.SubnetId {
			return subnet(sn), nil
		}
	}
	return nil, api.NewGetSubnetError(err, networkID, subnetID)
}
