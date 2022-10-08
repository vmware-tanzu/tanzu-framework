// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package csp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/fakes"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

const issuerURL = "https://auth0.com/"

var JWTHeader = `{"alg":"HS256","typ":"JWT"}`

func TestGetAccessTokenFromAPIToken(t *testing.T) {
	assert := assert.New(t)
	fakeHTTPClient = &fakes.FakeHTTPClient{}
	responseBody := io.NopCloser(bytes.NewReader([]byte(`{
		"id_token": "abc",
		"token_type": "Test",
		"expires_in": 86400,
		"scope": "Test",
		"access_token": "LetMeIn",
		"refresh_token": "LetMeInAgain"}`)))
	fakeHTTPClient.DoReturns(&http.Response{
		StatusCode: 200,
		Body:       responseBody,
	}, nil)
	httpRestClient = fakeHTTPClient
	token, err := GetAccessTokenFromAPIToken("asdas", issuerURL)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Error...................................")
	}
	assert.Nil(err)
	assert.Equal("LetMeIn", token.AccessToken)
}

func TestGetAccessTokenFromAPIToken_Err(t *testing.T) {
	assert := assert.New(t)

	token, err := GetAccessTokenFromAPIToken("asdas", "example.com")
	assert.NotNil(err)
	assert.Nil(token)
}
func TestGetAccessTokenFromAPIToken_FailStatus(t *testing.T) {
	assert := assert.New(t)
	fakeHTTPClient := &fakes.FakeHTTPClient{}
	responseBody := io.NopCloser(bytes.NewReader([]byte(``)))
	fakeHTTPClient.DoReturns(&http.Response{
		StatusCode: 403,
		Body:       responseBody,
	}, nil)
	httpRestClient = fakeHTTPClient
	token, err := GetAccessTokenFromAPIToken("asdas", issuerURL)
	assert.NotNil(err)
	assert.Contains(err.Error(), "obtain access token")
	assert.Nil(token)
}

func TestGetAccessTokenFromAPIToken_InvalidResponse(t *testing.T) {
	assert := assert.New(t)
	fakeHTTPClient := &fakes.FakeHTTPClient{}
	responseBody := io.NopCloser(bytes.NewReader([]byte(`[{
		"id_token": "abc",
		"token_type": "Test",
		"expires_in": 86400,
		"scope": "Test",
		"access_token": "LetMeIn",
		"refresh_token": "LetMeInAgain"}]`)))
	fakeHTTPClient.DoReturns(&http.Response{
		StatusCode: 200,
		Body:       responseBody,
	}, nil)
	httpRestClient = fakeHTTPClient

	token, err := GetAccessTokenFromAPIToken("asdas", issuerURL)
	assert.NotNil(err)
	assert.Contains(err.Error(), "could not unmarshal")
	assert.Nil(token)
}
func TestIsExpired(t *testing.T) {
	assert := assert.New(t)

	testTime := time.Now().Add(-time.Minute)
	assert.True(IsExpired(testTime))

	testTime = time.Now().Add(time.Minute * 30)
	assert.False(IsExpired(testTime))
}

func generateJWTToken(claims string) string {
	hm := hmac.New(sha256.New, []byte("secret"))
	_, _ = hm.Write([]byte(fmt.Sprintf(
		"%s.%s",
		base64.RawURLEncoding.EncodeToString([]byte(JWTHeader)),
		base64.RawURLEncoding.EncodeToString([]byte(claims)),
	)))
	sha := hex.EncodeToString(hm.Sum(nil))
	return fmt.Sprintf(
		"%s.%s.%s",
		base64.RawURLEncoding.EncodeToString([]byte(JWTHeader)),
		base64.RawURLEncoding.EncodeToString([]byte(claims)),
		sha,
	)
}

func TestParseToken_ParseFailure(t *testing.T) {
	assert := assert.New(t)

	// Pass in incorrectly formatted AccessToken
	tkn := oauth2.Token{
		AccessToken:  "LetMeIn",
		TokenType:    "Bearer",
		RefreshToken: "LetMeInAgain",
		Expiry:       time.Now().Add(time.Minute * 30),
	}

	context, err := ParseToken(&tkn)
	assert.NotNil(err)
	assert.Contains(err.Error(), "invalid")
	assert.Nil(context)
}

