// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aunum/log"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/auth/csp"
	tkgauth "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/auth/tkg"
	wcpauth "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/auth/wcp"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	cliconfig "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/pluginmanager"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

var (
	stderrOnly, forceCSP, staging, onlyCurrent                                  bool
	ctxName, ctxType, endpoint, apiToken, kubeConfig, kubeContext, getOutputFmt string
)

const (
	knownGlobalHost = "cloud.vmware.com"
)

var contextCmd = &cobra.Command{
	Use:     "context",
	Short:   "Configure and manage contexts for the Tanzu CLI",
	Aliases: []string{"ctx", "contexts"},
	Annotations: map[string]string{
		"group": string(cliapi.SystemCmdGroup),
	},
}

func init() {
	contextCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	contextCmd.AddCommand(
		createCtxCmd,
		listCtxCmd,
		getCtxCmd,
		deleteCtxCmd,
		useCtxCmd,
	)

	initCreateCtxCmd()

	listCtxCmd.Flags().StringVarP(&ctxType, "type", "t", "", "context type (k8s|tmc)")
	listCtxCmd.Flags().BoolVar(&onlyCurrent, "current", false, "list only current active contexts")
	listCtxCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "output format: table|yaml|json")

	getCtxCmd.Flags().StringVarP(&getOutputFmt, "output", "o", "yaml", "output format: yaml|json")

	deleteCtxCmd.Flags().BoolVarP(&unattended, "yes", "y", false, "delete the context entry without confirmation")
}

var createCtxCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Tanzu CLI context",
	RunE:  createCtx,
	Example: `
	# Create a TKG management cluster context using endpoint
	tanzu context create --endpoint "https://k8s.example.com" --name mgmt-cluster

	# Create a TKG management cluster context using kubeconfig path and context
	tanzu context create --kubeconfig path/to/kubeconfig --kubecontext kubecontext --name mgmt-cluster

	# Create a TKG management cluster context using default kubeconfig path and a kubeconfig context
	tanzu context create --kubecontext kubecontext --name mgmt-cluster

	[*] : User has two options to create a kubernetes cluster context. User can choose the control
	plane option by providing 'endpoint', or use the kubeconfig for the cluster by providing
	'kubeconfig' and 'context'. If only '--context' is set and '--kubeconfig' is not set
	$KUBECONFIG env variable would be used and, if $KUBECONFIG env is also not set default
	kubeconfig($HOME/.kube/config) would be used.
	`,
}

func initCreateCtxCmd() {
	createCtxCmd.Flags().StringVar(&ctxName, "name", "", "name of the context")
	createCtxCmd.Flags().StringVar(&endpoint, "endpoint", "", "endpoint to create a context for")
	createCtxCmd.Flags().StringVar(&apiToken, "api-token", "", "API token for the SaaS context")
	createCtxCmd.Flags().StringVar(&kubeConfig, "kubeconfig", "", "path to the kubeconfig file; valid only if user doesn't choose 'endpoint' option.(See [*])")
	createCtxCmd.Flags().StringVar(&kubeContext, "kubecontext", "", "the context in the kubeconfig to use; valid only if user doesn't choose 'endpoint' option.(See [*]) ")
	createCtxCmd.Flags().BoolVar(&stderrOnly, "stderr-only", false, "send all output to stderr rather than stdout")
	createCtxCmd.Flags().BoolVar(&forceCSP, "force-csp", false, "force the context to use CSP auth")
	createCtxCmd.Flags().BoolVar(&staging, "staging", false, "use CSP staging issuer")
	_ = createCtxCmd.Flags().MarkHidden("api-token")
	_ = createCtxCmd.Flags().MarkHidden("stderr-only")
	_ = createCtxCmd.Flags().MarkHidden("force-csp")
	_ = createCtxCmd.Flags().MarkHidden("staging")
}

func createCtx(_ *cobra.Command, _ []string) (err error) {
	ctx, err := createNewContext()
	if err != nil {
		return err
	}
	if ctx.Type == configapi.CtxTypeK8s {
		err = k8sLogin(ctx)
	} else {
		err = globalLogin(ctx)
	}

	if err != nil {
		return err
	}

	// Sync all required plugins if the "features.global.context-aware-cli-for-plugins" feature is enabled
	if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
		if err = pluginmanager.SyncPlugins(); err != nil {
			log.Warning("unable to automatically sync the plugins from target context. Please run 'tanzu plugin sync' command to sync plugins manually")
		}
	}

	return nil
}

