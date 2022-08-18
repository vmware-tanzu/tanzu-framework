// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ClientTestSuite is the set of tests to run for the v1alpha1 Client.
type ClientTestSuite struct {
	suite.Suite
	ClientConfig     ClientConfig
	GlobalServer     Server
	ManagementServer Server
}

// SetupTest performs setup for each test.
func (suite *ClientTestSuite) SetupTest() {
	suite.GlobalServer = Server{
		Name: "GlobalServer",
		Type: GlobalServerType,
	}
	suite.ManagementServer = Server{
		Name: "ManagementServer",
		Type: ManagementClusterServerType,
	}
	suite.ClientConfig = ClientConfig{
		KnownServers: []*Server{
			&suite.GlobalServer,
			&suite.ManagementServer,
		},
		CurrentServer: suite.GlobalServer.Name,
		KnownContexts: []*Context{
			{
				Name: suite.GlobalServer.Name,
				Type: CtxTypeTMC,
			},
			{
				Name: suite.ManagementServer.Name,
				Type: CtxTypeK8s,
			},
		},
		CurrentContext: map[ContextType]string{
			CtxTypeTMC: suite.GlobalServer.Name,
			CtxTypeK8s: suite.ManagementServer.Name,
		},
	}
}

func (suite *ClientTestSuite) TestGetCurrentServer() {
	server, err := suite.ClientConfig.GetCurrentServer()
	suite.Nil(err)
	suite.Equal(server.Name, "GlobalServer")
}

func (suite *ClientTestSuite) TestGetCurrentServer_NotFound() {
	suite.ClientConfig.CurrentServer = "InvalidServer"
	_, err := suite.ClientConfig.GetCurrentServer()
	suite.Error(err)
	suite.Contains(err.Error(), "not found")
}

func (suite *ClientTestSuite) TestHasServer_GlobalServer() {
	ok := suite.ClientConfig.HasServer("GlobalServer")
	suite.True(ok)
}

func (suite *ClientTestSuite) TestHasServer_ManagementServer() {
	ok := suite.ClientConfig.HasServer("ManagementServer")
	suite.True(ok)
}

func (suite *ClientTestSuite) TestHasServer_NotFound() {
	ok := suite.ClientConfig.HasServer("TestServer")
	suite.False(ok)
}

func (suite *ClientTestSuite) TestGetContext_TMC() {
	c, err := suite.ClientConfig.GetContext(suite.GlobalServer.Name)
	suite.Nil(err)
	suite.Equal(c.Name, suite.GlobalServer.Name)
}

func (suite *ClientTestSuite) TestGetContext_K8s() {
	c, err := suite.ClientConfig.GetContext(suite.ManagementServer.Name)
	suite.Nil(err)
	suite.Equal(c.Name, suite.ManagementServer.Name)
}

func (suite *ClientTestSuite) TestGetContext_NotFound() {
	_, err := suite.ClientConfig.GetContext("TestServer")
	suite.Error(err)
	suite.Contains(err.Error(), "could not find context")
}

func (suite *ClientTestSuite) TestHasContext_GlobalServer() {
	ok := suite.ClientConfig.HasContext("GlobalServer")
	suite.True(ok)
}

func (suite *ClientTestSuite) TestHasContext_ManagementServer() {
	ok := suite.ClientConfig.HasContext("ManagementServer")
	suite.True(ok)
}

func (suite *ClientTestSuite) TestHasContext_NotFound() {
	ok := suite.ClientConfig.HasContext("TestServer")
	suite.False(ok)
}

func (suite *ClientTestSuite) TestGetCurrentContext_TMC() {
	c, err := suite.ClientConfig.GetCurrentContext(CtxTypeTMC)
	suite.Nil(err)
	suite.Equal(c.Name, suite.GlobalServer.Name)
}

func (suite *ClientTestSuite) TestGetCurrentContext_K8s() {
	c, err := suite.ClientConfig.GetCurrentContext(CtxTypeK8s)
	suite.Nil(err)
	suite.Equal(c.Name, suite.ManagementServer.Name)
}

func (suite *ClientTestSuite) TestGetCurrentContext_NotFound() {
	_, err := suite.ClientConfig.GetCurrentContext("test")
	suite.Error(err)
	suite.EqualError(err, "no current context set for type \"test\"")
}

func (suite *ClientTestSuite) TestSetCurrentContext_TMC() {
	delete(suite.ClientConfig.CurrentContext, CtxTypeTMC)
	err := suite.ClientConfig.SetCurrentContext(CtxTypeTMC, suite.GlobalServer.Name)
	suite.NoError(err)
	suite.Equal(suite.GlobalServer.Name, suite.ClientConfig.CurrentContext[CtxTypeTMC])
}

func (suite *ClientTestSuite) TestSetCurrentContext_K8s() {
	delete(suite.ClientConfig.CurrentContext, CtxTypeK8s)
	err := suite.ClientConfig.SetCurrentContext(CtxTypeK8s, suite.ManagementServer.Name)
	suite.NoError(err)
	suite.Equal(suite.ManagementServer.Name, suite.ClientConfig.CurrentContext[CtxTypeK8s])
}

func (suite *ClientTestSuite) TestIsGlobal_True() {
	suite.True(suite.GlobalServer.IsGlobal())
}

func (suite *ClientTestSuite) TestIsGlobal_False() {
	suite.False(suite.ManagementServer.IsGlobal())
}

func (suite *ClientTestSuite) TestIsManagementCluster_True() {
	suite.True(suite.ManagementServer.IsManagementCluster())
}

func (suite *ClientTestSuite) TestIsManagementCluster_False() {
	suite.False(suite.GlobalServer.IsManagementCluster())
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
