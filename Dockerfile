# syntax=docker/dockerfile:1
FROM docker.io/library/alpine:3.20.3@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb737fc57cd1367effd16dc0d06d AS builder

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
