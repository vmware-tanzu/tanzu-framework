// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package ldap provides ldap configuration verification
package ldap

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// Client is a LDAP client interface
type Client interface {
	LdapConnect(params *models.LdapParams) (*models.LdapTestResult, error)
	LdapBind() (*models.LdapTestResult, error)
	LdapUserSearch() (*models.LdapTestResult, error)
	LdapGroupSearch() (*models.LdapTestResult, error)
	LdapCloseConnection()
}