func isGlobalContext(endpoint string) bool {
	if strings.Contains(endpoint, knownGlobalHost) {
		return true
	}
	if forceCSP {
		return true
	}
	return false
}

func getPromptOpts() []component.PromptOpt {
	var promptOpts []component.PromptOpt
	if stderrOnly {
		// This uses stderr because it needs to work inside the kubectl exec plugin flow where stdout is reserved.
		promptOpts = append(promptOpts, component.WithStdio(os.Stdin, os.Stderr, os.Stderr))
	}
	return promptOpts
}

func createNewContext() (context *configapi.Context, err error) {
	// user provided command line options to create a context using kubeconfig[optional] and context
	if kubeContext != "" {
		return createContextWithKubeconfig()
	}
	// user provided command line options to create a context using endpoint
	if endpoint != "" {
		return createContextWithEndpoint()
	}
	promptOpts := getPromptOpts()

	var ctxCreationType string

	err = component.Prompt(
		&component.PromptConfig{
			Message: "Select context creation type",
			Options: []string{"Control plane endpoint", "Local kubeconfig"},
			Default: "Control plane endpoint",
		},
		&ctxCreationType,
		promptOpts...,
	)
	if err != nil {
		return context, err
	}

	if ctxCreationType == "Control plane endpoint" {
		return createContextWithEndpoint()
	}

	return createContextWithKubeconfig()
}

func createContextWithKubeconfig() (context *configapi.Context, err error) {
	promptOpts := getPromptOpts()
	if kubeConfig == "" && kubeContext == "" {
		err = component.Prompt(
			&component.PromptConfig{
				Message: "Enter path to kubeconfig (if any)",
			},
			&kubeConfig,
			promptOpts...,
		)
		if err != nil {
			return
		}
	} else if kubeConfig == "" {
		kubeConfig = getDefaultKubeconfigPath()
	}

	if kubeConfig != "" && kubeContext == "" {
		err = component.Prompt(
			&component.PromptConfig{
				Message: "Enter kube context to use",
			},
			&kubeContext,
			promptOpts...,
		)
		if err != nil {
			return
		}
	}

	if ctxName == "" {
		err = component.Prompt(
			&component.PromptConfig{
				Message: "Give the context a name",
			},
			&ctxName,
			promptOpts...,
		)
		if err != nil {
			return
		}
	}
	exists, err := config.ContextExists(ctxName)
	if err != nil {
		return context, err
	}
	if exists {
		err = fmt.Errorf("context %q already exists", ctxName)
		return
	}

	context = &configapi.Context{
		Name: ctxName,
		Type: configapi.CtxTypeK8s,
		ClusterOpts: &configapi.ClusterServer{
			Path:                kubeConfig,
			Context:             kubeContext,
			Endpoint:            endpoint,
			IsManagementCluster: true,
		},
	}
	return context, err
}

