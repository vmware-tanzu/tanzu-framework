package kubeconfig

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/vmware-tanzu/tanzu-framework/pinniped-components/common/pkg/pinnipedinfo"
)

const (
	// ConciergeNamespace is the namespace where pinniped concierge is deployed
	ConciergeNamespace = "pinniped-concierge"

	// ConciergeAuthenticatorType is the pinniped concierge authenticator type
	ConciergeAuthenticatorType = "jwt"

	// ConciergeAuthenticatorName is the pinniped concierge authenticator object name
	ConciergeAuthenticatorName = "tkg-jwt-authenticator"

	// PinnipedOIDCScopes are the scopes of pinniped oidc
	PinnipedOIDCScopes = "offline_access,openid,pinniped:request-audience"

	// DefaultPinnipedLoginTimeout is the default login timeout
	DefaultPinnipedLoginTimeout = time.Minute
)

// GetPinnipedKubeconfig generate kubeconfig given cluster-info and pinniped-info and the requested audience
func GetPinnipedKubeconfig(cluster *api.Cluster, pinnipedInfo *pinnipedinfo.PinnipedInfo, clustername, audience string) (*api.Config, error) {
	execConfig := api.ExecConfig{
		APIVersion: v1beta1.SchemeGroupVersion.String(),
		Args:       []string{},
		Env:        []api.ExecEnvVar{},
	}

	execConfig.Command = "tanzu"
	execConfig.Args = append([]string{"pinniped-auth", "login"}, execConfig.Args...)

	conciergeEndpoint := cluster.Server
	if pinnipedInfo.ConciergeEndpoint != "" {
		conciergeEndpoint = pinnipedInfo.ConciergeEndpoint
	}

	// configure concierge
	execConfig.Args = append(execConfig.Args,
		"--enable-concierge",
		"--concierge-authenticator-name="+ConciergeAuthenticatorName,
		"--concierge-authenticator-type="+ConciergeAuthenticatorType,
		"--concierge-endpoint="+conciergeEndpoint,
		"--concierge-ca-bundle-data="+base64.StdEncoding.EncodeToString(cluster.CertificateAuthorityData),
		"--issuer="+pinnipedInfo.Issuer, // configure OIDC
		"--scopes="+PinnipedOIDCScopes,
		"--ca-bundle-data="+pinnipedInfo.IssuerCABundleData,
		"--request-audience="+audience,
	)

	if os.Getenv("TANZU_CLI_PINNIPED_AUTH_LOGIN_SKIP_BROWSER") != "" {
		execConfig.Args = append(execConfig.Args, "--skip-browser")
	}

	username := "tanzu-cli-" + clustername
	contextName := fmt.Sprintf("%s@%s", username, clustername)

	return &api.Config{
		Kind:           "Config",
		APIVersion:     api.SchemeGroupVersion.Version,
		Clusters:       map[string]*api.Cluster{clustername: cluster},
		AuthInfos:      map[string]*api.AuthInfo{username: {Exec: &execConfig}},
		Contexts:       map[string]*api.Context{contextName: {Cluster: clustername, AuthInfo: username}},
		CurrentContext: contextName,
	}, nil
}
