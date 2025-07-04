package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsc "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type AWSProcessorConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}

func (c *AWSProcessorConfig) AWSConfig(ctx context.Context, overrideRegion string) (aws.Config, error) {
	loadOptions := make([]func(*awsc.LoadOptions) error, 0)

	loadOptions = append(loadOptions, awsc.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, c.SessionToken),
	))

	if overrideRegion != "" {
		loadOptions = append(loadOptions, awsc.WithRegion(overrideRegion))
	} else if c.Region != "" {
		loadOptions = append(loadOptions, awsc.WithRegion(c.Region))
	}

	return awsc.LoadDefaultConfig(ctx, loadOptions...)
}
