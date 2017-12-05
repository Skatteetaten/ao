#!/bin/bash

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

set -o errexit
set -o nounset
set -o pipefail
if [ -z "${PKG}" ]; then
    echo "PKG must be set"
    exit 1
fi
if [ -z "${ARCH}" ]; then
    echo "ARCH must be set"
    exit 1
fi
if [ -z "${VERSION}" ]; then
    echo "VERSION must be set"
    exit 1
fi
if [ -z "${GITHASH}" ]; then
    echo "GITHASH must be set"
    exit 1
fi
if [ -z "${BUILDSTAMP}" ]; then
    echo "BUILDSTAMP must be set"
    exit 1
fi
if [ -z "${BRANCH}" ]; then
    echo "BRANCH must be set"
    exit 1
fi


export CGO_ENABLED=0
export GOARCH="${ARCH}"

#
# We have a lot of dependencies. If we use ./... we need to import the whole world.
# So for now, we filter on our own dependencies
#

PACKAGES=$(go list ./... | grep "ao/pkg\|ao$\|ao/cmd" | xargs echo)
go install                                                         \
    -ldflags "-X \"${PKG}/pkg/config.Version=${VERSION}\" -X \"${PKG}/pkg/config.Branch=${BRANCH}\" -X \"${PKG}/pkg/config.BuildStamp=${BUILDSTAMP}\" -X \"${PKG}/pkg/config.GitHash=${GITHASH}\"" \
    -gcflags='-B -l' \
    -pkgdir=${GOPATH}/pkg \
    ${PACKAGES}

