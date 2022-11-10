// Copyright 2021-2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aunum/log"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/auth/csp"
	tkgauth "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/auth/tkg"
	wcpauth "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/auth/wcp"
	cliconfig "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/pluginmanager"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
)

// nolint: gocritic
// Note: Shall be deprecated in a future version. Superseded by 'tanzu context' command.
var descriptor = cliapi.PluginDescriptor{
	Name:        "login",
	Description: "Login to the platform",
	Group:       cliapi.SystemCmdGroup,
	Aliases:     []string{"lo", "logins"},
	Version:     buildinfo.Version,
	BuildSHA:    buildinfo.SHA,
}

var (
	stderrOnly, forceCSP, staging                             bool
	endpoint, name, apiToken, server, kubeConfig, kubecontext string
)

const (
	knownGlobalHost = "cloud.vmware.com"
)

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.Cmd.Flags().StringVar(&endpoint, "endpoint", "", "endpoint to login to")
	p.Cmd.Flags().StringVar(&name, "name", "", "name of the server")
	p.Cmd.Flags().StringVar(&apiToken, "apiToken", "", "API token for global login")
	p.Cmd.Flags().StringVar(&server, "server", "", "login to the given server")
	p.Cmd.Flags().StringVar(&kubeConfig, "kubeconfig", "", "path to kubeconfig management cluster. Valid only if user doesn't choose 'endpoint' option.(See [*])")
	p.Cmd.Flags().StringVar(&kubecontext, "context", "", "the context in the kubeconfig to use for management cluster. Valid only if user doesn't choose 'endpoint' option.(See [*]) ")
	p.Cmd.Flags().BoolVar(&stderrOnly, "stderr-only", false, "send all output to stderr rather than stdout")
	p.Cmd.Flags().BoolVar(&forceCSP, "force-csp", false, "force the endpoint to be logged in as a csp server")
	p.Cmd.Flags().BoolVar(&staging, "staging", false, "use CSP staging issuer")
	p.Cmd.Flags().MarkHidden("stderr-only") // nolint
	p.Cmd.Flags().MarkHidden("force-csp")   // nolint
	p.Cmd.Flags().MarkHidden("staging")     // nolint
	p.Cmd.RunE = login
	p.Cmd.Example = `
	# Login to TKG management cluster using endpoint
	tanzu login --endpoint "https://login.example.com"  --name mgmt-cluster

	# Login to TKG management cluster by using kubeconfig path and context for the management cluster
	tanzu login --kubeconfig path/to/kubeconfig --context path/to/context --name mgmt-cluster

	# Login to TKG management cluster by using default kubeconfig path and context for the management cluster
	tanzu login  --context path/to/context --name mgmt-cluster

	# Login to an existing server
	tanzu login --server mgmt-cluster

	[*] : User has two options to login to TKG. User can choose the login endpoint option
	by providing 'endpoint', or user can choose to use the kubeconfig for the management cluster by
	providing 'kubeconfig' and 'context'. If only '--context' is set and '--kubeconfig' is unset
	$KUBECONFIG env variable would be used and, if $KUBECONFIG env is also unset default 
	kubeconfig($HOME/.kube/config) would be used
	`

	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func login(cmd *cobra.Command, args []string) (err error) {
	cfg, err := config.GetClientConfig()
	if err != nil {
		return err
	}

	newServerSelector := "+ new server"
	var serverTarget *configapi.Server
	if name != "" {
		serverTarget, err = createNewServer()
		if err != nil {
			return err
		}
	} else if server == "" {
		serverTarget, err = getServerTarget(cfg, newServerSelector)
		if err != nil {
			return err
		}
	} else {
		serverTarget, err = config.GetServer(server)
		if err != nil {
			return err
		}
	}

	if server == newServerSelector {
		serverTarget, err = createNewServer()
		if err != nil {
			return err
		}
	}

	if serverTarget.Type == configapi.GlobalServerType {
		err = globalLogin(serverTarget)
	} else {
		err = managementClusterLogin(serverTarget)
	}

	if err != nil {
		return err
	}

	// Sync all required plugins if the "features.global.context-aware-cli-for-plugins" feature is enabled
	if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
		if err = pluginmanager.SyncPlugins(); err != nil {
			log.Warning("unable to automatically sync the plugins from target server. Please run 'tanzu plugin sync' command to sync plugins manually")
		}
	}

	return nil
}

func getServerTarget(cfg *configapi.ClientConfig, newServerSelector string) (*configapi.Server, error) {
	promptOpts := getPromptOpts()
	servers := map[string]*configapi.Server{}
	for _, server := range cfg.KnownServers {
		ep, err := config.EndpointFromServer(server)
		if err != nil {
			return nil, err
		}

		s := rpad(server.Name, 20)
		s = fmt.Sprintf("%s(%s)", s, ep)
		servers[s] = server
	}
	if endpoint == "" {
		endpoint, _ = os.LookupEnv(config.EnvEndpointKey)
	}
	// If there are no existing servers
	if len(servers) == 0 {
		return createNewServer()
	}
	serverKeys := getKeys(servers)
	serverKeys = append(serverKeys, newServerSelector)
	servers[newServerSelector] = &configapi.Server{}
	err := component.Prompt(
		&component.PromptConfig{
			Message: "Select a server",
			Options: serverKeys,
			Default: serverKeys[0],
		},
		&server,
		promptOpts...,
	)
	if err != nil {
		return nil, err
	}
	return servers[server], nil
}

func getKeys(m map[string]*configapi.Server) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func isGlobalServer(endpoint string) bool {
	if strings.Contains(endpoint, knownGlobalHost) {
		return true
	}
	if forceCSP {
		return true
	}
	return false
}

