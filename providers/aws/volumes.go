package aws

import (
	"fmt"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"sort"
)

//VolumeManager defines volume management functions an anyclouds provider must provide
type VolumeManager struct {
	Provider *Provider
}

func (mgr *VolumeManager) selectVolumeType(options *api.CreateVolumeOptions) string {
	if options.MinIOPS < 250 && options.MinDataRate < 250 && options.Size >= 500 {
		return "sc1"
	}
	if options.MinIOPS < 500 && options.MinDataRate < 500 && options.Size >= 500 {
		return "st1"
	}
	if options.MinIOPS <= 16000 && options.Size >= 4 &&
		((options.MinDataRate <= cMBToMiB(250) && options.Size >= cGBToGiB(334)) || (options.MinDataRate <= cMBToMiB(128))) {
		return "gp2"
	}

	return "io1"
}

//Create creates a volume with options
func (mgr *VolumeManager) Create(options api.CreateVolumeOptions) (*api.Volume, api.CreateVolumeError) {
	out, err := mgr.Provider.AWSServices.EC2Client.CreateVolume(&ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(mgr.Provider.Configuration.AvailabilityZone),
		DryRun:           aws.Bool(false),
		Encrypted:        aws.Bool(false),
		Iops:             aws.Int64(options.MinIOPS),
		KmsKeyId:         nil,
		Size:             aws.Int64(options.Size),
		SnapshotId:       nil,
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("volume"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String(options.Name),
					},
				},
			},
		},
		VolumeType: aws.String(mgr.selectVolumeType(&options)),
	})
	if err != nil {
		return nil, api.NewCreateVolumeError(err, options)
	}
	err = mgr.Provider.AWSServices.EC2Client.WaitUntilVolumeAvailable(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{out.VolumeId},
	})
	if err != nil {
		err2 := mgr.Delete(*out.VolumeId)
		err = api.NewErrorStackFromError(err, err2)
		return nil, api.NewCreateVolumeError(err, options)
	}
	return volume(out), nil
}

//Delete deletes volume identified by id
func (mgr *VolumeManager) Delete(id string) api.DeleteVolumeError {
	_, err := mgr.Provider.AWSServices.EC2Client.DeleteVolume(&ec2.DeleteVolumeInput{
		DryRun:   aws.Bool(false),
		VolumeId: aws.String(id),
	})
	return api.NewDeleteVolumeError(err, id)
}

func name(tags []*ec2.Tag) string {
	i := sort.Search(len(tags), func(i int) bool {
		return *tags[i].Key == "name"
	})
	if i < len(tags) {
		return *tags[i].Value
	}
	return ""

}

func cMBToMiB(v int64) int64 {
	return int64(float64(v) / 1.04858)
}

func cGBToGiB(v int64) int64 {
	return int64(float64(v) / 1.07374)
}

func cMiBToMB(v int64) int64 {
	return int64(float64(v) * 1.04858)
}

func cGiBToGB(v int64) int64 {
	return int64(float64(v) * 1.07374)
}

func volume(v *ec2.Volume) *api.Volume {
	var dataRate int64
	if *v.VolumeType == "gp2" {
		if *v.Size <= cGiBToGB(334) {
			dataRate = cMiBToMB(128)
		} else {
			dataRate = cMiBToMB(250)
		}
	} else if *v.VolumeType == "st1" {
		dataRate = 500
	} else if *v.VolumeType == "sc1" {
		dataRate = 250
	} else if *v.VolumeType == "io1" {
		//TODO see how to manage non Nitro based instance
		//https://docs.aws.amazon.com/fr_fr/AWSEC2/latest/UserGuide/instance-types.html#ec2-nitro-instances
		dataRate = cMiBToMB(1000)
	}
	return &api.Volume{
		ID:       *v.VolumeId,
		Name:     name(v.Tags),
		Size:     *v.Size,
		IOPS:     *v.Iops,
		DataRate: dataRate,
	}
}

//List lists volumes along filter
func (mgr *VolumeManager) List() ([]api.Volume, api.ListVolumesError) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeVolumes(&ec2.DescribeVolumesInput{
		DryRun: aws.Bool(false),
		Filters: []*ec2.Filter{
			{
				Name: aws.String("availability-zone"),
				Values: []*string{
					aws.String(mgr.Provider.Configuration.Region),
				},
			},
		},
		MaxResults: nil,
		NextToken:  nil,
		VolumeIds:  nil,
	})
	if err != nil {
		return nil, api.NewListVolumesError(err)
	}
	var volumes []api.Volume
	for _, res := range out.Volumes {
		volumes = append(volumes, *volume(res))
	}
	return nil, nil
}

