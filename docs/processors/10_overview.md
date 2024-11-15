# Processors

In this section, there will be the processors available to be used to transform data
in the Integration Connector Agent.

If no processor are set in the configuration, the data will be sent to the sink as it is in input.

Each iteration of the processor will be applied to the data on the last iteration output.

The supported processors are:

- [**Mapper**](./20_mapper.md): Transform the data to an output event, based on the input.