func createContextWithEndpoint() (context *configapi.Context, err error) {
	promptOpts := getPromptOpts()
	if endpoint == "" {
		err = component.Prompt(
			&component.PromptConfig{
				Message: "Enter control plane endpoint",
			},
			&endpoint,
			promptOpts...,
		)
		if err != nil {
			return
		}
	}

	if ctxName == "" {
		err = component.Prompt(
			&component.PromptConfig{
				Message: "Give the context a name",
			},
			&ctxName,
			promptOpts...,
		)
		if err != nil {
			return
		}
	}
	exists, err := config.ContextExists(ctxName)
	if err != nil {
		return context, err
	}
	if exists {
		err = fmt.Errorf("context %q already exists", ctxName)
		return
	}

	if isGlobalContext(endpoint) {
		context = &configapi.Context{
			Name:       ctxName,
			Type:       configapi.CtxTypeTMC,
			GlobalOpts: &configapi.GlobalServer{Endpoint: sanitizeEndpoint(endpoint)},
		}
	} else {
		// While this would add an extra HTTP round trip, it avoids the need to
		// add extra provider specific login flags.
		isVSphereSupervisor, err := wcpauth.IsVSphereSupervisor(endpoint, getDiscoveryHTTPClient())
		// Fall back to assuming non vSphere supervisor.
		if err != nil {
			log.Fatalf("Error creating kubeconfig with tanzu pinniped-auth login plugin: %v", err)
			return nil, err
		}
		if isVSphereSupervisor {
			log.Info("Detected a vSphere Supervisor being used")
			kubeConfig, kubeContext, err = vSphereSupervisorLogin(endpoint)
			if err != nil {
				log.Fatalf("Error logging in to vSphere Supervisor: %v", err)
				return nil, err
			}
		} else {
			kubeConfig, kubeContext, err = tkgauth.KubeconfigWithPinnipedAuthLoginPlugin(endpoint, nil, tkgauth.DiscoveryStrategy{ClusterInfoConfigMap: tkgauth.DefaultClusterInfoConfigMap})
			if err != nil {
				log.Fatalf("Error creating kubeconfig with tanzu pinniped-auth login plugin: %v", err)
				return nil, err
			}
		}

		context = &configapi.Context{
			Name: ctxName,
			Type: configapi.CtxTypeK8s,
			ClusterOpts: &configapi.ClusterServer{
				Path:                kubeConfig,
				Context:             kubeContext,
				Endpoint:            endpoint,
				IsManagementCluster: true,
			},
		}
	}
	return context, err
}

func globalLogin(c *configapi.Context) (err error) {
	apiToken, apiTokenExists := os.LookupEnv(config.EnvAPITokenKey)

	issuer := csp.ProdIssuer
	if staging {
		issuer = csp.StgIssuer
	}
	if apiTokenExists {
		log.Debug("API token env var is set")
	} else {
		apiToken, err = promptAPIToken()
		if err != nil {
			return err
		}
	}
	token, err := csp.GetAccessTokenFromAPIToken(apiToken, issuer)
	if err != nil {
		return err
	}
	claims, err := csp.ParseToken(&oauth2.Token{AccessToken: token.AccessToken})
	if err != nil {
		return err
	}

	a := configapi.GlobalServerAuth{}
	a.Issuer = issuer
	a.UserName = claims.Username
	a.Permissions = claims.Permissions
	a.AccessToken = token.AccessToken
	a.IDToken = token.IDToken
	a.RefreshToken = apiToken
	a.Type = "api-token"
	expiresAt := time.Now().Local().Add(time.Second * time.Duration(token.ExpiresIn))
	a.Expiration = metav1.NewTime(expiresAt)
	c.GlobalOpts.Auth = a

	err = config.AddContext(c, true)
	if err != nil {
		return err
	}

	// format
	fmt.Println()
	log.Success("successfully created a TMC context")
	return nil
}

// Interactive way to create a TMC context. User will be prompted for CSP API token.
func promptAPIToken() (apiToken string, err error) {
	consoleURL := url.URL{
		Scheme:   "https",
		Host:     "console.cloud.vmware.com",
		Path:     "/csp/gateway/portal/",
		Fragment: "/user/tokens",
	}

	// format
	fmt.Println()
	log.Infof(
		"If you don't have an API token, visit the VMware Cloud Services console, select your organization, and create an API token with the TMC service roles:\n  %s\n",
		consoleURL.String(),
	)
	promptOpts := getPromptOpts()

	// format
	fmt.Println()
	err = component.Prompt(
		&component.PromptConfig{
			Message:   "API Token",
			Sensitive: true,
		},
		&apiToken,
		promptOpts...,
	)
	return
}

func k8sLogin(c *configapi.Context) error {
	if c.ClusterOpts.Path != "" && c.ClusterOpts.Context != "" {
		_, err := tkgauth.GetServerKubernetesVersion(c.ClusterOpts.Path, c.ClusterOpts.Context)
		if err != nil {
			log.Fatalf("failed to create context %q for a kubernetes cluster, %v", c.Name, err)
			return err
		}
		err = config.AddContext(c, true)
		if err != nil {
			return err
		}
		log.Successf("successfully created a kubernetes context using the kubeconfig %s", c.ClusterOpts.Path)
		return nil
	}

	return fmt.Errorf("not yet implemented")
}

