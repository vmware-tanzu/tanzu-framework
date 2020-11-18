
# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Directories
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

GOBINDATA := $(TOOLS_BIN_DIR)/go-bindata-$(GOOS)-$(GOARCH)

PINNIPED_GIT_REPOSITORY = https://github.com/vmware-tanzu/pinniped.git
ifeq ($(strip $(PINNIPED_GIT_COMMIT)),)
PINNIPED_GIT_COMMIT = main
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

PRIVATE_REPOS="github.com/vmware-tanzu-private"
GO := GOPRIVATE=${PRIVATE_REPOS} go

.DEFAULT_GOAL:=help

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

all: manager

test: generate fmt vet manifests build-cli-mocks ## Run tests
	$(GO) test ./... -coverprofile cover.out

manager: generate fmt vet ## Build manager binary
	$(GO) build -o bin/manager main.go

run: generate fmt vet manifests ## Run against the configured Kubernetes cluster in ~/.kube/config
	$(GO) run ./main.go

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
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

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
	$(GO) mod init tmp ;\
	$(GO) get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
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


ARTIFACTS_DIR ?= ./artifacts

ifeq ($(build_OS), Linux)
XDG_DATA_HOME := ${HOME}/.local/share
endif
ifeq ($(build_OS), Darwin)
XDG_DATA_HOME := ${HOME}/Library/ApplicationSupport
endif

export XDG_DATA_HOME


.PHONY: install-cli
install-cli: ## Install Tanzu CLI
	$(GO) install -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu

.PHONY: build-cli
build-cli: generate-pinniped-bindata ## Build Tanzu CLI
	$(GO) run ./cmd/cli/compiler/main.go --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu"
	$(GO) run ./cmd/cli/compiler/main.go --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin

.PHONY: build-cli-local
build-cli-local: generate-pinniped-bindata ## Build Tanzu CLI
	$(GO) run ./cmd/cli/compiler/main.go --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --target local
	$(GO) run ./cmd/cli/compiler/main.go --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin --target local

.PHONY: build-cli-mocks
build-cli-mocks: ## Build Tanzu CLI mocks
	$(GO) run ./cmd/cli/compiler/main.go --version 0.0.1 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-old --artifacts ./test/cli/mock/artifacts-old
	$(GO) run ./cmd/cli/compiler/main.go --version 0.0.2 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-new --artifacts ./test/cli/mock/artifacts-new
	$(GO) run ./cmd/cli/compiler/main.go --version 0.0.3 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-alt --artifacts ./test/cli/mock/artifacts-alt

.PHONY: test-cli
test-cli: build-cli-mocks ## Run tests
	$(GO) test ./...

.PHONY: build-install-cli-all
build-install-cli-all: clean-cli-plugins build-cli install-cli-plugins install-cli ## Build and install Tanzu CLI plugins

install-cli-plugins: TANZU_CLI_NO_INIT=true

.PHONY: install-cli-plugins
install-cli-plugins:  ## Install Tanzu CLI plugins
	$(GO) run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go \
		plugin install all --local $(ARTIFACTS_DIR)
	$(GO) run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go \
		plugin install all --local $(ARTIFACTS_DIR)-admin
	$(GO) run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go \
		test fetch --local $(ARTIFACTS_DIR)
	$(GO) run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go \
		test fetch --local $(ARTIFACTS_DIR)-admin

.PHONY: clean-cli-plugins
clean-cli-plugins: ## Remove Tanzu CLI plugins
	- rm -rf ${XDG_DATA_HOME}/tanzu-cli/*

.PHONY: generate-pinniped-bindata
generate-pinniped-bindata: $(GOBINDATA)
	@rm -rf pinniped
	@mkdir -p pinniped
	@GIT_TERMINAL_PROMPT=0 git clone ${PINNIPED_GIT_REPOSITORY} pinniped
	cd pinniped && GIT_TERMINAL_PROMPT=0 git checkout -f $(PINNIPED_GIT_COMMIT) && go build -o pinniped ./cmd/pinniped
	$(GOBINDATA) -mode=420 -modtime=1 -o=pkg/v1/auth/tkg/zz_generated.bindata.go -pkg=tkgauth pinniped/pinniped
	@rm -rf pinniped


$(GOBINDATA): $(TOOLS_DIR)/go.mod # Build go-bindata from tools folder
	mkdir -p $(TOOLS_BIN_DIR)
	cd $(TOOLS_DIR); go build -tags=tools -o ../../$(TOOLS_BIN_DIR) github.com/shuLhan/go-bindata/... ; mv ../../$(TOOLS_BIN_DIR)/go-bindata ../../$(GOBINDATA)
