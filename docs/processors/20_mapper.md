# Mapper

The Mapper processor allows you to transform input data into a desired output format using a template.  
This is useful for reformatting data before sending it to a sink.

## Configuration

To configure the Mapper processor, you need to define the `outputEvent` template in your configuration file.
The template uses placeholders to map input data fields to the desired output structure.

It is possible to create nested objects, or to point to a specific field in the input data (also nested).
If the field value is an object or an array, it will be copied as is in the output structure.

To configure mapper processor, you need to provide the following parameters in your configuration file:

- `type` (*string*): The type of the processor, which should be set to `mapper`.
- `outputEvent` (*object*): The output event template, it's an object which map the input event to the desired output event.

### Template

All data inside the `{{ }}` template will be evaluated. Inside the template it is
expected to have the field path to take the value from.

For example, given the following input:

```json
{
  "issue": {
    "key": "ISSUE-123",
    "fields": {
      "summary": "This is a summary",
      "created": "2021-01-01T00:00:00Z"
    }
  }
}
```

To reach the `summary` field, the template should be `{{ issue.fields.summary }}`.
To reach the `created` field, the template should be `{{ issue.fields.created }}`.

### Static data

It is also possible to set static data, for example:

```json
{
  "type": "mapper",
  "outputEvent": {
    "key": "static-value",
    "createdAt": "{{ issue.fields.created }}"
  }
}
```

In this case, the `key` field will always have the value `static-value`.

### Type Casting

The Mapper processor supports type casting to ensure that field values are stored in the correct data type. 
This is particularly useful when working with databases like MongoDB that don't enforce type constraints.

To cast a field value, use an object with `value` and `castTo` properties instead of a simple template string:

```json
{
  "type": "mapper",
  "outputEvent": {
    "key": "{{ issue.key }}",
    "issueId": {
      "value": "{{ issue.id }}",
      "castTo": "string"
    },
    "priority": {
      "value": "{{ issue.fields.priority }}",
      "castTo": "number"
    }
  }
}
```

#### Supported Cast Types

- **`string`**: Converts any value to its string representation
- **`number`**: Converts strings, integers, and booleans to numbers (float64)
  - Booleans: `true` becomes `1`, `false` becomes `0`
  - Strings: Must be valid numeric strings, otherwise an error is thrown

#### Error Handling

If a cast operation fails (e.g., trying to cast "not-a-number" to a number), the processor will return an error 
and the event will not be saved to the sink. This ensures data integrity.

#### Example with Casting

Given an input event where numeric IDs are sent as numbers but need to be stored as strings:

```json
{
  "issue": {
    "id": 155331,
    "key": "ISSUE-123",
    "fields": {
      "priority": "3"
    }
  }
}
```

With this configuration:

```json
{
  "type": "mapper",
  "outputEvent": {
    "key": "{{ issue.key }}",
    "issueId": {
      "value": "{{ issue.id }}",
      "castTo": "string"
    },
    "priority": {
      "value": "{{ issue.fields.priority }}",
      "castTo": "number"
    }
  }
}
```

The output will be:

```json
{
  "key": "ISSUE-123",
  "issueId": "155331",
  "priority": 3
}
```

### Limitation

It is not possible combine multiple fields in a single output field.

For example, given the following input:

```json
{
  "name": "value1",
  "surname": "value2"
}
```

The following configuration is not valid:

```json
{
  "type": "mapper",
  "outputEvent": {
    "fullName": "{{ name }} {{ surname }}"
  }
}
```

### Example

This configuration will map the input data to the `outputEvent` structure.

```json
{
  "type": "mapper",
  "outputEvent": {
    "key": "{{ issue.key }}",
    "createdAt": "{{ issue.fields.created }}",
    "summary": "{{ issue.fields.summary }}",
    "metadata": {
      "description": "{{ issue.fields.description }}"
    }
  }
}
```

Given an input event:

```json
{
  "id": "message-id",
  "issue": {
    "key": "ISSUE-123",
    "fields": {
      "summary": "This is a summary",
      "created": "2021-01-01T00:00:00Z",
      "description": "This is a description"
    }
  }
}
```

It will be transformed into:

```json
{
  "key": "ISSUE-123",
  "summary": "This is a summary",
  "createdAt": "2021-01-01T00:00:00Z",
  "metadata": {
    "description": "This is a description"
  }
}
```
