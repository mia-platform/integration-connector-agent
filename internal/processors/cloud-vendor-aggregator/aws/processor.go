package aws

import (
	"context"
	"encoding/json"
	"fmt"

	rpcprocessor "github.com/mia-platform/integration-connector-agent/adapters/rpc-processor"
	"github.com/mia-platform/integration-connector-agent/entities"
	lambdaclient "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws/clients/lambda"
	s3client "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws/clients/s3"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws/services/lambda"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws/services/s3"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"
	"github.com/sirupsen/logrus"
)

type AWSProcessor struct {
	ctx    context.Context
	logger rpcprocessor.Logger

	config AWSProcessorConfig
}

func New(logger *logrus.Logger, authOptions config.AuthOptions) entities.Processor {
	return &AWSProcessor{
		logger: logger,
		config: AWSProcessorConfig{
			AccessKeyID:     authOptions.AccessKeyID,
			SecretAccessKey: authOptions.SecretAccessKey.String(),
			SessionToken:    authOptions.SessionToken.String(),
			Region:          authOptions.Region,
		},
	}
}

func (p *AWSProcessor) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	cloudTrailEvent := new(awssqsevents.CloudTrailEvent)
	if err := json.Unmarshal(input.Data(), &cloudTrailEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input data: %w", err)
	}

	dataProcessor, err := p.EventDataProcessor(cloudTrailEvent)
	if err != nil {
		p.logger.WithError(err).Error("Failed to get event data processor")
		return nil, fmt.Errorf("failed to get event data processor: %w", err)
	}

	newData, err := dataProcessor.GetData(p.ctx, cloudTrailEvent)
	if err != nil {
		p.logger.WithError(err).Error("Failed to get data from AWS service")
		return nil, fmt.Errorf("failed to get data from AWS service: %w", err)
	}

	output := input.Clone()
	output.WithData(newData)
	return output, nil
}

func (p *AWSProcessor) EventDataProcessor(cloudTrailEvent *awssqsevents.CloudTrailEvent) (commons.DataAdapter[*awssqsevents.CloudTrailEvent], error) {
	awsConf, err := p.config.AWSConfig(p.ctx, cloudTrailEvent.Detail.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	eventSource := cloudTrailEvent.Detail.EventSource
	switch eventSource {
	case s3.EventSource:
		return s3.New(p.logger, s3client.NewS3Client(awsConf)), nil
	case lambda.EventSource:
		return lambda.New(p.logger, lambdaclient.NewS3Client(awsConf)), nil
	default:
		return nil, fmt.Errorf("unsupported event source: %s", eventSource)
	}
}