func sanitizeEndpoint(endpoint string) string {
	if len(strings.Split(endpoint, ":")) == 1 {
		return fmt.Sprintf("%s:443", endpoint)
	}
	return endpoint
}

func getDefaultKubeconfigPath() string {
	kubeConfigPath := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	// fallback to default kubeconfig file location if no env variable set
	if kubeConfigPath == "" {
		kubeConfigPath = clientcmd.RecommendedHomeFile
	}
	return kubeConfigPath
}

func getDiscoveryHTTPClient() *http.Client {
	// XXX: Insecure, but follows the existing tanzu login discovery patterns. If
	// there's something tracking not TOFUing, it might be good to follow that
	// eventually.
	tr := &http.Transport{
		// #nosec
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	return &http.Client{Transport: tr}
}

func vSphereSupervisorLogin(endpoint string) (mergeFilePath, currentContext string, err error) {
	port := 443
	kubeConfig, kubecontext, err := tkgauth.KubeconfigWithPinnipedAuthLoginPlugin(endpoint, nil, tkgauth.DiscoveryStrategy{DiscoveryPort: &port, ClusterInfoConfigMap: wcpauth.SupervisorVIPConfigMapName})
	if err != nil {
		log.Fatalf("Error creating kubeconfig with tanzu pinniped-auth login plugin: %v", err)
		return "", "", err
	}
	return kubeConfig, kubecontext, err
}

var listCtxCmd = &cobra.Command{
	Use:   "list",
	Short: "List contexts",
	RunE:  listCtx,
}

func listCtx(cmd *cobra.Command, _ []string) error {
	cfg, err := config.GetClientConfig()
	if err != nil {
		return err
	}

	if outputFormat == "" || outputFormat == string(component.TableOutputType) {
		displayContextListOutputSplitViewTarget(cfg, cmd.OutOrStdout())
	} else {
		displayContextListOutputListView(cfg, cmd.OutOrStdout())
	}

	return nil
}

var getCtxCmd = &cobra.Command{
	Use:   "get CONTEXT_NAME",
	Short: "Display a context from the config",
	RunE:  getCtx,
}

func getCtx(cmd *cobra.Command, args []string) error {
	var ctx *configapi.Context
	var err error
	if len(args) == 0 {
		ctx, err = promptCtx()
		if err != nil {
			return err
		}
	} else {
		ctx, err = config.GetContext(args[0])
		if err != nil {
			return err
		}
	}

	op := component.NewObjectWriter(cmd.OutOrStdout(), getOutputFmt, ctx)
	op.Render()
	return nil
}

func promptCtx() (*configapi.Context, error) {
	cfg, err := config.GetClientConfig()
	if err != nil {
		return nil, err
	}
	if len(cfg.KnownContexts) == 0 {
		return nil, errors.New("no contexts found")
	}

	promptOpts := getPromptOpts()
	contexts := make(map[string]*configapi.Context)
	for _, ctx := range cfg.KnownContexts {
		info, err := config.EndpointFromContext(ctx)
		if err != nil {
			return nil, err
		}
		if info == "" && ctx.Type == configapi.CtxTypeK8s {
			info = fmt.Sprintf("%s:%s", ctx.ClusterOpts.Path, ctx.ClusterOpts.Context)
		}

		ctxKey := rpad(ctx.Name, 20)
		ctxKey = fmt.Sprintf("%s(%s)", ctxKey, info)
		contexts[ctxKey] = ctx
	}

	ctxKeys := getKeys(contexts)
	ctxKey := ctxKeys[0]
	err = component.Prompt(
		&component.PromptConfig{
			Message: "Select a context",
			Options: ctxKeys,
			Default: ctxKey,
		},
		&ctxKey,
		promptOpts...,
	)
	if err != nil {
		return nil, err
	}
	return contexts[ctxKey], nil
}

func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func getKeys(m map[string]*configapi.Context) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var deleteCtxCmd = &cobra.Command{
	Use:   "delete CONTEXT_NAME",
	Short: "Delete a context from the config",
	RunE:  deleteCtx,
}

func deleteCtx(_ *cobra.Command, args []string) error {
	var name string
	if len(args) == 0 {
		ctx, err := promptCtx()
		if err != nil {
			return err
		}
		name = ctx.Name
	} else {
		name = args[0]
	}

	if !unattended {
		isAborted := component.AskForConfirmation("Deleting the context entry from the config will remove it from the list of tracked contexts. " +
			"You will need to use `tanzu context create` to re-create this context. Are you sure you want to continue?")
		if isAborted != nil {
			return nil
		}
	}

	log.Infof("Deleting entry for cluster %s", name)
	err := config.RemoveContext(name)
	if err != nil {
		return err
	}

	return nil
}

var useCtxCmd = &cobra.Command{
	Use:   "use CONTEXT_NAME",
	Short: "Set the context to be used by default",
	RunE:  useCtx,
}

func useCtx(_ *cobra.Command, args []string) error {
	var name string
	if len(args) == 0 {
		ctx, err := promptCtx()
		if err != nil {
			return err
		}
		name = ctx.Name
	} else {
		name = args[0]
	}

	err := config.SetCurrentContext(name)
	if err != nil {
		return err
	}
	return nil
}

func displayContextListOutputListView(cfg *configapi.ClientConfig, writer io.Writer) {
	op := component.NewOutputWriter(writer, outputFormat, "Name", "Type", "IsManagementCluster", "IsCurrent", "Endpoint", "KubeConfigPath", "KubeContext")
	for _, ctx := range cfg.KnownContexts {
		if ctxType != "" && ctx.Type != configapi.ContextType(ctxType) {
			continue
		}
		isMgmtCluster := ctx.IsManagementCluster()
		isCurrent := ctx.Name == cfg.CurrentContext[ctx.Type]
		if onlyCurrent && !isCurrent {
			continue
		}

		var endpoint, path, context string
		switch ctx.Type {
		case configapi.CtxTypeTMC:
			endpoint = ctx.GlobalOpts.Endpoint
		default:
			endpoint = ctx.ClusterOpts.Endpoint
			path = ctx.ClusterOpts.Path
			context = ctx.ClusterOpts.Context
		}
		op.AddRow(ctx.Name, ctx.Type, isMgmtCluster, isCurrent, endpoint, path, context)
	}
	op.Render()
}

func displayContextListOutputSplitViewTarget(cfg *configapi.ClientConfig, writer io.Writer) {
	outputWriterK8sTarget := component.NewOutputWriter(writer, outputFormat, "Name", "IsActive", "Endpoint", "KubeConfigPath", "KubeContext")
	outputWriterTMCTarget := component.NewOutputWriter(writer, outputFormat, "Name", "IsActive", "Endpoint")
	for _, ctx := range cfg.KnownContexts {
		if ctxType != "" && ctx.Type != configapi.ContextType(ctxType) {
			continue
		}
		isCurrent := ctx.Name == cfg.CurrentContext[ctx.Type]
		if onlyCurrent && !isCurrent {
			continue
		}

		var endpoint, path, context string
		switch ctx.Type {
		case configapi.CtxTypeTMC:
			endpoint = ctx.GlobalOpts.Endpoint
			outputWriterTMCTarget.AddRow(ctx.Name, isCurrent, endpoint)
		default:
			endpoint = ctx.ClusterOpts.Endpoint
			path = ctx.ClusterOpts.Path
			context = ctx.ClusterOpts.Context
			outputWriterK8sTarget.AddRow(ctx.Name, isCurrent, endpoint, path, context)
		}
	}

	cyanBold := color.New(color.FgCyan).Add(color.Bold)
	cyanBoldItalic := color.New(color.FgCyan).Add(color.Bold, color.Italic)
	_, _ = cyanBold.Println("Target: ", cyanBoldItalic.Sprintf("%s", configapi.CtxTypeK8s))
	outputWriterK8sTarget.Render()

	_, _ = cyanBold.Println("Target: ", cyanBoldItalic.Sprintf("%s", configapi.CtxTypeTMC))
	outputWriterTMCTarget.Render()
}
