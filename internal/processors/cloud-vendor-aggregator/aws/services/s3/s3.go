package s3

import (
	"context"
	"encoding/json"
	"fmt"

	rpcprocessor "github.com/mia-platform/integration-connector-agent/adapters/rpc-processor"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws/clients/s3"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"
)

type S3 struct {
	logger rpcprocessor.Logger
	client s3.Client
}

func New(logger rpcprocessor.Logger, client s3.Client) *S3 {
	return &S3{
		logger: logger,
		client: client,
	}
}

func (s *S3) GetData(ctx context.Context, event *awssqsevents.CloudTrailEvent) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)

	bucketName, ok := event.Detail.RequestParameters["bucketName"].(string)
	if !ok {
		s.logger.Warn("bucketName not found in request parameters")
		return nil, fmt.Errorf("%w: bucket name not found in request parameters", commons.ErrInvalidEvent)
	}

	tags, err := s.client.GetTags(ctx, bucketName)
	if err != nil {
		s.logger.WithError(err).Warn("failed to get S3 bucket tags")
		tags = make(commons.Tags)
	}

	bucket := &commons.Asset{
		Name:          bucketName,
		Type:          event.Detail.EventSource,
		Location:      event.Detail.AWSRegion,
		Tags:          tags,
		Provider:      commons.AWSAssetProvider,
		Relationships: []string{"account/" + event.Account},
		RawData:       data,
	}

	return json.Marshal(bucket)
}
