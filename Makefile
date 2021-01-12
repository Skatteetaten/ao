# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# The binary to build (just the basename).
BIN := ao

# This repo's root import path (under GOPATH).
PKG := github.com/skatteetaten/ao

#  Build OS detection
BUILD_OS := $(shell uname -s | tr A-Z a-z)

# Which architecture to build - see $(ALL_ARCH) for options.
ARCH ?= amd64

###
### These variables should not need tweaking.
###

SRC_DIRS := cmd pkg # directories which hold app source (not vendored)

ARCH := amd64

GOPATH := $(shell pwd)/.go

GOSRC := $(shell pwd)/.go/src

VERSION := $(shell git describe --tags --always --dirty)

BRANCH := $(shell git branch | sed -n -e 's/^\* \(.*\)/\1/p')

BUILDSTAMP := $(shell date '+%Y-%m-%d_%H:%M:%S%p')

GITHASH := $(shell git rev-parse HEAD)

# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
all: build

build: build-dirs bin-file-linux bin-file-darwin bin-file-windows

bin-file-linux:
	@echo "Building for Linux with GoPath: $(GOPATH) , GoSrc: $(GOSRC) , Version: $(VERSION)"
ifneq (,$(findstring dirty,$(VERSION)))
	@/bin/sh -c "git status"
	@/bin/sh -c "git diff"
endif
	@/bin/sh -c "                                                          \
	        cd .go/src/$(PKG);                                             \
	        GOPATH=$(GOPATH)                                               \
	        GOSRC=$(GOSRC)                                                 \
			OS=linux													   \
	        ARCH=$(ARCH)                                                   \
	        PKG=$(PKG)                                                     \
	        VERSION=$(VERSION)                                             \
	        BRANCH=$(BRANCH)                                               \
	        BUILDSTAMP=$(BUILDSTAMP)                                       \
	        GITHASH=$(GITHASH)                                             \
	        ./build/build.sh                                               \
	    "

bin-file-darwin:
	@echo "Building for Darwin with GoPath: $(GOPATH) , GoSrc: $(GOSRC) , Version: $(VERSION)"
	@/bin/sh -c "                                                          \
	        cd .go/src/$(PKG);                                             \
	        GOPATH=$(GOPATH)                                               \
	        GOSRC=$(GOSRC)                                                 \
			OS=darwin													   \
	        ARCH=$(ARCH)                                                   \
	        PKG=$(PKG)                                                     \
	        VERSION=$(VERSION)                                             \
	        BRANCH=$(BRANCH)                                               \
	        BUILDSTAMP=$(BUILDSTAMP)                                       \
	        GITHASH=$(GITHASH)                                             \
	        ./build/build.sh                                               \
	    "

bin-file-windows:
	@echo "Building for Windows with GoPath: $(GOPATH) , GoSrc: $(GOSRC) , Version: $(VERSION)"
	@/bin/sh -c "                                                          \
	        cd .go/src/$(PKG);                                             \
	        GOPATH=$(GOPATH)                                               \
	        GOSRC=$(GOSRC)                                                 \
			OS=windows													   \
	        ARCH=$(ARCH)                                                   \
	        PKG=$(PKG)                                                     \
	        VERSION=$(VERSION)                                             \
	        BRANCH=$(BRANCH)                                               \
	        BUILDSTAMP=$(BUILDSTAMP)                                       \
	        GITHASH=$(GITHASH)                                             \
	        ./build/build.sh                                               \
	    "


test: build-dirs
	    @/bin/sh -c "                                                      \
	    cd .go/src/$(PKG);                                                 \
	    GOPATH=$(GOPATH)                                                   \
	    GOSRC=$(GOSRC)                                                     \
	    ./build/test.sh $(SRC_DIRS)                                        \
	    "

build-dirs: .go/src/$(PKG)
	@mkdir -p bin/amd64
	@mkdir -p bin/darwin_amd64
	@mkdir -p bin/windows_amd64
	@mkdir -p .go/pkg .go/bin .go/bin/linux_$(ARCH) .go/bin/darwin_$(ARCH) .go/bin/windows_$(ARCH)

.go/src/$(PKG):
	@mkdir -p .go/src/$(PKG)
	@rmdir .go/src/$(PKG)
	@echo "Build OS: $(BUILD_OS)"
ifeq ($(BUILD_OS),darwin)
	@ln -s ../../../../ .go/src/$(PKG)
else
	@ln -s -r . .go/src/$(PKG)
endif

clean:
	rm -rf .go bin
