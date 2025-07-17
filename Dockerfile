# syntax=docker/dockerfile:1
FROM docker.io/library/alpine:3.22.1@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1 AS builder

ARG TARGETPLATFORM

WORKDIR /app

COPY bin/${TARGETPLATFORM}/integration-connector-agent .
COPY LICENSE .

FROM scratch

# Import the certs from the builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app /app

# Use an unprivileged user.
USER 1000

CMD ["/app/integration-connector-agent"]
