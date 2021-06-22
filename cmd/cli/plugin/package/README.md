# Package

Manage package lifecycle operations.

## Usage

Package and Repository Operations ( Subject to Change ):

```sh
>>> tanzu package --help
Tanzu package management

Usage:
  tanzu package [command]

Available Commands:
  install       Install a package
  get           Get details for a package or installed package
  install       Install a package
  list          List a package
  repository    Repository operations
  uninstall     Uninstall a package

Flags:
  -h, --help              help for package
      --log-file string   Log file path
  -v, --verbose int32     Number for the log level verbosity(0-9)

Use "tanzu package [command] --help" for more information about a command.
```

```sh
>>> tanzu package repository --help
Add, list, get or delete a repository for tanzu packages

Usage:
  tanzu package repository [command]

Available Commands:
  add         Add a repository
  delete      Delete a repository
  get         Get repository status
  list        List repository
  update      Update repository

Flags:
  -h, --help   help for repository

Global Flags:
      --log-file string   Log file path
  -v, --verbose int32     Number for the log level verbosity(0-9)

Use "tanzu package repository [command] --help" for more information about a command.
```

## Test

1. Create a management cluster

Note: Steps 2 & 3 are applicable until the kapp controller alpha release is built into daily build

2. Delete kapp-controller deployment 
3. Apply kapp controller and carvel CRD's from https://github.com/vmware-tanzu/carvel-kapp-controller/blob/develop/alpha-releases/v0.20.0-rc.1.yml
4. Build cli
5. Install cli and plugins
6. Use package commands to:
   - add a repository
   - list a repository
   - get a repository status
   - list packages
   - delete a repository
   
   For v0.20.0-rc.1 release, use the following image package bundles for testing:
   1. Repository URL  => index.docker.io/k8slt/kc-e2e-test-repo@sha256:62d187c044fd6a5c57ac870733fe4413ebf7e2909d8b6267707c5dd2080821e6
   
   Here is an example workflow
   
7. Add a repository

   ```sh
   >>> tanzu package repository add testrepo index.docker.io/k8slt/kc-e2e-test-repo@sha256:62d187c044fd6a5c57ac870733fe4413ebf7e2909d8b6267707c5dd2080821e6 -n test-ns --create-namespace
   Package Repository 'testrepo' added
   ```

8. Update a repository

   ```sh
   >>> tanzu package repository update testrepo2 projects-stg.registry.vmware.com/tkg/shivaani/package-bundle:1.0.0 -n test-ns
   Updated package repository 'testrepo2'
   ```

9. List the repository
   ```sh
   >>> tanzu package repository list -n test-ns
   NAME      REPOSITORY                                                                                                        STATUS               DETAILS  
   testrepo  index.docker.io/k8slt/kc-e2e-test-repo@sha256:62d187c044fd6a5c57ac870733fe4413ebf7e2909d8b6267707c5dd2080821e6    Reconcile succeeded          
   ```

10. Install a package

    Example 1: Install the specified version for package name "fluent-bit.tkg-standard.tanzu.vmware" and while providing the values.yaml file.
    ```sh
    
    >>> tanzu package install fluentbit --package-name fluent-bit.tkg-standard.tanzu.vmware --namespace test-ns --create-namespace --version 1.7.5-vmware1 --values-file values.yaml
    Installing package 'fluentbit' in namespace 'test-ns'
    Added installed package 'fluentbit' in namespace 'test-ns'
    ```

    An example values.yaml is as follows:
    ```sh
    #@data/values
    #@overlay/match-child-defaults missing_ok=True
    ---
    fluent_bit:
      outputs: |
        [OUTPUT]
          Name     stdout
          Match    *
    ```
    
    Example 2: Install the latest version for package name "simple-app.corp.com". If the namespace does not exist beforehand, it gets created.
    ```sh
    >>> tanzu package install simple-app --package-name simple-app.corp.com --namespace test-ns-2 --create-namespace
    Added installed package 'simple-app' in namespace 'test-ns-2'
    ```

11. Uninstall a package

    ```sh
    >>> tanzu package uninstall fluentbit --namespace test-ns
    Uninstalling package 'fluentbit' from namespace 'test-ns'
    Uninstalled package 'fluentbit' from namespace 'test-ns'
    ```

12. List the packages

   ```sh
   #List installed packages in the default namespace
   >>> tanzu package list
   NAME  VERSION
  
   #List installed packages across all namespaces
   >>> tanzu package list -A
   NAME        VERSION        NAMESPACE      
   fluent-bit  1.7.5-vmware1  tanzu-logging  
   mysa        2.0.0          test-sa      
  
   #List installed packages in user provided namespace
   >>> tanzu package list --namespace tanzu-logging
   NAME        VERSION        
   fluent-bit  1.7.5-vmware1
  
   #List all available package CRs
   >>> tanzu package list --available --namespace test-ns
   NAME                                    DISPLAYNAME        SHORTDESCRIPTION                                             
   cert-manager.tkg-standard.tanzu.vmware  cert-manager       This package provides certificate management functionality.  
   fluent-bit.tkg-standard.tanzu.vmware    fluent-bit         This package provides logging functionality.                 
   simple-app.corp.com                     simple-app v1.0.0  Simple app consisting of a k8s deployment and service
  
   #List all available packages for package name
   >>> tanzu package list --available fluent-bit.tkg-standard.tanzu.vmware --namespace test-ns
   or
   >>> tanzu package list fluent-bit.tkg-standard.tanzu.vmware --available --namespace test-ns
   NAME                                  VERSION        
   fluent-bit.tkg-standard.tanzu.vmware  1.7.5-vmware1  
  
   With kubeconfig flag

   >>> tanzu package list --kubeconfig wc-kc-alpha8                       
   NAME  VERSION  
  
   >>> tanzu package list -A --kubeconfig wc-kc-alpha8
   NAME  VERSION        NAMESPACE        
   mycm  1.1.0-vmware1  test-1           
   myfb  1.7.5-vmware1  test-logging-wc  
      
   >>> tanzu package list --namespace test-logging-wc --kubeconfig wc-kc-alpha8
   NAME  VERSION        
   myfb  1.7.5-vmware1  
  
   >>> tanzu package list --available --namespace test-ns --kubeconfig wc-kc-alpha8 
   NAME                                    DISPLAYNAME   SHORTDESCRIPTION                                             
   cert-manager.tkg-standard.tanzu.vmware  cert-manager  This package provides certificate management functionality.  
   fluent-bit.tkg-standard.tanzu.vmware    fluent-bit    This package provides logging functionality.                 
  
   >>> tanzu package list --available cert-manager.tkg-standard.tanzu.vmware --namespace test-ns --kubeconfig wc-kc-alpha8
   NAME                                    VERSION        
   cert-manager.tkg-standard.tanzu.vmware  1.1.0-vmware1  

   ```

13. Delete the repository
   ```sh
   >>> tanzu package repository delete testrepo --namespace test-ns
   Deleted package repository 'testrepo' in namespace 'test-ns'
   ```
