// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template

import "github.com/vmware-tanzu-private/core/pkg/v1/builder/template/plugintemplates"

// GoMod target
var GoMod = Target{
	Filepath: "go.mod",
	Template: plugintemplates.Gomod,
}

// BuildVersion target
var BuildVersion = Target{
	Filepath: "BUILD_VERSION",
	Template: `0.0.1`,
}

// GitIgnore target
var GitIgnore = Target{
	Filepath: ".gitignore",
	Template: `/artifacts`,
}

// GitLabCI target
var GitLabCI = Target{
	Filepath: ".gitlab-ci.yaml",
	Template: `
buildpush:
  only:
    - master
stage: deploy
image: golang:1.16.0
script:
  # Note: this is all one step because the artifacts were too large to copy over.
  - make

  # Download and install Google Cloud SDK
  - wget https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz
  - tar zxvf google-cloud-sdk.tar.gz && ./google-cloud-sdk/install.sh --usage-reporting=false --path-update=true
  - PATH="google-cloud-sdk/bin:${PATH}"
  - gcloud --quiet components update

  - echo $GCP_BUCKET_SA > ${HOME}/gcloud-service-key.json
  - gcloud auth activate-service-account --key-file ${HOME}/gcloud-service-key.json
  - gcloud config set project $GCP_PROJECT_ID

  - gsutil -m cp -R artifacts gs://{{ .RepositoryName }}
`,
}

// GitHubCI target
// TODO (pbarker): should we push everything to a single repository, or at least make that possible?
// TODO (pbarker): should report stats
var GitHubCI = Target{
	Filepath: ".github/workflows/release.yaml",
	Template: `name: Release

on:
  push:
  tags:
    - v*
  branches: [ master ]

jobs:

  build:
  name: Release
  runs-on: ubuntu-latest
  steps:

  - name: Set up Go 1.x
    uses: actions/setup-go@v2
    with:
      go-version: ^1.16
      id: go

  - name: Check out code into the Go module directory
    uses: actions/checkout@v2

  - name: Get dependencies
    run: |
    go get -v -t -d ./...
    curl -o tanzu https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-linux_amd64 && \
        mv tanzu /usr/local/bin/tanzu && \
      chmod +x /usr/local/bin/tanzu
    tanzu plugin repo add -b tanzu-cli-admin-plugins -n admin

  - name: init
    run: make init

  - name: Build
    run: make build

  - name: Test
    run: make test

  - name: Test
    run: make lint

  - id: upload-cli-artifacts
    uses: GoogleCloudPlatform/github-actions/upload-cloud-storage@master
    with:
      path: ./artifacts
      destination: {{ .RepositoryName }}
      credentials: {{"${{ secrets.GCP_BUCKET_SA }}"}}
`,
}

// Makefile target
var Makefile = Target{
	Filepath: "Makefile",
	Template: `BUILD_VERSION ?= $$(cat BUILD_VERSION)
BUILD_SHA ?= $$(git rev-parse --short HEAD)
BUILD_DATE ?= $$(date -u +"%Y-%m-%d")

LD_FLAGS = -X 'github.com/vmware-tanzu-private/core/pkg/v1/cli.BuildDate=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu-private/core/pkg/v1/cli.BuildSHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu-private/core/pkg/v1/cli.BuildVersion=$(BUILD_VERSION)'

TOOLS_DIR := tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint

build:
	tanzu builder cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/plugin

lint: golangci-lint
	$(GOLANGCI_LINT) run -v

init:
	GOPRIVATE=github.com/vmware-tanzu-private go mod download
	mkdir $(TOOLS_BIN_DIR)

test:
	go test ./...

golangci-lint:
	go build -tags=tools -o $(GOLANGCI_LINT) github.com/golangci/golangci-lint/cmd/golangci-lint
`,
}

// Codeowners target
// TODO (pbarker): replace with the CLI reviewers group
var Codeowners = Target{
	Filepath: "CODEOWNERS",
	Template: `* @pbarker @vuil`,
}

// Tools target.
var Tools = Target{
	Filepath: "tools/tools.go",
	Template: `// +build tools

package tools

import (
	_ "github.com/vmware-tanzu-private/core/cmd/cli/compiler"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
	`,
}

// MainReadMe target
var MainReadMe = Target{
	Filepath: "README.md",
	Template: plugintemplates.PluginReadme,
}

// GolangCIConfig target.
var GolangCIConfig = Target{
	Filepath: ".golangci.yaml",
	Template: plugintemplates.GolangCIConfig,
}
