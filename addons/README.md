# How to add an addon template
This directory hosts the addon templates. If you want to add a new addon template, create a new folder named after the new addon.

Current addon templates:
- addon-manager
- antrea
- calico
- kapp-controller
- metrics-server
- pinniped
- vsphere_cpi
- vsphere_csi
  
## Folder Structure

```
.
+-- examples (optinal)
|   +-- example_values.yaml
|   +-- ....
+-- tempaltes
|   +-- base-files
|   +-- libs
|   +-- overlays
|   +-- values.star
|   +-- values.yaml
+-- Makefile
```
### examples (optional)
Put sample `values.yaml` files here. These files can be examples that represent some specific use cases, for more details refer to [pinniped examples](pinniped/examples)

### templates
This is where to put the ytt template files.

#### base files
YTT base files. Ideally, they should not contain any YTT functions or variables. All customizations to the manifests should better happen in Overlays.

#### libs
Custom YTT fuctions that might be used by the Overlays.

#### overlays
YTT overlays that customize the yaml manifests. For more details about TYY overlay, please refer to [this link](https://carvel.dev/ytt/#example:example-overlay-files)

#### values.star and values.yaml
Define the input validation funtion in `values.star`. Follow the `starlarks` language syntax in this file. Input validation is optional.

Define all the configurable values in `values.yaml`. These values will be used by overlay to customize the yaml manifests.

### Makefile
The Makefile should follow the current examples. Change the following variables to match your addon.
- `IMG_CATEGORY := csi`
- `IMG_CLUSTER_TYPE := management workload` (What types of cluster should the addon run on)
- `ADDON_NAME := vsphere-csi`
- `IMG_NAME ?= vsphere-csi-templates`

## Template image and BOM
After the template is successfully created, you need to change two files to make sure the template image and BOM contents are correctly generated.

### ./Makefile
- After line 19, add the newly added template image name `<ADDON_NAME>_TEMPLATES_IMAGE_NAME ?= <addon-name>-templates`
- In `Addon templates` section, follow the examples to add the new addon to all three targets, which are `build-addon-template-images`, `save-addon-template-images` and `push-addon-template-images`

### ./bom/Makefile
In line 8 of `./bom/Makefile` add your newly added `ADDON_NAME` to the list of `OBJECTS`