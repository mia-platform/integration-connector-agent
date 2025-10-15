// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package kafka

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"
)

type Config struct {
	ProducerConfig *kafka.ConfigMap `json:"producerConfig"`
	Topic          string           `json:"topic"`
}

func (c *Config) Validate() error {
	if c.ProducerConfig == nil {
		return errors.New("producerConfig is required")
	}

	if len(c.Topic) == 0 {
		return errors.New("topic is required")
	}
	return nil
}

type Sink[T entities.PipelineEvent] struct {
	producer *kafka.Producer
	topic    string
}

func New[T entities.PipelineEvent](cfg *Config) (sinks.Sink[T], error) {
	p, err := kafka.NewProducer(cfg.ProducerConfig)
	if err != nil {
		return nil, err
	}

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			case kafka.Error:
				// Generic client instance-level errors, such as
				// broker connection failures, authentication issues, etc.
				//
				// These errors should generally be considered informational
				// as the underlying client will automatically try to
				// recover from any errors encountered, the application
				// does not need to take action on them.
				fmt.Printf("Error: %v\n", ev)
			default:
				fmt.Printf("Ignored event: %s\n", ev)
			}
		}
	}()

	return &Sink[T]{
		producer: p,
		topic:    cfg.Topic,
	}, nil
}

func (k *Sink[T]) WriteData(_ context.Context, data T) error {
	keys, err := json.Marshal(data.GetPrimaryKeys())
	if err != nil {
		return fmt.Errorf("failed to serialize primary keys: %w", err)
	}

	hasher := sha256.New()
	if _, err := hasher.Write(keys); err != nil {
		return fmt.Errorf("failed to hash primary keys: %w", err)
	}

	return k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &k.topic,
			Partition: kafka.PartitionAny,
		},
		Key: hasher.Sum(nil),
		Headers: []kafka.Header{
			{
				Key:   "operation_type",
				Value: []byte(data.Operation().String()),
			},
			{
				Key:   "event_type",
				Value: []byte(data.GetType()),
			},
			{
				Key:   "primary_key",
				Value: keys,
			},
		},
		Value: data.Data(),
	}, nil)
}

func (k *Sink[T]) Close(_ context.Context) error {
	k.producer.Flush(1000) // waith max 1 second for message deliveries
	k.producer.Close()
	return nil
}
