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

//go:build integration
// +build integration

package kafka

import (
	"context"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcKafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

func TestKafkaConsumer(t *testing.T) {
	ctx := context.Background()
	kafkaClusterID := testutils.RandomString(t, 6)
	kafkaContainer, err := tcKafka.Run(ctx, "confluentinc/confluent-local:7.5.0", tcKafka.WithClusterID(kafkaClusterID))
	testcontainers.CleanupContainer(t, kafkaContainer)
	require.NoError(t, err)

	servers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": servers,
	})
	require.NoError(t, err)

	t.Run("produce kafka message", func(t *testing.T) {
		sink, err := New[*entities.Event](&Config{
			Topic: "test",
			ProducerConfig: &kafka.ConfigMap{
				"bootstrap.servers": servers,
			},
		})
		require.NoError(t, err)

		err = sink.WriteData(context.Background(), &entities.Event{
			OriginalRaw: []byte("test"),
		})
		require.NoError(t, err)

		for {
			if event := consumer.Poll(10); event != nil {
				require.Equal(t, event.String(), "test")
			}
		}
	})
}
