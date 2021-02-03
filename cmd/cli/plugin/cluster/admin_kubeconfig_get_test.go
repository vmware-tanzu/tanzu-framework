// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"
)

func Test_getClusterAdminKubeconfigCmd(t *testing.T) {
	tests := []struct {
		name                      string
		args                      []string
		getCredentailsErr         error
		gotGetCredentialsOptions  tkgctl.GetWorkloadClusterCredentialsOptions
		wantGetCredentialsOptions tkgctl.GetWorkloadClusterCredentialsOptions
		wantError                 bool
		wantStdout                string
		wantStderr                string
	}{
		{
			name: "With out arguments should get admin-kubeconfig of management cluster with name extracted from server kubeconfig",
			args: []string{},
			wantGetCredentialsOptions: tkgctl.GetWorkloadClusterCredentialsOptions{
				ClusterName: "horse-cluster",
				Namespace:   TKGSystemNamespace,
			},
			wantError: false,
		},
		{
			name: "With 'workload-clustername' as option without 'namespace'",
			args: []string{"--workload-clustername", "fake-workload-clustername"},
			wantGetCredentialsOptions: tkgctl.GetWorkloadClusterCredentialsOptions{
				ClusterName: "fake-workload-clustername",
				Namespace:   DefaultNamespace,
			},
			wantError: false,
		},
		{
			name: "With 'workload-clustername' and 'export-file' as options ",
			args: []string{"--workload-clustername", "fake-workload-clustername", "--export-file", "./fake-export-file"},
			wantGetCredentialsOptions: tkgctl.GetWorkloadClusterCredentialsOptions{
				ClusterName: "fake-workload-clustername",
				Namespace:   DefaultNamespace,
				ExportFile:  "./fake-export-file",
			},
			wantError: false,
		},
		{
			name: "With 'workload-clustername' and namespace as options",
			args: []string{"--workload-clustername", "fake-workload-clustername", "--namespace", "fake-namespace"},
			wantGetCredentialsOptions: tkgctl.GetWorkloadClusterCredentialsOptions{
				ClusterName: "fake-workload-clustername",
				Namespace:   "fake-namespace",
			},
			wantError: false,
		},
		{
			name: "If getCredentials()  return error",
			args: []string{"--workload-clustername", "fake-workload-clustername"},
			wantGetCredentialsOptions: tkgctl.GetWorkloadClusterCredentialsOptions{
				ClusterName: "fake-workload-clustername",
				Namespace:   DefaultNamespace,
			},
			getCredentailsErr: errors.New("fake-error from getCredentials()"),
			wantError:         true,
			wantStderr:        "Error: fake-error from getCredentials()\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := getClusterAdminKubeconfigCmd(getClusterAdminKubeconfigCmdDeps{
				getCurrentServer: func() (*v1alpha1.Server, error) {
					return &v1alpha1.Server{
						ManagementClusterOpts: &v1alpha1.ManagementClusterServer{
							Path:    "./testdata/kubeconfig1.yaml",
							Context: "federal-context"},
					}, nil
				},
				getClusterCredentials: func(tkgctlClient tkgctl.TKGClient) getClusterCredentialFunc {
					return func(getCredentialOptions tkgctl.GetWorkloadClusterCredentialsOptions) error {
						tt.gotGetCredentialsOptions = getCredentialOptions
						return tt.getCredentailsErr
					}
				},
			})
			require.NotNil(t, cmd)
			var stdout, stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantStdout, stdout.String(), "unexpected stdout")
			require.Equal(t, tt.wantStderr, stderr.String(), "unexpected stderr")
			require.Equal(t, tt.wantGetCredentialsOptions, tt.gotGetCredentialsOptions, "unexpected GetCredentialOptions")
		})
	}
}
