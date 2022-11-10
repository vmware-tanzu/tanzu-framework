// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pluginmanager

import (
	"fmt"
	"strings"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func defaultDiscoverySourceBasedOnServer(server *configapi.Server) []configapi.PluginDiscovery {
	var defaultDiscoveries []configapi.PluginDiscovery
	// If current server type is management-cluster, then add
	// the default kubernetes discovery endpoint pointing to the
	// management-cluster kubeconfig
	if server.Type == configapi.ManagementClusterServerType && server.ManagementClusterOpts != nil {
		defaultDiscoveries = append(defaultDiscoveries, defaultDiscoverySourceForK8sContextType(server.Name, server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context))
	}
	return defaultDiscoveries
}

func defaultDiscoverySourceBasedOnContext(context *configapi.Context) []configapi.PluginDiscovery {
	var defaultDiscoveries []configapi.PluginDiscovery

	// If current context is of type k8s, then add the default
	// kubernetes discovery endpoint pointing to the cluster kubeconfig
	// If the current context is of type tmc, then add the default REST endpoint
	// for the tmc discovery service
	if context.Type == configapi.CtxTypeK8s && context.ClusterOpts != nil {
		defaultDiscoveries = append(defaultDiscoveries, defaultDiscoverySourceForK8sContextType(context.Name, context.ClusterOpts.Path, context.ClusterOpts.Context))
	} else if context.Type == configapi.CtxTypeTMC && context.GlobalOpts != nil {
		defaultDiscoveries = append(defaultDiscoveries, defaultDiscoverySourceForTMCContextType(context))
	}
	return defaultDiscoveries
}

func defaultDiscoverySourceForK8sContextType(name, kubeconfig, context string) configapi.PluginDiscovery {
	return configapi.PluginDiscovery{
		Kubernetes: &configapi.KubernetesDiscovery{
			Name:    fmt.Sprintf("default-%s", name),
			Path:    kubeconfig,
			Context: context,
		},
	}
}

func defaultDiscoverySourceForTMCContextType(context *configapi.Context) configapi.PluginDiscovery {
	return configapi.PluginDiscovery{
		REST: &configapi.GenericRESTDiscovery{
			Name:     fmt.Sprintf("default-%s", context.Name),
			Endpoint: appendURLScheme(context.GlobalOpts.Endpoint),
			BasePath: "v1alpha1/system/binaries/plugins",
		},
	}
}

func appendURLScheme(endpoint string) string {
	e := strings.Split(endpoint, ":")[0]
	if !strings.Contains(e, "https") {
		return fmt.Sprintf("https://%s", e)
	}
	return e
}
