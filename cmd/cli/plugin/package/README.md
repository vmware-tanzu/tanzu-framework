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
3. Apply kapp controller and carvel CRD's from core/cmd/cli/plugin/package/kapp-package/v0.19.0-alpha.9.yaml
4. Build cli
5. Install cli and plugins
6. Use package commands to:
   - add a repository
   - list a repository
   - get a repository status
   - list packages
   - delete a repository
   
   For alpha.9 release, use the following image package bundles for testing:
   ```
   1. Repository URL ( Simple App ) => index.docker.io/k8slt/kc-e2e-repo-bundle@sha256:388d353574446eea0bba4e3f656079963660704e0d474fbc87b3a9bc6efb1688
   2. Repository URL ( Fluent Bit, Cert Manager ) => projects-stg.registry.vmware.com/tkg/tkgextensions-dev/tkg-standard-repo:d60aeb6
   3. Repository URL ( Contour, Harbor, Cert Manager ) => projects-stg.registry.vmware.com/tkg/tkgextensions-dev/tkg-standard-repo@sha256:e5a307190145ccb92eecf86d3c863d3d37e82c7f8c3383ecd8d2d5640e9b9649
   ```
   
   Here is an example workflow
   
7. Add a repository

   ```sh
   >>> tanzu package repository add testrepo index.docker.io/k8slt/kc-e2e-repo-bundle@sha256:388d353574446eea0bba4e3f656079963660704e0d474fbc87b3a9bc6efb1688
   Package Repository 'testrepo' added
   ```

8. Update a repository

   ```sh
   >>> tanzu package repository update testrepo projects-stg.registry.vmware.com/tkg/shivaani/package-bundle:1.0.0
   Successfully updated package repository 'testrepo'
   ```

9. List the repository
   ```sh
   >>> tanzu package repository list
   NAME      REPOSITORY                                                                                                        STATUS               DETAILS  
   testrepo  index.docker.io/k8slt/kc-e2e-repo-bundle@sha256:388d353574446eea0bba4e3f656079963660704e0d474fbc87b3a9bc6efb1688  Reconcile succeeded          
   ```

10. Install a package

    Example 1: Install the specified version for package name "fluent-bit.tkg-standard.tanzu.vmware" and while providing the values.yaml file.
    ```sh
    
    >>> tanzu package install fluentbit --package-name fluent-bit.tkg-standard.tanzu.vmware --namespace test-ns --create-namespace --version 1.7.5-vmware1 --values-file values.yaml
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
    Uninstalled package 'fluentbit' in namespace 'test-ns'
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
   >>> tanzu package list --available
   NAME                                    DISPLAYNAME        SHORTDESCRIPTION                                             
   cert-manager.tkg-standard.tanzu.vmware  cert-manager       This package provides certificate management functionality.  
   fluent-bit.tkg-standard.tanzu.vmware    fluent-bit         This package provides logging functionality.                 
   simple-app.corp.com                     simple-app v1.0.0  Simple app consisting of a k8s deployment and service
  
   #List all available packages for package name
   >>> tanzu package list --available fluent-bit.tkg-standard.tanzu.vmware
   or
   >>> tanzu package list fluent-bit.tkg-standard.tanzu.vmware --available 
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
  
   >>> tanzu package list --available --kubeconfig wc-kc-alpha8 
   NAME                                    DISPLAYNAME   SHORTDESCRIPTION                                             
   cert-manager.tkg-standard.tanzu.vmware  cert-manager  This package provides certificate management functionality.  
   fluent-bit.tkg-standard.tanzu.vmware    fluent-bit    This package provides logging functionality.                 
  
   >>> tanzu package list --available cert-manager.tkg-standard.tanzu.vmware --kubeconfig wc-kc-alpha8
   NAME                                    VERSION        
   cert-manager.tkg-standard.tanzu.vmware  1.1.0-vmware1  

   ```

13. Delete the repository
   ```sh
   >>> tanzu package repository delete testrepo
   Successfully deleted package repository 'testrepo'
   ```
