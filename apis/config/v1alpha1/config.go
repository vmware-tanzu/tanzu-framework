// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import "fmt"

// IsGlobal tells if the server is global.
func (s *Server) IsGlobal() bool {
	return s.Type == GlobalServerType
}

// IsManagementCluster tells if the server is a management cluster.
func (s *Server) IsManagementCluster() bool {
	return s.Type == ManagementClusterServerType
}

// GetCurrentServer returns the current server/
func (c *Config) GetCurrentServer() (*Server, error) {
	for _, server := range c.KnownServers {
		if server.Name == c.CurrentServer {
			return server, nil
		}
	}
	return nil, fmt.Errorf("current server %q not found", c.CurrentServer)
}
