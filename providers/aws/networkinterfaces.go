package aws

import (
	"fmt"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
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

func (mgr *NetworkInterfaceManager) create(options api.CreateNetworkInterfaceOptions) (*api.NetworkInterface, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.CreateNetworkInterface(&ec2.CreateNetworkInterfaceInput{
		Description:      &options.Name,
		Groups:           []*string{&options.SecurityGroupID},
		PrivateIpAddress: options.PrivateIPAddress,
		SubnetId:         &options.SubnetID,
	})
	if err != nil {
		return nil, err
	}
	if options.ServerID == nil {
		err = mgr.Provider.AWSServices.EC2Client.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{out.NetworkInterface.NetworkInterfaceId},
		})
		if err != nil {
			if out.NetworkInterface != nil {
				err2 := mgr.Delete(*out.NetworkInterface.NetworkInterfaceId)
				err = api.NewErrorStackFromError(err, err2)
			}
			return nil, err
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
		return nil, api.NewErrorStackFromError(err, err2)
	}
	err = mgr.Provider.AWSServices.EC2Client.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{out.NetworkInterface.NetworkInterfaceId},
	})
	if err != nil {
		_, err2 := mgr.Provider.AWSServices.EC2Client.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
			AttachmentId: att.AttachmentId,
			Force:        aws.Bool(true),
		})
		err = api.NewErrorStackFromError(err, err2)
		_, err2 = mgr.Provider.AWSServices.EC2Client.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: out.NetworkInterface.NetworkInterfaceId,
		})
		return nil, api.NewErrorStackFromError(err, err2)
	}
	return mgr.Get(*out.NetworkInterface.NetworkInterfaceId)
}

func (mgr *NetworkInterfaceManager) Create(options api.CreateNetworkInterfaceOptions) (*api.NetworkInterface, api.CreateNetworkInterfaceError) {
	ni, err := mgr.create(options)
	return ni, api.NewCreateNetworkInterfaceError(err, options)
}

func (mgr *NetworkInterfaceManager) delete(id string) error {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{&id},
	})
	if err != nil {
		return err
	}
	if out.NetworkInterfaces == nil {
		return fmt.Errorf("network interface %s not found", id)
	}
	ni := out.NetworkInterfaces[0]
	var err2 error
	if ni.Attachment != nil && ni.Attachment.AttachmentId != nil {
		_, err2 = mgr.Provider.AWSServices.EC2Client.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
			AttachmentId: ni.Attachment.AttachmentId,
			Force:        aws.Bool(true),
		})
	}

	_, err = mgr.Provider.AWSServices.EC2Client.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: ni.NetworkInterfaceId,
	})
	//if detach fails but delete succeeds then no error is raised
	if err != nil {
		err = api.NewErrorStackFromError(err, err2)
	}
	return nil
}

func (mgr *NetworkInterfaceManager) Delete(id string) api.DeleteNetworkInterfaceError {
	return api.NewDeleteNetworkInterfaceError(mgr.delete(id), id)
}

func (mgr *NetworkInterfaceManager) get(id string) (*api.NetworkInterface, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{&id},
	})
	if err != nil {
		return nil, err
	}
	return mgr.convert(out.NetworkInterfaces[0]), nil
}

func (mgr *NetworkInterfaceManager) Get(id string) (*api.NetworkInterface, api.GetNetworkInterfaceError) {
	ni, err := mgr.get(id)
	return ni, api.NewGetNetworkInterfaceError(err, id)
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

func (mgr *NetworkInterfaceManager) list(options *api.ListNetworkInterfacesOptions) ([]api.NetworkInterface, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		Filters: createFilters(options),
	})
	if err != nil {
		return nil, err
	}
	var list []api.NetworkInterface
	for _, ni := range out.NetworkInterfaces {
		list = append(list, *mgr.convert(ni))
	}
	return list, nil
}

func (mgr *NetworkInterfaceManager) List(options *api.ListNetworkInterfacesOptions) ([]api.NetworkInterface, api.ListNetworkInterfacesError) {
	nis, err := mgr.list(options)
	return nis, api.NewListNetworkInterfacesError(err, options)
}

func (mgr *NetworkInterfaceManager) update(options api.UpdateNetworkInterfaceOptions) (*api.NetworkInterface, error) {
	if options.ServerID != nil {
		out, err := mgr.Provider.AWSServices.EC2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{&options.ID},
		})
		if err != nil {
			return nil, err
		}
		if out.NetworkInterfaces != nil {
			return nil, fmt.Errorf("network interface %s not found", options.ID)
		}
		ni := out.NetworkInterfaces[0]
		if ni.Attachment != nil && ni.Attachment.AttachmentId != nil {
			_, err = mgr.Provider.AWSServices.EC2Client.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
				AttachmentId: ni.Attachment.AttachmentId,
				Force:        aws.Bool(true),
			})
			if err != nil {
				return nil, err
			}
		}
		_, err = mgr.Provider.AWSServices.EC2Client.AttachNetworkInterface(&ec2.AttachNetworkInterfaceInput{
			InstanceId:         options.ServerID,
			NetworkInterfaceId: ni.NetworkInterfaceId,
		})
		if err != nil {
			return nil, err
		}
		err = mgr.Provider.AWSServices.EC2Client.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{&options.ID},
		})
		return nil, err
	}
	if options.SecurityGroupID != nil {
		_, err := mgr.Provider.AWSServices.EC2Client.ModifyNetworkInterfaceAttribute(&ec2.ModifyNetworkInterfaceAttributeInput{
			Groups:             []*string{options.SecurityGroupID},
			NetworkInterfaceId: aws.String(options.ID),
		})
		if err != nil {
			return nil, err
		}
		err = mgr.Provider.AWSServices.EC2Client.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []*string{&options.ID},
		})
		return nil, err
	}
	return mgr.Get(options.ID)
}

func (mgr *NetworkInterfaceManager) Update(options api.UpdateNetworkInterfaceOptions) (*api.NetworkInterface, api.UpdateNetworkInterfaceError) {
	ni, err := mgr.update(options)
	return ni, api.NewUpdateNetworkInterfaceError(err, options)
}
