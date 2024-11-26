# Sinks

Sinks are the destinations where the Integration Connector Agent sends the processed data.
They are responsible for storing or further processing the data received from the external sources.
The Integration Connector Agent supports multiple types of sinks, allowing for flexible and scalable data integration solutions.

The supported sinks are:

- [**Console Catalog**](15_console-catalog.md): The Console Catalog sink allows you to save data into the Mia-Platform Console Catalog.
- [**CRUD Service**](30_crudservice.md): Useful to save events using
- [**MongoDB**](20_mongodb.md): A NoSQL database that stores data in a flexible, JSON-like format.
  [Mia-Platform CRUD Service](https://docs.mia-platform.eu/docs/runtime_suite/crud-service/overview_and_usage) HTTP API.
- [Apache Kafka](40_kafka.md): A distributed event streaming platform
