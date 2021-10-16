// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/require"
)

//nolint:funlen
func TestLoginOIDCCommand(t *testing.T) {
	sessionsCacheFilePath := filepath.Join(mustGetConfigDir(), "sessions.yaml")
	tests := []struct {
		name            string
		args            []string
		execReturnError error
		wantError       bool
		wantStdout      string
		wantStderr      string
		wantArgs        []string
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
			name: "cli exec returns error from login",
			args: []string{
				"--issuer", "test-issuer",
			},
			execReturnError: errors.New("pinniped cli exec fake error"),
			wantError:       true,
			wantStderr: Doc(`
			Error: pinniped-auth login failed: pinniped cli exec fake error
			`),
			wantArgs: []string{
				"login",
				"oidc",
				"--issuer=test-issuer",
				"--client-id=pinniped-cli",
				"--listen-port=0",
				"--skip-browser=false",
				fmt.Sprintf("--session-cache=%s", sessionsCacheFilePath),
				"--debug-session-cache=false",
				"--scopes=offline_access, openid, pinniped:request-audience",
				"--ca-bundle=",
				"--ca-bundle-data=",
				"--request-audience=",
				"--enable-concierge=false",
				"--concierge-namespace=pinniped-concierge",
				"--concierge-authenticator-type=",
				"--concierge-authenticator-name=",
				"--concierge-endpoint=",
				"--concierge-ca-bundle-data=",
			},
		},
		{
			name: "test options allow multiple values",
			args: []string{
				"--issuer", "test-issuer",
				"--scopes", "offline_access, openid",
				"--ca-bundle", "/some/path, /another/path",
				"--ca-bundle-data", "somebase64encodeddata, morebase64encodeddata",
			},
			wantArgs: []string{
				"login",
				"oidc",
				"--issuer=test-issuer",
				"--client-id=pinniped-cli",
				"--listen-port=0",
				"--skip-browser=false",
				fmt.Sprintf("--session-cache=%s", sessionsCacheFilePath),
				"--debug-session-cache=false",
				"--scopes=offline_access, openid",
				"--ca-bundle=/some/path, /another/path",
				"--ca-bundle-data=somebase64encodeddata, morebase64encodeddata",
				"--request-audience=",
				"--enable-concierge=false",
				"--concierge-namespace=pinniped-concierge",
				"--concierge-authenticator-type=",
				"--concierge-authenticator-name=",
				"--concierge-endpoint=",
				"--concierge-ca-bundle-data=",
			},
		},
		{
			name: "test options are populated correctly with defaults",
			args: []string{
				"--issuer", "test-issuer",
			},
			wantArgs: []string{
				"login",
				"oidc",
				"--issuer=test-issuer",
				"--client-id=pinniped-cli",
				"--listen-port=0",
				"--skip-browser=false",
				fmt.Sprintf("--session-cache=%s", sessionsCacheFilePath),
				"--debug-session-cache=false",
				"--scopes=offline_access, openid, pinniped:request-audience",
				"--ca-bundle=",
				"--ca-bundle-data=",
				"--request-audience=",
				"--enable-concierge=false",
				"--concierge-namespace=pinniped-concierge",
				"--concierge-authenticator-type=",
				"--concierge-authenticator-name=",
				"--concierge-endpoint=",
				"--concierge-ca-bundle-data=",
			},
		},
		{
			name: "test options are populated correctly with given user values",
			args: []string{
				"--issuer", "different-issuer",
				"--client-id", "test-client",
				"--listen-port", "3737",
				"--skip-browser", "true",
				"--debug-session-cache", "true",
				"--scopes", "openid",
				"--session-cache", "/some/path",
				"--ca-bundle", "/some/path",
				"--ca-bundle-data", "somebase64encodeddata",
				"--request-audience", "alternateaudience",
				"--enable-concierge", "true",
				"--concierge-namespace", "test-namespace",
				"--concierge-authenticator-type", "webhook",
				"--concierge-authenticator-name", "concierge-authenticator",
				"--concierge-endpoint", "test-endpoint",
				"--concierge-ca-bundle-data", "test-bundle",
				"--concierge-api-group-suffix", "tuna.io",
				"--concierge-is-cluster-scoped", "true",
			},
			wantArgs: []string{
				"login",
				"oidc",
				"--issuer=different-issuer",
				"--client-id=test-client",
				"--listen-port=3737",
				"--skip-browser=true",
				"--session-cache=/some/path",
				"--debug-session-cache=true",
				"--scopes=openid",
				"--ca-bundle=/some/path",
				"--ca-bundle-data=somebase64encodeddata",
				"--request-audience=alternateaudience",
				"--enable-concierge=true",
				"--concierge-namespace=test-namespace",
				"--concierge-authenticator-type=webhook",
				"--concierge-authenticator-name=concierge-authenticator",
				"--concierge-endpoint=test-endpoint",
				"--concierge-ca-bundle-data=test-bundle",
			},
		},
	}

	for _, test := range tests {
		test := test
		// Resetting and setting the flags again is necessary in these tests to ensure new
		// flags get picked up
		loginCommand.ResetFlags()
		setLoginCommandFlags()
		t.Run(test.name, func(t *testing.T) {
			var (
				gotArgs []string
			)
			cmd := loginOIDCCommand(func(args []string) error {
				gotArgs = args

				return test.execReturnError
			})
			require.NotNil(t, cmd)

			var stdout, stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(test.args)
			err := cmd.Execute()
			if test.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.wantStdout, stdout.String(), "unexpected stdout")
			require.Equal(t, test.wantStderr, stderr.String(), "unexpected stderr")
			require.Equal(t, test.wantArgs, gotArgs, "Given CLI arguments did not match actual CLI arguments")
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
