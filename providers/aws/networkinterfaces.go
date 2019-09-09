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

func convert(ni *ec2.NetworkInterface) *api.NetworkInterface {
	srvID := ""
	if ni.Attachment != nil {
		srvID = *ni.Attachment.InstanceId
	}
	ipAddr := ""
	if ni.PrivateIpAddress != nil {
		ipAddr = *ni.PrivateIpAddress
	}
	return &api.NetworkInterface{
		ID:         *ni.NetworkInterfaceId,
		Name:       *ni.Description,
		MacAddress: *ni.MacAddress,
		NetworkID:  *ni.VpcId,
		SubnetID:   *ni.SubnetId,
		ServerID:   srvID,
		IPAddress:  ipAddr,
		Primary:    false,
	}
}

func (mgr *NetworkInterfaceManager) Create(options *api.NetworkInterfaceOptions) (*api.NetworkInterface, error) {
	out, err := mgr.Provider.EC2Client.CreateNetworkInterface(&ec2.CreateNetworkInterfaceInput{
		Description:      &options.Name,
		Groups:           []*string{&options.SecurityGroupID},
		PrivateIpAddress: options.IPAddress,
		SubnetId:         &options.SubnetID,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	att, err := mgr.Provider.EC2Client.AttachNetworkInterface(&ec2.AttachNetworkInterfaceInput{
		InstanceId:         &options.ServerID,
		NetworkInterfaceId: out.NetworkInterface.NetworkInterfaceId,
	})
	if err != nil {
		_, err2 := mgr.Provider.EC2Client.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: out.NetworkInterface.NetworkInterfaceId,
		})
		err = errors.Wrapf(err2, err.Error())
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	err = mgr.Provider.EC2Client.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{out.NetworkInterface.NetworkInterfaceId},
	})
	if err != nil {
		_, _ = mgr.Provider.EC2Client.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
			AttachmentId: att.AttachmentId,
			Force:        aws.Bool(true),
		})

		_, _ = mgr.Provider.EC2Client.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: out.NetworkInterface.NetworkInterfaceId,
		})
		return nil, errors.Wrapf(err, "error creating network interface between server %s and subnet %s on network %s",
			options.ServerID,
			options.SubnetID,
			options.NetworkID,
		)
	}
	return mgr.Get(*out.NetworkInterface.NetworkInterfaceId)
}

func (mgr *NetworkInterfaceManager) Delete(id string) error {
	out, err := mgr.Provider.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{&id},
	})
	if err != nil {
		return errors.Wrapf(err, "error deleting network interface %s", id)
	}
	ni := out.NetworkInterfaces[0]
	_, _ = mgr.Provider.EC2Client.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
		AttachmentId: ni.Attachment.AttachmentId,
		Force:        aws.Bool(true),
	})
	_, err = mgr.Provider.EC2Client.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: ni.NetworkInterfaceId,
	})
	return errors.Wrapf(err, "error deleting network interface %s", id)

}

func (mgr *NetworkInterfaceManager) Get(id string) (*api.NetworkInterface, error) {
	out, err := mgr.Provider.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{&id},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error deleting network interface %s", id)
	}
	return convert(out.NetworkInterfaces[0]), nil
}

func (mgr *NetworkInterfaceManager) List() ([]api.NetworkInterface, error) {
	out, err := mgr.Provider.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{})
	if err != nil {
		return nil, errors.Wrap(err, "error listing network interfaces")
	}
	var list []api.NetworkInterface
	for _, ni := range out.NetworkInterfaces {
		list = append(list, *convert(ni))
	}
	return list, nil
}

func (mgr *NetworkInterfaceManager) Update(options *api.NetworkInterfacesUpdateOptions) (*api.NetworkInterface, error) {
	if options.ServerID != nil {
		out, err := mgr.Provider.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{&options.ID},
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
		}
		ni := out.NetworkInterfaces[0]
		if ni.Attachment != nil {
			_, err = mgr.Provider.EC2Client.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
				AttachmentId: ni.Attachment.AttachmentId,
				Force:        aws.Bool(true),
			})
			if err != nil {
				return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
			}
		}
		_, err = mgr.Provider.EC2Client.AttachNetworkInterface(&ec2.AttachNetworkInterfaceInput{
			InstanceId:         options.ServerID,
			NetworkInterfaceId: ni.NetworkInterfaceId,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
		}
		err = mgr.Provider.EC2Client.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{&options.ID},
		})
		return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
	}
	if options.SecurityGroupID != nil {
		_, err := mgr.Provider.EC2Client.ModifyNetworkInterfaceAttribute(&ec2.ModifyNetworkInterfaceAttributeInput{
			Groups:             []*string{options.SecurityGroupID},
			NetworkInterfaceId: aws.String(options.ID),
			SourceDestCheck:    nil,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error updating network interface %s", options.ID)
		}
		err = mgr.Provider.EC2Client.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
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
