# MongoDB Sink

The MongoDB sink allows you to save and delete data from a MongoDB instance.
It supports upserting data using a specified ID as the primary key.

Each record will be updated based on a unique ID field, which is statically the `_eventId` one.

For each kind of input event, there is a chosen event id. In this table, we can see the event id for each kind of event.
[See the interested source](../sources/10_overview.md) to view the event id and how the events are mapped.

## Flow

Depending on the source event, it is possible to create two different actions:

- **Upsert**: The sink will insert the data into the collection if not present, or completely replace it if
it is already present. The update is based on the `_eventId` field.
- **Delete**: The sink will delete the data from the collection. The delete is based on the `_eventId` field.

[See how different events are managed](../sources/10_overview.md)  in the sources documentation.

## Configuration

To configure the MongoDB sink, you need to provide the following parameters in your configuration file:

- `type` (*string*): The type of the sink, which should be set to `mongo`.
- `url` ([*SecretSource*](../20_install.md#secretsource)): The MongoDB connection URL
- `collection` (*string*): The name of the MongoDB collection where data will be stored.

Example configuration:

```json
{
  "type": "mongo",
  "url": {
    "fromEnv": "MONGO_URL"
  },
  "collection": "sink-target-collection"
}
```

The db will be taken from the URL.

:::info
If not present in db, the collection will be created.

It is highly recommended to set an unique index on the `_eventId` field.
:::
