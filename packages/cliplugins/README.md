# cliplugins Package

This package provides the `CLIPlugin` CRD that defines the CLIPlugin resource which is used to define CLIPlugins used by the Tanzu CLI when talking to a management cluster. This is a required package that needs to be installed on each management cluster so that the Tanzu CLI can understand the recommended set of plugins to use when talking to the cluster.

## Components

* CLIPlugin CRD

## Details

To learn more about the CLIPlugins and how it is getting used with Tanzu CLI, refer to this [doc](../../docs/design/context-aware-plugin-discovery-design.md)
