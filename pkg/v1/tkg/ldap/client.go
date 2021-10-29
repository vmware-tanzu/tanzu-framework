// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ldap

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	ldapApi "github.com/go-ldap/ldap/v3"

	tkg_models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

const (
	testSuccessCode = 1
	testSuccessDesc = "The test has passed"
	testSkippedCode = 2
	testSkippedDesc = "The test is skipped"
)

var (
	resultSuccess = &tkg_models.LdapTestResult{
		Code: testSuccessCode,
		Desc: testSuccessDesc,
	}

	resultSkipped = &tkg_models.LdapTestResult{
		Code: testSkippedCode,
		Desc: testSkippedDesc,
	}
)

type client struct {
	params *tkg_models.LdapParams
	conn   *ldapApi.Conn
}

// New creates a new LDAP client
func New() Client {
	return &client{
		params: nil,
		conn:   nil,
	}
}

// LdapConnect verifies the reachability of the LDAP server
func (c *client) LdapConnect(params *tkg_models.LdapParams) (*tkg_models.LdapTestResult, error) {
	c.params = params
	var err error

	// Verify if we can connect to the server.
	// If we are unable to connect, we expect the error message
	// would contain meaningful information, be it a a URL issue
	// or a certificate one.
	ca := strings.TrimSpace(params.LdapRootCa)
	if len(ca) > 0 {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(ca))
		tlsConfig := &tls.Config{ //nolint:gosec
			RootCAs:            caCertPool,
			InsecureSkipVerify: false,
		}

		c.conn, err = ldapApi.DialURL(params.LdapURL, ldapApi.DialWithTLSConfig(tlsConfig))
	} else {
		c.conn, err = ldapApi.DialURL(params.LdapURL)
	}

	if err != nil {
		return nil, errors.Wrap(err, `unable to connect to `+params.LdapURL)
	}

	return resultSuccess, nil
}

// LdapBind verifies username and password is properly authenticated via bind
// Note that BindDn amd BindPassword are not must-haves in order to create
// a connection
func (c *client) LdapBind() (*tkg_models.LdapTestResult, error) {
	dn := c.params.LdapBindDn
	password := strings.TrimSpace(c.params.LdapBindPassword)
	if len(dn) > 0 && len(password) > 0 {
		err := c.conn.Bind(dn, password)
		if err != nil {
			return nil, errors.Wrap(err, `unable to connect to `+c.params.LdapURL)
		}
	} else {
		return resultSkipped, nil
	}

	return resultSuccess, nil
}

// LdapUserSearch verifies user search function is able to locate the user
func (c *client) LdapUserSearch() (*tkg_models.LdapTestResult, error) {
	if c.params.LdapTestUser == "" {
		return resultSkipped, nil
	}

	baseDn := strings.TrimSpace(c.params.LdapUserSearchBaseDn)
	filter := strings.TrimSpace(c.params.LdapUserSearchFilter)
	uid := strings.TrimSpace(c.params.LdapUserSearchUsername)

	attrs := []string{}
	if filter != "" && c.params.LdapTestUser != "" {
		filter = fmt.Sprintf("(&%s(CN=%s))", filter, ldapApi.EscapeFilter(c.params.LdapTestUser))
	} else if filter == "" && c.params.LdapTestUser != "" {
		filter = fmt.Sprintf("(CN=%s)", ldapApi.EscapeFilter(c.params.LdapTestUser))
	}

	if len(baseDn+filter+uid) > 0 { // perform this verification iff we have at least one input
		if len(uid) > 0 {
			attrs = []string{uid}
		}

		scope := ldapApi.ScopeBaseObject
		if c.params.LdapTestUser != "" {
			scope = ldapApi.ScopeWholeSubtree
		}

		searchReq := ldapApi.NewSearchRequest(
			baseDn,
			scope,
			0,
			0,
			0,
			false,
			filter,
			attrs,
			[]ldapApi.Control{})

		rs, err := c.conn.Search(searchReq)

		if err != nil {
			return nil, errors.Wrap(err, `user search test failed`)
		}

		if c.params.LdapTestUser != "" && len(rs.Entries) == 0 {
			return nil, errors.Errorf(`Unable to find user: '%s'`, c.params.LdapTestUser)
		}
	} else {
		return resultSkipped, nil
	}

	return resultSuccess, nil
}

// LdapGroupSearch verifies group search is able to locate group
func (c *client) LdapGroupSearch() (*tkg_models.LdapTestResult, error) {
	if c.params.LdapTestGroup == "" {
		return resultSkipped, nil
	}

	attrs := []string{}
	if len(c.params.LdapGroupSearchNameAttr) > 0 {
		attrs = append(attrs, c.params.LdapGroupSearchNameAttr)
	}

	if len(c.params.LdapGroupSearchUserAttr) > 0 {
		attrs = append(attrs, c.params.LdapGroupSearchUserAttr)
	}

	if len(c.params.LdapGroupSearchGroupAttr) > 0 {
		attrs = append(attrs, c.params.LdapGroupSearchGroupAttr)
	}

	if len(c.params.LdapGroupSearchBaseDn+
		c.params.LdapGroupSearchFilter+
		c.params.LdapGroupSearchNameAttr+
		c.params.LdapGroupSearchUserAttr+
		c.params.LdapGroupSearchGroupAttr) > 0 {
		filter := strings.TrimSpace(c.params.LdapGroupSearchFilter)

		if filter != "" && c.params.LdapTestGroup != "" {
			filter = fmt.Sprintf("(&%s(OU=%s))", filter, ldapApi.EscapeFilter(c.params.LdapTestGroup))
		} else if filter == "" && c.params.LdapTestGroup != "" {
			filter = fmt.Sprintf("(OU=%s)", ldapApi.EscapeFilter(c.params.LdapTestGroup))
		}

		scope := ldapApi.ScopeSingleLevel
		if c.params.LdapTestGroup != "" {
			scope = ldapApi.ScopeWholeSubtree
		}

		searchReq := ldapApi.NewSearchRequest(c.params.LdapGroupSearchBaseDn,
			scope,
			0,
			0,
			0,
			false,
			filter,
			attrs,
			[]ldapApi.Control{})

		rs, err := c.conn.Search(searchReq)

		if err != nil {
			return nil, errors.Wrap(err, `group search test failed`)
		}

		if c.params.LdapTestGroup != "" && len(rs.Entries) == 0 {
			return nil, errors.Errorf(`Unable to find group: '%s'`, c.params.LdapTestGroup)
		}
	} else {
		return resultSkipped, nil
	}

	return resultSuccess, nil
}

// LdapCloseConnection closes the LDAP connection
func (c *client) LdapCloseConnection() {
	if c.conn != nil {
		c.conn.Close()
	}
}
