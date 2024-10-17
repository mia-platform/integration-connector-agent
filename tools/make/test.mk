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

.PHONY: test/unit
test/unit:
	$(info Running tests...)
	go test $(GO_TEST_DEBUG_FLAG) -race ./...

.PHONY: test/coverage
test/coverage:
	$(info Running tests with coverage on...)
	go test $(GO_TEST_DEBUG_FLAG) -race -coverprofile=coverage.txt -covermode=atomic ./...

test/show/coverage:
	go tool cover -func=coverage.txt

.PHONY: test
test: test/unit

.PHONY: test-coverage
test-coverage: test/coverage

.PHONY: show-coverage
show-coverage: test-coverage test/show/coverage
