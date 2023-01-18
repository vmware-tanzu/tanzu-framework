# How to update the version of a component in a TKG Release?

## Updating the version of a managed component
Follow these steps to update the version of a managed component:
* Create a new config file of type [_Managed Component_](README.md#managed-component) with the new version under `component/<name-of-the-component>` 
  * The name of the file should be `<version>.yaml`
  * The version should start with `v`
* Set the `upstreamTag` or `upstreamCommit` (only one) to the desired tag/commit to pull from upstream
* Make sure the correct publish artifacts are all captured, if there are any changes from the previous versions you should make a new `vX.publish.yaml` file representing the new collection of published artifacts. For example, `v1.publish.yaml` or `v2.publish.yaml`. bolt-cli will automatically pick up the new `vX.publish.yaml` for the new version of config for this component. There is **no** need to update the `publishVersionMap.yaml` file.
* Make sure you are using the correct url config, if there are any changes from the previous versions of `url.yaml`, you should make a new `vX.url.yaml` file representing the new collection of url config. For example, `v1.url.yaml` or `v2.url.yaml`. bolt-cli will automatically pick up the new `vX.url.yaml` for the new version of config for this component. There is **no** need to update the `urlVersionMap.yaml` file.
* If you want the parent component(s) to use the new version of this component make sure to update the reference in the parent
  * Example:  
    If you are updating `kubernetes_autoscaler` and want the new version to be picked up by `tkg-bom` make sure to update the corresponding entry for `kubernetes_autoscaler` under `compoenentReferences` of `tkg-bom` 
  * Note: You can use the latest release graph image stored in the `release` folder to help identify all the parents of a component   
    
## Updating the version of an unmanaged component
Follow these steps to update the version of an unmanaged component:
* Create the required branch on the cayman repo in the `core-build` namespace. 
  * Make sure to update the git submodule, if applicable, to the correct branch on the mirror repo.
  * Make sure the mirror repo is tagged with the appropriate version to reflect the expectations in the build logic of cayman repo
  * Make sure the correct publish artifacts are all captured, if there are any changes from the previous versions you should make a new `vX.publish.yaml` file representing the new collection of published artifacts. For example, `v1.publish.yaml` or `v2.publish.yaml`. bolt-cli will automatically pick up the new `vX.publish.yaml` for the new version of config for this component. There is **no** need to update the `publishVersionMap.yaml` file.
  * Make sure you are using the correct url config, if there are any changes from the previous versions of `url.yaml`, you should make a new `vX.url.yaml` file representing the new collection of url config. For example, `v1.url.yaml` or `v2.url.yaml`. bolt-cli will automatically pick up the new `vX.url.yaml` for the new version of config for this component. There is no need to update the `urlVersionMap.yaml` file.
* Create a new config file of type [_Unmanaged Component_](README.md#unmanaged-component) with the new version under `component/<name-of-the-component>` 
  * The name of the file should be `<version>.yaml`
  * The version should start with `v` 
  
## Updating the version of `tanzu_core`
Updating the version of `tanzu_core` is a special case. Since `tanzu_core` is a _managed component_ to update the version of `tanzu_core` one should follow all the steps under [updating the version of a managed component](#updating-the-version-of-a-managed component) and then do the following:
* Update `componentTanzuCoreVersion` in the configuration of the latest tkg-bom
* Update `componentTanzuCoreVersion` in all tkr-bom configs that are included in the current TKG release
* Add an entry for the new version in the latest tkg-compatibility configuration