func TestIDTokenFromTokenSource_getIDToken(t *testing.T) {
	assert := assert.New(t)

	// Pass in incorrectly formatted AccessToken
	tkn := oauth2.Token{
		AccessToken:  "LetMeIn",
		TokenType:    "Bearer",
		RefreshToken: "LetMeInAgain",
		Expiry:       time.Now().Add(time.Minute * 30),
	}
	tknExt := tkn.WithExtra(map[string]interface{}{
		"id_token": "idtoken",
	})
	idtoken := IDTokenFromTokenSource(tknExt)
	assert.Contains(idtoken, "idtoken")
}

func TestParseToken_MissingUsername(t *testing.T) {
	assert := assert.New(t)

	accessToken := generateJWTToken(
		`{"sub":"1234567890","name":"John Doe","iat":1516239022}`,
	)
	tkn := oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		RefreshToken: "LetMeInAgain",
		Expiry:       time.Now().Add(time.Minute * 30),
	}

	context, err := ParseToken(&tkn)
	assert.NotNil(err)
	assert.Contains(err.Error(), "could not parse username")
	assert.Nil(context)
}

func TestParseToken_MissingContextName(t *testing.T) {
	assert := assert.New(t)

	accessToken := generateJWTToken(
		`{"sub":"1234567890","username":"John Doe","orgID":1516239022}`,
	)
	tkn := oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		RefreshToken: "LetMeInAgain",
		Expiry:       time.Now().Add(time.Minute * 30),
	}

	context, err := ParseToken(&tkn)
	assert.NotNil(err)
	assert.Contains(err.Error(), "could not parse orgID")
	assert.Nil(context)
}

func TestParseToken(t *testing.T) {
	assert := assert.New(t)

	accessToken := generateJWTToken(
		`{"sub":"1234567890","username":"John Doe","context_name":"1516239022"}`,
	)
	tkn := oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		RefreshToken: "LetMeInAgain",
		Expiry:       time.Now().Add(time.Minute * 30),
	}

	claim, err := ParseToken(&tkn)
	assert.Nil(err)
	assert.NotNil(claim)

	assert.Equal("John Doe", claim.Username)
	assert.Equal("1516239022", claim.OrgID)
	assert.Empty(claim.Permissions)
}

func TestGetToken_Valid_NotExpired(t *testing.T) {
	assert := assert.New(t)

	accessToken := generateJWTToken(
		`{"sub":"1234567890","username":"joe","context_name":"1516239022"}`,
	)
	expireTime := time.Now().Add(time.Minute * 30)

	serverAuth := configapi.GlobalServerAuth{
		Issuer:       "https://oidc.example.com",
		UserName:     "jdoe",
		AccessToken:  accessToken,
		IDToken:      "xxyyzz",
		RefreshToken: "sprite",
		Expiration:   v1.NewTime(expireTime),
		Type:         "client",
	}

	tok, err := GetToken(&serverAuth)
	assert.Nil(err)
	assert.NotNil(tok)
	assert.Equal(accessToken, tok.AccessToken)
	assert.Equal(expireTime, tok.Expiry)
}

func TestGetToken_Expired(t *testing.T) {
	assert := assert.New(t)

	accessToken := generateJWTToken(
		`{"sub":"1234567890","username":"joe","context_name":"1516239022"}`,
	)
	expireTime := time.Now().Add(-time.Minute * 30)

	serverAuth := configapi.GlobalServerAuth{
		Issuer:       "https://oidc.example.com",
		UserName:     "jdoe",
		AccessToken:  accessToken,
		IDToken:      "xxyyzz",
		RefreshToken: "sprite",
		Expiration:   v1.NewTime(expireTime),
		Type:         "client",
	}

	fakeHTTPClient := &fakes.FakeHTTPClient{}
	responseBody := io.NopCloser(bytes.NewReader([]byte(`{
		"id_token": "abc",
		"token_type": "Test",
		"expires_in": 86400,
		"scope": "Test",
		"access_token": "LetMeIn",
		"refresh_token": "LetMeInAgain"}`)))
	fakeHTTPClient.DoReturns(&http.Response{
		StatusCode: 200,
		Body:       responseBody,
	}, nil)
	httpRestClient = fakeHTTPClient

	tok, err := GetToken(&serverAuth)
	assert.Nil(err)
	assert.NotNil(tok)
	assert.Equal(tok.AccessToken, "LetMeIn")
	assert.Equal(tok.RefreshToken, "LetMeInAgain")
}
