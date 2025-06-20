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
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/sirupsen/logrus"
)

type ListenerFunc func(ctx context.Context, data []byte) error

type PubSub interface {
	Listen(ctx context.Context, handler ListenerFunc) error
	Close() error
}

type ConcretePubSub struct {
	c      *pubsub.Client
	config PubSubConfig
	log    *logrus.Logger
}

type PubSubConfig struct {
	ProjectID          string
	AckDeadlineSeconds int
	TopicName          string
	SubscriptionID     string
}

func New(ctx context.Context, log *logrus.Logger, config PubSubConfig) (PubSub, error) {
	client, err := pubsub.NewClient(ctx, config.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	return &ConcretePubSub{
		c:      client,
		log:    log,
		config: config,
	}, nil
}

func (p *ConcretePubSub) Listen(ctx context.Context, handler ListenerFunc) error {
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

func (p *ConcretePubSub) Close() error {
	if err := p.c.Close(); err != nil {
		return fmt.Errorf("failed to close pubsub client: %w", err)
	}
	return nil
}

func (p *ConcretePubSub) ensureSubscription(ctx context.Context, topicName, subscriptionID string) (*pubsub.Subscription, error) {
	subscription := p.c.Subscription(subscriptionID)
	exists, err := subscription.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if subscription exists: %w", err)
	}
	if exists {
		return subscription, nil
	}

	subscription, err = p.c.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
		Topic:       p.c.Topic(topicName),
		AckDeadline: time.Duration(p.config.AckDeadlineSeconds) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}
	return subscription, nil
}
