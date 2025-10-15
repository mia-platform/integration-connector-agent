# syntax=docker/dockerfile:1
FROM docker.io/library/golang:1.25.3-bookworm@sha256:51b6b12427dc03451c24f7fc996c43a20e8a8e56f0849dd0db6ff6e9225cc892 AS builder

ARG TARGETPLATFORM

WORKDIR /build

COPY go.* .
RUN go mod download
COPY . .

RUN make build

RUN mkdir /app && cp -r LICENSE bin/${TARGETPLATFORM}/integration-connector-agent /app

FROM gcr.io/distroless/base-debian12:nonroot@sha256:10136f394cbc891efa9f20974a48843f21a6b3cbde55b1778582195d6726fa85

COPY --from=builder /app /app

# Use an unprivileged user.
USER 1000

CMD ["/app/integration-connector-agent"]
