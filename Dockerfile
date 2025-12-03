# syntax=docker/dockerfile:1
FROM docker.io/library/golang:1.25.5-trixie@sha256:4f9d98ebaa759f776496d850e0439c48948d587b191fc3949b5f5e4667abef90 AS builder

ARG TARGETPLATFORM

WORKDIR /build

COPY go.* .
RUN go mod download
COPY . .

RUN make build

RUN mkdir /app && cp -r LICENSE bin/${TARGETPLATFORM}/integration-connector-agent /app

FROM gcr.io/distroless/base-debian13:nonroot@sha256:4179ca36333695b889c9e6664ba26a627775a7978a8d5b6cd95d5b3b6a84b1e6

COPY --from=builder /app /app

# Use an unprivileged user.
USER 1000

CMD ["/app/integration-connector-agent"]
