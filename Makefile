# Copyright Mia srl
# SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
# See LICENSE.md for more details

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

ifeq ($(origin BUILD_OUTPUT),undefined)
BUILD_OUTPUT:= $(PROJECT_DIR)/bin
endif

# Set here the name of the package you want to build
CMDNAME:= integration-connector-agent
BUILD_PATH:= .
CONFORMANCE_TEST_PATH:= $(PROJECT_DIR)/tests/e2e

# enable modules
GO111MODULE:= on
GOOS:= $(shell go env GOOS)
GOARCH:= $(shell go env GOARCH)
GOARM:= $(shell go env GOARM)

## Build Variables
GIT_REV:= $(shell git rev-parse --short HEAD 2>/dev/null)
VERSION:= $(shell git describe --tags --exact-match 2>/dev/null || (echo $(GIT_REV) | cut -c1-12))
# insert here the go module where to add the version metadata
VERSION_MODULE_NAME:= main

# Add additional targets that you want to run when calling make without arguments
.PHONY: all
all: lint test

## Includes
include tools/make/build.mk
include tools/make/clean.mk
include tools/make/generate.mk
include tools/make/lint.mk
include tools/make/test.mk

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

test/build-plugin:
	$(info Building RPC mock plugin for tests...)
	go build -o ./internal/processors/hcgp/testdata/mockplugin/mockplugin ./internal/processors/hcgp/testdata/mockplugin/*.go

test/unit: test/build-plugin
test/coverage: test/build-plugin
test/integration: test/build-plugin

test/integration/setup:
	$(info Setup mongo...)
	docker run --rm --name mongo -p 27017:27017 -d mongo
	$(info Setup gcloud pubsub emulator...)
	docker run --rm --name gcloud-pubsub-emulator -p 8085:8085 -d gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators gcloud beta emulators pubsub start --project=test-project-id
	$(eval PUBSUB_EMULATOR_HOST:= localhost:8085)

test/integration/teardown:
	$(info Teardown integration tests...)
	docker rm mongo --force
	docker rm gcloud-pubsub-emulator --force

clean-mockplugin:
	$(info Cleaning up mock plugin binary...)
	rm -f ./internal/processors/hcgp/testdata/mockplugin/mockplugin

clean: clean-mockplugin
