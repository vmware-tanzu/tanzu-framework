package main

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/vmware-tanzu-private/core/pkg/v1/client"

	"golang.org/x/oauth2"

	"github.com/AlecAivazis/survey/v2"
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
	Version:     "v0.0.1",
	Group:       cli.SystemCmdGroup,
}

var (
	stderrOnly                       bool
	endpoint, name, apiToken, server string
)

const (
	knownTMCHost = "tmc.cloud.vmware.com"
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
	p.Cmd.Flags().BoolVar(&stderrOnly, "stderr-only", false, "send all output to stderr rather than stdout")
	p.Cmd.Flags().MarkHidden("stderr-only")
	p.Cmd.RunE = login
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

	surveyOpts := getSurveyOpts()

	newServerSelector := "+ new server"
	var serverTarget *clientv1alpha1.Server
	if server == "" {
		servers := map[string]*clientv1alpha1.Server{}
		for _, server := range cfg.KnownServers {
			endpoint, err := client.EndpointFromServer(server)
			if err != nil {
				return err
			}

			s := rpad(server.Name, 20)
			s = fmt.Sprintf("%s(%s)", s, endpoint)
			servers[s] = server
		}
		if endpoint == "" {
			endpoint, _ = os.LookupEnv(client.EnvEndpointKey)
		}
		if len(servers) == 0 {
			serverTarget, err = createNewServer()
			if err != nil {
				return err
			}
		} else {
			serverKeys := getKeys(servers)
			serverKeys = append(serverKeys, newServerSelector)
			servers[newServerSelector] = &clientv1alpha1.Server{}
			err = survey.AskOne(
				&survey.Select{
					Message: "Select a server",
					Options: serverKeys,
				},
				&server,
				surveyOpts...,
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
	if strings.Contains(endpoint, knownTMCHost) {
		return true
	}
	return false
}

func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func getSurveyOpts() []survey.AskOpt {
	var surveyOpts []survey.AskOpt
	if stderrOnly {
		// This uses stderr because it needs to work inside the kubectl exec plugin flow where stdout is reserved.
		surveyOpts = append(surveyOpts, survey.WithStdio(os.Stdin, os.Stderr, os.Stderr))
	}
	return surveyOpts
}

func createNewServer() (server *clientv1alpha1.Server, err error) {
	surveyOpts := getSurveyOpts()
	if endpoint == "" {
		err = survey.AskOne(
			&survey.Input{
				Message: "Enter server endpoint",
			},
			&endpoint,
			surveyOpts...,
		)
		if err != nil {
			return
		}
	}
	if name == "" {
		err = survey.AskOne(
			&survey.Input{
				Message: "Give the server a name",
			},
			&name,
			surveyOpts...,
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
		server = &clientv1alpha1.Server{
			Name: name,
			Type: clientv1alpha1.ManagementClusterServerType,
		}
	}
	return
}

func globalLogin(s *clientv1alpha1.Server) (err error) {
	a := clientv1alpha1.GlobalServerAuth{}
	apiToken, apiTokenExists := os.LookupEnv(client.EnvAPITokenKey)

	// TODO (pbarker): configurable issuer
	issuer := csp.ProdIssuer
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

	var surveyOpts []survey.AskOpt
	if stderrOnly {
		// This uses stderr because it needs to work inside the kubectl exec plugin flow where stdout is reserved.
		surveyOpts = append(surveyOpts, survey.WithStdio(os.Stdin, os.Stderr, os.Stderr))
	}

	// format
	fmt.Println()
	err = survey.AskOne(
		&survey.Password{
			Message: "API Token"},
		&apiToken,
		surveyOpts...,
	)
	return
}

// TODO (pbarker): need pinniped story more fleshed out
func managementClusterLogin(s *clientv1alpha1.Server, endpoint string) error {
	return fmt.Errorf("not yet implemented")
}

func sanitizeEndpoint(endpoint string) string {
	if len(strings.Split(endpoint, ":")) == 1 {
		return fmt.Sprintf("%s:443", endpoint)
	}
	return endpoint
}
