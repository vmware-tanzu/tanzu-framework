# Add or Update a component in TKG

The TKG release graph looks like the following:  
<img src="tkg_example_graph.png" alt="Example TKG Release Graph" width="460" height="500"/>

## Add new component to TKG
The following points outline the steps needed to add a new component to TKG:
* Define a new component using a suitable component type definition. Checkout [this doc](README.md) to lear how to define a new component.
* Add the component to the current TKG release graph by adding an entry under `componentReferences` of the corresponding [tkg-bom configuration](tkg-bom).
* If this is a gobuild type component (ummanaged/managed), add this component to the `tkg_release` [filters](../filters/filters.yaml) list

## Update the version of a component in TKG
The following points outline the steps needed to update the version of a component in TKG:
* Create a new config file for the new version of the component by following [these instructions](UPDATE_COMPONENT_VERSION.md#how-to-update-the-version-of-a-component-in-a-tkg-release?)
* Update the entry for the component under `componentReferences` of the corresponding [tkg-bom configuration](tkg-bom) to use the new version

## Add a new component to TKG Core
The following points outline the steps needed to add a new component to TKG:
* Define a new component using a suitable component type definition. Checkout [this doc](README.md) to lear how to define a new component.
* Add the component to the current TKG Core release graph by adding an entry under `componentReferences` of the corresponding [kscom_release configuration](kscom_release).

## Update the version of a component in TKG Core
The following points outline the steps needed to update the version of a component in TKG:
* Create a new config file for the new version of the component by following [these instructions](UPDATE_COMPONENT_VERSION.md#how-to-update-the-version-of-a-component-in-a-tkg-release?)
* Update the entry for the component under `componentReferences` of the corresponding [kscom_release configuration](kscom_release) to use the new version