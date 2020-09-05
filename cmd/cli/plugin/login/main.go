package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/vmware-tanzu-private/core/pkg/v1/client"

	"golang.org/x/oauth2"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aunum/log"
	"github.com/spf13/cobra"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
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

func init() {
	globalLoginCmd.Flags().BoolVar(&stderrOnly, "stderr-only", false, "send all output to stderr rather than stdout")
	globalLoginCmd.Flags().MarkHidden("stderr-only")
}

func main() {
	p, err := plugin.NewPlugin(descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.AddCommands(
		globalLoginCmd,
		regionLoginCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

// tanzu login
func login(cmd *cobra.Command, args []string) (err error) {

	newEndpointSelector := "+ new endpoint"
	questions := []*survey.Question{
		{
			Name: "endpoint",
			Prompt: &survey.Select{
				Message: "Select an endpoint",
				Options: []string{},
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Endpoint string `survey:"endpoint"`
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
}

var globalLoginCmd = &cobra.Command{
	Use:   "global",
	Short: "login to the global control plane",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := &clientv1alpha1.Context{}
		apiToken, apiTokenExists := os.LookupEnv(csp.APITokenKey)
		endpoint, endpointExists := os.LookupEnv(csp.APITokenKey)

		// TODO (pbarker): make this configurable
		issuer := csp.ProdIssuer
		if apiTokenExists && endpointExists {
			log.Debug("API token and endpoint env vars are set")
		} else {
			apiToken, endpoint, err = interactiveGlobalLogin(c)
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
		c.Spec = clientv1alpha1.ContextSpec{}
		c.Spec.OrgID = claims.OrgID
		c.Spec.GlobalAuth.Endpoint = endpoint
		c.Spec.GlobalAuth.Issuer = issuer
		c.Status = clientv1alpha1.ContextStatus{}

		c.Status.GlobalAuth.UserName = claims.Username
		c.Status.GlobalAuth.Permissions = claims.Permissions
		c.Status.GlobalAuth.AccessToken = token.AccessToken
		c.Status.GlobalAuth.IDToken = token.IDToken
		c.Status.GlobalAuth.RefreshToken = apiToken
		c.Status.GlobalAuth.Type = "api-token"

		expiresAt := time.Now().Local().Add(time.Second * time.Duration(token.ExpiresIn))
		c.Status.GlobalAuth.Expiration = metav1.NewTime(expiresAt)

		err = client.StoreContext(c)
		if err != nil {
			return err
		}

		// format
		fmt.Println()
		log.Success("successfully logged into global control plane")
		return nil
	},
}

var regionLoginCmd = &cobra.Command{
	Use:   "region",
	Short: "login to a regional control plane",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// Interactive way to login to TMC. User will be prompted for token and context name.
func interactiveGlobalLogin(c *clientv1alpha1.Context) (apiToken string, endpoint string, err error) {
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
		// TODO (pbarker): remove this once global tenancy api is public
		{
			Name: "endpoint",
			Prompt: &survey.Input{
				Message: "Endpoint for global control plane",
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		APIToken string `survey:"apiToken"`
		Endpoint string `survey:"endpoint"`
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

	return answers.APIToken, answers.Endpoint, nil
}
