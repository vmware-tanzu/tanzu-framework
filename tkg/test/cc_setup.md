# Deploy management and workload cluster on AWS using the ClusterClass mechanism

- Build the CLI
  - make build-install-cli-local

- Export below environment variables: (This is needed until TKR source controller is able to deploy TKR related resources on the cluster.)

```bash
export _MANAGEMENT_PACKAGE_REPO_IMAGE=gcr.io/eminent-nation-87317/tanzu_framework/github-actions/main/management:v0.21.0
export _MANAGEMENT_PACKAGE_VERSION=0.21.0
```

- Make sure to enable the feature-flag for `package-based-cc`

```bash
tanzu config set features.management-cluster.package-based-cc true
```

By default, it's true. You can check it use `tanzu config get`

- Create aws management-cluster using existing VPC.

```bash
tanzu management-cluster create --ui
```

- Created management-cluster will have all the ClusterClass and latest components deployed on the management-cluster. So, you can use this management-cluster to create ClusterClass based workload cluster

```bash
tanzu cluster create --file <config-file.yaml>
```
