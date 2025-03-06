# syntax=docker/dockerfile:1
FROM docker.io/library/alpine:3.21.2@sha256:56fa17d2a7e7f168a043a2712e63aed1f8543aeafdcee47c58dcffe38ed51099 AS builder

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
