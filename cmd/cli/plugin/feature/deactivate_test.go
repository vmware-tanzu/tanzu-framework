// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestExecuteFeatureDectivateCommand(t *testing.T) {
	test := struct {
		description string
		cmd         *cobra.Command
		featName    string
		errMsg      string
	}{
		description: "FeatureGateClient cannot connect due to bad configuration",
		cmd:         FeatureDeactivateCmd,
		featName:    "bazzies",
		errMsg:      "could not get FeatureGateClient",
	}

	storedKubeConfig, lastEnvVarOK := os.LookupEnv("KUBECONFIG")
	if err := os.Setenv("KUBECONFIG", "test/k8s_config.kube"); err != nil {
		t.Fatalf("failed to set test kubeconfig: %v", err)
	}

	test.cmd.SetArgs([]string{test.featName})

	err := test.cmd.Execute()

	if len(test.errMsg) > 0 {
		if err == nil {
			t.Fatal("expected error, but did not get one")
		}

		if !strings.Contains(err.Error(), test.errMsg) {
			t.Errorf("got: %s, want error to contain: %s", err, test.errMsg)
		}
	}

	// Restore previous kubeconfig if it was set.
	if err := restoreEnvVar("KUBECONFIG", storedKubeConfig, lastEnvVarOK); err != nil {
		t.Errorf("unable to restore previous KUBECONFIG envvar: %v", err)
	}
}
