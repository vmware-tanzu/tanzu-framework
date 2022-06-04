# Support for ClusterClass-based LCM

## Intro

The ClusterClass functionality in Cluster API supports the server-side
management CAPI cluster topologies. By leveraging this functionality
not only will we be able to encapsulate common cluster configuration
scenarios at the server end in a reusable fashion, we will also be able
eliminate a significant fraction of client-side processing required to create a
workload cluster today.

## Approach

To introduce ClusterClass in a backward-compatible fashion, all the existing
ytt-base client-side templates are retained. The CLI will accept the legacy
cluster configuratiion file (based on flat key-value pairs) as well as a CAPI
ClusterClass-based Cluster resource as appropriate for the target
infrastructure. The latter is sometimes referred to as the 'CCluster.yaml'

Where possible, legacy configuration will be translated into a CCluster.yaml,
and users will be urged to use the new format where possible in the future.

Existing ytt overlays are retained. New files to support ClusterClass
are introduced in a non-overlapping fashion as separate directories (name
'yttcc' and 'cconly') across all supported target infrastructures.

yttcc/ : contains overlays required to translate legacy config values into
CCluster.yaml. They will be included in an upcoming version of the CLI which
supports ClusterClass-based cluster creation.

cconly/ : contains overlays required configure the ClusterClass and associated
templates themselves. They are included for development purposes for now, and
will eventually not be installed on the CLI host.

The reason for the above is that ClusterClass's and their associate templates
are being assembled as Carvel packages and deployed to the bootstrap cluster
and management cluster through these packages (to enable ClusterClass-based
creation of the management cluster and workload clusters respectively). One can
assume that there will always be a default set of ClusterClass and templates
on each Management Cluster deployed.

Note: To enable the deployment of ClusterClass package on the bootstrap cluster,
kapp-controller is now installed into the bootstrap cluster as well.
