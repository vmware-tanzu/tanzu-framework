# Tips when developing ClusterClassc for the Tanzu CLI

## Deploying new CC/templates

Any new set of ClusterClass definition, along with their associated templates can
be directly applied to the management cluster.

## Updating existing CC/templates

Any CC (and templates associated with it) is immutable if it is being
referenced by an existing Cluster object. To make modifications to them, said
reference has to be removed (e.g. by deleting the Cluster)

## When debugging/modifying CCs deployed via Carvel packages

Further complications arise for CC/templates deployed through a package, as
any modifications successfully done to them can subsequently be undone when
packages reconciles the CC/templates back to its 'expected' state.
To prevent that from happening, ensure that the "paused: true" field is set on
the .spec of the appropriate PackageInstall resources. For instance, in the
case of the AWS infra provider, this means updating the following
PackageInstall resources: tkg, tkg-clusterclass, tkg-clusterclass-aws

## Others

(more to be added)
