
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

all: manager

# Run tests
test: generate fmt vet manifests build-cli-mocks
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
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
install-cli:
	go install -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu

.PHONY: build-cli
build-cli:
	go run ./cmd/cli/compiler/main.go --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu"
	go run ./cmd/cli/compiler/main.go --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin

.PHONY: build-cli-mocks
build-cli-mocks:
	go run ./cmd/cli/compiler/main.go --version 0.0.1 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-old --artifacts ./test/cli/mock/artifacts-old 
	go run ./cmd/cli/compiler/main.go --version 0.0.2 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-new --artifacts ./test/cli/mock/artifacts-new
	go run ./cmd/cli/compiler/main.go --version 0.0.3 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-alt --artifacts ./test/cli/mock/artifacts-alt

.PHONY: test-cli
test-cli: build-cli-mocks
	go test ./...

.PHONY: build-install-cli-plugins
build-install-cli-plugins: clean-cli-plugins build-cli install-cli-plugins install-cli

.PHONY: install-cli-plugins
install-cli-plugins:
	go run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go \
		plugin install all --local $(ARTIFACTS_DIR)

.PHONY: clean-cli-plugins
clean-cli-plugins:
	- rm ${XDG_DATA_HOME}/tanzu-cli/*
