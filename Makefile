# Copyright 2015 Google Inc. All rights reserved.
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

REGISTRY   := dfkozlov
GIT_REPO   := $$(basename -s .git `git config --get remote.origin.url`)
GIT_BRANCH := $$(if [ -n "$$BRANCH_NAME" ]; then echo "$$BRANCH_NAME"; else git rev-parse --abbrev-ref HEAD; fi)
GIT_BRANCH := $$(echo "${GIT_BRANCH}" | tr '[:upper:]' '[:lower:]')
GIT_SHA1   := $$(git rev-parse HEAD)
NAME       := ${REGISTRY}/${GIT_REPO}/${GIT_BRANCH}/shared-gpu-gcp-k8s-device-plugin
IMG        := "${NAME}:${GIT_SHA1}"
LATEST     := "${NAME}:latest"
DOCKER_CMD := docker

GO := go
pkgs  = $(shell $(GO) list ./... | grep -v vendor)

all: presubmit

test:
	@echo ">> running tests"
	@$(GO) test -short -race $(pkgs)

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

presubmit: vet
	@echo ">> checking go formatting"
	@./build/check_gofmt.sh .
	@echo ">> checking file boilerplate"
	@./build/check_boilerplate.sh

build:
	cd cmd/nvidia_gpu; go build nvidia_gpu.go

container:
	${DOCKER_CMD} build --pull -t ${IMG} -t ${LATEST} .

push:
	${DOCKER_CMD} push ${LATEST}
	${DOCKER_CMD} push ${IMG}

.PHONY: all format test vet presubmit build container push
