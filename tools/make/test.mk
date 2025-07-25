# Copyright Mia srl
# SPDX-License-Identifier: Apache-2.0

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#    http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

##@ Go Tests Goals

DEBUG_TEST?=
ifeq ($(DEBUG_TEST),1)
GO_TEST_DEBUG_FLAG:= -v
else
GO_TEST_DEBUG_FLAG:=
endif

.PHONY: test/build-plugin
test/build-plugin:
	$(info Building RPC mock plugin for tests...)
	go build -o ./internal/processors/hcgp/testdata/mockplugin/mockplugin ./internal/processors/hcgp/testdata/mockplugin/*.go

.PHONY: test/unit
test/unit:
	$(info Running tests...)
	go test $(GO_TEST_DEBUG_FLAG) -race ./...

.PHONY: test/integration/setup test/integration test/integration/teardown
test/integration/setup:
	$(info Setup mongo...)
	docker run --rm --name mongo -p 27017:27017 -d mongo
	$(info Setup gcloud pubsub emulator...)
	docker run --rm --name gcloud-pubsub-emulator -p 8085:8085 -d gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators gcloud beta emulators pubsub start --project=test-project-id
test/integration:
	$(info Running integration tests...)
	PUBSUB_EMULATOR_HOST=localhost:8085 go test $(GO_TEST_DEBUG_FLAG) -tags=integration -count=1 -cover -race ./...
test/integration/teardown:
	$(info Teardown integration tests...)
	docker rm mongo --force
	docker rm gcloud-pubsub-emulator --force

.PHONY: test/coverage
test/coverage:
	$(info Running tests with coverage on...)
	go test $(GO_TEST_DEBUG_FLAG) -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: test/integration/coverage
test/integration/coverage:
	$(info Running ci tests with coverage on...)
	PUBSUB_EMULATOR_HOST=localhost:8085 go test $(GO_TEST_DEBUG_FLAG) -tags=integration -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: test/conformance test/conformance/setup test/conformance/teardown
test/conformance/setup:
test/conformance:
	$(info Running conformance tests...)
	go test $(GO_TEST_DEBUG_FLAG) -tags=conformance -race -count=1 $(CONFORMANCE_TEST_PATH)
test/conformance/teardown:

test/show/coverage:
	go tool cover -func=coverage.txt

.PHONY: test
test: test/build-plugin test/unit

.PHONY: test-coverage
test-coverage: test/build-plugin test/coverage

.PHONY: test-integration
test-integration: test/build-plugin test/integration/setup test/integration test/integration/teardown

.PHONY: test-integration-coverage
test-integration-coverage: test/build-plugin test/integration/setup test/integration/coverage test/integration/teardown

.PHONY: test-conformance
test-conformance: test/build-plugin test/conformance/setup test/conformance test/conformance/teardown

.PHONY: show-coverage
show-coverage: test/build-plugin test-coverage test/show/coverage

.PHONY: show-integration-coverage
show-integration-coverage: test/build-plugin test-integration-coverage test/show/coverage
