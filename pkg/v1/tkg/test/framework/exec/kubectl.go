// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"bytes"
	"context"

	. "github.com/onsi/ginkgo" // nolint:stylecheck,golint
)

// KubectlApplyWithArgs applies config with args
func KubectlApplyWithArgs(ctx context.Context, kubeconfigPath string, resources []byte, args ...string) error {
	aargs := append([]string{"apply", "--kubeconfig", kubeconfigPath, "-f", "-"}, args...)
	rbytes := bytes.NewReader(resources)
	applyCmd := NewCommand(
		WithCommand("kubectl"),
		WithArgs(aargs...),
		WithStdin(rbytes),
		WithStdout(GinkgoWriter),
	)
	err := applyCmd.RunAndRedirectOutput(ctx)
	return err
}
