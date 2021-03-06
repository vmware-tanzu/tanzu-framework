// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template

// GoMod target
var GoMod = Target{
	Filepath: "go.mod",
	Template: `module {{ .RepositoryName }}

go 1.14

require (
	github.com/aunum/log v0.0.0-20200821225356-38d2e2c8b489
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/vmware-tanzu-private/core v0.0.0-20201105155058-3739a04e35ae
	gopkg.in/yaml.v2 v2.3.0
	sigs.k8s.io/yaml v1.2.0
)
`,
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
image: golang:1.15.2
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
      go-version: ^1.13
      id: go

  - name: Check out code into the Go module directory
    uses: actions/checkout@v2

  - name: Get dependencies
    run: |
    go get -v -t -d ./...
    if [ -f Gopkg.toml ]; then
      curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
      dep ensure
    fi
    curl -o tanzu https://storage.googleapis.com/tanzu-cli/artifacts/core/latest/tanzu-core-linux_amd64 && \
        mv tanzu /usr/local/bin/tanzu && \
      chmod +x /usr/local/bin/tanzu
    tanzu plugin repo add -b tanzu-cli-admin-plugins -n admin

  - name: Make
    run: make

  - name: Build
    run: make build-cli

  - name: Test
    run: make test

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

build:
	tanzu builder cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/plugin
	`,
}

// Codeowners target
// TODO (pbarker): replace with the CLI reviewers group
var Codeowners = Target{
	Filepath: "CODEOWNERS",
	Template: `*       @pbarker @vuil`,
}

// Tools target.
var Tools = Target{
	Filepath: "tools/tools.go",
	Template: `package tools

import (
	_ "github.com/vmware-tanzu-private/core/cmd/cli/compiler"
)	
	`,
}

// MainReadMe target
var MainReadMe = Target{
	Filepath: "README.md",
	Template: `# {{ .RepositoryName }}
	`,
}
