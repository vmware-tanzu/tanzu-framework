// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

/*
Manage cluster lifecycle operations.

# Create a cluster

Usage:

	tanzu cluster create CLUSTER_NAME [flags]

Flags:

	-d, --dry-run       Does not create cluster but show the deployment YAML instead
	-f, --file string   Cluster configuration file from which to create a Cluster
	-h, --help          help for create
	    --tkr string    TanzuKubernetesRelease(TKr) to be used for creating the workload cluster

# List clusters

Usage:

	tanzu cluster list [flags]

Flags:

	-h, --help                         help for list
	    --include-management-cluster   Show active management cluster information as well
	-n, --namespace string             The namespace from which to list workload clusters. If not provided clusters from all namespaces will be returned
	-o, --output string                Output format. Supported formats: json|yaml

# Delete a cluster

Usage:

	tanzu cluster delete CLUSTER_NAME [flags]

Flags:

	-h, --help               help for delete
	-n, --namespace string   The namespace where the workload cluster was created. Assumes 'default' if not specified.
	-y, --yes                Delete workload cluster without asking for confirmation

# Scale a cluster

Usage:

	tanzu cluster scale CLUSTER_NAME [flags]

Flags:

	-c, --controlplane-machine-count int32   The number of control plane nodes to scale to. Assumes unchanged if not specified
	-h, --help                               help for scale
	-n, --namespace string                   The namespace where the workload cluster was created. Assumes 'default' if not specified.
	-w, --worker-machine-count int32         The number of worker nodes to scale to. Assumes unchanged if not specified

# Upgrade a cluster

Usage:

	tanzu cluster upgrade CLUSTER_NAME [flags]

Flags:

	-h, --help                        help for upgrade
	-n, --namespace string            The namespace where the workload cluster was created. Assumes 'default' if not specified
	-t, --timeout duration            Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s) (default 30m0s)
	    --tkr string                  TanzuKubernetesRelease(TKr) to upgrade to
	-y, --yes                         Upgrade workload cluster without asking for confirmation

# Get,set, or delete a MachineHealthCheck object for a Tanzu Kubernetes cluster

Usage:

	tanzu cluster machinehealthcheck [command]

Available Commands:

	delete      Delete a MachineHealthCheck object of a cluster
	get         Get a MachineHealthCheck object of a cluster
	set         Create or update a MachineHealthCheck for a cluster

Flags:

	-h, --help   help for machinehealthcheck

Global Flags:

	    --log-file string   Log file path
	-v, --verbose int32     Number for the log level verbosity(0-9)

# Get a MachineHealthCheck object for the given cluster

Usage:

	tanzu cluster machinehealthcheck get CLUSTER_NAME [flags]

Flags:

	-h, --help               help for get
	-m, --mhc-name string    Name of the MachineHealthCheck object
	-n, --namespace string   The namespace where the MachineHealthCheck object was created.

# Create or update a MachineHealthCheck object for a cluster

Usage:

	tanzu cluster machinehealthcheck set CLUSTER_NAME [flags]

Flags:

	  -h, --help                      	  help for set
		  --match-labels string           Label selector to match machines whose health will be exercised
	  -m, --mhc-name string               Name of the MachineHealthCheck object
	  -n, --namespace string              Namespace of the cluster
		  --node-startup-timeout string   Any machine being created that takes longer than this duration to join the cluster is considered to have failed and will be remediated
		  --unhealthy-conditions string   A list of the conditions that determine whether a node is considered unhealthy. Available condition types: [Ready, MemoryPressure,DiskPressure,PIDPressure, NetworkUnavailable], Available condition status: [True, False, Unknown]heck object was created.

# Delete a MachineHealthCheck object for the given cluster

Usage:

	tanzu cluster machinehealthcheck delete CLUSTER_NAME [flags]

Flags:

	-h, --help               help for delete
	-m, --mhc-name string    Name of the MachineHealthCheck object
	-n, --namespace string   The namespace where the MachineHealthCheck object was created, default to the cluster's namespace
	-y, --yes                Delete the MachineHealthCheck object without asking for confirmation

# Update credentials for a cluster

Usage:

	tanzu cluster credentials [command]

Available Commands:

	update      Update credentials for a cluster

Flags:

	-h, --help   help for credentials

Use "cluster credentials [command] --help" for more information about a command.

# Update credentials for a cluster

Usage:

	tanzu cluster credentials update CLUSTER_NAME [flags]

Flags:

	-h, --help                      help for update
	-n, --namespace string          The namespace of cluster whose credentials have to be updated
	    --vsphere-password string   Password for vSphere provider
	    --vsphere-user string       Username for vSphere provider

# Getting clusters details

Usage:

	tanzu cluster get CLUSTER_NAME [flags]

Flags:

	-h, --help                         help for get
	-n, --namespace string             The namespace from which to get workload clusters. If not provided clusters from all namespaces will be returned
	    --show-all-conditions string   List of comma separated kind or kind/name for which we should show all the object's conditions (all to show conditions for all the objects)
	    --show-details                 Show details of MachineInfrastructure and BootstrapConfig when ready condition is true or it has the Status, Severity and Reason of the machine's object
	    --show-group-members           Expand machine groups whose ready condition has the same Status, Severity and Reason

# Get kubeconfig of a cluster and merge the context into the default kubeconfig file

Usage:

	tanzu cluster kubeconfig get CLUSTER_NAME [flags]

Examples:

	# Get workload cluster kubeconfig
	tanzu cluster kubeconfig get CLUSTER_NAME

	# Get workload cluster admin kubeconfig
	tanzu cluster kubeconfig get CLUSTER_NAME --admin

Flags:

	    --admin                Get admin kubeconfig of the workload cluster
	    --export-file string   File path to export a standalone kubeconfig for workload cluster
	-h, --help                 help for get
	-n, --namespace string     The namespace where the workload cluster was created. Assumes 'default' if not specified.
*/
package main
