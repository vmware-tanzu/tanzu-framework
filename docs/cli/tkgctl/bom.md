# Details on BoM usage in the TKG library

## Introduction

BoM stands for Bill of Materials and there are 2 types of BoM files:

* a TKG BoM file   (1 file per TKG release)
* TKR BoM files (multiple files per TKG release)

### TKG BoM file

* This file contains information about TKG related components which mainly get associated with management cluster creation.
  * It specifies the TKG version under the `release.version` key and the default TKR version under the `default.k8sVersion` key.
  * It also specifies cluster-api provider components, pinniped components, image repo information, and many more.

[A sample TKG BoM file is available here.](example-boms/tkg-bom.yaml)

### TKR BoM files

* These files contain information related to:
  * the specific k8s version,
  * node images to use with (vsphere, aws, azure) for the matching k8s version,
  * add-ons specific components.

[A sample TKR BoM file is available here.](example-boms/tkr-bom.yaml)

## Bundling BoM files into the TKG library

* The TKG BoM file (and TKR BoM files) are decoupled from the Tanzu CLI and the CLI runtime library.
* Instead, the Tanzu CLI is compiled with a TKG Compatibility Image path (which is a build-time constant). At runtime, the CLI downloads this compatibility file and uses it to determine the location of the appropriate TKG/TKR BoM files (choosing the version that matches the version of the management-cluster plugin currently installed).  The CLI then downloads the correct BoM files to the user's machine. [Reference.](../../../tkg/tkgconfigupdater/ensure.go)

### Why do we need a TKG Compatibility Image?

If the BoM files were bundled with the CLI, and if, at a later time, there were a CVE fix for any images/components that are part of the TKG/TKR BoM files for an existing release, this would require a new CLI release.
By using a TKG Compatibility Image, a new CLI release can be avoided in such a case. The release team can publish a new Compatibility Image (using a new git tag) with updated TKG BoM paths for a given version of the management-cluster plugin.  The CLI would then download the new Compatibility Image (using the latest Image tag) and thereby be able to download and use the updated TKG/TKR BoM files while creating a new management cluster.

A sample compatibility file is shown below:

```yaml
version: v2
managementClusterPluginVersions:
- version: v1.5.1
  supportedTKGBomVersions:
  - imagePath: tkg-bom
    tag: v1.5.1
- version: v1.5.0
  supportedTKGBomVersions:
  - imagePath: tkg-bom
     tag: v1.5.0
```

Below is a new sample compatibility file after updating the BoM version (new BoM which has images/components with CVE fixes for v1.5.0). The new Compatibility Image would be pushed to an image repository along with an updated tag (e.g., **tkg/tkg-compatibility:v3**)

```yaml
version: v3
managementClusterPluginVersions:
  - version: v1.5.1
    supportedTKGBomVersions:
      - imagePath: tkg-bom
        tag: v1.5.1
  - version: v1.5.0
    supportedTKGBomVersions:
      - imagePath: tkg-bom
          tag: v1.5.0-patch
```

Notes:

* Once a management cluster is created, it runs the TKR controller which reconciles all the supported TKR and updates TKR compatibility.
* For workload cluster creation, the CLI can download the necessary TKR BoM files to the user's local machine directly from the management cluster (using an existing ConfigMap on the management cluster).

### Why do we need to download the TKR BoM files locally?

* This is required because the cluster template creation logic is still running locally with the TKG library which internally uses YTT
* To generate cluster templates using YTT overlays easily we are downloading the TKR BoM files locally before generating the cluster templates
* This allows users to read and understand the content of the TKR BoM files for debugging purpose

## Updating the TKG Compatibility Image Path into the TKG library

1. Update the `TKG_DEFAULT_IMAGE_REPOSITORY` and `TKG_DEFAULT_COMPATIBILITY_IMAGE_PATH` variables inside [Makefile](../../../Makefile)
2. Run `make configure-bom` which will update the build-time constants for downloading the TKG compatibility file.
3. Commit the `Makefile` changes alongside the changes generated to the constants file.

## How are the BoM files getting used in the CLI and the cluster template creation with ytt

### How are the BoM files getting used in the TKG CLI?

* TKG Compatibility file metadata is bundled into TKG library as build time constants
* When the tkgctl client gets created, as part of ensuring prerequisites, tkg-compatibility file and BoM files are extracted to the BoM file location. For tanzu cli it will be,$HOME/.tanzu/tkg/compatibility and $HOME/.tanzu/tkg/bom respectively. If there is a compatibility file already present in the user's local file system, the new compatibility file would not be downloaded. User can delete the compatibility file so that tanzu CLI would download the latest compatibility file(or user can do `tanzu config init` or `tanzu management-cluster create --force-config-update -f <filename>`).
* Library implements [tkgconfigbom](../../../tkg/tkgconfigbom/client.go) package which implements methods to read TKG and TKR BoM files.
* TKG CLI reads in these BoM files from the user's local filesystem and uses the content of TKG and TKR BoM files for the various purpose, a few of those are listed below:
  * Uses TKG BoM file to determine image repository to use for provider installation and updates images section under TKG settings file $HOME/.tanzu/tkg/config.yaml
  * Reads TKR BoM file to select correct AMI, Azure image to use. Uses the same information for the vSphere VM template verification purpose
  * Uses  information in TKR BoM to set config variables like KUBERNETES_VERSION, AMI_ID, AZURE_IMAGE_* etc
* Another use is in cluster template generation but for that, we will be using ytt library to directly read in BoM file as text files as described in next section.

### How the cluster template creation with ytt uses BoM file?

* When creating cluster template with ytt we read files mentioned in `cluster-template-definition-<plan>.yaml` which includes BoM file with file-mark as text-plain. Meaning BoM files will be read in as plain text file instead of base-template or overlay files.
* Once the BoM files are read during the cluster template generation process into the ytt engine, the code mentioned in config_default.yaml converts it to data_values as part of the boms map. As mentioned below

```yaml
#! ---------------------------------------------------------------------
#! BoM file processing, internal use only
#! ---------------------------------------------------------------------

#@ load("@ytt:yaml", "yaml")
#@ load("@ytt:data", "data")

#@ files = data.list()
boms:
  #@ for/end file in [ f for f in files if f.startswith("tkg-bom") or f.startswith("tkr-bom") or f.startswith("bom")]:
  - bom_name: #@ file
  bom_data: #@ yaml.decode(data.read(file))
```

* This will store all BoM files present under $HOME/.tanzu/tkg/bom into the boms array
* TKG library sets TKG_DEFAULT_BOM config variable before generating cluster template which gets used to determine default BoM file with helper functions
* ytt overlays have some helper functions to get `get_default_tkr_bom_data`, `get_bom_data_for_tkr_name`, `get_default_tkg_bom_data` which gets used during overlays
* Using the functions mentioned above, ytt overlays are written in a way that reads the correct BoM file based on the given TKR, and files in the correct image name and image tag.

## How TKG library determines the BoM files in the user's file system is outdated and needs to be replaced or not?

* As part of [ensurePrerequisite](../../../tkg/tkgctl/client.go) whenever we create tkgctl client, [EnsureBOMFiles](../../../tkg/tkgconfigupdater/ensure.go) function is invoked
* This function checks the default TKG BoM file name (determined from the tkg-compatibility file) and compares it with the BoM file present in the user's local filesystem
* If the default TKG BoM filename does not exist, TKG will back up the old BoM directory and extract bundled BoM file into the user's BoM directory
