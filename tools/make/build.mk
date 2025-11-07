# Copyright Mia srl
# SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
# See LICENSE.md for more details

##@ Go Builds Goals

.PHONY: build
build:

BUILD_DATE:= $(shell date -u "+%Y-%m-%d")
GO_LDFLAGS+= -s -w

ifdef VERSION_MODULE_NAME
GO_LDFLAGS+= -X $(VERSION_MODULE_NAME).Version=$(VERSION)
GO_LDFLAGS+= -X $(VERSION_MODULE_NAME).BuildDate=$(BUILD_DATE)
endif

.PHONY: go/build/%
go/build/%:
	$(eval OS:= $(word 1,$(subst /, ,$*)))
	$(eval ARCH:= $(word 2,$(subst /, ,$*)))
	$(eval ARM:= $(word 3,$(subst /, ,$*)))
	$(info Building image for $(OS) $(ARCH) $(ARM))

	mkdir -p "$(BUILD_OUTPUT)/$(OS)/$(ARCH)$(if $(ARM),/v$(ARM),)"
	GOOS=$(OS) GOARCH=$(ARCH) GOARM=$(ARM) go build -trimpath \
		-ldflags "$(GO_LDFLAGS)" -o $(BUILD_OUTPUT)/$(OS)/$(ARCH)$(if $(ARM),/v$(ARM),)/$(CMDNAME) $(BUILD_PATH)

.PHONY: build
build: clean/bin go/build/$(GOOS)/$(GOARCH)/$(GOARM)
