// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/adrg/xdg"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

var (
	// DefaultPluginRoot is the default plugin root.
	DefaultPluginRoot = filepath.Join(xdg.DataHome, "tanzu-cli")
)

type loginOIDCOptions struct {
	scopes                     []string
	caBundlePaths              []string
	caBundleData               []string
	issuer                     string
	clientID                   string
	sessionCachePath           string
	requestAudience            string
	conciergeAuthenticatorType string
	conciergeAuthenticatorName string
	conciergeEndpoint          string
	conciergeCABundle          string
	credentialCachePath        string
	listenPort                 uint16
	skipBrowser                bool
	debugSessionCache          bool
	conciergeEnabled           bool
}

// We currently embed only the 0.12.1 CLI
var (
	//go:embed asset/pinniped-v0.12.1
	pinnipedv0121Binary []byte
)

var loginOptions = &loginOIDCOptions{}

var loginCommand = &cobra.Command{
	Use:          "login",
	Short:        "Login using an OpenID Connect provider",
	SilenceUsage: true,
	Example: `
    # pinniped-auth login using OpenID Connect provider
    tanzu pinniped-auth login  --issuer https://issuer.example.com --client-id tanzu-cli

    # pinniped-auth login using OpenID Connect provider with TCP port for local host listener (authorization code flow only)
    tanzu pinniped-auth login  --issuer https://issuer.example.com --client-id tanzu-cli --listen-port=48095`,
}

func init() {
	setLoginCommandFlags()
}

func loginOIDCCommand(
	getPinnipedCLICmdFunc func(
		args []string,
		pluginRoot, buildVersion, buildSHA string,
	) (*exec.Cmd, error),
) *cobra.Command {

	loginCommand.RunE = func(cmd *cobra.Command, args []string) error {
		oidcLoginArgs := []string{
			"login", "oidc",
			fmt.Sprintf("--issuer=%s", loginOptions.issuer),
			fmt.Sprintf("--client-id=%s", loginOptions.clientID),
			fmt.Sprintf("--listen-port=%d", loginOptions.listenPort),
			fmt.Sprintf("--skip-browser=%s", strconv.FormatBool(loginOptions.skipBrowser)),
			fmt.Sprintf("--session-cache=%s", loginOptions.sessionCachePath),
			fmt.Sprintf("--debug-session-cache=%s", strconv.FormatBool(loginOptions.debugSessionCache)),
			fmt.Sprintf("--scopes=%s", strings.Join(loginOptions.scopes, ",")),
			fmt.Sprintf("--ca-bundle=%s", strings.Join(loginOptions.caBundlePaths, ",")),
			fmt.Sprintf("--ca-bundle-data=%s", strings.Join(loginOptions.caBundleData, ",")),
			fmt.Sprintf("--request-audience=%s", loginOptions.requestAudience),
			fmt.Sprintf("--enable-concierge=%s", strconv.FormatBool(loginOptions.conciergeEnabled)),
			fmt.Sprintf("--concierge-authenticator-type=%s", loginOptions.conciergeAuthenticatorType),
			fmt.Sprintf("--concierge-authenticator-name=%s", loginOptions.conciergeAuthenticatorName),
			fmt.Sprintf("--concierge-endpoint=%s", loginOptions.conciergeEndpoint),
			fmt.Sprintf("--concierge-ca-bundle-data=%s", loginOptions.conciergeCABundle),
			fmt.Sprintf("--credential-cache=%s", loginOptions.credentialCachePath),
		}

		pinnipedCliCmd, err := getPinnipedCLICmdFunc(oidcLoginArgs, DefaultPluginRoot, buildinfo.Version, buildinfo.SHA)
		if err != nil {
			return fmt.Errorf("cannot construct pinniped cli command: %w", err)
		}
		if err := pinnipedCliCmd.Run(); err != nil {
			return fmt.Errorf("pinniped-auth login failed: %w", err)
		}
		return nil
	}
	return loginCommand
}

