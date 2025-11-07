# Copyright Mia srl
# SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
# See LICENSE.md for more details

##@ Deepcopy Goals

.PHONY: generate-deps
generate-deps:

.PHONY: generate
generate: generate-deps
	go generate -x -ldflags "$(GO_LDFLAGS)" ./...
