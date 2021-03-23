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
	ClientConfig     Config
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
	suite.ClientConfig = Config{
		KnownServers: []*Server{
			&suite.GlobalServer,
			&suite.ManagementServer,
		},
		CurrentServer: "GlobalServer",
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
