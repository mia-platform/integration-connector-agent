// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
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

package awsclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/sirupsen/logrus"
)

type concrete struct {
	sqs *sqs.Client
	s3  *s3.Client
	l   *lambda.Client

	log     *logrus.Logger
	config  Config
	stopped bool
	mu      sync.Mutex
}

type Config struct {
	QueueURL        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func New(ctx context.Context, log *logrus.Logger, c Config) (AWS, error) {
	loadOptions := make([]func(*config.LoadOptions) error, 0)

	if c.AccessKeyID != "" && c.SecretAccessKey != "" {
		credentialOptions := config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				c.AccessKeyID,
				c.SecretAccessKey,
				c.SessionToken,
			),
		)
		loadOptions = append(loadOptions, credentialOptions)
	} else {
		log.Warn("AccessKeyID and SecretAccessKey are not provided: using default credentials")
	}

	if c.Region != "" {
		loadOptions = append(loadOptions, config.WithRegion(c.Region))
	}

	sdkConfig, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqs client: %w", err)
	}

	return &concrete{
		sqs:     sqs.NewFromConfig(sdkConfig),
		s3:      s3.NewFromConfig(sdkConfig),
		l:       lambda.NewFromConfig(sdkConfig),
		log:     log,
		config:  c,
		stopped: false,
	}, nil
}

func (s *concrete) Listen(ctx context.Context, handler ListenerFunc) error {
	for {
		s.mu.Lock()
		if s.stopped {
			s.mu.Unlock()
			s.log.WithField("queueUrl", s.config.QueueURL).Info("stopped processing messages")
			return nil
		}
		s.mu.Unlock()

		result, err := s.sqs.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            &s.config.QueueURL,
			MaxNumberOfMessages: 5,
			WaitTimeSeconds:     5,
		})
		if err != nil {
			s.log.WithField("queueUrl", s.config.QueueURL).WithError(err).Warn("error receiving messages")
			continue
		}

		if len(result.Messages) == 0 {
			continue
		}

		s.log.WithFields(logrus.Fields{
			"queueUrl": s.config.QueueURL,
			"count":    len(result.Messages),
		}).Debug("received messages from SQS")

		for _, message := range result.Messages {
			if err := handler(ctx, []byte(*message.Body)); err != nil {
				s.log.WithFields(logrus.Fields{
					"queueUrl":  s.config.QueueURL,
					"messageId": message.MessageId,
				}).WithError(err).Warn("error processing message")
				continue
			}

			s.log.WithFields(logrus.Fields{
				"queueUrl":  s.config.QueueURL,
				"messageId": message.MessageId,
			}).Debug("message processed successfully")
			_, err := s.sqs.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      &s.config.QueueURL,
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil {
				s.log.WithFields(logrus.Fields{
					"queueUrl":  s.config.QueueURL,
					"messageId": message.MessageId,
				}).Warn("error deleting message from queue, it may be processed again later")
				continue
			}

			s.log.WithFields(logrus.Fields{
				"queueUrl":  s.config.QueueURL,
				"messageId": message.MessageId,
			}).Debug("message deleted successfully")
		}
	}
}

func (s *concrete) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopped = true
	return nil
}

func (s *concrete) ListBuckets(ctx context.Context) ([]*Bucket, error) {
	// BucketRegion is only returned when at least a valid parameter is set in the request.
	// Apparently, passing an empty prefix is considered a valid request parameter...
	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3@v1.83.0/types#Bucket
	buckets, err := s.s3.ListBuckets(ctx, &s3.ListBucketsInput{
		Prefix: aws.String(""),
	})
	if err != nil {
		return nil, err
	}

	result := make([]*Bucket, 0, len(buckets.Buckets))
	for _, bucket := range buckets.Buckets {
		b := &Bucket{
			Name: *bucket.Name,
		}

		if bucket.BucketRegion != nil {
			b.Region = *bucket.BucketRegion
		}

		// BucketArn is only returned for directory buckets.
		// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3@v1.83.0/types#Bucket
		if bucket.BucketArn != nil {
			parsedArn, err := arn.Parse(*bucket.BucketArn)
			if err == nil {
				b.AccountID = parsedArn.AccountID
			}
		}

		result = append(result, b)
	}

	return result, nil
}

func (s *concrete) ListFunctions(ctx context.Context) ([]*Function, error) {
	functions, err := s.l.ListFunctions(ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		return nil, err
	}

	result := make([]*Function, 0, len(functions.Functions))
	for _, function := range functions.Functions {
		f := &Function{
			Name: *function.FunctionName,
		}

		if function.FunctionArn != nil {
			parsedArn, err := arn.Parse(*function.FunctionArn)
			if err == nil {
				f.Region = parsedArn.Region
				f.AccountID = parsedArn.AccountID
			}
		}

		result = append(result, f)
	}
	return result, nil
}

func (c *concrete) GetBucketTags(ctx context.Context, bucketName string) (commons.Tags, error) {
	tags, err := c.s3.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
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

func (c *concrete) GetFunction(ctx context.Context, functionName string) (*Function, error) {
	function, err := c.l.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: &functionName,
	})
	if err != nil {
		return nil, err
	}

	result := &Function{
		Name: *function.Configuration.FunctionName,
		ARN:  *function.Configuration.FunctionArn,
		Tags: function.Tags,
	}
	return result, nil
}
