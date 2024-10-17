# Copyright Mia srl
# SPDX-License-Identifier: Apache-2.0

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
CMDNAME:= REPO_NAME
BUILD_PATH:= .
CONFORMANCE_TEST_PATH:= $(PROJECT_DIR)/tests/e2e

# enable modules
GO111MODULE:= on
GOOS:= $(shell go env GOOS)
GOARCH:= $(shell go env GOARCH)
GOARM:= $(shell go env GOARM)

# Add additional targets that you want to run when calling make without arguments
.PHONY: all
all: build

## Includes
include tools/make/clean.mk
include tools/make/lint.mk
include tools/make/test.mk
include tools/make/build.mk

# Uncomment the correct test suite to run during CI
.PHONY: ci
ci: test-coverage

### Put your custom import, define or goals under here ###
