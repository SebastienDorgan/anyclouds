package aws

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

type NetworkInterfaceManager struct {
	Provider *Provider
}

func (mgr *NetworkInterfaceManager) convert(ni *ec2.NetworkInterface) *api.NetworkInterface {
	srvID := ""
	if ni.Attachment != nil {
		srvID = *ni.Attachment.InstanceId
	}
	ipAddr := ""
	if ni.PrivateIpAddress != nil {
		ipAddr = *ni.PrivateIpAddress
	}
	publicIP := ""
	if ni.Association != nil {
		publicIP = *ni.Association.PublicIp
	}
	return &api.NetworkInterface{
		ID:               *ni.NetworkInterfaceId,
		Name:             *ni.Description,
		MacAddress:       *ni.MacAddress,
		NetworkID:        *ni.VpcId,
		SubnetID:         *ni.SubnetId,
		ServerID:         srvID,
		PrivateIPAddress: ipAddr,
		PublicIPAddress:  publicIP,
		SecurityGroupID:  *ni.Groups[0].GroupId,
	}
}

func (mgr *NetworkInterfaceManager) Create(options api.CreateNetworkInterfaceOptions) (*api.NetworkInterface, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.CreateNetworkInterface(&ec2.CreateNetworkInterfaceInput{
		Description:      &options.Name,
		Groups:           []*string{&options.SecurityGroupID},
		PrivateIpAddress: options.PrivateIPAddress,
		SubnetId:         &options.SubnetID,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			*options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	if options.ServerID == nil {
		err = mgr.Provider.AWSServices.EC2Client.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{out.NetworkInterface.NetworkInterfaceId},
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
				*options.ServerID,
				options.SubnetID,
				options.NetworkID,
			)
		}
		return mgr.Get(*out.NetworkInterface.NetworkInterfaceId)
	}
	att, err := mgr.Provider.AWSServices.EC2Client.AttachNetworkInterface(&ec2.AttachNetworkInterfaceInput{
		InstanceId:         options.ServerID,
		NetworkInterfaceId: out.NetworkInterface.NetworkInterfaceId,
	})
	if err != nil {
		_, err2 := mgr.Provider.AWSServices.EC2Client.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: out.NetworkInterface.NetworkInterfaceId,
		})
		err = errors.Wrapf(err2, err.Error())
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			*options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	err = mgr.Provider.AWSServices.EC2Client.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{out.NetworkInterface.NetworkInterfaceId},
	})
	if err != nil {
		_, _ = mgr.Provider.AWSServices.EC2Client.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
			AttachmentId: att.AttachmentId,
			Force:        aws.Bool(true),
		})

		_, _ = mgr.Provider.AWSServices.EC2Client.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: out.NetworkInterface.NetworkInterfaceId,
		})
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			*options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	ni, err := mgr.Get(*out.NetworkInterface.NetworkInterfaceId)
	return ni, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
		*options.ServerID,
		options.SubnetID,
		options.NetworkID,
	)
}

func (mgr *NetworkInterfaceManager) Delete(id string) error {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{&id},
	})
	if err != nil {
		return errors.Wrapf(err, "error deleting network interface %s", id)
	}
	ni := out.NetworkInterfaces[0]
	_, _ = mgr.Provider.AWSServices.EC2Client.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
		AttachmentId: ni.Attachment.AttachmentId,
		Force:        aws.Bool(true),
	})
	_, err = mgr.Provider.AWSServices.EC2Client.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: ni.NetworkInterfaceId,
	})
	return errors.Wrapf(err, "error deleting network interface %s", id)

}

func (mgr *NetworkInterfaceManager) Get(id string) (*api.NetworkInterface, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{&id},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error deleting network interface %s", id)
	}
	return mgr.convert(out.NetworkInterfaces[0]), nil
}

func createFilters(options *api.ListNetworkInterfacesOptions) []*ec2.Filter {
	if options == nil {
		return nil
	}
	var filters []*ec2.Filter
	if options.NetworkID != nil {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("vpc-id"),
			Values: []*string{options.NetworkID},
		})
	}
	if options.ServerID != nil {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("attachment.instance-id"),
			Values: []*string{options.ServerID},
		})
	}
	if options.SubnetID != nil {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("subnet-id"),
			Values: []*string{options.SubnetID},
		})
	}
	if options.PrivateIPAddress != nil {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("private-ip-address"),
			Values: []*string{options.PrivateIPAddress},
		})
	}
	if options.SecurityGroupID != nil {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("group-id"),
			Values: []*string{options.SecurityGroupID},
		})
	}
	//
	return filters
}

func (mgr *NetworkInterfaceManager) List(options *api.ListNetworkInterfacesOptions) ([]api.NetworkInterface, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		Filters: createFilters(options),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error listing network interfaces")
	}
	var list []api.NetworkInterface
	for _, ni := range out.NetworkInterfaces {
		list = append(list, *mgr.convert(ni))
	}
	return list, nil
}

func (mgr *NetworkInterfaceManager) Update(options api.UpdateNetworkInterfacesOptions) (*api.NetworkInterface, error) {
	if options.ServerID != nil {
		out, err := mgr.Provider.AWSServices.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{&options.ID},
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
		}
		ni := out.NetworkInterfaces[0]
		if ni.Attachment != nil {
			_, err = mgr.Provider.AWSServices.EC2Client.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
				AttachmentId: ni.Attachment.AttachmentId,
				Force:        aws.Bool(true),
			})
			if err != nil {
				return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
			}
		}
		_, err = mgr.Provider.AWSServices.EC2Client.AttachNetworkInterface(&ec2.AttachNetworkInterfaceInput{
			InstanceId:         options.ServerID,
			NetworkInterfaceId: ni.NetworkInterfaceId,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
		}
		err = mgr.Provider.AWSServices.EC2Client.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{&options.ID},
		})
		return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
	}
	if options.SecurityGroupID != nil {
		_, err := mgr.Provider.AWSServices.EC2Client.ModifyNetworkInterfaceAttribute(&ec2.ModifyNetworkInterfaceAttributeInput{
			Groups:             []*string{options.SecurityGroupID},
			NetworkInterfaceId: aws.String(options.ID),
			SourceDestCheck:    nil,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
		}
		err = mgr.Provider.AWSServices.EC2Client.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{&options.ID},
		})
		return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
	}
	ni, err := mgr.Get(options.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
	}
	return ni, nil

}
