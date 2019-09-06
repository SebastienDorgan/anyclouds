package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

func createAWSTags(tags map[string]string) []*ec2.Tag {
	var awsTags []*ec2.Tag
	for k, v := range tags {
		awsTags = append(awsTags, &ec2.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	return awsTags
}

func (p *Provider) AddTags(resourceId string, tags map[string]string) error {
	_, err := p.EC2Client.CreateTags(&ec2.CreateTagsInput{
		DryRun: aws.Bool(false),
		Resources: []*string{
			aws.String(resourceId),
		},
		Tags: createAWSTags(tags),
	})

	return errors.Wrapf(err, "error adding tags to resource %s", resourceId)

}
