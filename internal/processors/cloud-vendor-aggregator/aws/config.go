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

package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsc "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type processorConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}

func (c *processorConfig) AWSConfig(ctx context.Context, overrideRegion string) (aws.Config, error) {
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
