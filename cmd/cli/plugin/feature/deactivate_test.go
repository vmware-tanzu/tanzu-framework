// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
	crclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/featuregates/client/pkg/featuregateclient"
	"github.com/vmware-tanzu/tanzu-framework/featuregates/client/pkg/featuregateclient/fake"
)

func TestDeactivateFeature(t *testing.T) {
	tests := []struct {
		description  string
		featureName  string
		wantErr      error
		wantActivate bool
	}{
		{
			description:  "deactivate a deprecated feature",
			featureName:  "biz",
			wantErr:      nil,
			wantActivate: false,
		},
		{
			description:  "deactivate a technical preview feature",
			featureName:  "tuna",
			wantErr:      nil,
			wantActivate: false,
		},
		{
			description:  "deactivate experimental feature",
			featureName:  "cloud-event-listener",
			wantErr:      nil,
			wantActivate: false,
		},
		{
			description:  "cannot deactivate a stable feature",
			featureName:  "super-toaster",
			wantErr:      featuregateclient.ErrTypeForbidden,
			wantActivate: true,
		},
		{
			description:  "cannot deactivate a feature that is not gated by feature gate",
			featureName:  "hard-to-get",
			wantErr:      featuregateclient.ErrTypeNotFound,
			wantActivate: false,
		},
		{
			description:  "cannot deactivate a feature that is gated by more than one feature gate",
			featureName:  "bazzies",
			wantErr:      featuregateclient.ErrTypeTooMany,
			wantActivate: true,
		},
	}

	objs, _, _ := fake.GetTestObjects()
	s := scheme.Scheme
	if err := corev1alpha2.AddToScheme(s); err != nil {
		t.Fatalf("unable to add config scheme: (%v)", err)
	}

	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	fgClient, err := featuregateclient.NewFeatureGateClient(featuregateclient.WithClient(cl))
	if err != nil {
		t.Fatalf("unable to get FeatureGate client: %v", err)
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			_, err := deactivateFeature(context.Background(), fgClient, tc.featureName)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got error: %v, want: %v", err, tc.wantErr)
			}

			gates, err := fgClient.GetFeatureGateList(context.Background())
			if err != nil {
				t.Errorf("get FeatureGate List: %v", err)
			}
			_, ref := featuregateclient.FeatureRefFromGateList(gates, tc.featureName)
			if ref.Activate != tc.wantActivate {
				t.Errorf("got activate: %t, want: %t", ref.Activate, tc.wantActivate)
			}
		})
	}
}

func TestExecuteFeatureDeactivateCommand(t *testing.T) {
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
