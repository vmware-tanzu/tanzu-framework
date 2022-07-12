# Support for ClusterClass-based Cluster Lifecycle Management

## Introduction

The ClusterClass functionality in Cluster API supports the server-side
management CAPI cluster topologies. By leveraging this functionality
not only will we be able to encapsulate common cluster configuration
scenarios at the server end in a reusable fashion, we will also be able
eliminate a significant fraction of client-side processing required to create a
workload cluster today.

Readers interested in finding out more about Cluster API ClusterClass
are welcomed to check out the Cluster API Book [section](https://cluster-api.sigs.k8s.io/tasks/experimental-features/cluster-class/index.html)
as well as additional documentation on [writing a clusterclass](https://cluster-api.sigs.k8s.io/tasks/experimental-features/cluster-class/write-clusterclass.html)

## Status

ClusterClass-based Cluster Lifecycle Management is not enable by default, but
gated by a CLI-side feature flag `global.package-based-lcm-beta`. The work to
enable it by default is being tracked in issue [2772](https://github.com/vmware-tanzu/tanzu-framework/issues/2772).

## Approach

To introduce ClusterClass in a backward-compatible fashion, the CLI will
continue to accept the legacy cluster configuratiion file (based on flat
key-value pairs) as well as a CAPI ClusterClass-based Cluster resource as
appropriate for the target infrastructure. The latter is sometimes referred to
as the 'CCluster.yaml'

Where possible, legacy configuration will be translated into a CCluster.yaml,
and users will be urged to use the new format where possible in the future.
See [this document](legacy-to-cc-variable-mapping.md) for more details on how
the legacy configuration values are being translated.

All the existing ytt-based client-side templates are retained for situations
where they still need to be used to generate the cluster configuration
entirely at the CLI end.

Since existing ytt overlays are retained verbatim, new files to support
ClusterClass are introduced in a non-overlapping fashion as separate
directories (name 'yttcc' and 'cconly') across all supported target
infrastructures.

*yttcc/* : contains overlays required to translate legacy config values into
CCluster.yaml. They will be included in an upcoming version of the CLI which
supports ClusterClass-based cluster creation.

*cconly/* : contains overlays required configure the ClusterClass and associated
templates themselves. They are included for development purposes for now, but
are not expected to be installed on the CLI host.  The reason for this is that
ClusterClass's and their associate templates are being assembled as Carvel
packages and deployed to the bootstrap cluster and management cluster through
these packages (to enable ClusterClass-based creation of the management cluster
and workload clusters respectively). One can assume that there will always be a
default set of ClusterClass and templates on each Management Cluster deployed.

Note: To enable the deployment of Carvel packageson the bootstrap cluster,
[kapp-controller](https://carvel.dev/kapp-controller/) is now installed into
the bootstrap cluster as well.

See [here](../../packages/README.md) for more information regarding defining
new packages in this repository.

## ClusterClass Based Management Cluster Creation

Just like workload cluster creation, management cluster creation has also been
updated to leverage ClusterClass. To enable this, the bootstrap cluster
will be deployed with the default clusterclass appropriate for the target
infrastructure, and updated to support Carvel package installation (so as to be
able to install the clusterclass package in the first place).

More details on the updated manage cluster creation workflow is available
[here](../tkgctl/management-cluster-using-clusterclass.md).
