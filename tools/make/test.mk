# Copyright Mia srl
# SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
# See LICENSE.md for more details

##@ Go Tests Goals

DEBUG_TEST?=
ifeq ($(DEBUG_TEST),1)
GO_TEST_DEBUG_FLAG:= -v
else
GO_TEST_DEBUG_FLAG:=
endif

.PHONY: test/unit
test/unit:
	$(info Running tests...)
	go test $(GO_TEST_DEBUG_FLAG) -race ./...

.PHONY: test/integration/setup test/integration test/integration/teardown
test/integration/setup:
test/integration:
	$(info Running integration tests...)
	go test $(GO_TEST_DEBUG_FLAG) -tags=integration -race ./...
test/integration/teardown:

.PHONY: test/coverage
test/coverage:
	$(info Running tests with coverage on...)
	go test $(GO_TEST_DEBUG_FLAG) -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: test/integration/coverage
test/integration/coverage:
	$(info Running ci tests with coverage on...)
	go test $(GO_TEST_DEBUG_FLAG) -tags=integration -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: test/conformance test/conformance/setup test/conformance/teardown
test/conformance/setup:
test/conformance:
	$(info Running conformance tests...)
	go test $(GO_TEST_DEBUG_FLAG) -tags=conformance -race -count=1 $(CONFORMANCE_TEST_PATH)
test/conformance/teardown:

test/show/coverage:
	go tool cover -func=coverage.txt

.PHONY: test
test: test/unit

.PHONY: test-coverage
test-coverage: test/coverage

.PHONY: test-integration
test-integration: test/integration/setup test/integration test/integration/teardown

.PHONY: test-integration-coverage
test-integration-coverage: test/integration/setup test/integration/coverage test/integration/teardown

.PHONY: test-conformance
test-conformance: test/conformance/setup test/conformance test/conformance/teardown

.PHONY: show-coverage
show-coverage: test-coverage test/show/coverage

.PHONY: show-integration-coverage
show-integration-coverage: test-integration-coverage test/show/coverage
