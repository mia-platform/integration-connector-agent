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

package internal

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/sirupsen/logrus"
)

type ListenerFunc func(ctx context.Context, data []byte) error

type SQS interface {
	Listen(ctx context.Context, handler ListenerFunc) error
	Close() error
}

type concreteSQS struct {
	c       *sqs.Client
	log     *logrus.Logger
	config  Config
	stopped bool
	mu      sync.Mutex
}

type Config struct {
	QueueURL string
	Region   string
}

func New(ctx context.Context, log *logrus.Logger, c Config) (SQS, error) {
	loadOptions := make([]func(*config.LoadOptions) error, 0)
	if c.Region != "" {
		loadOptions = append(loadOptions, config.WithRegion(c.Region))
	}

	sdkConfig, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqs client: %w", err)
	}

	client := sqs.NewFromConfig(sdkConfig)
	return &concreteSQS{
		c:       client,
		log:     log,
		config:  c,
		stopped: false,
	}, nil
}

func (s *concreteSQS) Listen(ctx context.Context, handler ListenerFunc) error {
	for {
		s.mu.Lock()
		if s.stopped {
			s.mu.Unlock()
			s.log.WithField("queueUrl", s.config.QueueURL).Info("stopped processing messages")
			return nil
		}
		s.mu.Unlock()

		result, err := s.c.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
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
			_, err := s.c.DeleteMessage(ctx, &sqs.DeleteMessageInput{
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

func (s *concreteSQS) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopped = true
	return nil
}
