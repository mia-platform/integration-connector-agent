package lambda

import (
	"context"
	"encoding/json"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws/clients/lambda"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	awssqsevents "github.com/mia-platform/integration-connector-agent/internal/sources/aws-sqs/events"
	"github.com/sirupsen/logrus"
)

type Lambda struct {
	logger *logrus.Logger
	client lambda.Client
}

func New(logger *logrus.Logger, client lambda.Client) *Lambda {
	return &Lambda{
		logger: logger,
		client: client,
	}
}

func (l *Lambda) GetData(ctx context.Context, event *awssqsevents.CloudTrailEvent) ([]byte, error) {
	// it cannot fail because the event is already validated from the main processor
	data, _ := json.Marshal(event)

	name := lambdaName(event)
	if name == "" {
		l.logger.Error("functionName not found in request parameters or response elements")
		return nil, commons.ErrInvalidEvent
	}

	functionDetails, err := l.client.GetFunction(ctx, name)
	if err != nil {
		l.logger.WithError(err).Warn("failed to get Lambda function details")
	}

	tags := make(commons.Tags)
	if functionDetails != nil {
		tags = functionDetails.Tags
	}

	lambda := &commons.Asset{
		Name:          name,
		Type:          event.Detail.EventSource,
		Provider:      commons.AWSAssetProvider,
		Location:      event.Detail.AWSRegion,
		Tags:          tags,
		Relationships: []string{"account/" + event.Account},
		RawData:       data,
	}

	return json.Marshal(lambda)
}

func lambdaName(event *awssqsevents.CloudTrailEvent) string {
	if event.Detail.ResponseElements != nil {
		if name, ok := event.Detail.ResponseElements["functionName"]; ok {
			nameStr, ok := name.(string)
			if ok {
				return nameStr
			}
		}
	}

	if event.Detail.RequestParameters != nil {
		name := event.Detail.RequestParameters["functionName"]
		nameStr, ok := name.(string)
		if ok {
			return nameStr
		}
	}

	return ""
}
