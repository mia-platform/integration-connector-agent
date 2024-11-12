# syntax=docker/dockerfile:1
FROM docker.io/library/alpine:3.20.3@sha256:1e42bbe2508154c9126d48c2b8a75420c3544343bf86fd041fb7527e017a4b4a AS builder

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
