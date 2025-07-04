package s3

import (
	"context"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client interface {
	GetTags(ctx context.Context, bucketName string) (commons.Tags, error)
}

type client struct {
	c *s3.Client
}

func NewS3Client(awsConfig aws.Config) Client {
	return &client{
		c: s3.NewFromConfig(awsConfig),
	}
}

func (c *client) GetTags(ctx context.Context, bucketName string) (commons.Tags, error) {
	tags, err := c.c.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return nil, err
	}

	tagMap := make(commons.Tags, len(tags.TagSet))
	for _, tag := range tags.TagSet {
		tagMap[*tag.Key] = *tag.Value
	}
	return tagMap, nil
}
