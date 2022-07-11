## What is tanzu addons manager: 
https://github.com/vmware-tanzu/tanzu-framework/tree/main/addons

The Tanzu Addons Manager gets deployed on the Tanzu management cluster by default.

## Configure the TKR bundle
1. Get an existing TKR bundle from Tanzu release page.
2. Customize
    - TKR bundle is constituted by yaml files, which contains image locations and configs, secrets, etc
    - `Packages` folder contains the addons
    - `TanzuKubernetesRelease.yml` stores the TKR version and packages name (which are same as the addon names in the manifests under `packages` folder)
    - `ClusterBootstrapTemplate.yml` contains the configs of addons which will be leveraged to bootstrap/updated workload clusters
    ![tkr-bundle](https://github.com/yzaccc/public_uploads/blob/main/images/tkr_bundle.png?raw=ture)
    - To customize, you can modify the manifests of existing addons(image url, configs, etc), you can also add new packages
    - To add new packages
        - You need to use Carvel Imgpkg tooling to create a imgpkg bundle, refer https://carvel.dev/imgpkg/docs/v0.29.0/basic-workflow/
        - Publish your imgpkg bundle to your own registry
        - Create a new package in the `packages` folder, according to the current format, and configure with the info of your imgpkg bundle
        - Insert your new package’s name into `TanzuKubernetesRelease.yml`
        - Insert your new package’s name into `ClusterBootstrapTemplate.yml`
        - (Optional) If you have secret or config to pass into the imgpkg bundle, refer `config/AntreaConfig.yaml` for the config, refer `config/PinnipetConfig.yml` for secrets. And transit the configs/secrets into the `ClusterBootstrapTemplate.yml` using `valuesFrom` field.
     - Great! Now you have your own TKR bundle, publish it to your registry
    - Configure the Tanzu Addons Manager to consume your TKR bundle.
    - Enjoy the automation, that your addon will be reconciled on workload clusters!
## Cluster upgrade
    - During a cluster upgrade, which is triggered by setting TKR label of cluster object:
        - cluster.Labels[constants.TKRLabelClassyClusters]
        - 1.22.3 -> 1.23.1 
    - The addons manager compares the existing ClusterBootstrapTempalte with the new one in your registry, to determine if there is any addon that needs to be added or updated.
        - If needed, the addons manager will pause the workload cluster, by leveraging cluster api `pause` attribute -> upgrade/add the addons -> and resume the workload clusters
