# syntax=docker/dockerfile:1
FROM docker.io/library/golang:1.25.6-trixie@sha256:fb4b74a39c7318d53539ebda43ccd3ecba6e447a78591889c0efc0a7235ea8b3 AS builder

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
