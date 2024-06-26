
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Versions
KUSTOMIZE_VERSION ?= v5.3.0
CONTROLLER_TOOLS_VERSION ?= v0.14.0

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: test
test: ## Run tests locally
	go test -cover ./... -v 2

.PHONY: build
build: ## Build the Golang CLI binary
	go build -o swdt .

##@ Packer 

WINDOWS_ISO = './packer/isos/windows.iso'
WINDOWS_SHA256 = 'sha256:3e4fa6d8507b554856fc9ca6079cc402df11a8b79344871669f0251535255325'
VIRTIO_ISO = '${PWD}/packer/isos/virtio.iso'
KEY_PATH = '${HOME}/.ssh/id_rsa.pub'

.PHONY: packer
packer: ## Build the Windows QCOW2 image from ISO
	rm -fr packer/output
	packer init packer/kvm
	ln -s $(KEY_PATH) ./packer/kvm/floppy/ssh_key.pub || true
	PACKER_LOG=1 packer build \
		-var 'windows_iso=$(WINDOWS_ISO)' \
		-var 'windows_sha256=$(WINDOWS_SHA256)' \
		-var 'virtio_iso=$(VIRTIO_ISO)' \
		packer/kvm/

##@ Commands
start: ## Execute swdt start to bootstrap the domains
	./swdt start -v=5

setup: ## Execute swdt start to bootstrap the domains
	./swdt setup -v=5

destroy: ## Execute swdt destroy to remove all domains
	./swdt destroy -v=5
