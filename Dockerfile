# syntax=docker/dockerfile:1
FROM docker.io/library/alpine:3.21.0@sha256:21dc6063fd678b478f57c0e13f47560d0ea4eeba26dfc947b2a4f81f686b9f45 AS builder

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
