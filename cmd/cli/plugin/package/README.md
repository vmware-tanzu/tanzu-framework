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
    available   Manage available packages
    install     Install a package
    installed   Manage installed packages
    repository  Manage registered package repositories


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
   - get a package information
   - get an installed package information
   - update a package
   - delete a repository
   
   For v0.20.0-rc.1 release, use the following image package bundles for testing:
   1. Repository URL  => index.docker.io/k8slt/kc-e2e-test-repo@sha256:62d187c044fd6a5c57ac870733fe4413ebf7e2909d8b6267707c5dd2080821e6
   
   Here is an example workflow
   
7. Add a repository

   ```sh
   >>> tanzu package repository add testrepo --url index.docker.io/k8slt/kc-e2e-test-repo@sha256:62d187c044fd6a5c57ac870733fe4413ebf7e2909d8b6267707c5dd2080821e6 -n test-ns --create-namespace
   Package Repository 'testrepo' added
   ```

8. Get repository status
   ```sh
   >>> tanzu package repository get testrepo -n test-ns
   NAME      VERSION  REPOSITORY                                                                                                        STATUS               REASON
   testrepo  627590   index.docker.io/k8slt/kc-e2e-repo-bundle@sha256:388d353574446eea0bba4e3f656079963660704e0d474fbc87b3a9bc6efb1688  Reconcile succeeded
   ```

9. Update a repository

   ```sh
   >>> tanzu package repository update testrepo2 --url projects-stg.registry.vmware.com/tkg/shivaani/package-bundle:1.0.0 -n test-ns
   Updated package repository 'testrepo2'
   ```

10. List the repository
   ```sh
   >>> tanzu package repository list -n test-ns
   NAME      REPOSITORY                                                                                                        STATUS               DETAILS  
   testrepo  index.docker.io/k8slt/kc-e2e-test-repo@sha256:62d187c044fd6a5c57ac870733fe4413ebf7e2909d8b6267707c5dd2080821e6    Reconcile succeeded          
   ```

11. Get information of a package
   ```sh
   >>> tanzu package available get simple-app.corp.com/version 1.0.0
    NAME: simple-app.corp.com
    VERSION: 1.0.0
    RELEASED-AT: 2021-Jun-23 10:00:00Z
    DISPLAY-NAME: Simple app
    SHORT-DESCRIPTION: Simple app consisting of a k8s deployment...
    PACKAGE-PROVIDER:
    MINIMUM-CAPACITY-REQUIREMENTS:
    LONG-DESCRIPTION: ...
    MAINTAINERS: ...
    RELEASE-NOTES: ...
    LICENSE: Apache 2.0

   ```

12. Install a package

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

13. Get information of an installed package
   ```sh
   >>> tanzu package installed get simple-app --namespace test-ns
   NAME        VERSION     NAMESPACE  STATUS                REASON
   simple-app  3.0.0-rc.1  test-ns    Reconcile succeeded
   ```

14. Update a package

    Example 1: Update a package with different version
    ```sh
    >>> tanzu package update fluent-bit --namespace test-ns --version 2.0.0
    Updated package 'fluent-bit' in namespace 'test-ns'
    ```

    Example 2: Update a package which is not installed
    ```sh
    >>> tanzu package installed update fluent-bit --package-name fluent-bit.tkg-standard.tanzu.vmware --version 1.0.0 --namespace test-ns --install
    Updated package 'fluent-bit' in namespace 'test-ns'
    ```

15. Uninstall a package

    ```sh
    >>> tanzu package installed delete fluentbit --namespace test-ns
    Uninstalling package 'fluentbit' from namespace 'test-ns'
    Uninstalled package 'fluentbit' from namespace 'test-ns'
    ```

16. List the packages

   ```sh
   #List installed packages in the default namespace
   >>> tanzu installed list
   NAME  DISPLAY-NAME  SHORT-DESCRIPTION
  
   #List installed packages across all namespaces
   >>> tanzu installed list -A
   NAME       DISPLAY-NAME  SHORT-DESCRIPTION  NAMESPACE

  
   #List installed packages in user provided namespace
   >>> tanzu package installed list --namespace tanzu-logging
   NAME        VERSION        
   fluent-bit  1.7.5-vmware1
  
   #List all available package CRs
   >>> tanzu package available list
   NAME                                    DISPLAYNAME        SHORTDESCRIPTION                                             
   cert-manager.tkg-standard.tanzu.vmware  cert-manager       This package provides certificate management functionality.  
   fluent-bit.tkg-standard.tanzu.vmware    fluent-bit         This package provides logging functionality.                 
   simple-app.corp.com                     simple-app v1.0.0  Simple app consisting of a k8s deployment and service
  
   #List all available packages for package name
   >>> tanzu package available list fluent-bit.tkg-standard.tanzu.vmware --namespace test-ns

  
   With kubeconfig flag

   >>> tanzu package installed list --kubeconfig wc-kc-alpha8                       
   NAME  VERSION  
  
   >>> tanzu package installed list -A --kubeconfig wc-kc-alpha8
   NAME  VERSION        NAMESPACE        
   mycm  1.1.0-vmware1  test-1           
   myfb  1.7.5-vmware1  test-logging-wc  
      
   >>> tanzu package installed list --namespace test-logging-wc --kubeconfig wc-kc-alpha8
   NAME  VERSION        
   myfb  1.7.5-vmware1  
  
   >>> tanzu package available list --namespace test-ns --kubeconfig wc-kc-alpha8 
   NAME                                    DISPLAYNAME   SHORTDESCRIPTION                                             
   cert-manager.tkg-standard.tanzu.vmware  cert-manager  This package provides certificate management functionality.  
   fluent-bit.tkg-standard.tanzu.vmware    fluent-bit    This package provides logging functionality.                 
  
   >>> tanzu package available list cert-manager.tkg-standard.tanzu.vmware --namespace test-ns --kubeconfig wc-kc-alpha8
   NAME                                    VERSION        
   cert-manager.tkg-standard.tanzu.vmware  1.1.0-vmware1  

   ```

17. Delete the repository
   ```sh
   >>> tanzu package repository delete testrepo --namespace test-ns
   Deleted package repository 'testrepo' in namespace 'test-ns'
   ```
