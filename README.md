# data-connector-agent

This is a simple Go application template with a pre-configured [logger] and a
[library] to handle configuration file and env variables.
It also contains basic dependencies for testing and http request.
By default the module name is "service", if you want to change it, please remember to change it in the imports too.

## Development Local

To develop the service locally you need:
	- Go 1.23+

To start the application locally

```go
go run data-connector-agent
```

By default the service will run on port 8080, to change the port please set `HTTP_PORT` env variable

## Testing

To test the application use:

```go
make test
```

[logger]: https://github.com/mia-platform/glogger
[library]: https://github.com/mia-platform/configlib
