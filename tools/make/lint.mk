# Copyright Mia srl
# SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
# See LICENSE.md for more details

##@ Lint Goals

# if not already installed in the system install a pinned version in tools folder
GOLANGCI_PATH:= $(shell command -v golangci-lint 2> /dev/null)
ifndef GOLANGCI_PATH
	GOLANGCI_PATH:=$(TOOLS_BIN)/golangci-lint
endif

.PHONY: lint
lint:

.PHONY: lint-deps
lint-deps:

.PHONY: golangci-lint
lint: golangci-lint
golangci-lint: $(GOLANGCI_PATH)
	$(info Running golangci-lint with .golangci.yaml config file...)
	$(GOLANGCI_PATH) run --config=.golangci.yaml

lint-deps: $(GOLANGCI_PATH)
$(TOOLS_BIN)/golangci-lint: $(TOOLS_DIR)/GOLANGCI_LINT_VERSION
	$(eval GOLANGCI_LINT_VERSION:= $(shell cat $<))
	mkdir -p $(TOOLS_BIN)
	$(info Installing golangci-lint $(GOLANGCI_LINT_VERSION) bin in $(TOOLS_BIN))
	GOBIN=$(TOOLS_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: gomod-lint
lint: gomod-lint
gomod-lint:
	$(info Running go mod tidy)
# Always keep this version to latest -1 version of Go
	go mod tidy -compat=1.24

.PHONY: ci-lint
ci-lint: lint
# Block the lint during ci if the go.mod and go.sum will be changed by go mod tidy
	git diff --exit-code go.mod;
	git diff --exit-code go.sum;
