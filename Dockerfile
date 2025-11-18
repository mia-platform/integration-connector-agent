# syntax=docker/dockerfile:1
FROM docker.io/library/golang:1.25.3-trixie@sha256:5e43883cbcb374df63de6fb695a86218a852b829055d2aa88c260a4189be46c5 AS builder

ARG TARGETPLATFORM

WORKDIR /build

COPY go.* .
RUN go mod download
COPY . .

RUN make build

RUN mkdir /app && cp -r LICENSE bin/${TARGETPLATFORM}/integration-connector-agent /app

FROM gcr.io/distroless/base-debian13:nonroot@sha256:f01077109844dbd3627d3ffbf9a16dc37c13016d96bfa1781761aa056a46f412

COPY --from=builder /app /app

# Use an unprivileged user.
USER 1000

CMD ["/app/integration-connector-agent"]
