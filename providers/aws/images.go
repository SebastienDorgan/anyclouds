package aws

import (
	"fmt"
	"time"

	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

//ImageManager defines image management functions a anyclouds provider must provide
type ImageManager struct {
	AWS *Provider
}

func values(values ...string) []*string {
	var result []*string
	for _, v := range values {
		result = append(result, aws.String(v))
	}
	return result
}

func (mgr *ImageManager) search(owner string, name string) ([]api.Image, error) {
	out, err := mgr.AWS.EC2Client.DescribeImages(&ec2.DescribeImagesInput{
		DryRun: aws.Bool(false),
		Owners: values(owner),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("virtualization-type"),
				Values: values("hvm"),
			},
			{
				Name:   aws.String("architecture"),
				Values: values("x86_64"),
			},
			{
				Name:   aws.String("root-device-type"),
				Values: values("ebs"),
			},
			{
				Name:   aws.String("block-device-mapping.volume-type"),
				Values: values("gp2"),
			},
			{
				Name:   aws.String("name"),
				Values: values(name),
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error getting image")
	}
	var result []api.Image
	for _, img := range out.Images {
		result = append(result, *image(img))
	}
	return result, nil
}

//List returns available image list
func (mgr *ImageManager) List() ([]api.Image, error) {
	imgsUbuntu, err := mgr.search("099720109477", "ubuntu/images/hvm-ssd/ubuntu-*-*-*-*-????????")
	if err != nil {
		return nil, errors.Wrap(err, "Error listing image")
	}
	imgsRHEL, err := mgr.search("309956199498", "RHEL-?.?_HVM_GA*")
	if err != nil {
		return nil, errors.Wrap(err, "Error listing image")
	}
	imgsDebian, err := mgr.search("379101102735", "debian-*")
	if err != nil {
		return nil, errors.Wrap(err, "Error listing image")
	}
	imgsCentOS, err := mgr.search("410186602215", "CentOS*")
	if err != nil {
		return nil, errors.Wrap(err, "Error listing image")
	}

	var result []api.Image
	result = append(result, imgsUbuntu...)
	result = append(result, imgsRHEL...)
	result = append(result, imgsDebian...)
	result = append(result, imgsCentOS...)
	return result, nil
}

func image(img *ec2.Image) *api.Image {
	creationDate, _ := time.Parse(time.RFC3339, *img.CreationDate)
	return &api.Image{
		CreatedAt: creationDate,
		ID:        *img.ImageId,
		Name:      *img.Name,
		MinDisk:   0,
		MinRAM:    0,
		UpdatedAt: creationDate,
	}
}

//Get returns the image identified by id
func (mgr *ImageManager) Get(id string) (*api.Image, error) {
	out, err := mgr.AWS.EC2Client.DescribeImages(&ec2.DescribeImagesInput{
		DryRun: aws.Bool(false),
		ImageIds: []*string{
			aws.String(id),
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "error getting image")
	}
	if len(out.Images) == 0 {
		return nil, fmt.Errorf("image %s not found", id)
	}
	if len(out.Images) > 1 {
		return nil, fmt.Errorf("multiple images with the same id: %s", id)
	}
	return image(out.Images[0]), nil
}
