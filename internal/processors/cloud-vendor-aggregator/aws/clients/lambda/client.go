package lambda

import (
	"context"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type Client interface {
	GetFunction(ctx context.Context, functionName string) (*Function, error)
}

type client struct {
	c *lambda.Client
}

func NewS3Client(awsConfig aws.Config) Client {
	return &client{
		c: lambda.NewFromConfig(awsConfig),
	}
}

type Function struct {
	FunctionName string
	FunctionArn  string
	Tags         commons.Tags
}

func (c *client) GetFunction(ctx context.Context, functionName string) (*Function, error) {
	function, err := c.c.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: &functionName,
	})
	if err != nil {
		return nil, err
	}

	result := &Function{
		FunctionName: *function.Configuration.FunctionName,
		FunctionArn:  *function.Configuration.FunctionArn,
		Tags:         function.Tags,
	}
	return result, nil
}
