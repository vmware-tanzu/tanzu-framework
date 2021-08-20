ROOT_DIR_RELATIVE := .

include $(ROOT_DIR_RELATIVE)/common.mk

BUILD_VERSION ?= $$(cat BUILD_VERSION)
BUILD_SHA ?= $$(git rev-parse --short HEAD)
BUILD_DATE ?= $$(date -u +"%Y-%m-%d")

LD_FLAGS = -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli.BuildDate=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli.BuildSHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli.BuildVersion=$(BUILD_VERSION)'

TOOLS_DIR := tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint

PRIVATE_REPOS="github.com/vmware-tanzu/tanzu-framework"

export GOPRIVATE := $(PRIVATE_REPOS)

GO_SRCS := $(call rwildcard,.,*.go)

go.mod go.sum: $(GO_SRCS)
	go mod download
	go mod tidy

.PHONY: build
build: ## Build the plugin
	tanzu builder cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/plugin --goprivate "$(PRIVATE_REPOS)"

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Lint the plugin
	$(GOLANGCI_LINT) run -v

.PHONY: init
init:go.mod go.sum  ## Initialise the plugin

.PHONY: test
test: $(GO_SRCS) go.sum
	go test ./...

$(TOOLS_BIN_DIR):
	-mkdir -p $@

$(GOLANGCI_LINT): $(TOOLS_BIN_DIR)
	go build -tags=tools -o $@ github.com/golangci/golangci-lint/cmd/golangci-lint
