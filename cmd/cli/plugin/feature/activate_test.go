// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
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

func TestActivateFeature(t *testing.T) {
	type nullBool struct {
		wasSet bool
		setTo  bool
	}

	tests := []struct {
		description  string
		featureName  string
		userAllows   nullBool
		wantErr      error
		wantActivate bool
	}{
		{
			description:  "activate a technical preview feature",
			featureName:  "tuna",
			wantErr:      nil,
			userAllows:   nullBool{wasSet: false},
			wantActivate: true,
		},
		{
			description:  "activate a deprecated feature",
			featureName:  "biz",
			wantErr:      nil,
			userAllows:   nullBool{wasSet: false},
			wantActivate: true,
		},
		{
			description:  "cannot activate a feature that was not found in cluster",
			featureName:  "hard-to-get",
			userAllows:   nullBool{wasSet: false},
			wantErr:      featuregateclient.ErrTypeNotFound,
			wantActivate: false,
		},
		{
			description:  "activate a stable feature that is already activated",
			featureName:  "super-toaster",
			userAllows:   nullBool{wasSet: false},
			wantErr:      nil,
			wantActivate: true,
		},
		{
			description:  "cannot activate a feature that is not gated by feature gate",
			featureName:  "specialized-toaster",
			userAllows:   nullBool{wasSet: false},
			wantErr:      featuregateclient.ErrTypeNotFound,
			wantActivate: false,
		},
		{
			description:  "cannot activate a feature that is gated by more than one feature gate",
			featureName:  "baz",
			userAllows:   nullBool{wasSet: false},
			wantErr:      featuregateclient.ErrTypeTooMany,
			wantActivate: false,
		},
		{
			description:  "activate experimental feature with user permission",
			featureName:  "cloud-event-speaker",
			userAllows:   nullBool{wasSet: true, setTo: true},
			wantErr:      nil,
			wantActivate: true,
		},
		{
			description:  "don't activate experimental feature when user disallows it",
			featureName:  "cloud-event-relayer",
			userAllows:   nullBool{wasSet: true, setTo: false},
			wantErr:      featuregateclient.ErrTypeForbidden,
			wantActivate: false,
		},
		{
			description:  "don't activate experimental feature when user does not give permission",
			featureName:  "cloud-event-relayer",
			userAllows:   nullBool{wasSet: false},
			wantErr:      featuregateclient.ErrTypeForbidden,
			wantActivate: false,
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
			var userAllows *bool
			if tc.userAllows.wasSet {
				userAllows = &tc.userAllows.setTo
			}
			_, err := activateFeature(context.Background(), fgClient, tc.featureName, userAllows)

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

func TestUserAllowsByInteractiveCLI(t *testing.T) {
	tests := []struct {
		description string
		userInput   string
		want        bool
	}{
		{
			description: "return true for user input yes",
			userInput:   "yes",
			want:        true,
		},
		{
			description: "return true for user input Y",
			userInput:   "Y",
			want:        true,
		},
		{
			description: "return true for user input no",
			userInput:   "no",
			want:        false,
		},
		{
			description: "return true for user input N",
			userInput:   "N",
			want:        false,
		},
		{
			description: "return false for no user input",
			userInput:   "",
			want:        false,
		},
		{
			description: "return false for bad user input on first try and no user input on second",
			userInput:   "bad-input",
			want:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			in, err := os.CreateTemp("", "tmp_test_feature_activate")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := in.Write([]byte(tc.userInput)); err != nil {
				t.Fatal(err)
			}
			if _, err := in.Seek(0, 0); err != nil {
				t.Fatal(err)
			}

			got, err := userAllowsByInteractiveCLI(in)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got: %t, want: %t", got, tc.want)
			}

			// Clean up temporary file.
			if err := in.Close(); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(in.Name()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestExecuteFeatureActivateCommand(t *testing.T) {
	test := struct {
		description string
		cmd         *cobra.Command
		featName    string
		errMsg      string
	}{
		description: "FeatureGateClient cannot connect due to bad configuration",
		cmd:         FeatureActivateCmd,
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

func restoreEnvVar(env string, previousVal string, wasPreviouslySet bool) error {
	if !wasPreviouslySet {
		if err := os.Unsetenv(env); err != nil {
			return fmt.Errorf("failed to unset environment variable %q: %w", env, err)
		}
		return nil
	}

	if err := os.Setenv(env, previousVal); err != nil {
		return fmt.Errorf("failed to set environment variable %q: %w", env, err)
	}

	return nil
}
