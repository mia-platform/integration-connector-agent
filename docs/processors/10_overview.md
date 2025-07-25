# Processors

In this section, there will be the processors available to be used to transform data
in the Integration Connector Agent.

If no processor are set in the configuration, the data will be sent to the sink as it is in input.

Each iteration of the processor will be applied to the data on the last iteration output.

The supported processors are:

- [**Filter**](./15_filter.md): Filter the event based on a condition. If the event is filtered,
it will not be sent to the sink.
- [**Mapper**](./20_mapper.md): Transform the data to an output event, based on the input.
- [**RPC Plugin**](./30_rpc_plugin.md): Transform data to the desired output using a custom-built RPC Plugin ([example usage](https://github.com/mia-platform/integration-connector-agent/blob/main/examples/rpc-processor-plugin/plugin.go)).
- [**Cloud Vendor Aggregator**](./40_cloud_vendor_aggregator.md): Aggregate events from cloud vendors into a standardized
asset shape.
