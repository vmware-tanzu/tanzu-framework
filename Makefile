
# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.DEFAULT_GOAL:=help

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

all: manager

test: generate fmt vet manifests build-cli-mocks ## Run tests
	go test ./... -coverprofile cover.out

manager: generate fmt vet ## Build manager binary
	go build -o bin/manager main.go

run: generate fmt vet manifests ## Run against the configured Kubernetes cluster in ~/.kube/config
	go run ./main.go

install: manifests ## Install CRDs into a cluster
	kustomize build config/crd | kubectl apply -f -

uninstall: manifests ## Uninstall CRDs from a cluster
	kustomize build config/crd | kubectl delete -f -

deploy: manifests ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

manifests: controller-gen ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

fmt: ## Run go fmt
	go fmt ./...

vet: ## Run go vet
	go vet ./...

generate: controller-gen ## Generate code via controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

docker-build: test ## Build the docker image
	docker build . -t ${IMG}

docker-push: ## Push the docker image
	docker push ${IMG}

controller-gen: ## Download controller-gen
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

BUILD_VERSION ?= $$(cat BUILD_VERSION)
BUILD_SHA ?= $$(git rev-parse --short HEAD)
BUILD_DATE ?= $$(date -u +"%Y-%m-%d")

build_OS := $(shell uname 2>/dev/null || echo Unknown)

LD_FLAGS = -X 'github.com/vmware-tanzu-private/core/pkg/v1/cli.BuildDate=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu-private/core/pkg/v1/cli.BuildSHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu-private/core/pkg/v1/cli.BuildVersion=$(BUILD_VERSION)'


ARTIFACTS_DIR ?= "./artifacts"

ifeq ($(build_OS), Linux)
XDG_DATA_HOME := ${HOME}/.local/share
endif
ifeq ($(build_OS), Darwin)
XDG_DATA_HOME := ${HOME}/Library/ApplicationSupport
endif

export XDG_DATA_HOME


.PHONY: install-cli
install-cli: ## Install Tanzu CLI
	go install -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu

.PHONY: build-cli
build-cli: ## Build Tanzu CLI
	go run ./cmd/cli/compiler/main.go --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu"

.PHONY: build-cli-mocks
build-cli-mocks: ## Build Tanzu CLI mocks
	go run ./cmd/cli/compiler/main.go --version 0.0.1 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-old --artifacts ./test/cli/mock/artifacts-old 
	go run ./cmd/cli/compiler/main.go --version 0.0.2 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-new --artifacts ./test/cli/mock/artifacts-new
	go run ./cmd/cli/compiler/main.go --version 0.0.3 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-alt --artifacts ./test/cli/mock/artifacts-alt

.PHONY: test-cli
test-cli: ## Run tests
	go test ./...

.PHONY: build-install-cli-plugins
build-install-cli-plugins: clean-cli-plugins build-cli install-cli-plugins install-cli ## Build and install Tanzu CLI plugins

.PHONY: install-cli-plugins
install-cli-plugins: ## Install Tanzu CLI plugins
	go run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go \
		plugin install all --local $(ARTIFACTS_DIR)

.PHONY: clean-cli-plugins
clean-cli-plugins: ## Remove Tanzu CLI plugins
	- rm ${XDG_DATA_HOME}/tanzu-cli/*
