## prerequisite knowledge

### How bolt isolate the sandbox container images from the official images
Each component's `publish.yaml` defines container images' `sourcePath` and `destinationPath`. For example,
```yaml
  - name: veleroImage
    type: image
    sourcePath: https://build-squid.eng.vmware.com/build/mts/release/$BUILDID/publish/lin64/velero/images/velero-$UNDERSCOREVERSION.tar.gz
    checksumPath: https://build-squid.eng.vmware.com/build/mts/release/$BUILDID/publish/lin64/velero/images/velero-$UNDERSCOREVERSION-images-checksums.txt
    destinationPath: velero/velero:$UNDERSCOREVERSION
```
For sandbox build, bolt-cli will publish the image to the `sandbox/` namespace in the container registry, for example: `sandbox/velero/velero:$UNDERSCOREVERSION`.   
For official build, bolt-cli will publish the image to both `sandbox/` namespace and the official namespace, for example: `sandbox/velero/velero:$UNDERSCOREVERSION` and `velero/velero:$UNDERSCOREVERSION`.  

This is applicable to all components, including bom (bolt-cli treat tkr-compatibility as bom component).

## How bolt build tkr-compatibility

Given a new tkr-compatibility config file (`v2.yaml`), bolt-cli will combine this new config file with the existing global compatibility file (`tkr-compatibility.yaml`), and store the new compatibility file in `tkr-compatibility-v2.yaml`.  
Bolt-cli will update the global compatibility file only when `--rtm=true` flag is set.

## How bolt publish tkr-compatibility image

Design goal:
1. official builds of tkr-compatibility should share the imagePath only when the release-graphs are same
2. every sandbox build of tkr-compatibility should have its own unique imagePath

Here is an example of the new tkr-compatibility config for tkg-1.3.1 dailybuild:
```yaml
name: tkr-compatibility
boms:
- version: v2-v1.3.1-zlatest
  type: compatibility
  dependencies:
    rtmVersions:
      - v1.3.1
  compatibilityMetadata:
    version: v2
    managementClusterVersions:
    - version: v1.3.1-zlatest
      supportedKubernetesVersions:
      - v1.20.4+vmware.1-tkg.2-zlatest
      - v1.19.8+vmware.1-tkg.2-zlatest
      - v1.18.16+vmware.1-tkg.2-zlatest
      - v1.17.16+vmware.2-tkg.2-zlatest
```
Here is an example of the new tkr-compatibility config for tkg-1.4.0 dailybuild:
```yaml
name: tkr-compatibility
boms:
- version: v2-v1.4.0-zlatest
  type: compatibility
  dependencies:
    rtmVersions:
      - v1.4.0
  compatibilityMetadata:
    version: v2
    managementClusterVersions:
    - version: v1.4.0-zlatest
      supportedKubernetesVersions:
      - v1.20.4+vmware.1-tkg.3-zlatest
      - v1.19.8+vmware.1-tkg.3-zlatest
      - v1.18.16+vmware.1-tkg.3-zlatest
      - v1.17.16+vmware.2-tkg.3-zlatest
```

### To solve goal `1`:  
All official build of tkr-compatibility will be published to `boms[0].version/tkr-compatibility:$VERSION`  
Note:  
`$VERSION` here is `boms[0].compatibilityMetadata.version`(`v2`), NOT config's version(`boms[0],version`, aka `v2-v1.3.1-zlatest`)  

### To solve goal `2`:  
All sandbox build of tkr-compatibility will be published to `sandbox/$boms[0].version/$CHANNEL/tkr-compatibility:$VERSION`  

### What about RC?
Create a new tkr-compatibility config:
```yaml
name: tkr-compatibility
boms:
- version: v2-v1.3.1-rc.1
  type: compatibility
  dependencies:
    rtmVersions:
      - v1.3.1
  compatibilityMetadata:
    version: v2
    managementClusterVersions:
    - version: v1.3.1-rc.1
      supportedKubernetesVersions:
      - v1.20.4+vmware.1-tkg.2-rc.1
      - v1.19.8+vmware.1-tkg.2-rc.1
      - v1.18.16+vmware.1-tkg.2-rc.1
      - v1.17.16+vmware.2-tkg.2-rc.1
```
this means that for RC sandbox build, tkr-compatibility will be published to `sandbox/v2-v1.3.1-rc.1/$CHANNEL/tkr-compatibility:v2`  
for RC official build, tkr-compatibility will be published to `v2-v1.3.1-rc.1/tkr-compatibility:v2`

### What about RTM candidate?
Create a new tkr-compatibility config:
```yaml
name: tkr-compatibility
boms:
- version: v2-v1.3.1
  type: compatibility
  dependencies:
    rtmVersions:
      - v1.3.1
  compatibilityMetadata:
    version: v2
    managementClusterVersions:
    - version: v1.3.1
      supportedKubernetesVersions:
      - v1.20.4+vmware.1-tkg.2
      - v1.19.8+vmware.1-tkg.2
      - v1.18.16+vmware.1-tkg.2
      - v1.17.16+vmware.2-tkg.2
```
Set `--rtm=true` for bolt-cli  
this means that for RTM candidate sandbox build, tkr-compatibility will be published to `sandbox/$CHANNEL/tkr-compatibility:v2`  
for RTM candidate official build, tkr-compatibility will be published to `tkr-compatibility:v2`(only at this time, we have a change to test that tkr-controller can handle the case that there are multiple version/tag of tkr-compatibility image)  
the global compatibility file will be generated, but the image tarball is only stored in channel, it won't be published to any image registry.

### What about RTM?
Nothing need to do about tkr-compatibility, we can reuse the RTM candidate tkr-compatibility build directly.  
CIP will publish the global conpatibility image from RTM candidate channel to prod registry.

### What need to do after RTM?
Since 1.3.1 RTM files are checked in to bolt-release-yaml main branch, now the main branch's global compatibility file (`component/tkr-compatibility/tkr-compatibility.yaml`) looks like this:
```yaml
version: v2
managementClusterVersions:
- version: v1.3.0
  supportedKubernetesVersions:
  - v1.20.4+vmware.1-tkg.1
  - v1.19.8+vmware.1-tkg.1
  - v1.18.16+vmware.1-tkg.1
  - v1.17.16+vmware.2-tkg.1
- version: v1.3.1
  supportedKubernetesVersions:
  - v1.20.4+vmware.1-tkg.2
  - v1.19.8+vmware.1-tkg.2
  - v1.18.16+vmware.1-tkg.2
  - v1.17.16+vmware.2-tkg.2
```
Since global compatibility file has version `v2`, now we need to bump `1.4.0` tkr-compatibility's version to `v3`, now the `1.4.0` dailybuild config file looks like this:
```yaml
name: tkr-compatibility
boms:
- version: v3-v1.4.0-zlatest
  type: compatibility
  dependencies:
    rtmVersions:
      - v1.4.0
  compatibilityMetadata:
    version: v3
    managementClusterVersions:
    - version: v1.4.0-zlatest
      supportedKubernetesVersions:
      - v1.20.4+vmware.1-tkg.3-zlatest
      - v1.19.8+vmware.1-tkg.3-zlatest
      - v1.18.16+vmware.1-tkg.3-zlatest
      - v1.17.16+vmware.2-tkg.3-zlatest
```
NOTE:  
if the release team forget to bump the 1.4.0 dailybuild's version from `v2` to `v3`, then the next `1.4.0-zlatest` dailybuild will fail,  
because bolt-cli do sanity check to make sure that the new compatibility's `boms[0].compatibilityMetadata.version` should be strictly larger than global compatibility file's version
