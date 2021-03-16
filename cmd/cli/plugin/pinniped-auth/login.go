// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	tkgauth "github.com/vmware-tanzu-private/core/pkg/v1/auth/tkg"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
)

type loginOIDCOptions struct {
	scopes                     []string
	caBundlePaths              []string
	caBundleData               []string
	issuer                     string
	clientID                   string
	sessionCachePath           string
	requestAudience            string
	conciergeNamespace         string
	conciergeAuthenticatorType string
	conciergeAuthenticatorName string
	conciergeEndpoint          string
	conciergeCABundle          string
	listenPort                 uint16
	skipBrowser                bool
	debugSessionCache          bool
	conciergeEnabled           bool
}

var lo = &loginOIDCOptions{}

var loCmd = &cobra.Command{
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
	loCmd.Flags().StringVar(&lo.issuer, "issuer", "", "OpenID Connect issuer URL.")
	loCmd.Flags().StringVar(&lo.clientID, "client-id", "pinniped-cli", "OpenID Connect client ID.")
	loCmd.Flags().Uint16Var(&lo.listenPort, "listen-port", 0, "TCP port for localhost listener (authorization code flow only).")
	loCmd.Flags().StringSliceVar(&lo.scopes, "scopes", []string{"offline_access, openid, pinniped:request-audience"}, "OIDC scopes to request during login.")
	loCmd.Flags().BoolVar(&lo.skipBrowser, "skip-browser", false, "Skip opening the browser (just print the URL).")
	loCmd.Flags().StringVar(&lo.sessionCachePath, "session-cache", filepath.Join(mustGetConfigDir(), "sessions.yaml"), "Path to session cache file.")
	loCmd.Flags().StringSliceVar(&lo.caBundlePaths, "ca-bundle", nil, "Path to TLS certificate authority bundle (PEM format, optional, can be repeated).")
	loCmd.Flags().StringSliceVar(&lo.caBundleData, "ca-bundle-data", nil, "Base64 endcoded TLS certificate authority bundle (base64 encoded PEM format, optional, can be repeated)")
	loCmd.Flags().BoolVar(&lo.debugSessionCache, "debug-session-cache", false, "Print debug logs related to the session cache.")
	loCmd.Flags().StringVar(&lo.requestAudience, "request-audience", "", "Request a token with an alternate audience using RFC8693 token exchange")
	loCmd.Flags().BoolVar(&lo.conciergeEnabled, "enable-concierge", false, "Exchange the OIDC ID token with the Pinniped concierge during login")
	loCmd.Flags().StringVar(&lo.conciergeNamespace, "concierge-namespace", "pinniped-concierge", "Namespace in which the concierge was installed")
	loCmd.Flags().StringVar(&lo.conciergeAuthenticatorType, "concierge-authenticator-type", "", "Concierge authenticator type (e.g., 'webhook', 'jwt')")
	loCmd.Flags().StringVar(&lo.conciergeAuthenticatorName, "concierge-authenticator-name", "", "Concierge authenticator name")
	loCmd.Flags().StringVar(&lo.conciergeEndpoint, "concierge-endpoint", "", "API base for the Pinniped concierge endpoint")
	loCmd.Flags().StringVar(&lo.conciergeCABundle, "concierge-ca-bundle-data", "", "CA bundle to use when connecting to the concierge")
	loCmd.Flags().MarkHidden("debug-session-cache") //nolint
	loCmd.MarkFlagRequired("issuer")                //nolint
}

func loginoidcCmd(pinnipedloginCliExec func(args []string) error) *cobra.Command {
	loCmd.RunE = func(cmd *cobra.Command, args []string) error {
		oidcLoginArgs := []string{
			"login", "oidc",
			fmt.Sprintf("--issuer=%s", lo.issuer),
			fmt.Sprintf("--client-id=%s", lo.clientID),
			fmt.Sprintf("--listen-port=%d", lo.listenPort),
			fmt.Sprintf("--skip-browser=%s", strconv.FormatBool(lo.skipBrowser)),
			fmt.Sprintf("--session-cache=%s", lo.sessionCachePath),
			fmt.Sprintf("--debug-session-cache=%s", strconv.FormatBool(lo.debugSessionCache)),
			fmt.Sprintf("--scopes=%s", strings.Join(lo.scopes, ",")),
			fmt.Sprintf("--ca-bundle=%s", strings.Join(lo.caBundlePaths, ",")),
			fmt.Sprintf("--ca-bundle-data=%s", strings.Join(lo.caBundleData, ",")),
			fmt.Sprintf("--request-audience=%s", lo.requestAudience),
			fmt.Sprintf("--enable-concierge=%s", strconv.FormatBool(lo.conciergeEnabled)),
			fmt.Sprintf("--concierge-namespace=%s", lo.conciergeNamespace),
			fmt.Sprintf("--concierge-authenticator-type=%s", lo.conciergeAuthenticatorType),
			fmt.Sprintf("--concierge-authenticator-name=%s", lo.conciergeAuthenticatorName),
			fmt.Sprintf("--concierge-endpoint=%s", lo.conciergeEndpoint),
			fmt.Sprintf("--concierge-ca-bundle-data=%s", lo.conciergeCABundle),
		}

		err := pinnipedloginCliExec(oidcLoginArgs)
		if err != nil {
			return errors.Wrapf(err, "pinniped-auth login failed")
		}
		return nil
	}
	return loCmd
}

// pinnipedLoginExec executes embedded pinniped cli binary
func pinnipedLoginExec(oidcLoginArgs []string) error {
	buildSHA := strings.ReplaceAll(cli.BuildSHA, "-dirty", "")
	pinnipedCLIBinFile := fmt.Sprintf("tanzu-pinniped-client-%s-%s", cli.BuildVersion, buildSHA)
	if runtime.GOOS == "windows" {
		pinnipedCLIBinFile += ".exe"
	}

	pinnipedGoClientRoot := filepath.Join(cli.DefaultPluginRoot, "tanzu-pinniped-go-client")
	if err := ensurePinnipedGoClientRoot(pinnipedGoClientRoot); err != nil {
		return err
	}

	pinnipedCLIBinFilePath := filepath.Join(pinnipedGoClientRoot, pinnipedCLIBinFile)
	if err := ensurePinnipedCLIBinFile(pinnipedCLIBinFilePath); err != nil {
		return err
	}

	cmd := exec.Command(pinnipedCLIBinFilePath, oidcLoginArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

// mustGetConfigDir returns the pinniped config directory
func mustGetConfigDir() string {
	const pinnipedConfigDir = "pinniped"

	tanzuLocalDir, err := client.LocalDir()
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
func ensurePinnipedCLIBinFile(pinnipedCLIBinFilePath string) error {
	pinnipedBinary, err := tkgauth.Asset("pinniped/pinniped")
	if err != nil {
		return err
	}
	_, err = os.Stat(pinnipedCLIBinFilePath)
	if os.IsNotExist(err) {
		err = os.WriteFile(pinnipedCLIBinFilePath, pinnipedBinary, 0755)
		if err != nil {
			return errors.Wrap(err, "could not write pinniped binary to file")
		}
	}
	return nil
}
