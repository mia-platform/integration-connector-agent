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

package kafka

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"
)

type Config struct {
	ProducerConfig *kafka.ConfigMap `json:"producerConfig"`
	Topic          string           `json:"topic"`
}

func (c *Config) Validate() error {
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
	return k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &k.topic,
			Partition: kafka.PartitionAny,
		},
		Value: data.Data(),
	}, nil)
}

func (k *Sink[T]) Close(_ context.Context) error {
	k.producer.Close()
	return nil
}
