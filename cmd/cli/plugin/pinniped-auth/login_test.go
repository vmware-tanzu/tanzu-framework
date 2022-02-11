// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
)

const (
	tanzuCLIBuildVersion = "1.2.3"
	tanzuCLIBuildSHA     = "abc123"
)

//nolint:funlen
func TestLoginOIDCCommand(t *testing.T) {
	sessionsCacheFilePath := filepath.Join(mustGetConfigDir(), "sessions.yaml")
	credentialCacheFilePath := filepath.Join(mustGetConfigDir(), "credentials.yaml")
	tests := []struct {
		name                 string
		args                 []string
		getPinnipedCLICmdErr error
		execReturnExitCode   int
		wantError            bool
		wantStdout           string
		wantStderr           string
		wantArgs             []string
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
			name: "getting pinniped cli cmd fails",
			args: []string{
				"--issuer", "test-issuer",
			},
			getPinnipedCLICmdErr: errors.New("some construction error"),
			wantError:            true,
			wantStderr: Doc(`
				Error: cannot construct pinniped cli command: some construction error
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
				"--concierge-authenticator-type=",
				"--concierge-authenticator-name=",
				"--concierge-endpoint=",
				"--concierge-ca-bundle-data=",
				"--concierge-namespace=pinniped-concierge",
			},
		},
		{
			name: "cli exec returns error from login",
			args: []string{
				"--issuer", "test-issuer",
			},
			execReturnExitCode: 88,
			wantError:          true,
			wantStderr: Doc(`
			Error: pinniped-auth login failed: exit status 88
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
				"--concierge-authenticator-type=",
				"--concierge-authenticator-name=",
				"--concierge-endpoint=",
				"--concierge-ca-bundle-data=",
				"--concierge-namespace=pinniped-concierge",
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
				"--concierge-authenticator-type=",
				"--concierge-authenticator-name=",
				"--concierge-endpoint=",
				"--concierge-ca-bundle-data=",
				"--concierge-namespace=pinniped-concierge",
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
				"--concierge-authenticator-type=",
				"--concierge-authenticator-name=",
				"--concierge-endpoint=",
				"--concierge-ca-bundle-data=",
				"--concierge-namespace=pinniped-concierge",
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
				"--session-cache", "/path/to/session",
				"--credential-cache", "/path/to/cred",
				"--ca-bundle", "/some/path",
				"--ca-bundle-data", "somebase64encodeddata",
				"--request-audience", "alternateaudience",
				"--enable-concierge", "true",
				"--concierge-namespace", "test-namespace",
				"--concierge-authenticator-type", "webhook",
				"--concierge-authenticator-name", "concierge-authenticator",
				"--concierge-endpoint", "test-endpoint",
				"--concierge-ca-bundle-data", "test-bundle",
				"--concierge-is-cluster-scoped", "true",
			},
			wantArgs: []string{
				"login",
				"oidc",
				"--issuer=different-issuer",
				"--client-id=test-client",
				"--listen-port=3737",
				"--skip-browser=true",
				"--session-cache=/path/to/session",
				"--debug-session-cache=true",
				"--scopes=openid",
				"--ca-bundle=/some/path",
				"--ca-bundle-data=somebase64encodeddata",
				"--request-audience=alternateaudience",
				"--enable-concierge=true",
				"--concierge-authenticator-type=webhook",
				"--concierge-authenticator-name=concierge-authenticator",
				"--concierge-endpoint=test-endpoint",
				"--concierge-ca-bundle-data=test-bundle",
				"--credential-cache=/path/to/cred",
			},
		},
		{
			name: "test concierge-namespace not included in login args when concierge-is-cluster-scoped is true",
			args: []string{
				"--issuer", "test-issuer",
				"--concierge-is-cluster-scoped", "true",
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
				"--concierge-authenticator-type=",
				"--concierge-authenticator-name=",
				"--concierge-endpoint=",
				"--concierge-ca-bundle-data=",
				fmt.Sprintf("--credential-cache=%s", credentialCacheFilePath),
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
			cmd := loginOIDCCommand(func(args []string, loginOptions *loginOIDCOptions, pluginRoot, buildVersion, buildSHA string) (*exec.Cmd, error) {
				gotArgs = args

				require.Equal(t, cli.DefaultPluginRoot, pluginRoot)
				require.Equal(t, buildinfo.Version, buildVersion)
				require.Equal(t, buildinfo.SHA, buildSHA)

				cmd := exec.Command("./testdata/fake-pinniped-cli.sh")
				cmd.Env = []string{fmt.Sprintf("FAKE_PINNIPED_CLI_EXIT_CODE=%d", test.execReturnExitCode)}
				return cmd, test.getPinnipedCLICmdErr
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

func TestGetPinnipedCLICmd(t *testing.T) {
	tests := []struct {
		name                string
		isBuildSHADirty     bool
		loginOptions        *loginOIDCOptions
		wantError           string
		wantPinnipedVersion string
		wantBinary          []byte
	}{
		{
			name:                "0.4.4 cli",
			loginOptions:        &loginOIDCOptions{conciergeIsClusterScoped: false},
			wantPinnipedVersion: "v0.4.4",
			wantBinary:          pinnipedv044Binary,
		},
		{
			name:                "0.4.4 cli with dirty build sha",
			isBuildSHADirty:     true,
			loginOptions:        &loginOIDCOptions{conciergeIsClusterScoped: false},
			wantPinnipedVersion: "v0.4.4",
			wantBinary:          pinnipedv044Binary,
		},
		{
			name:                "0.12.1 cli",
			loginOptions:        &loginOIDCOptions{conciergeIsClusterScoped: true},
			wantPinnipedVersion: "v0.12.1",
			wantBinary:          pinnipedv0121Binary,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			pluginRoot, err := os.MkdirTemp("", "pinniped-auth-get-pinniped-cli-cmd-test-*")
			require.NoError(t, err)
			defer require.NoError(t, os.RemoveAll(pluginRoot))

			buildSHA := tanzuCLIBuildSHA
			wantBuildSHA := tanzuCLIBuildSHA
			if test.isBuildSHADirty {
				buildSHA = buildSHA + "-dirty"
			}

			wantPathBasename := fmt.Sprintf("tanzu-pinniped-%s-client-%s-%s", test.wantPinnipedVersion, tanzuCLIBuildVersion, wantBuildSHA)
			wantPath := filepath.Join(pluginRoot, "tanzu-pinniped-go-client", wantPathBasename)
			wantArgs := []string{"some", "args", "here"}

			cmd, err := getPinnipedCLICmd(wantArgs, test.loginOptions, pluginRoot, tanzuCLIBuildVersion, buildSHA)
			if test.wantError != "" {
				require.EqualError(t, err, test.wantError)
				return
			}
			require.NoError(t, err)

			require.Equal(t, wantPath, cmd.Path)
			require.FileExists(t, cmd.Path)
			requireFileContents(t, cmd.Path, test.wantBinary)
			require.Equal(t, append([]string{cmd.Path}, wantArgs...), cmd.Args)
			require.Equal(t, os.Stdout, cmd.Stdout)
			require.Equal(t, os.Stderr, cmd.Stderr)
			require.Equal(t, os.Stdin, cmd.Stdin)
			require.Empty(t, cmd.Env)
			require.Empty(t, cmd.Dir)
		})
	}
}

// TestGetPinnipedCLICmdMultipleUsage validates that the pinniped CLI binary logic works when used
// multiple times with the same plugin root.
func TestGetPinnipedCLICmdMultipleUsage(t *testing.T) {
	pluginRoot, err := os.MkdirTemp("", "pinniped-auth-get-pinniped-cli-cmd-test-*")
	require.NoError(t, err)
	defer require.NoError(t, os.RemoveAll(pluginRoot))

	steps := []struct {
		loginOptions        *loginOIDCOptions
		wantPinnipedVersion string
		wantBinary          []byte
	}{
		{
			loginOptions:        &loginOIDCOptions{conciergeIsClusterScoped: false},
			wantPinnipedVersion: "v0.4.4",
			wantBinary:          pinnipedv044Binary,
		},
		{
			loginOptions:        &loginOIDCOptions{conciergeIsClusterScoped: true},
			wantPinnipedVersion: "v0.12.1",
			wantBinary:          pinnipedv0121Binary,
		},
	}
	for _, step := range steps {
		wantPathBasename := fmt.Sprintf("tanzu-pinniped-%s-client-%s-%s", step.wantPinnipedVersion, tanzuCLIBuildVersion, tanzuCLIBuildSHA)
		wantPath := filepath.Join(pluginRoot, "tanzu-pinniped-go-client", wantPathBasename)

		cmd, err := getPinnipedCLICmd([]string{}, step.loginOptions, pluginRoot, tanzuCLIBuildVersion, tanzuCLIBuildSHA)
		require.NoError(t, err)
		require.Equal(t, wantPath, cmd.Path)
		require.FileExists(t, cmd.Path)
		requireFileContents(t, cmd.Path, step.wantBinary)
	}
}

func Doc(s string) string {
	const (
		tab       = "\t"
		twoSpaces = "  "
	)
	return strings.ReplaceAll(heredoc.Doc(s), tab, twoSpaces)
}

func requireFileContents(t *testing.T, path string, wantContents []byte) {
	t.Helper()

	gotContents, err := os.ReadFile(path)
	require.NoError(t, err)

	// These binary files can be really large (~50M), so comparing them via digest will speed up the
	// tests considerably.
	hash := sha256.New()

	hash.Reset()
	hash.Write(wantContents)
	wantDigest := hash.Sum(nil)

	hash.Reset()
	hash.Write(gotContents)
	gotDigest := hash.Sum(nil)

	require.Equalf(t, hex.EncodeToString(wantDigest[:]), hex.EncodeToString(gotDigest[:]), "path %q does not have expected contents", path)
}
