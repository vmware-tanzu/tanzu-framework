# Simple-App Addon
Simple App addon

## Prerequisites
* kapp-controller deployed on cluster where addon is to be installed.

## Steps to deploy simple-app addon

### Management Cluster

1. Create TKR CRD
   
    ```shell
    kubectl apply -f ../../../config/crd/bases/run.tanzu.vmware.com_tanzukubernetesreleases.yaml
    ```

2. Deploy Addon controller

    ```shell
    kubectl apply -f ../tanzu-addons-manager.yaml
    ```

3. Create TKR instance and BOM configmap

   ```shell
   kubectl apply -f tkr.yaml
   kubectl apply -f bom_configmap.yaml
   ``` 

4. Create simple-app addon secret.
 
   Replace cluster-name (tkg.tanzu.vmware.com/cluster-name: cluster1) and 
   namespace (same as cluster's namespace).
   
    ```shell
    kubectl apply -f simple-app_addon_secret.yaml
    ```

### Workload cluster

1. Check if app is reconciled successfully on workload cluster

    ```shell
    kubectl get app simple-app -n <cluster-namespace>
    ```