func getPinnipedCLICmd(args []string, pluginRoot, buildVersion, buildSHA string) (*exec.Cmd, error) {
	pinnipedBinary := pinnipedv0121Binary
	pinnipedVersion := "v0.12.1"

	buildSHA = strings.ReplaceAll(buildSHA, "-dirty", "")
	pinnipedCLIBinFile := fmt.Sprintf("tanzu-pinniped-%s-client-%s-%s", pinnipedVersion, buildVersion, buildSHA)
	if runtime.GOOS == "windows" {
		pinnipedCLIBinFile += ".exe"
	}

	pinnipedGoClientRoot := filepath.Join(pluginRoot, "tanzu-pinniped-go-client")
	if err := ensurePinnipedGoClientRoot(pinnipedGoClientRoot); err != nil {
		return nil, err
	}

	pinnipedCLIBinFilePath := filepath.Join(pinnipedGoClientRoot, pinnipedCLIBinFile)
	if err := ensurePinnipedCLIBinFile(pinnipedCLIBinFilePath, pinnipedBinary); err != nil {
		return nil, err
	}

	cmd := exec.Command(pinnipedCLIBinFilePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd, nil
}

func setLoginCommandFlags() {
	loginCommand.Flags().StringVar(&loginOptions.issuer, "issuer", "", "OpenID Connect issuer URL.")
	loginCommand.Flags().StringVar(&loginOptions.clientID, "client-id", "pinniped-cli", "OpenID Connect client ID.")
	loginCommand.Flags().Uint16Var(&loginOptions.listenPort, "listen-port", 0, "TCP port for localhost listener (authorization code flow only).")
	loginCommand.Flags().StringSliceVar(&loginOptions.scopes, "scopes", []string{"offline_access, openid, pinniped:request-audience"}, "OIDC scopes to request during login.")
	loginCommand.Flags().BoolVar(&loginOptions.skipBrowser, "skip-browser", false, "Skip opening the browser (just print the URL).")
	loginCommand.Flags().StringVar(&loginOptions.sessionCachePath, "session-cache", filepath.Join(mustGetConfigDir(), "sessions.yaml"), "Path to session cache file.")
	loginCommand.Flags().StringSliceVar(&loginOptions.caBundlePaths, "ca-bundle", nil, "Path to TLS certificate authority bundle (PEM format, optional, can be repeated).")
	loginCommand.Flags().StringSliceVar(&loginOptions.caBundleData, "ca-bundle-data", nil, "Base64 endcoded TLS certificate authority bundle (base64 encoded PEM format, optional, can be repeated)")
	loginCommand.Flags().BoolVar(&loginOptions.debugSessionCache, "debug-session-cache", false, "Print debug logs related to the session cache.")
	loginCommand.Flags().StringVar(&loginOptions.requestAudience, "request-audience", "", "Request a token with an alternate audience using RFC8693 token exchange")
	loginCommand.Flags().BoolVar(&loginOptions.conciergeEnabled, "enable-concierge", false, "Exchange the OIDC ID token with the Pinniped concierge during login")
	loginCommand.Flags().StringVar(&loginOptions.conciergeAuthenticatorType, "concierge-authenticator-type", "", "Concierge authenticator type (e.g., 'webhook', 'jwt')")
	loginCommand.Flags().StringVar(&loginOptions.conciergeAuthenticatorName, "concierge-authenticator-name", "", "Concierge authenticator name")
	loginCommand.Flags().StringVar(&loginOptions.conciergeEndpoint, "concierge-endpoint", "", "API base for the Pinniped concierge endpoint")
	loginCommand.Flags().StringVar(&loginOptions.conciergeCABundle, "concierge-ca-bundle-data", "", "CA bundle to use when connecting to the concierge")
	loginCommand.Flags().StringVar(&loginOptions.credentialCachePath, "credential-cache", filepath.Join(mustGetConfigDir(), "credentials.yaml"), "Path to cluster-specific credentials cache (\"\" disables the cache)")
	loginCommand.Flags().MarkHidden("debug-session-cache") //nolint
	loginCommand.MarkFlagRequired("issuer")                //nolint
}

// mustGetConfigDir returns the pinniped config directory
func mustGetConfigDir() string {
	const pinnipedConfigDir = "pinniped"

	tanzuLocalDir, err := config.LocalDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(tanzuLocalDir, pinnipedConfigDir)
}

// Ensure the pinniped go client directory exists.
func ensurePinnipedGoClientRoot(pinnipedGoClientRoot string) error {
	_, err := os.Stat(pinnipedGoClientRoot)
	if os.IsNotExist(err) {
		err := os.MkdirAll(pinnipedGoClientRoot, 0755)
		return errors.Wrap(err, "could not make pinniped go client directory")
	}
	return err
}

// Ensure the pinniped cli binary file exists.
func ensurePinnipedCLIBinFile(pinnipedCLIBinFilePath string, pinnipedBinary []byte) error {
	_, err := os.Stat(pinnipedCLIBinFilePath)
	if os.IsNotExist(err) {
		err = os.WriteFile(pinnipedCLIBinFilePath, pinnipedBinary, 0755)
		if err != nil {
			return errors.Wrap(err, "could not write pinniped binary to file")
		}
	}
	return nil
}
