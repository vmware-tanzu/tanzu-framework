// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

/*
Kubernetes management cluster operations.

# Get or set ceip participation

Usage:

	tanzu management-cluster ceip-participation [command]

Available Commands:

	get         Get the current CEIP opt-in status of the current management cluster
	set         Set the opt-in preference for CEIP of the current management cluster

Flags:

	-h, --help   help for ceip-participation

Global Flags:

	--log-file string   	  Log file path
	-v, --verbose int32     Number for the log level verbosity(0-9)

Use "management-cluster ceip-participation [command] --help" for more information about a command.

# Create a Tanzu Kubernetes Grid management cluster

Includes initializing it with Cluster API components appropriate for the target infrastructure.

Usage:

	tanzu management-cluster create [flags]

Examples:

	# Create a management cluster on AWS infrastructure, initializing it with
	# components required to create workload clusters through it on the same infrastructure
	# by bootstrapping through a self-provisioned bootstrap cluster.
	tanzu management-cluster create --file ~/clusterconfigs/aws-mc-1.yaml
	# Launch an interactive UI to configure the settings necessary to create a
	# management cluster
	tanzu management-cluster create --ui
	# Create a management cluster on vSphere infrastructure by using an existing
	# bootstrapper cluster. The current kube context should point to that
	# of the existing bootstrap cluster.
	tanzu management-cluster create --use-existing-bootstrap-cluster --file vsphere-mc-1.yaml

Flags:

	-b, --bind string                      Specify the IP and port to bind the Kickstart UI against (e.g. 127.0.0.1:8080). (default "127.0.0.1:8080")
	    --browser string                   Specify the browser to open the Kickstart UI on. Use 'none' for no browser. Defaults to OS default browser. Supported: ['chrome', 'firefox', 'safari', 'ie', 'edge', 'none']
	-f, --file string                      Configuration file from which to create a management cluster
	-h, --help                             help for create
	-t, --timeout duration                 Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s) (default 30m0s)
	-u, --ui                               Launch interactive management cluster provisioning UI
	-e, --use-existing-bootstrap-cluster   Use an existing bootstrap cluster to deploy the management cluster
	-y, --yes                              Create management cluster without asking for confirmation

Global Flags:

	      --log-file string   Log file path
	-v, --verbose int32     Number for the log level verbosity(0-9)

# Update Credentials for Management Cluster

Usage:

	tanzu management-cluster credentials [command]

Available Commands:

	update      Update credentials for management cluster

Flags:

	-h, --help   help for credentials

Global Flags:

	--log-file string   	  Log file path
	-v, --verbose int32     Number for the log level verbosity(0-9)

Use "management-cluster credentials [command] --help" for more information about a command.

# Delete a management cluster and tears down the underlying infrastructure

Usage:

	tanzu management-cluster delete [flags]

Examples:

	# Deletes the management cluster of the current server
	tanzu management-cluster delete

Flags:

	--force                              Force deletion of the management cluster even if it is managing active Tanzu Kubernetes clusters
	-h, --help                           help for delete
	-t, --timeout duration               Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s) (default 30m0s)
	-e, --use-existing-cleanup-cluster   Use an existing cleanup cluster to delete the management cluster
	-y, --yes                            Delete management cluster without asking for confirmation

Global Flags:

	--log-file string       Log file path
	-v, --verbose int32     Number for the log level verbosity(0-9)

Retrieves details about the current management cluster. Requires the current server to be a management cluster

Usage:

	tanzu management-cluster get [flags]

Flags:

	-h, --help                         help for get
	    --show-all-conditions string   List of comma separated kind or kind/name for which we should show all the object's conditions (all to show conditions for all the objects)
	    --show-details                 Show details of MachineInfrastructure and BootstrapConfig when ready condition is true or it has the Status, Severity and Reason of the machine's object
	    --show-group-members           Expand machine groups whose ready condition has the same Status, Severity and Reason

Global Flags:

	--log-file string   	  Log file path
	-v, --verbose int32     Number for the log level verbosity(0-9)

# Import Tanzu Kubernetes Grid management cluster from TKG settings file

Usage:

	tanzu management-cluster import [flags]

Examples:

	# Import management cluster config from default config file
	tanzu management-cluster import

	# Import management cluster config from custom config file
	tanzu management-cluster import -f path/to/configfile.yaml

Flags:

	-f, --file string   TKG settings file (default '$HOME/.tkg/config.yaml')
	-h, --help          help for import

Global Flags:

	--log-file string       Log file path
	-v, --verbose int32     Number for the log level verbosity(0-9)

# Kubeconfig of management cluster

Usage:

	tanzu management-cluster kubeconfig [command]

Available Commands:

	get         Get Kubeconfig of a management cluster

Flags:

	-h, --help   help for kubeconfig

Global Flags:

	--log-file string       Log file path
	-v, --verbose int32     Number for the log level verbosity(0-9)

Use "management-cluster kubeconfig [command] --help" for more information about a command.

# Configure permissions on cloud providers

Usage:

	tanzu management-cluster permissions [command]

Available Commands:

	aws         Configure permissions on AWS

Flags:

	-h, --help   help for permissions

Global Flags:

	--log-file string   	  Log file path
	-v, --verbose int32     Number for the log level verbosity(0-9)

Use "management-cluster permissions [command] --help" for more information about a command.

# Upgrades the management cluster

Usage:

	tanzu management-cluster upgrade [flags]

Flags:

	-h, --help                              help for upgrade
	-t, --timeout duration                  Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s) (default 30m0s)
	    --vsphere-vm-template-name string   The vSphere VM template to be used with upgraded kubernetes version. Discovered automatically if not provided
	-y, --yes                               Upgrade management cluster without asking for confirmation

Global Flags:

	--log-file string       Log file path
	-v, --verbose int32     Number for the log level verbosity(0-9)
*/
package main
