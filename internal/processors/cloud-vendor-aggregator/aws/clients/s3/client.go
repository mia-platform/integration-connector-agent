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
