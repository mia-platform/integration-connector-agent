// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
