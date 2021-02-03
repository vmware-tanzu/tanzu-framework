// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/require"
)

func Test_loginoidcCmd(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		execResult      []byte
		execReturnError error
		wantError       bool
		wantStdout      string
		wantStderr      string
		wantIssuer      string
		wantClientID    string
	}{
		{
			name:      "missing required flags",
			args:      []string{},
			wantError: true,
			wantStderr: Doc(`
				Error: required flag(s) "issuer" not set
			`),
		},
		{
			name: "cli exec returns error",
			args: []string{
				"--client-id", "test-client-id",
				"--issuer", "test-issuer",
			},
			wantIssuer:      "--issuer=test-issuer",
			wantClientID:    "--client-id=test-client-id",
			execReturnError: errors.New("pinniped cli exec fake error"),
			wantError:       true,
			wantStderr: Doc(`
			Error: pinniped-auth login, output: : pinniped cli exec fake error
			`),
		},
		{
			name: "success- test options are populated correctly",
			args: []string{
				"--client-id", "test-client-id",
				"--issuer", "test-issuer",
			},
			wantIssuer:   "--issuer=test-issuer",
			wantClientID: "--client-id=test-client-id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotIssuer   string
				gotClientID string
			)
			cmd := loginoidcCmd(func(execargs []string) ([]byte, error) {
				gotIssuer = execargs[2]
				gotClientID = execargs[3]

				return tt.execResult, tt.execReturnError
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
			require.Equal(t, tt.wantIssuer, gotIssuer, "unexpected issuer")
			require.Equal(t, tt.wantClientID, gotClientID, "unexpected client ID")
		})
	}
}

func Doc(s string) string {
	const (
		tab       = "\t"
		twoSpaces = "  "
	)
	return strings.ReplaceAll(heredoc.Doc(s), tab, twoSpaces)
}
