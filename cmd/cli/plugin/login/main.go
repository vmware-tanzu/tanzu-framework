// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	tkgauth "github.com/vmware-tanzu-private/core/pkg/v1/auth/tkg"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"

	"golang.org/x/oauth2"

	"github.com/aunum/log"
	"github.com/spf13/cobra"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/auth/csp"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var descriptor = cli.PluginDescriptor{
	Name:        "login",
	Description: "Login to the platform",
	Version:     cli.BuildVersion,
	Group:       cli.SystemCmdGroup,
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
	p.Cmd.Flags().MarkHidden("stderr-only") //nolint
	p.Cmd.Flags().MarkHidden("force-csp")   //nolint
	p.Cmd.Flags().MarkHidden("staging")     //nolint
	p.Cmd.RunE = login
	p.Cmd.Example = `
	# Login to TKG management cluster using endpoint
	tanzu login --endpoint "https://login.example.com"  --name mgmt-cluster

	# Login to TKG management cluster by using kubeconfig path and context for the management cluster
	tanzu login --kubeconfig path/to/kubeconfig --context path/to/context --name mgmt-cluster

	# Login to an existing server 
	tanzu login --server mgmt-cluster
	
	[*] : User has two options to login to TKG. User can choose the login endpoint option
	by providing 'endpoint', or user can choose to use the kubeconfig for the management cluster by 
	providing 'kubeconfig' and 'context'
	`
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func login(cmd *cobra.Command, args []string) (err error) {
	cfg, err := client.GetConfig()
	if _, ok := err.(*client.ConfigNotExistError); ok {
		cfg, err = client.NewConfig()
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	promptOpts := getPromptOpts()

	newServerSelector := "+ new server"
	var serverTarget *clientv1alpha1.Server
	if name != "" {
		serverTarget, err = createNewServer()
		if err != nil {
			return err
		}
	} else if server == "" {
		servers := map[string]*clientv1alpha1.Server{}
		for _, server := range cfg.KnownServers {
			ep, err := client.EndpointFromServer(server)
			if err != nil {
				return err
			}

			s := rpad(server.Name, 20)
			s = fmt.Sprintf("%s(%s)", s, ep)
			servers[s] = server
		}
		if endpoint == "" {
			endpoint, _ = os.LookupEnv(client.EnvEndpointKey)
		}
		// If there are no existing servers
		if len(servers) == 0 {
			serverTarget, err = createNewServer()
			if err != nil {
				return err
			}
		} else {
			serverKeys := getKeys(servers)
			serverKeys = append(serverKeys, newServerSelector)
			servers[newServerSelector] = &clientv1alpha1.Server{}
			err = component.Prompt(
				&component.PromptConfig{
					Message: "Select a server",
					Options: serverKeys,
					Default: serverKeys[0],
				},
				&server,
				promptOpts...,
			)
			if err != nil {
				return
			}
			serverTarget = servers[server]
		}

	} else {
		serverTarget, err = client.GetServer(server)
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

	if serverTarget.Type == clientv1alpha1.GlobalServerType {
		return globalLogin(serverTarget)
	}

	return managementClusterLogin(serverTarget, endpoint)
}

func getKeys(m map[string]*clientv1alpha1.Server) []string {
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

func createNewServer() (server *clientv1alpha1.Server, err error) {
	// user provided command line options to create a server using kubeconfig and context
	if kubeConfig != "" && kubecontext != "" {
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

func createServerWithKubeconfig() (server *clientv1alpha1.Server, err error) {
	promptOpts := getPromptOpts()
	if kubeConfig == "" {
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
	nameExists, err := client.ServerExists(name)
	if err != nil {
		return server, err
	}
	if nameExists {
		err = fmt.Errorf("server %q already exists", name)
		return
	}
	server = &clientv1alpha1.Server{
		Name: name,
		Type: clientv1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &clientv1alpha1.ManagementClusterServer{
			Path:     kubeConfig,
			Context:  kubecontext,
			Endpoint: endpoint},
	}
	return
}

func createServerWithEndpoint() (server *clientv1alpha1.Server, err error) {
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
	nameExists, err := client.ServerExists(name)
	if err != nil {
		return server, err
	}
	if nameExists {
		err = fmt.Errorf("server %q already exists", name)
		return
	}
	if isGlobalServer(endpoint) {
		server = &clientv1alpha1.Server{
			Name:       name,
			Type:       clientv1alpha1.GlobalServerType,
			GlobalOpts: &clientv1alpha1.GlobalServer{Endpoint: sanitizeEndpoint(endpoint)},
		}
	} else {
		kubeConfig, kubecontext, err = tkgauth.KubeconfigWithPinnipedAuthLoginPlugin(endpoint, nil)
		if err != nil {
			log.Fatalf("Error creating kubeconfig with tanzu pinniped-auth login plugin: err-%v", err)
			return nil, err
		}
		server = &clientv1alpha1.Server{
			Name: name,
			Type: clientv1alpha1.ManagementClusterServerType,
			ManagementClusterOpts: &clientv1alpha1.ManagementClusterServer{
				Path:     kubeConfig,
				Context:  kubecontext,
				Endpoint: endpoint},
		}
	}
	return
}

func globalLogin(s *clientv1alpha1.Server) (err error) {
	a := clientv1alpha1.GlobalServerAuth{}
	apiToken, apiTokenExists := os.LookupEnv(client.EnvAPITokenKey)

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

	err = client.PutServer(s, true)
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
			Message: "API Token",
		},
		&apiToken,
		promptOpts...,
	)
	return
}

func managementClusterLogin(s *clientv1alpha1.Server, endpoint string) error {

	if s.ManagementClusterOpts.Path != "" && s.ManagementClusterOpts.Context != "" {
		_, err := tkgauth.GetServerKubernetesVersion(s.ManagementClusterOpts.Path, s.ManagementClusterOpts.Context)
		if err != nil {
			log.Fatalf("failed to login to the management cluster %s, err-%v", s.Name, err)
			return err
		}
		err = client.PutServer(s, true)
		if err != nil {
			return err
		}
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
