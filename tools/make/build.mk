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

GO_LDFLAGS+= -s -w

.PHONY: go/build/%
go/build/%:
	$(eval OS:= $(word 1,$(subst /, ,$*)))
	$(eval ARCH:= $(word 2,$(subst /, ,$*)))
	$(eval ARM:= $(word 3,$(subst /, ,$*)))
	$(info Building binary for $(OS) $(ARCH) $(ARM))

	GOOS=$(OS) GOARCH=$(ARCH) GOARM=$(ARM) go build -C $(BUILD_PATH) -trimpath \
		-ldflags "$(GO_LDFLAGS)" -o bin/$(OS)/$(ARCH)/$(ARM)/$(CMDNAME)

.PHONY: build
build: go/build/$(GOOS)/$(GOARCH)/$(GOARM)
