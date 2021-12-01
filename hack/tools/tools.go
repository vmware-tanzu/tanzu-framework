//go:build tools
// +build tools

// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tools imports things required by build scripts, to force `go mod` to see them as dependencies

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
	_ "github.com/onsi/ginkgo/ginkgo"
	_ "github.com/shuLhan/go-bindata"
	_ "golang.org/x/tools/cmd/goimports"
)
