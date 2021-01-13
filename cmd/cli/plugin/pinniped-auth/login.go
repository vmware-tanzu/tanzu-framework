// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/amenzhinsky/go-memexec"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	tkgauth "github.com/vmware-tanzu-private/core/pkg/v1/auth/tkg"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
)

type loginOIDCOptions struct {
	issuer            string
	clientID          string
	listenPort        uint16
	scopes            []string
	skipBrowser       bool
	sessionCachePath  string
	caBundlePaths     []string
	debugSessionCache bool
}

var lo = &loginOIDCOptions{}

var loginoidcCmd = &cobra.Command{
	Use:   "login",
	Short: "Login using an OpenID Connect provider",
	Example: `
	# pinniped-auth login using OpenID Connect provider
	tanzu pinniped-auth login  --issuer https://issuer.example.com --client-id tanzu-cli

	# pinniped-auth login using OpenID Connect provider with TCP port for local host listener (authorization code flow only)
	tanzu pinniped-auth login  --issuer https://issuer.example.com --client-id tanzu-cli --listen-port=48095`,
	RunE: get,
}

func init() {
	loginoidcCmd.Flags().StringVar(&lo.issuer, "issuer", "", "OpenID Connect issuer URL.")
	loginoidcCmd.Flags().StringVar(&lo.clientID, "client-id", "", "OpenID Connect client ID.")
	loginoidcCmd.Flags().Uint16Var(&lo.listenPort, "listen-port", 0, "TCP port for localhost listener (authorization code flow only).")
	loginoidcCmd.Flags().StringSliceVar(&lo.scopes, "scopes", []string{"offline_access", "openid", "email", "profile"}, "OIDC scopes to request during login.")
	loginoidcCmd.Flags().BoolVar(&lo.skipBrowser, "skip-browser", false, "Skip opening the browser (just print the URL).")
	loginoidcCmd.Flags().StringVar(&lo.sessionCachePath, "session-cache", filepath.Join(mustGetConfigDir(), "sessions.yaml"), "Path to session cache file.")
	loginoidcCmd.Flags().StringSliceVar(&lo.caBundlePaths, "ca-bundle", nil, "Path to TLS certificate authority bundle (PEM format, optional, can be repeated).")
	loginoidcCmd.Flags().BoolVar(&lo.debugSessionCache, "debug-session-cache", false, "Print debug logs related to the session cache.")
	loginoidcCmd.Flags().MarkHidden("debug-session-cache")
	loginoidcCmd.MarkFlagRequired("issuer")
	loginoidcCmd.MarkFlagRequired("client-id")
}

func get(cmd *cobra.Command, args []string) error {

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
	}

	pinnipedBinary, err := tkgauth.Asset("pinniped/pinniped")
	if err != nil {
		return err
	}
	// TODO: Improve latency by avoiding the binary bits copy by storing locally
	exe, err := memexec.New(pinnipedBinary)
	if err != nil {
		return err
	}
	defer exe.Close()

	pinnipedCmd := exe.Command(oidcLoginArgs...)
	out, err := pinnipedCmd.Output()
	if err != nil {
		return errors.Wrapf(err, "pinniped-auth login, output: %s", string(out))
	}
	fmt.Print(string(out))
	return nil
}

func mustGetConfigDir() string {
	const pinnipedConfigDir = "pinniped"

	tanzuLocalDir, err := client.LocalDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(tanzuLocalDir, pinnipedConfigDir)
}
