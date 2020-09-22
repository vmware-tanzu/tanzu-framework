package csp

import (
	"context"
	"fmt"
	"time"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"gitlab.eng.vmware.com/olympus/api/pkg/common/auth"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	grpc_oauth "google.golang.org/grpc/credentials/oauth"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	mdKeyAuthToken   = "Authorization"
	authTokenPrefix  = "Bearer "
	mdKeyAuthIDToken = "X-User-Id"
)

// IsExpired checks for the token expiry and returns true if the token has expired else will return false
func IsExpired(tokenExpiry time.Time) bool {
	// refresh at half token life
	now := time.Now().Unix()
	halfDur := -time.Duration((tokenExpiry.Unix()-now)/2) * time.Second
	if tokenExpiry.Add(halfDur).Unix() < now {
		return true
	}
	return false
}

// WithCredentialDiscovery returns a grpc.CallOption that adds credentials into gRPC calls.
// The credentials are loaded from the auth context found on the machine.
func WithCredentialDiscovery() (grpc.CallOption, error) {
	cfg, err := client.GetConfig()
	if err != nil {
		return nil, err
	}
	// Wrap our TokenSource to supply id tokens
	return grpc.PerRPCCredentials(&TokenSource{
		TokenSource: &configSource{cfg},
	}), nil
}

// WithStaticCreds will wrap a static access token into a grpc.CallOption
func WithStaticCreds(accessToken string) grpc.CallOption {
	return grpc.PerRPCCredentials(&grpc_oauth.TokenSource{
		TokenSource: oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: accessToken},
		),
	})
}

type configSource struct {
	*clientv1alpha1.Config
}

func (c *configSource) Token() (*oauth2.Token, error) {
	g, err := c.GetCurrentServer()
	if err != nil {
		return nil, err
	}
	if !g.IsGlobal() {
		return nil, fmt.Errorf("trying to fetch token for non global server")
	}
	if !IsExpired(g.GlobalOpts.Auth.Expiration.Time) {
		tok := &oauth2.Token{
			AccessToken: g.GlobalOpts.Auth.AccessToken,
			Expiry:      g.GlobalOpts.Auth.Expiration.Time,
		}
		return tok.WithExtra(map[string]interface{}{
			auth.ExtraIdToken: g.GlobalOpts.Auth.IDToken,
		}), nil
	}
	token, err := auth.GetCSPAccessTokenFromAPIToken(g.GlobalOpts.Auth.RefreshToken, ProdIssuer)
	if err != nil {
		return nil, err
	}

	g.GlobalOpts.Auth.Type = "api-token"
	expiration := time.Now().Local().Add(time.Second * time.Duration(token.ExpiresIn))
	g.GlobalOpts.Auth.Expiration = metav1.NewTime(expiration)
	g.GlobalOpts.Auth.RefreshToken = token.RefreshToken
	g.GlobalOpts.Auth.AccessToken = token.AccessToken
	g.GlobalOpts.Auth.IDToken = token.IDToken

	if err = client.StoreConfig(c.Config); err != nil {
		return nil, err
	}

	tok := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       expiration,
	}
	return tok.WithExtra(map[string]interface{}{
		auth.ExtraIdToken: token.IDToken,
	}), nil
}

// TokenSource supplies PerRPCCredentials from an oauth2.TokenSource using CSP as the IDP.
// It will supply access token through authorization header and id_token through user-Id header
type TokenSource struct {
	oauth2.TokenSource
}

// GetRequestMetadata gets the request metadata as a map from a TokenSource.
func (ts TokenSource) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := ts.Token()
	if err != nil {
		return nil, err
	}

	headers := map[string]string{mdKeyAuthToken: authTokenPrefix + " " + token.AccessToken}
	idTok := auth.GetIdTokenFromTokenSource(*token)
	if idTok != "" {
		headers[mdKeyAuthIDToken] = idTok
	}

	return headers, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security.
func (ts TokenSource) RequireTransportSecurity() bool {
	return true
}