func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func getPromptOpts() []component.PromptOpt {
	var promptOpts []component.PromptOpt
	if stderrOnly {
		// This uses stderr because it needs to work inside the kubectl exec plugin flow where stdout is reserved.
		promptOpts = append(promptOpts, component.WithStdio(os.Stdin, os.Stderr, os.Stderr))
	}
	return promptOpts
}

func createNewServer() (server *configapi.Server, err error) {
	// user provided command line options to create a server using kubeconfig[optional] and context
	if kubecontext != "" {
		return createServerWithKubeconfig()
	}
	// user provided command line options to create a server using endpoint
	if endpoint != "" {
		return createServerWithEndpoint()
	}
	promptOpts := getPromptOpts()

	var loginType string

	err = component.Prompt(
		&component.PromptConfig{
			Message: "Select login type",
			Options: []string{"Server endpoint", "Local kubeconfig"},
			Default: "Server endpoint",
		},
		&loginType,
		promptOpts...,
	)
	if err != nil {
		return server, err
	}

	if loginType == "Server endpoint" {
		return createServerWithEndpoint()
	}

	return createServerWithKubeconfig()
}

func createServerWithKubeconfig() (server *configapi.Server, err error) {
	promptOpts := getPromptOpts()
	if kubeConfig == "" && kubecontext == "" {
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
	}
	if kubeConfig == "" {
		kubeConfig = getDefaultKubeconfigPath()
	}

	if kubeConfig != "" && kubecontext == "" {
		err = component.Prompt(
			&component.PromptConfig{
				Message: "Enter kube context to use",
			},
			&kubecontext,
			promptOpts...,
		)
		if err != nil {
			return
		}
	}
	if name == "" {
		err = component.Prompt(
			&component.PromptConfig{
				Message: "Give the server a name",
			},
			&name,
			promptOpts...,
		)
		if err != nil {
			return
		}
	}
	nameExists, err := config.ServerExists(name)
	if err != nil {
		return server, err
	}
	if nameExists {
		err = fmt.Errorf("server %q already exists", name)
		return
	}

	endpointType := configapi.ManagementClusterServerType

	server = &configapi.Server{
		Name: name,
		Type: endpointType,
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Path:     kubeConfig,
			Context:  kubecontext,
			Endpoint: endpoint},
	}
	return server, err
}

func createServerWithEndpoint() (server *configapi.Server, err error) {
	promptOpts := getPromptOpts()
	if endpoint == "" {
		err = component.Prompt(
			&component.PromptConfig{
				Message: "Enter server endpoint",
			},
			&endpoint,
			promptOpts...,
		)
		if err != nil {
			return
		}
	}
	if name == "" {
		err = component.Prompt(
			&component.PromptConfig{
				Message: "Give the server a name",
			},
			&name,
			promptOpts...,
		)
		if err != nil {
			return
		}
	}
	nameExists, err := config.ServerExists(name)
	if err != nil {
		return server, err
	}
	if nameExists {
		err = fmt.Errorf("server %q already exists", name)
		return
	}
	if isGlobalServer(endpoint) {
		server = &configapi.Server{
			Name:       name,
			Type:       configapi.GlobalServerType,
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
			kubeConfig, kubecontext, err = vSphereSupervisorLogin(endpoint)
			if err != nil {
				log.Fatalf("Error logging in to vSphere Supervisor: %v", err)
				return nil, err
			}
		} else {
			kubeConfig, kubecontext, err = tkgauth.KubeconfigWithPinnipedAuthLoginPlugin(endpoint, nil, tkgauth.DiscoveryStrategy{ClusterInfoConfigMap: tkgauth.DefaultClusterInfoConfigMap})
			if err != nil {
				log.Fatalf("Error creating kubeconfig with tanzu pinniped-auth login plugin: %v", err)
				return nil, err
			}
		}

		server = &configapi.Server{
			Name: name,
			Type: configapi.ManagementClusterServerType,
			ManagementClusterOpts: &configapi.ManagementClusterServer{
				Path:     kubeConfig,
				Context:  kubecontext,
				Endpoint: endpoint},
		}
	}
	return server, err
}

func globalLogin(s *configapi.Server) (err error) {
	a := configapi.GlobalServerAuth{}
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

	a.Issuer = issuer

	a.UserName = claims.Username
	a.Permissions = claims.Permissions
	a.AccessToken = token.AccessToken
	a.IDToken = token.IDToken
	a.RefreshToken = apiToken
	a.Type = "api-token"

	expiresAt := time.Now().Local().Add(time.Second * time.Duration(token.ExpiresIn))
	a.Expiration = metav1.NewTime(expiresAt)

	s.GlobalOpts.Auth = a

	err = config.PutServer(s, true)
	if err != nil {
		return err
	}

	// format
	fmt.Println()
	log.Success("successfully logged into global control plane")
	return nil
}

// Interactive way to login to TMC. User will be prompted for token and context name.
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

func managementClusterLogin(s *configapi.Server) error {
	if s.ManagementClusterOpts.Path != "" && s.ManagementClusterOpts.Context != "" {
		_, err := tkgauth.GetServerKubernetesVersion(s.ManagementClusterOpts.Path, s.ManagementClusterOpts.Context)
		if err != nil {
			log.Fatalf("failed to login to the management cluster %s, %v", s.Name, err)
			return err
		}
		err = config.PutServer(s, true)
		if err != nil {
			return err
		}
		// XXX s.Name isn't necessarily a kubeconfig?
		log.Successf("successfully logged in to management cluster using the kubeconfig %s", s.Name)
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
	kubeConfigFilename := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	// fallback to default kubeconfig file location if no env variable set
	if kubeConfigFilename == "" {
		kubeConfigFilename = clientcmd.RecommendedHomeFile
	}
	return kubeConfigFilename
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