//Get returns volume details
func (mgr *VolumeManager) Get(id string) (*api.Volume, api.GetVolumeError) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeVolumes(&ec2.DescribeVolumesInput{
		DryRun: aws.Bool(false),
		Filters: []*ec2.Filter{
			{
				Name: aws.String("availability-zone"),
				Values: []*string{
					aws.String(mgr.Provider.Configuration.AvailabilityZone),
				},
			},
		},
		MaxResults: nil,
		NextToken:  nil,
		VolumeIds:  []*string{aws.String(id)},
	})
	if err != nil {
		return nil, api.NewGetVolumeError(err, id)
	}
	if out.Volumes != nil {
		return nil, api.NewGetVolumeError(err, id)
	}

	return volume(out.Volumes[0]), nil
}

//Attach attaches a volume to an Server
func (mgr *VolumeManager) Attach(options api.AttachVolumeOptions) (*api.VolumeAttachment, api.AttachVolumeError) {
	out, err := mgr.Provider.AWSServices.EC2Client.AttachVolume(&ec2.AttachVolumeInput{
		Device:     aws.String(options.DevicePath),
		DryRun:     aws.Bool(false),
		InstanceId: aws.String(options.ServerID),
		VolumeId:   aws.String(options.VolumeID),
	})
	if err != nil {
		return nil, api.NewAttachVolumeError(err, options)
	}

	return attachment(out), nil

}

func attachment(out *ec2.VolumeAttachment) *api.VolumeAttachment {
	return &api.VolumeAttachment{
		ID:       fmt.Sprintf("%s#%s#%s", *out.VolumeId, *out.InstanceId, *out.Device),
		VolumeID: *out.VolumeId,
		ServerID: *out.InstanceId,
		Device:   *out.Device,
	}
}

//Detach detach a volume from an Server
func (mgr *VolumeManager) Detach(options api.DetachVolumeOptions) api.DetachVolumeError {
	_, err := mgr.Provider.AWSServices.EC2Client.DetachVolume(&ec2.DetachVolumeInput{
		Device:     nil,
		DryRun:     aws.Bool(false),
		Force:      aws.Bool(options.Force),
		InstanceId: aws.String(options.ServerID),
		VolumeId:   aws.String(options.VolumeID),
	})
	if err != nil {
		return api.NewDetachVolumeError(err, options)
	}
	err = mgr.Provider.AWSServices.EC2Client.WaitUntilVolumeAvailable(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{&options.VolumeID},
	})
	return api.NewDetachVolumeError(err, options)
}

func (mgr *VolumeManager) createFilter(options *api.ListAttachmentsOptions) []*ec2.Filter {
	return []*ec2.Filter{
		{
			Name: aws.String("availability-zone"),
			Values: []*string{
				aws.String(mgr.Provider.Configuration.AvailabilityZone),
			},
		},
		{
			Name: aws.String("attachment.instance-id"),
			Values: []*string{
				options.ServerID,
			},
		},
		{
			Name: aws.String("volume-id"),
			Values: []*string{
				options.VolumeID,
			},
		},
	}
}

//attachment returns the attachment between a volume and an Server
func (mgr *VolumeManager) attachments(options *api.ListAttachmentsOptions) ([]api.VolumeAttachment, error) {
	out, err := mgr.Provider.AWSServices.EC2Client.DescribeVolumes(&ec2.DescribeVolumesInput{
		DryRun:  aws.Bool(false),
		Filters: mgr.createFilter(options),
	})
	if err != nil {
		return nil, err
	}
	if out.Volumes == nil {
		return nil, nil
	}
	var attachments []api.VolumeAttachment
	for _, v := range out.Volumes {
		for _, att := range v.Attachments {
			if options.ServerID == nil || *options.ServerID == *att.InstanceId {
				attachments = append(attachments, *attachment(att))
			}
		}
	}
	return attachments, nil
}

//ListAttachments returns all the attachments of an Server
func (mgr *VolumeManager) ListAttachments(options *api.ListAttachmentsOptions) ([]api.VolumeAttachment, api.ListVolumeAttachmentsError) {
	attachments, err := mgr.attachments(options)
	return attachments, api.NewListVolumeAttachmentsError(err, options)

}

func (mgr *VolumeManager) Resize(options api.ResizeVolumeOptions) (*api.Volume, api.ResizeVolumeError) {
	vType := mgr.selectVolumeType(&api.CreateVolumeOptions{
		Name:        "",
		Size:        options.Size,
		MinIOPS:     options.MinIOPS,
		MinDataRate: options.MinDataRate,
	})
	out, err := mgr.Provider.AWSServices.EC2Client.ModifyVolume(&ec2.ModifyVolumeInput{
		DryRun:     aws.Bool(false),
		Iops:       aws.Int64(options.MinIOPS),
		Size:       aws.Int64(options.Size),
		VolumeId:   aws.String(options.ID),
		VolumeType: aws.String(vType),
	})
	if err != nil {
		return nil, api.NewResizeVolumeError(err, options)
	}
	v, err := mgr.Get(*out.VolumeModification.VolumeId)
	return v, api.NewResizeVolumeError(err, options)
}
