# TODO: I'm messing up with the build chain as I need CGO_ENABLED and cross compilation.
# This dockerfile is WIP, we shall at least:
#   - don't use root user (pretty plz)
#   - consider whether building the golang binary outside the container (using goreleaser with additional setups)
#   - if not, consider removing goreleaser as we are not going to use multi-arch builds anymore (for now, we might want to build arm64 images)
#   - clean-up .dockerignore file as I've commented out some lines (go.mod and .go are pretty much needed for the build)

# syntax=docker/dockerfile:1

###### ORIGINAL DOCKERFILE FIRST STAGE ######

# FROM docker.io/library/alpine:3.22.0@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715 AS builder
# ARG TARGETPLATFORM
# WORKDIR /app
# COPY bin/${TARGETPLATFORM}/integration-connector-agent .
# COPY LICENSE .

###### END OF DOCKERFILE FIRST STAGE ######

FROM docker.io/library/golang:1.24.3 AS builder

WORKDIR /dist

ARG TARGETOS
ARG TARGETARCH

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download
RUN go mod verify

COPY . .

RUN GOOS="${TARGETOS}" CGO_ENABLED=1 GOARCH="${TARGETARCH}" go build -ldflags="-w -s" -o integration-connector-agent .

WORKDIR /app


RUN cp /dist/integration-connector-agent .

FROM alpine:3.22.0
# FROM scratch

RUN apk add --no-cache libc6-compat
# Import the certs from the builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

#COPY --from=builder /app /app
# Copia direttamente il binario e la licenza nella posizione finale
COPY --from=builder /app/integration-connector-agent /app/integration-connector-agent
COPY ./LICENSE /app/LICENSE

# Use an unprivileged user.
#USER 1000

CMD ["/app/integration-connector-agent"]
