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

# Which architecture to build - see $(ALL_ARCH) for options.
ARCH ?= amd64

###
### These variables should not need tweaking.
###

SRC_DIRS := cmd pkg # directories which hold app source (not vendored)

#ARCH := amd64
ARCH := amd64

VERSION := unset


GOPATH := $(shell pwd)/.go

GOSRC := $(shell pwd)/.go/src

GOBIN := $(shell pwd)/bin/$(ARCH)

# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
all: build

deps:
	@echo "installing deps"
	@glide install


build: build-dirs bin-file

bin-file:
	@echo "Building with GoPath : $(GOPATH) and GoSrc $(GOSRC)"
	@/bin/sh -c "                                                          \
	        cd .go/src/$(PKG);                                             \
	        GOPATH=$(GOPATH)                                               \
	        GOSRC=$(GOSRC)                                                 \
	        GOBIN=$(GOBIN)                                                 \
	        ARCH=$(ARCH)                                                   \
	        PKG=$(PKG)                                                     \
	        VERSION=$(VERSION)                                             \
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
	@mkdir -p bin/$(ARCH)
	@mkdir -p .go/pkg .go/bin .go/std/$(ARCH)

.go/src/$(PKG):
	@mkdir -p .go/src/$(PKG)
	@rmdir .go/src/$(PKG)
	@ln -s -r . .go/src/$(PKG)


clean:
	rm -rf .go bin
