# Apache Kafka Sink

The Apache Kafka sink allows you to write data to a specific Kafka topic.

## Flows

This sink will only produce kafka event and send them to the configured topic, it will be up to the consumers or the
Apache Kafka admin to setup additional logic like setting the topic compaction or implement an upsert logic for
subsequent events.

## Configuration

To configure the Apache Kafka sink, you need to provide the following parameters in your configuration file:

- `topic`: the name of the topic where to save the events received by the sink
- `producerConfig`: contains the kafka connection configuration, you can found the [supported keys and values] in the
  official documentation of librdkafka

Example configuration:

```json
{
	"topic": "topic-name",
	"producerConfig": {
		"bootstrap.servers": "localhost:9092"
	}
}
```

[supported keys and values]: https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
