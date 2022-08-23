// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package catalog implements catalog management functions
package catalog

import (
	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
)

// Catalog is the interface that maintains an index of the installed plugins as well as the active plugins.
type Catalog interface {
	// Upsert inserts/updates the given plugin.
	Upsert(plugin cliv1alpha1.PluginDescriptor)

	// Get looks up the descriptor of a plugin given its name.
	Get(pluginName string) (cliv1alpha1.PluginDescriptor, bool)

	// List returns the list of active plugins.
	// Active plugin means the plugin that are available to the user
	// based on the current logged-in server.
	List() []cliv1alpha1.PluginDescriptor

	// Delete deletes the given plugin from the catalog, but it does not delete the installation.
	Delete(plugin string)
}
