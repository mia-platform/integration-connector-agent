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

package gcpclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type ListenerFunc func(ctx context.Context, data []byte) error

type GCP interface {
	ListBuckets(ctx context.Context) ([]*Bucket, error)
	Listen(ctx context.Context, handler ListenerFunc) error
	Close() error
}

type concrete struct {
	config GCPConfig
	log    *logrus.Logger

	p *pubsub.Client
	s *storage.Client
}

type GCPConfig struct {
	ProjectID          string
	AckDeadlineSeconds int
	TopicName          string
	SubscriptionID     string
	CredentialsJSON    string
}

func New(ctx context.Context, log *logrus.Logger, config GCPConfig) (GCP, error) {
	options := make([]option.ClientOption, 0)
	if config.CredentialsJSON != "" {
		log.Debug("using credentials JSON for Pub/Sub client")
		options = append(options, option.WithCredentialsJSON([]byte(config.CredentialsJSON)))
	}

	pubSubClient, err := pubsub.NewClient(ctx, config.ProjectID, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	storageClient, err := storage.NewClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &concrete{
		log:    log,
		config: config,
		p:      pubSubClient,
		s:      storageClient,
	}, nil
}

func (p *concrete) Listen(ctx context.Context, handler ListenerFunc) error {
	subscription, err := p.ensureSubscription(ctx, p.config.TopicName, p.config.SubscriptionID)
	if err != nil {
		return err
	}

	p.log.WithFields(logrus.Fields{
		"projectId":      p.config.ProjectID,
		"topicName":      p.config.TopicName,
		"subscriptionId": p.config.SubscriptionID,
	}).Debug("starting to listen to Pub/Sub messages")

	return subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		p.log.WithFields(logrus.Fields{
			"projectId":      p.config.ProjectID,
			"topicName":      p.config.TopicName,
			"subscriptionId": p.config.SubscriptionID,
			"messageId":      msg.ID,
		}).Trace("received message from Pub/Sub")

		if err := handler(ctx, msg.Data); err != nil {
			p.log.
				WithFields(logrus.Fields{
					"projectId":      p.config.ProjectID,
					"topicName":      p.config.TopicName,
					"subscriptionId": p.config.SubscriptionID,
					"messageId":      msg.ID,
				}).
				WithError(err).
				Error("error handling message")

			msg.Nack()
			return
		}

		// TODO: message is Acked here once the pipelines have received the message for processing.
		// This means that if the pipeline fails after this point, the message will not be
		// retried. Consider implementing, either:
		// - a dead-letter queue or similar mechanism.
		// - a way to be notified here if all the pipelins have processed the
		//   message successfully in order to correctly ack/nack it.
		msg.Ack()
	})
}

func (p *concrete) Close() error {
	if err := p.p.Close(); err != nil {
		return fmt.Errorf("failed to close pubsub client: %w", err)
	}

	if err := p.s.Close(); err != nil {
		return fmt.Errorf("failed to close storage client: %w", err)
	}
	return nil
}

func (p *concrete) ensureSubscription(ctx context.Context, topicName, subscriptionID string) (*pubsub.Subscription, error) {
	subscription := p.p.Subscription(subscriptionID)
	exists, err := subscription.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if subscription exists: %w", err)
	}
	if exists {
		return subscription, nil
	}

	ackDeadline := 10 * time.Second
	if p.config.AckDeadlineSeconds > 0 {
		ackDeadline = time.Duration(p.config.AckDeadlineSeconds) * time.Second
	}

	subscription, err = p.p.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
		Topic:       p.p.Topic(topicName),
		AckDeadline: ackDeadline,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}
	return subscription, nil
}

type Bucket struct {
	Name string
}

func (p *concrete) ListBuckets(ctx context.Context) ([]*Bucket, error) {
	buckets := make([]*Bucket, 0)

	it := p.s.Buckets(ctx, p.config.ProjectID)
	for {
		bucket, err := it.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, fmt.Errorf("failed to list buckets: %w", err)
		}

		buckets = append(buckets, &Bucket{
			Name: bucket.Name,
		})
	}
	return buckets, nil
}
