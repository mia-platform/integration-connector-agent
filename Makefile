# Copyright Mia srl
# SPDX-License-Identifier: Ap# Set here the name of the package you want to build
CMDNAME:= integration-connector-agent
BUILD_PATH:= .

# Version for pipeline builds
VERSION ?= latest

# Create a variable that contains the current date in UTC
# Different flow if this script is running on Darwin or Linux machines.
ifeq (Darwin,$(shell uname))
	NOW_DATE = $(shell date -u +%d-%m-%Y)
else
	NOW_DATE = $(shell date -u -I)
endif

# enable modules
GO111MODULE:= on
GOOS:= $(shell go env GOOS)
GOARCH:= $(shell go env GOARCH)
GOARM:= $(shell go env GOARM)
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DEBUG_MAKEFILE?=
ifeq ($(DEBUG_MAKEFILE),1)
$(warning ***** executing goal(s) "$(MAKECMDGOALS)")
$(warning ***** $(shell date))
else
# If we're not debugging the Makefile, always hide the commands inside the goals
MAKEFLAGS+= -s
endif

# It's necessary to set this because some environments don't link sh -> bash.
# Using env is more portable than setting the path directly
SHELL:= /usr/bin/env bash

.EXPORT_ALL_VARIABLES:

.SUFFIXES:

## Set all variables
ifeq ($(origin PROJECT_DIR),undefined)
PROJECT_DIR:= $(abspath $(shell pwd -P))
endif

ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR:= $(PROJECT_DIR)/bin
endif

ifeq ($(origin TOOLS_DIR),undefined)
TOOLS_DIR:= $(PROJECT_DIR)/tools
endif

ifeq ($(origin TOOLS_BIN),undefined)
TOOLS_BIN:= $(TOOLS_DIR)/bin
endif

# Set here the name of the package you want to build
CMDNAME:= integration-connector-agent
BUILD_PATH:= .

# enable modules
GO111MODULE:= on
GOOS:= $(shell go env GOOS)
GOARCH:= $(shell go env GOARCH)
GOARM:= $(shell go env GOARM)

# supported platforms for container creation, these are a subset of the supported
# platforms of the base image.
# Or if you start from scratch the platforms you want to support in your image
# This link contains the rules on how the strings must be formed https://github.com/containerd/containerd/blob/v1.4.3/platforms/platforms.go#L63
SUPPORTED_PLATFORMS:= linux/amd64 linux/arm64
# Default platform for which building the docker image (darwin can run linux images for the same arch)
# as SUPPORTED_PLATFORMS it highly depends on which platform are supported by the base image
DEFAULT_DOCKER_PLATFORM:= linux/$(GOARCH)/$(GOARM)
# List of one or more container registries for tagging the resulting docker images
CONTAINER_REGISTRIES:= docker.io/miaplatform ghcr.io/mia-platform nexus.mia-platform.eu/plugins
# The description used on the org.opencontainers.description label of the container
DESCRIPTION:=
# The vendor name used on the org.opencontainers.image.vendor label of the container
VENDOR_NAME:= Mia s.r.l.
# The license used on the org.opencontainers.image.license label of the container
LICENSE:= Apache-2.0
# The documentation url used on the org.opencontainers.image.documentation label of the container
DOCUMENTATION_URL:= https://docs.mia-platform.eu
# The source url used on the org.opencontainers.image.source label of the container
SOURCE_URL:= https://github.com/mia-platform/integration-connector-agent
BUILDX_CONTEXT?= integration-connector-agent-build-context

# Add additional targets that you want to run when calling make without arguments
.PHONY: all
all: lint test

## Includes
include tools/make/clean.mk
include tools/make/lint.mk
include tools/make/test.mk
include tools/make/generate.mk
include tools/make/build.mk
include tools/make/container.mk
include tools/make/release.mk

# Uncomment the correct test suite to run during CI
.PHONY: ci
# ci: test-coverage
ci: test-integration-coverage

### Put your custom import, define or goals under here ###

generate-deps: $(TOOLS_BIN)/stringer
$(TOOLS_BIN)/stringer: $(TOOLS_DIR)/STRINGER_VERSION
	$(eval STRINGER_VERSION:= $(shell cat $<))
	mkdir -p $(TOOLS_BIN)
	$(info Installing stringer $(STRINGER_VERSION) bin in $(TOOLS_BIN))
	GOBIN=$(TOOLS_BIN) go install golang.org/x/tools/cmd/stringer@$(STRINGER_VERSION)

##@ Pipeline Targets

.PHONY: test-pipeline
test-pipeline: test/build-plugin
	$(info Running tests for pipeline with coverage...)
	go test ./... -coverprofile coverage.out

.PHONY: version
version:
	sed -i.bck "s|SERVICE_VERSION=\"[0-9]*.[0-9]*.[0-9]*.*\"|SERVICE_VERSION=\"${VERSION}\"|" "Dockerfile"
	rm -fr "Dockerfile.bck"
	git add "Dockerfile"
	git commit -m "bump v${VERSION}"
	git tag v${VERSION}
