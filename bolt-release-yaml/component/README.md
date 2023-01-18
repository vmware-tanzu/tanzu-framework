# Component

This directory holds all the configurations of every component used by BOLT.  

## What is a Component?

Any project that needs to be built as part of a release is considered a component. Each component is defined by a version specific config YAML, a `publish.yaml` and a `url.yaml`.  

BOLT supports the following types of components:

* Managed
* Unmanaged
* Image
* BOM
* Aggregator

## Types of Components

### Managed Component

**What is a Managed Component?**  
A Managed Component is a component that meets the following criteria:

* Should have a cayman repo in the `core-build` namespace
* A fork of the cayman repo is in the `k8s-common-core` namespace  
* Cayman repo should use a python based builder
* Should have a mirror repo in `core-build` namespace that mirrors an upstream project.
* A fork of the mirror repo is in the `k8s-common-core` namespace
* The cayman project should have a single git submodule pointing to the mirror repo.  
* The cayman project in `core-build` should have the build logic defined in a branch. This branch will be referred to as _base branch_.
* The cayman build target should be registered at https://bmps.eng.vmware.com/target/

**How is a Managed Component built?**  
A Managed Component build represents a downstream build of an upstream project.  
BOLT perform the following steps when building a Managed Component:

* Create a branch on the mirror repo based on a given upstream tag or upstream commit.
* Create a branch on the cayman repo out of the _base branch_.
* Update the git submodule in cayman repo to point to the newly created branch in the mirror repo
* Trigger a buildweb build for the cayman project
* Perform [publish actions](README.md#publishyaml) on the build outputs.

Note: Repos in the `core-build` namespace are used for official builds and repos in `k8s-common-core` namespace are used for sandbox builds.

**How to define a Managed Component?**  
A Managed Component is defined using 3 configuration YAMLs:

* `<version>.yaml`  
  The `<version.yaml>` file defines the configuration to build a specific version of the managed component.
  An example YAML configuration:

```yaml
name: kubernetes
managedComponents:
  - version: v1.20.1+vmware.2
    upstreamTag: v1.20.1
    mirrorCherryPicks:
      - sha: 4855dcf19390129e79b83c932c4be7e0c55fe8aa
        description: 'this commit changes the base image registry to point to vmware internal
                      artifactory: kscom-docker-local.artifactory.eng.vmware.com'
      - sha: 024275a23fa6b677b921d9659e02c1394db1846d
        description: fixes to 'docker save' to work in cayman env
      - sha: bf398923b181c5a147877d58ef9dd174e78d0894
        description: set pause version to 3.2 in makefile
      - sha: b05c74892c3e8770ce98f27edcbae3ebb0c86d61
        description: 'change build conformance image to pull debian image from harbor cache registry'
    gobuildBaseBranch: vmware-1.20.0+vmware.0~7823ebe232ad5b36b8d975a129590951616bb069
    dependencies:
      gobuild.golangVersion: 1.15.5
      gobuild.containerNamePrefix: projects.registry.vmware.com/tkg
      gobuild.debian_base_image: debian-base-amd64:buster-v1.2.0
      gobuild.debian_iptable_image: debian-iptables-amd64:buster-v1.3.0
      gobuild.linux_hosttype: linux-centos8
      gobuild.pause: 3.2
      generateDebRpmChangelog: true
      kubeadmAPIVersion: kubeadm.k8s.io/v1beta2
    componentReferences:
      - name: cni_plugins
        version:
          - v0.8.7+vmware.4
      - name: coredns
        version:
          - v1.7.0+vmware.7
      - name: cri_tools
        version:
          - v1.19.0+vmware.1
      - name: etcd
        version:
          - v3.4.13+vmware.6
      - name: containerd
        version:
          - v1.4.3+vmware.1
```

YAML Explained:

* `name` - Is the name of the component. This should match with the name in the config files and the name of the component directory.
* `managedComponents` - Holds the configuration for a Managed Component type component.
* `version` - Is the version of the component. It is suggested to prefix version with `v`.
* `upstreamTag`/`upstreamCommit` - Mutually exclusive fields that determine the upstream version.
* `mirrorCherryPicks` - The list of cherry-picks to apply on the target branch of the mirror repo. These cherry-picks are commit available on the downstream mirror.  
  Note: merge commits are NOT allowed as cherry-picks.
* `gobuildBaseBranch` - The _base branch_ in the cayman repo and the commit sha of the base branch, with `~` as separator
* `dependencies` - A list of key-value pair that can hold additional build information.
  * All keys that start with `gobuild.` will be added to the dependency.config file in the cayman repo. Example: `gobuild.foo: bar` in dependencies will be stored as `foo='bar'` in dependency.config.
  * `gobuild.linux_hosttype` is a mandatory key if the component is a descendant of an _Aggregator_ type component. The host type used in the gobuild environment.
* `componentReferences` - List of components that are children of the given component.  
  * `name` - Name of the child component
  * `version` - List of versions to use as child components  

**When to use a Managed Component?**  
A Managed component should be used when the component can meet all the criteria mention above. It is advised to use a manged component if the component needs to be updated frequently to pickup new upstream versions.

**Requirement on Managed component cayman repository**  
[Managed component cayman repo checklist](https://confluence.eng.vmware.com/display/TKG/Caymanization+checklist)

### Unmanaged Component

**What is an Unmanaged Component?**  
An Unmanaged Component can be any component that is a cayman project that can be built using buildweb.  
The cayman project should be a repo in the `core-build` namespace.  
The cayman build target should be registered at https://bmps.eng.vmware.com/target/  

**How is an Unmanaged Component built?**  
BOLT perform the following steps when building an Unmanaged Component:

* Trigger a buildweb build for the cayman project using the specified `gobuildBranch`
* Perform [publish actions](README.md#publishyaml) on the build outputs.

**How to define an Unmanaged Component?**  
An unmanaged component is defined in a `<version>.yaml`.  
The `<version>.yaml` file defines the configuration to build a specific version of an unmanaged component.  
An example YAML configuration:  

```yaml
name: containerd
unmanagedComponents:
  - version: v1.4.3+vmware.1
    gobuildBranch: vmware-1.4.3~7823ebe232ad5b36b8d975a129590951616bb069
    dependencies:
      gobuild.golangVersion: 1.15.5
      gobuild.linux_hosttype: linux-centos72-gc32
```

YAML Explained:

* `name` - Is the name of the component. This should match with the name in the config files and the name of the component directory.
* `unmanagedComponents` - Holds the configuration for an Unmanaged Component type component.
* `version` - Is the version of the component. It is suggested to prefix version with `v`.
* `gobuildBranch` - The cayman repo branch and the commit sha with `~` as separator.  
* `dependencies` - A list of key-value pair that can hold additional build information.
  * `gobuild.linux_hosttype` is a mandatory key if the component is a descendant of an _Aggregator_ type component. The host type used in the gobuild environment.
* `componentReferences` - List of components that are children of the given component.
  * `name` - Name of the child component
  * `version` - List of versions to use as child components  

**When to use an Unmanaged Component?**  
An Unmanaged component can be used to build any cayman project. It is the preferred solution then a component cannot meet the requirements of a Managed Component.

### Image

**What is an Image Component?**  
An Image component can be used to build a VM image using [Image Builder](https://github.com/kubernetes-sigs/image-builder).  

**How is an Image Component built?**  
An Image component builds a VM image by triggering a Jenkins pipeline. The Jenkins pipeline is expected to use Image Builder to build the VM image.

* Build a VM image by triggering a Jenkins pipeline. The Jenkins pipeline is used to build the VM image.
* Perform [publish actions](README.md#publishyaml) on the build outputs.

**How to define an Image Component?**  
An image component is defined in a `<version>.yaml`.  
The `<version>.yaml` file defines the configuration to build a specific version of an image component.  
An example YAML configuration:

```yaml
name: ova-ubuntu-2004
images:
  - version: v1.20.1+vmware.2-tkg.1.3-3161359046336510236
    osinfo:
      name: ubuntu
      version: 2004
      arch: amd64
    type: ova
    imageBuilderVersion: 675b064470578a38974bd52d30df5ec431da8474
    componentReferences:
      - name: coredns
        version:
        - v1.7.0+vmware.7
      - name: kubernetes
        version:
        - v1.20.1+vmware.2
      - name: etcd
        version:
        - v3.4.13+vmware.6
      - name: cri_tools
        version:
        - v1.19.0+vmware.1
      - name: cni_plugins
        version:
        - v0.8.7+vmware.4
      - name: containerd
        version:
        - v1.4.3+vmware.1
    dependencies:
      base_iso_url: https://build-artifactory.eng.vmware.com/kscom-generic-local/isos/ubuntu-20.04-server-amd64.iso
      base_iso_sha: 36f15879bd9dfd061cd588620a164a82972663fdd148cce1f70d57d314c21b73
```

YAML Explained:

* `name` - Is the name of the component. This should match with the name in the config files and the name of the component directory.
* `images` - Holds the configuration for an Image type component.
* `version` - Is the version of the component. It is suggested to prefix version with `v`.
* `osinfo` - Holds the OS information used to build the VM image.
* `type` - Type of VM image. Supported values:
  * `ova`
  * `ami`
  * `azure`  
* `dependencies` - A list of key-value pair that can hold additional build information.
* `componentReferences` - List of components that are children of the given component.
  * `name` - Name of the child component
  * `version` - List of versions to use as child components

**When to use a Image Component?**  
Use an Image Component when you want to build a VM Image on Jenkins using Image Builder.

### BOM

**What is a BOM Component?**
A BOM component builds a BOM file. The BOM component can build 3 types of BOM files:

* TKG BOM
* TKR BOM
* TKR Compatibility

**How to define a BOM Component?**  
A BOM component is defined in a `<version>.yaml`.  
The `<version>.yaml` file defines the configuration to build a specific version of an BOM component.  
An example YAML configuration:

```yaml
name: tkr-bom
boms:
  - version: v1.20.1+vmware.2
    type: tkr
    imageRepository: projects-stg.registry.vmware.com/tkg
    dependencies:
      componentTanzuCoreVersion: v1.3.0+vmware.1
    componentReferences:
      - name: kubernetes
        version:
          - v1.20.1+vmware.2
      - name: cni_plugins
        version:
          - v0.8.7+vmware.4
      ...
      - name: ova-photon-3
        version:
          - v1.20.1+vmware.2-tkg.1.3-3161359046336510236
      - name: ova-ubuntu-2004
        version:
          - v1.20.1+vmware.2-tkg.1.3-3161359046336510236
      ...
```

YAML Explained:

* `name` - Is the name of the component. This should match with the name in the config files and the name of the component directory.
* `bmms` - Holds the configuration for a BOM type component.
* `version` - Is the version of the component. It is suggested to prefix version with `v`.
* `type` - Defines the BOM type. Supported values are:
  * `tkg`
  * `tkr`
  * `tkr-compatibility`
* `imageRepository` - The `imageRepository` value in the generated BOM file.
* `k8sVersion` - The default TKR version for a TKG BOM. Used only by `tkg` type BOM.
* `compatibilityMetadata` - The compatibility information for building a `tkr-compatibility` BOM. Used only by `tkr-compatibility` BOM.
  * `version` - TKR Compatibility BOM version
  * `managementClusterVersions` - List of compatibility information from TKG to TKR BOMS.
    * `version` - TKG version
    * `supportedKubernetesVersions` - List of supported TKR versions
* `dependencies` - A list of key-value pair that can hold additional build information.
* `componentReferences` - List of components that are children of the given component.
  * `name` - Name of the child component
  * `version` - List of versions to use as child components

### Aggregator Component

An Aggregator Component is a special type of cayman project. An Aggregator is a cayman project that triggers one or more child cayman projects.

**How is an Aggregator Component built?**  

* Register all descendant components(by recursively resolving `componentReferences`) that are of type `managed` and `unmanaged` components as cayman children
* Trigger the aggregator cayman project on buildweb
* Aggregator cayman project triggers its child cayman projects
* Aggregator cayman project waits for all the child components to finish
* Copy over the artifacts of all child builds under the aggregator build's _deliverables_
* Perform [publish actions](README.md#publishyaml) on the build outputs

**How to define an Aggregator Component?**  
An aggregator component is defined in a `<version>.yaml`.  
The `<version>.yaml` file defines the configuration to build a specific version of an aggregator component.  
An example YAML configuration:

```yaml
name: tkg_release
aggregators:
  - version: v1.3.0+vmware.1
    gobuildBaseBranch: vmware-0.0.0+vmware.0~7823ebe232ad5b36b8d975a129590951616bb069
    dependencies:
      gobuild.linux_hosttype: linux-centos72-gc32
    componentReferences:
      - name: vmware-private_tanzu-cli-tkg-plugins
        version:
          - v1.3.0+vmware.1

```

YAML Explained:

* `name` - Is the name of the component. This should match with the name in the config files and the name of the component directory.
* `aggregators` - Holds the configuration for an Aggregator type component.
* `version` - Is the version of the component. It is suggested to prefix version with `v`.
* `gobuildBaseBranch` - The _base branch_ in the cayman repo and the commit sha of base branch, with `~` as separator
* `dependencies` - A list of key-value pair that can hold additional build information.
  * All keys that start with `gobuild.` will be added to the dependency.config file in the cayman repo. Example: `gobuild.foo: bar` in dependencies will be stored as `foo='bar'` in dependency.config.
  * `gobuild.linux_hostyype` is a mandatory key for an _Aggregator_ type component. The host type used in the gobuild environment.
* `componentReferences` - List of components that are children of the given component.
  * `name` - Name of the child component
  * `version` - List of versions to use as child components

## `url.yaml`

Each component should have a `url.yaml` file. The `url.yaml` file captures the git URL for all the repos used in the build of the component.  
**Note**: Different versions of the same component will share `url.yaml`  
An example YAML for a Managed type component:

```yaml
name: kubernetes
upstreamURL: "https://github.com/kubernetes/kubernetes.git"
mirrorSandboxURL: "git@gitlab.eng.vmware.com:k8s-common-core/mirrors_github_kubernetes.git"
mirrorOfficialURL: "git@gitlab.eng.vmware.com:core-build/mirrors_github_kubernetes.git"
gobuildBaseURL: "git@gitlab.eng.vmware.com:core-build/cayman_kubernetes.git"
gobuildSandboxURL: "git@gitlab.eng.vmware.com:k8s-common-core/cayman_kubernetes.git"
gobuildOfficialURL: "git@gitlab.eng.vmware.com:core-build/cayman_kubernetes.git"
gobuildTarget: "cayman_kubernetes"
```

YAML file explained:

* `name` - Is the name of the component. This should match with the name in the config files and the name of the component directory.
* `upstreamURL` - Git url of the upstream project to be used. This is required for _Managed_ type components. Is optional for _Unmanaged_ and _Aggregator_ type components.  
  Note: This URL has to be the https url. If the upstream repo is private, grant access to [@tkg-serviceaccount](https://github.com/tkg-serviceaccount) and use the ssh git url.
* `mirrorSandboxURL` - The ssh git url for the sandbox repo of the mirror. It is suggested to create the sandbox fork in the `k8s-common-core` namespace. This is required for _Managed_ type components. Is optional for _Unmanaged_ and _Aggregator_ type components.
* `mirrorOfficialURL` - The ssh git url for the official mirror repo. This has to be a repo in the `core-build` namespace. This is required for _Managed_ type components. Is optional for _Unmanaged_ and _Aggregator_ type components.
* `gobuildBaseURL` - The ssh url to the gitlab cayman repo that has the _base branch_. This is required for _Managed_ type components.
* `gobuildSandboxURL` - The ssh url for the gitlab cayman repo to use for sandbox builds. It is suggested to create the sandbox fork in the `k8s-common-core` namespace for _Managed_ and _Aggregator_ components. For unmanaged components this value should match `gobuildOfficialURL`.  This is required for _Managed_, _Unmanaged_ and _Aggregator_ type components.
* `gobuildOfficialURL` - The ssh url for the gitlab cayman repo to use for official builds. This has to be a repo in the `core-build` namespace. This is required for _Managed_, _Unmanaged_ and _Aggregator_ type components.
* `gobuildTarget` - The cayman target name as registered at https://bmps.eng.vmware.com/target/

An example YAML for a Unmanaged type component:

```yaml
name: tkg_telemetry
gobuildBaseURL: git@gitlab.eng.vmware.com:core-build/tkg-telemetry.git
gobuildSandboxURL: git@gitlab.eng.vmware.com:core-build/tkg-telemetry.git
gobuildOfficialURL: git@gitlab.eng.vmware.com:core-build/tkg-telemetry.git
gobuildTarget: tkg_telemetry
```

YAML file explained:

* `name` - Is the name of the component. This should match with the name in the config files and the name of the component directory.
* `gobuildSandboxURL` - same as `gobuildOfficialURL`
* `gobuildOfficialURL` - The ssh url for the gitlab cayman repo to use for official builds. This has to be a repo in the `core-build` namespace.
* `gobuildTarget` - The cayman target name as registered at https://bmps.eng.vmware.com/target/

An example YAML for Image type component:

```yaml
name: ova-ubuntu-2004
jenkinsJobName: build-ova-v1alpha2
```

YAML file explained:

* `name` - Is the name of the component. This should match with the name in the config files and the name of the component directory.
* `jenkinsJobName` - Name of the Jenkins job to build the VM Image. The jenkins job should be defined in the Jenkins server at : https://kscom.svc.eng.vmware.com

Since BOM type components do not engage with any external URLs their `url.yaml` are very simple.
An example YAML for BOM type component:

```yaml
name: tkg-bom
```

## `publish.yaml`

Each component should have a `publish.yaml` file. The `publish.yaml` file defines all the artifacts generated by a build, and the publish actions to perform on each of the build artifacts.  
**Note**: Different versions of the same component will share `publish.yaml`  
An example YAML:

```yaml
name: kubernetes
publish:
  - name: kubeAPIServer
    type: image
    sourcePath: https://build-squid.eng.vmware.com/build/mts/release/$BUILDID/publish/lin64/kubernetes/images/kube-apiserver-$UNDERSCOREVERSION.tar.gz
    checksumPath: https://build-squid.eng.vmware.com/build/mts/release/$BUILDID/publish/lin64/kubernetes/images/kubernetes-$UNDERSCOREVERSION-image-checksums.txt
    destinationPath: kube-apiserver:$UNDERSCOREVERSION
  - name: kube-apiserver_file
    type: file
    sourcePath: https://build-squid.eng.vmware.com/build/mts/release/$BUILDID/publish/lin64/kubernetes/images/kube-apiserver-$UNDERSCOREVERSION.tar.gz
    checksumPath: https://build-squid.eng.vmware.com/build/mts/release/$BUILDID/publish/lin64/kubernetes/images/kubernetes-$UNDERSCOREVERSION-image-checksums.txt
    destinationPath: $CHANNEL/image-builder/kubernetes/$VERSION/bin/linux/amd64/kube-apiserver.tar
  ...
nonpublish:
  - name: kubeadm-linux
    type: file
    destinationPath: https://build-squid.eng.vmware.com/build/mts/release/$BUILDID/publish/lin64/kubernetes/executables/kubeadm-linux-$VERSION.gz
    checksumPath: https://build-squid.eng.vmware.com/build/mts/release/$BUILDID/publish/lin64/kubernetes/executables/kubernetes-$VERSION-executables-checksums.txt
  - name: kubectl-linux
    type: file
    destinationPath: https://build-squid.eng.vmware.com/build/mts/release/$BUILDID/publish/lin64/kubernetes/executables/kubectl-linux-$VERSION.gz
    checksumPath: https://build-squid.eng.vmware.com/build/mts/release/$BUILDID/publish/lin64/kubernetes/executables/kubernetes-$VERSION-executables-checksums.txt
  ...
```

The `publish.yaml` captures 2 kinds of build artifacts: `publish` and `nonpublish`.  
`nonpublish` artifacts that dont need any publish action.  
Any additional publish action is performed on `publish` artifacts.
A `publish` artifact can be of 2 types:

* type `file` - The file is copied from `sourcePath` and placed in `destinationPath`
* type `image` - The file in `sourcePath` is converted to a container image and tagged with the value of `destinationPath` and pushed to the `stagingImageRepo.imageRepository` registry mentioned in the release yaml.

YAML file explained:

* `name` - Is the name of the component. This should match with the name in the config files and the name of the component directory.
* `publish` - This section is a list of artifacts that need an additional publish action.
* `nonpublish` - This sections is a list of artifacts that do not need any additional publish action.
* `type` - Can be one of `file` or `image`
* `sourcePath` - The source path of the artifact. Required for `publish` artifacts.
* `destinationPath` - The final location the artifacts are placed.
* `checksumPath` - Path to the checksum file that holds the sha256 value of the artifact.

All the URLs captured in `publish.yaml` are template URL with template variable that are replaced with real values at run time.

* `$BUILDID` - Buildweb build id. Will be resolved to `sb-xxx` for sandbox builds and `bora-xxx` for official builds
* `$VERSION` - The exact version of the component
* `$UNDERSCOREVERSION` - The underscored valued of `$VERSION`. All "+" characters will be replaced with "_".  
   Example: v1.3.0+vmware.2 is converted to v1.3.0_vmware.2
* `$CHANNEL` - The URL to a folder in artifactory that is uniquely generated for every run.  

## `publishVersionMap.yaml`

Each component has a `publishVersionMap`, which defines the version map between config file and publish file.  
When you create a new component, please create a `publishVersionMap.yaml` file with the following empty content:
```yaml
{}
```

## `urlVersionMap.yaml`

Each component has a `urlVersionMap`, which defines the version map between config file and url file.  
When you create a new component, please create a `urlVersionMap.yaml` file with the following empty content:
```yaml
{}
```

## `_boltArtifacts` folder

Each component has a `_boltArtifacts` folder, which saves the lock/cache/buildinfo/refbuildinfo files automatically generated by bolt-cli.  
When you create a new component, please create a `_boltArtifacts` folder and make add `.keep` file to it. The content of the `.keep` file is empty.
