package csp

// import (
// 	"context"
// 	"time"

// 	"gitlab.eng.vmware.com/olympus/api/pkg/common/auth"
// 	"golang.org/x/oauth2"
// 	"google.golang.org/grpc"
// 	grpc_oauth "google.golang.org/grpc/credentials/oauth"
// )

// const (
// 	mdKeyAuthToken   = "Authorization"
// 	authTokenPrefix  = "Bearer "
// 	mdKeyAuthIDToken = "X-User-Id"
// )

// // IsExpired checks for the token expiry and returns true if the token has expired else will return false
// func IsExpired(tokenExpiry time.Time) bool {
// 	// refresh at half token life
// 	now := time.Now().Unix()
// 	halfDur := -time.Duration((tokenExpiry.Unix()-now)/2) * time.Second
// 	if tokenExpiry.Add(halfDur).Unix() < now {
// 		return true
// 	}
// 	return false
// }

// // WithCredentialDiscovery returns a grpc.CallOption that adds credentials into gRPC calls.
// // The credentials are loaded from the auth context found on the machine.
// func WithCredentialDiscovery() (grpc.CallOption, error) {
// 	ctx, err := client.GetCurrentContext()
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Wrap our TokenSource to supply id tokens
// 	return grpc.PerRPCCredentials(&CSPTokenSource{
// 		TokenSource: ctx,
// 	}), nil
// }

// // WithStaticCreds will wrap a static access token into a grpc.CallOption
// func WithStaticCreds(accessToken string) grpc.CallOption {
// 	return grpc.PerRPCCredentials(&grpc_oauth.TokenSource{
// 		TokenSource: oauth2.StaticTokenSource(
// 			&oauth2.Token{AccessToken: accessToken},
// 		),
// 	})
// }

// // CSPTokenSource supplies PerRPCCredentials from an oauth2.TokenSource using CSP as the IDP.
// // It will supply access token through authorization header and id_token through user-Id header
// type CSPTokenSource struct {
// 	oauth2.TokenSource
// }

// // GetRequestMetadata gets the request metadata as a map from a TokenSource.
// func (ts CSPTokenSource) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
// 	token, err := ts.Token()
// 	if err != nil {
// 		return nil, err
// 	}

// 	headers := map[string]string{mdKeyAuthToken: authTokenPrefix + " " + token.AccessToken}
// 	idTok := auth.GetIdTokenFromTokenSource(*token)
// 	if idTok != "" {
// 		headers[mdKeyAuthIDToken] = idTok
// 	}

// 	return headers, nil
// }

// // RequireTransportSecurity indicates whether the credentials requires transport security.
// func (ts CSPTokenSource) RequireTransportSecurity() bool {
// 	return true
// }
