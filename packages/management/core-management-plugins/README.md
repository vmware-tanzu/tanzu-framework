# core-management-plugins Package

This package provides the `CLIPlugin` CRD and default resources needed for the Tanzu CLI to understand `CLIPlugin` resources. This is a required package that needs to be installed on each management cluster so that the Tanzu CLI can understand the recommended set of plugins to use when talking to the cluster.

## Components

* CLIPlugin CRD
* Required CLIPlugin resources for `cluster`, `kubernetes-release` plugins, etc.

## Details

To learn more about the CLIPlugins and how it is getting used with Tanzu CLI, refer to this [doc](../../docs/design/context-aware-plugin-discovery-design.md)
