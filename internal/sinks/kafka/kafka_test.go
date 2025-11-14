// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package kafka

import (
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKafkaConsumer(t *testing.T) {
	topicName := "test-topic"
	testMessage := `{"key":"value"}`
	server, err := kafka.NewMockCluster(1)
	require.NoError(t, err)
	defer server.Close()

	bootstrapServers := server.BootstrapServers()
	require.NoError(t, server.CreateTopic(topicName, 1, 1))

	sink, err := New[entities.PipelineEvent](&Config{
		ProducerConfig: &kafka.ConfigMap{
			"bootstrap.servers": bootstrapServers,
		},
		Topic: topicName,
	})

	require.NoError(t, err)
	defer sink.Close(t.Context())

	primaryKeys := entities.PkFields{
		{
			Key:   "id",
			Value: "123",
		},
	}
	err = sink.WriteData(t.Context(), &entities.Event{
		PrimaryKeys:   primaryKeys,
		Type:          "eventType",
		OperationType: entities.Write,
		OriginalRaw:   json.RawMessage(testMessage),
	})
	require.NoError(t, err)

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":     bootstrapServers,
		"broker.address.family": "v4",
		"group.id":              "group",
		"session.timeout.ms":    6000,
		"auto.offset.reset":     "earliest",
	})
	require.NoError(t, err)

	require.NoError(t, consumer.SubscribeTopics([]string{topicName}, nil))
	for {
		message, err := consumer.ReadMessage(1 * time.Second)
		if err != nil {
			t.Log(err.Error())
			continue
		}

		expectedHeaders := []kafka.Header{
			{Key: "operation_type", Value: []byte(entities.Write.String())},
			{Key: "event_type", Value: []byte("eventType")},
			{Key: "primary_key", Value: []byte(`[{"Key":"id","Value":"123"}]`)},
		}

		assert.Equal(t, topicName, *message.TopicPartition.Topic)
		assert.Equal(t, "a94ca16651d330029359101fafc1f9fd35413da8185dd93e1d5a80ef933a027b", hex.EncodeToString(message.Key))
		assert.Equal(t, testMessage, string(message.Value))
		assert.Equal(t, expectedHeaders, message.Headers)
		break
	}
}
