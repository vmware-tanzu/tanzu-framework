// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kubernetes contains utilities for interacting with kubernetes clusters and unit testing code that
// deals with clusters
package kubernetes

const (
	NamespacesURI = "/api/v1/namespaces"
	ConfigMapsURI = "/api/v1/namespaces/vmware-system-telemetry/configmaps"

	TelemetryNamespace    = "vmware-system-telemetry"
	TelemetryNamespaceURI = "/api/v1/namespaces/vmware-system-telemetry"

	CeipConfigMapName = "vmware-telemetry-cluster-ceip"
	CeipConfigMapURI  = "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-cluster-ceip"

	SharedIdsConfigMapName = "vmware-telemetry-identifiers"
	SharedIdsConfigMapURI  = "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-identifiers"
)
