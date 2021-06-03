// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"

	ldapClient "github.com/vmware-tanzu-private/core/pkg/v1/tkg/ldap"
	ldap "github.com/vmware-tanzu-private/core/pkg/v1/tkg/web/server/restapi/operations/ldap"
)

// VerifyLdapConnect checks LDAP server reachability
func (app *App) VerifyLdapConnect(params ldap.VerifyLdapConnectParams) middleware.Responder {
	app.ldapClient = ldapClient.New()
	success, err := app.ldapClient.LdapConnect(params.Credentials)

	if err != nil {
		return ldap.NewVerifyLdapConnectBadRequest().WithPayload(Err(errors.Wrap(err, "unable to connect to LDAP server as configed")))
	}

	return ldap.NewVerifyLdapConnectOK().WithPayload(success)
}

// VerifyLdapBind verifies LDAP authentication
func (app *App) VerifyLdapBind(params ldap.VerifyLdapBindParams) middleware.Responder {
	if app.ldapClient == nil {
		return ldap.NewVerifyLdapBindInternalServerError().WithPayload(Err(errors.New("LDAP client is not initialized properly")))
	}

	result, err := app.ldapClient.LdapBind()
	if err != nil {
		return ldap.NewVerifyLdapBindBadRequest().WithPayload(Err(errors.Wrap(err, "unable to perform LDAP bind")))
	}

	return ldap.NewVerifyLdapBindOK().WithPayload(result)
}

// VerifyUserSearch verifies LDAP user search capability
func (app *App) VerifyUserSearch(params ldap.VerifyLdapUserSearchParams) middleware.Responder {
	if app.ldapClient == nil {
		return ldap.NewVerifyLdapUserSearchInternalServerError().WithPayload(Err(errors.New("LDAP client is not initialized properly")))
	}

	success, err := app.ldapClient.LdapUserSearch()
	if err != nil {
		return ldap.NewVerifyLdapUserSearchBadRequest().WithPayload(Err(errors.Wrap(err, "unable to perform LDAP User Search")))
	}

	return ldap.NewVerifyLdapUserSearchOK().WithPayload(success)
}

// VerifyGroupSearch verifies LDAP group search capability
func (app *App) VerifyGroupSearch(params ldap.VerifyLdapGroupSearchParams) middleware.Responder {
	if app.ldapClient == nil {
		return ldap.NewVerifyLdapGroupSearchInternalServerError().WithPayload(Err(errors.New("LDAP client is not initialized properly")))
	}

	success, err := app.ldapClient.LdapGroupSearch()
	if err != nil {
		return ldap.NewVerifyLdapGroupSearchBadRequest().WithPayload(Err(errors.Wrap(err, "unable to perform LDAP Group Search")))
	}

	return ldap.NewVerifyLdapGroupSearchOK().WithPayload(success)
}

// VerifyLdapCloseConnection disconnect from a LDAP server
func (app *App) VerifyLdapCloseConnection(params ldap.VerifyLdapCloseConnectionParams) middleware.Responder {
	if app.ldapClient == nil {
		return ldap.NewVerifyLdapCloseConnectionInternalServerError().WithPayload(Err(errors.New("LDAP client is not initialized properly")))
	}

	app.ldapClient.LdapCloseConnection()

	return ldap.NewVerifyLdapCloseConnectionCreated()
}
