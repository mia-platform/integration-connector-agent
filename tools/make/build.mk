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
