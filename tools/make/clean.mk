# Copyright Mia srl
# SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
# See LICENSE.md for more details

##@ Lint Goals

.PHONY: clean
clean:

.PHONY: clean/coverage
clean: clean/coverage
clean/coverage:
	$(info Clean coverage file...)
	rm -fr coverage.txt

.PHONY: clean/bin
clean: clean/bin
clean/bin:
	$(info Clean artifacts files...)
	rm -fr $(OUTPUT_DIR)

.PHONY: clean/tools
clean/tools:
	$(info Clean tools folder...)
	[ -d $(TOOLS_BIN)/k8s ] && chmod +w $(TOOLS_BIN)/k8s/* || true
	rm -fr $(TOOLS_BIN)

.PHONY: clean/go
clean/go:
	$(info Clean golang cache...)
	go clean -cache

.PHONY: clean-all
clean-all: clean clean/tools clean/go
