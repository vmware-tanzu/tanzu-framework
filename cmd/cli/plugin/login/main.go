package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/vmware-tanzu-private/core/pkg/v1/client"

	"golang.org/x/oauth2"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aunum/log"
	"github.com/spf13/cobra"

	authv1alpha1 "github.com/vmware-tanzu-private/core/apis/auth/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/auth/csp"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/commands/plugin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var descriptor = cli.PluginDescriptor{
	Name:        "login",
	Description: "Login to the platform",
	Version:     "v0.0.1",
	Group:       cli.SystemCmdGroup,
}

var (
	stderrOnly bool
)

const (
	knownTMCHost = "tmc.cloud.vmware.com"
)

func init() {
	// globalLoginCmd.Flags().BoolVar(&stderrOnly, "stderr-only", false, "send all output to stderr rather than stdout")
	// globalLoginCmd.Flags().MarkHidden("stderr-only")
}

func main() {
	p, err := plugin.NewPlugin(descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.AddCommands()
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

// tanzu login
func login(cmd *cobra.Command, args []string) (err error) {
	cfg, err := client.GetConfig()
	if err != nil {
		return err
	}
	servers := []string{}
	for _, server := range cfg.Spec.Servers {
		endpoint, err := csp.EndpointFromServer(server)
		if err != nil {
			return err
		}
		s := fmt.Sprintf("%s (%s)", server.Name, endpoint)
		servers = append(servers, s)
	}

	tmcServerSelector := knownTMCHost
	newServerSelector := "+ new server"
	servers = append(servers, tmcServerSelector, newServerSelector)
	questions := []*survey.Question{
		{
			Name: "server",
			Prompt: &survey.Select{
				Message: "Select a server",
				Options: []string{},
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Server string `survey:"server"`
	}{}

	var surveyOpts []survey.AskOpt
	if stderrOnly {
		// This uses stderr because it needs to work inside the kubectl exec plugin flow where stdout is reserved.
		surveyOpts = append(surveyOpts, survey.WithStdio(os.Stdin, os.Stderr, os.Stderr))
	}

	// format
	fmt.Println()
	err = survey.Ask(questions, &answers, surveyOpts...)
	if err != nil {
		return
	}
	if answers.Server == newServerSelector {
		questions := []*survey.Question{
			{
				Name: "endpoint",
				Prompt: &survey.Input{
					Message: "Enter an endpoint",
				},
			},
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Give the server a name",
				},
			},
		}

		answers := struct {
			Endpoint string `survey:"endpoint"`
			Name     string `survey:"name"`
		}{}

		var surveyOpts []survey.AskOpt
		if stderrOnly {
			// This uses stderr because it needs to work inside the kubectl exec plugin flow where stdout is reserved.
			surveyOpts = append(surveyOpts, survey.WithStdio(os.Stdin, os.Stderr, os.Stderr))
		}

		// format
		fmt.Println()
		err = survey.Ask(questions, &answers, surveyOpts...)
		if err != nil {
			return
		}

		if isTanzuServer(answers.Endpoint) {
			err = tanzuLogin(answers.Name, answers.Endpoint)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isTanzuServer(endpoint string) bool {
	if strings.Contains(endpoint, knownTMCHost) {
		return true
	}
	return false
}

// var globalLoginCmd = &cobra.Command{
// 	Use:   "global",
// 	Short: "login to the global control plane",
// 	RunE: func(cmd *cobra.Command, args []string) (err error) {

// 	},
// }

func tanzuLogin(name, endpoint string) (err error) {
	c := &authv1alpha1.CSPConfig{}
	apiToken, apiTokenExists := os.LookupEnv(csp.APITokenKey)
	endpoint, endpointExists := os.LookupEnv(csp.APITokenKey)

	// TODO (pbarker): configurable issuer
	issuer := csp.ProdIssuer
	if apiTokenExists && endpointExists {
		log.Debug("API token and endpoint env vars are set")
	} else {
		apiToken, err = interactiveGlobalLogin(c)
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
	c.Spec = authv1alpha1.CSPConfigSpec{}
	c.Spec.Endpoint = endpoint
	c.Spec.Issuer = issuer
	c.Status = authv1alpha1.CSPConfigStatus{}

	c.Status.UserName = claims.Username
	c.Status.Permissions = claims.Permissions
	c.Status.AccessToken = token.AccessToken
	c.Status.IDToken = token.IDToken
	c.Status.RefreshToken = apiToken
	c.Status.Type = "api-token"

	expiresAt := time.Now().Local().Add(time.Second * time.Duration(token.ExpiresIn))
	c.Status.Expiration = metav1.NewTime(expiresAt)

	err = csp.StoreConfig(c)
	if err != nil {
		return err
	}

	// format
	fmt.Println()
	log.Success("successfully logged into global control plane")
	return nil
}

// var regionLoginCmd = &cobra.Command{
// 	Use:   "region",
// 	Short: "login to a regional control plane",
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		return nil
// 	},
// }

// Interactive way to login to TMC. User will be prompted for token and context name.
func interactiveGlobalLogin(c *authv1alpha1.CSPConfig) (apiToken string, err error) {
	consoleURL := url.URL{
		Scheme:   "https",
		Host:     "console.cloud.vmware.com",
		Path:     "/csp/gateway/portal/",
		Fragment: "/user/tokens",
	}

	log.Infof(
		"If you don't have an API token, visit the VMware Cloud Services console, select your organization, and create an API token with the TMC service roles:\n  %s\n",
		consoleURL.String(),
	)

	questions := []*survey.Question{
		{
			Name: "apiToken",
			Prompt: &survey.Password{
				Message: "API Token",
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		APIToken string `survey:"apiToken"`
	}{}

	var surveyOpts []survey.AskOpt
	if stderrOnly {
		// This uses stderr because it needs to work inside the kubectl exec plugin flow where stdout is reserved.
		surveyOpts = append(surveyOpts, survey.WithStdio(os.Stdin, os.Stderr, os.Stderr))
	}

	// format
	fmt.Println()
	err = survey.Ask(questions, &answers, surveyOpts...)
	if err != nil {
		return
	}

	return answers.APIToken, nil
}
