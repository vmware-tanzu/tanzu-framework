# Package

Manage package lifecycle operations.

## Usage

Package and Repository Operations ( Subject to Change ):

```sh
>>> tanzu package --help
Tanzu package management (available, install, installed, repository)

Usage:
  tanzu package [flags]
  tanzu package [command]

Available Commands:
  available     Manage available packages (get, list)
  completion    Generate the autocompletion script for the specified shell
  install       Install package
  installed     Manage installed packages (create, delete, get, list, update)
  repository    Manage package repositories (add, delete, get, list, update)

Flags:
      --column strings              Filter to show only given columns
      --debug                       Include debug output
  -h, --help                        help for package
      --kube-api-burst int          Set Kubernetes API client burst limit (default 1000)
      --kube-api-qps float32        Set Kubernetes API client QPS limit (default 1000)
      --kubeconfig string           Path to the kubeconfig file ($TANZU_KUBECONFIG)
      --kubeconfig-context string   Kubeconfig context override ($TANZU_KUBECONFIG_CONTEXT)
      --kubeconfig-yaml string      Kubeconfig contents as YAML ($TANZU_KUBECONFIG_YAML)
      --tty                         Force TTY-like output (default true)
  -y, --yes                         Assume yes for any prompt

Use "tanzu package [command] --help" for more information about a command.
```

```sh
>>> tanzu package repository --help
Manage package repositories (add, delete, get, list, update)

Usage:
  tanzu package repository [flags]
  tanzu package repository [command]

Aliases:
  repository, repo, r

Available Commands:
  add         Add a package repository
  delete      Delete a package repository
  get         Get details for a package repository
  list        List package repositories in a namespace
  update      Update a package repository

Flags:
  -h, --help   help for repository

Global Flags:
      --column strings              Filter to show only given columns
      --debug                       Include debug output
      --kube-api-burst int          Set Kubernetes API client burst limit (default 1000)
      --kube-api-qps float32        Set Kubernetes API client QPS limit (default 1000)
      --kubeconfig string           Path to the kubeconfig file ($TANZU_KUBECONFIG)
      --kubeconfig-context string   Kubeconfig context override ($TANZU_KUBECONFIG_CONTEXT)
      --kubeconfig-yaml string      Kubeconfig contents as YAML ($TANZU_KUBECONFIG_YAML)
      --tty                         Force TTY-like output (default true)
  -y, --yes                         Assume yes for any prompt

Use "tanzu package repository [command] --help" for more information about a command.
```

## Test

1. Create a management cluster using latest tanzu cli

2. Use package commands to:
   * add a repository
   * list a repository
   * get a repository status
   * list packages
   * get a package information
   * get an installed package information
   * update a package
   * delete a repository

   Use the following image package bundles for testing:

   | S.no |                        Repository URL                                    |
   | :----| :------------------------------------------------------------------------|
   |  1.  |  projects-stg.registry.vmware.com/tkg/test-packages/test-repo:v1.0.0     |
   |  2.  |  projects-stg.registry.vmware.com/tkg/test-packages/standard-repo:v1.0.0 |

   Here is an example workflow

3. Add a repository

   ```sh
   >>> tanzu package repository add standard-repo --url projects-stg.registry.vmware.com/tkg/test-packages/standard-repo:v1.0.0 -n test-ns
   Waiting for package repository to be added
   4:04:00PM: packagerepository/standard-repo (packaging.carvel.dev/v1alpha1) namespace: test-ns: Reconciling
   4:04:14PM: packagerepository/standard-repo (packaging.carvel.dev/v1alpha1) namespace: test-ns: ReconcileSucceeded
   ```

4. Get repository status

   ```sh
   >>> tanzu package repository get standard-repo -n test-ns
   NAMESPACE:               test-ns
   NAME:                    standard-repo
   SOURCE:                  (imgpkg) projects-stg.registry.vmware.com/tkg/test-packages/standard-repo:v1.0.0
   STATUS:                  Reconcile succeeded
   CONDITIONS:              - type: ReconcileSucceeded
     status: "True"
     reason: ""
     message: ""
   ```

5. Update a repository

   ```sh
   >>> tanzu package repository update standard-repo --url projects-stg.registry.vmware.com/tkg/test-packages/standard-repo:v1.0.0 -n test-ns
   Waiting for package repository to be updated
   4:13:30PM: packagerepository/standard-repo (packaging.carvel.dev/v1alpha1) namespace: test-ns: ReconcileSucceeded
   ```

6. List the repository

   ```sh
   >>> tanzu package repository list -A
   NAMESPACE                  NAME            SOURCE                                                                            STATUS
   test-ns                    repo            (imgpkg) projects-stg.registry.vmware.com/tkg/test-packages/test-repo:v1.0.0      Reconcile succeeded
   test-ns                    standard-repo   (imgpkg) projects-stg.registry.vmware.com/tkg/test-packages/standard-repo:v1.0.0  Reconcile succeeded
   ```

7. Get information of a package

   Example 1: Get detailed information of a package

   ```sh
   >>> tanzu package available get contour.tanzu.vmware.com/1.15.1+vmware.1-tkg.1 --namespace test-ns
     NAME:                         contour.tanzu.vmware.com
     DISPLAY-NAME:                 contour
     CATEGORIES:                   - ingress
     SHORT-DESCRIPTION:            This package provides ingress functionality.
     LONG-DESCRIPTION:             This package provides ingress functionality.
     PROVIDER:
     MAINTAINERS:
     SUPPORT-DESCRIPTION:
     VERSION:                      1.15.1+vmware.1-tkg.1
     RELEASED-AT:                  0001-01-01 00:00:00 +0000 UTC
     MIN-CAPACITY-REQUIREMENTS:
     RELEASE-NOTES:
     LICENSES:
   ```

   Example 2: Get openAPI schema of a package

   ```sh
   >>> tanzu package available get external-dns.tanzu.vmware.com/0.8.0+vmware.1-tkg.1 -n external-dns --values-schema
    KEY                         DEFAULT                         TYPE    DESCRIPTION
    deployment.args                                             array   List of arguments passed via command-line to external-dns.
                                                                         For more guidance on configuration options for your
                                                                        desired DNS provider, consult the ExternalDNS docs at
                                                                        https://github.com/kubernetes-sigs/external-dns#running-externaldns

    deployment.env                                              array   List of environment variables to set in the external-dns container.
    deployment.securityContext                                          SecurityContext defines the security options the
                                                                        external-dns container should be run with. More info:
                                                                        https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
    deployment.volumeMounts                                     array   Pod volumes to mount into the external-dns container's filesystem.
    deployment.volumes                                          array   List of volumes that can be mounted by containers belonging to the external-dns
                                                                        pod. More info: https://kubernetes.io/docs/concepts/storage/volumes
    namespace                   tanzu-system-service-discovery  string  The namespace in which to deploy ExternalDNS.
   ```

   Example 3: Generate default values.yaml for a package

   ```sh
   >>> tanzu package available get contour.tanzu.vmware.com/1.18.2+vmware.1-tkg.1 --default-values-file-output contour-default-values.yaml
    Created default values file at contour-default-values.yaml
    NAME:                         contour.tanzu.vmware.com
    DISPLAY-NAME:                 contour
    CATEGORIES:                   - ingress
    SHORT-DESCRIPTION:            An ingress controller
    LONG-DESCRIPTION:             An Envoy-based ingress controller that supports dynamic configuration updates
    and multi-team ingress delegation. See https://projectcontour.io for more
    information.
    PROVIDER:                     VMware
    MAINTAINERS:                  - name: Steve Kriss
    - name: Steve Sloka
    - name: Nick Young
    - name: Sunjay Bhatia
    - name: Nicholas Seemiller
    SUPPORT-DESCRIPTION:          Support provided by VMware for deployment on TKG 1.4+ clusters. Best-effort
    support for deployment on any conformant Kubernetes cluster. Contact support by
    opening a support request via VMware Cloud Services or my.vmware.com.
    VERSION:                      1.18.2+vmware.1-tkg.1
    RELEASED-AT:                  2021-10-05 05:30:00 +0530 IST
    MIN-CAPACITY-REQUIREMENTS:    Varies significantly based on number of Services, Ingresses/HTTPProxies, etc. A
    starting point is 128MB RAM and 0.5 CPU for each Contour and Envoy pod, but this
    can and should be tuned based on observed usage.
    RELEASE-NOTES:                contour 1.18.2 https://github.com/projectcontour/contour/releases/tag/v1.18.2
    LICENSES:                     VMware’s End User License Agreement (Underlying OSS license: Apache License 2.0)
   ```

   contour-default-values.yaml

   ```yaml
    # certificates:
    #   duration: 8760h
    #   renewBefore: 360h
    # contour:
    #   logLevel: info
    #   replicas: 2
    #   useProxyProtocol: false
    # envoy:
    #   hostNetwork: false
    #   hostPorts:
    #     enable: true
    #     http: 80
    #     https: 443
    #   logLevel: info
    #   service:
    #     aws:
    #       LBType: classic
    #     externalTrafficPolicy: Cluster
    #   terminationGracePeriodSeconds: 300
    # namespace: tanzu-system-ingress  
   ```

8. Install a package

   Example 1: Install the specified version for package name "fluent-bit.tkg-standard.tanzu.vmware", while providing the values.yaml file and without waiting for package reconciliation to complete

   ```sh

   >>> tanzu package install fluentbit --package fluent-bit.tanzu.vmware.com --namespace test-ns --version 1.7.5+vmware.1-tkg.1 --values-file values.yaml --wait=false
   Creating service account 'fluentbit-test-ns-sa'
   Creating cluster admin role 'fluentbit-test-ns-cluster-role'
   Creating cluster role binding 'fluentbit-test-ns-cluster-rolebinding'
   Creating secret 'fluentbit-test-ns-values'
   Creating package install resource
   ```

   An example values.yaml is as follows:

   ```yaml
   fluent_bit:
      config:
        outputs: |
          [OUTPUT]
            Name     stdout
            Match    *
    ```

    Example 2: Install the latest version for package name "contour.tanzu.vmware.com".

    ```sh
    >>> tanzu package install contour-pkg --package contour.tanzu.vmware.com --namespace test-ns --version 1.15.1+vmware.1-tkg.1
    11:02:01AM: Creating service account 'contour-pkg-test-ns-sa'
    11:02:01AM: Creating cluster admin role 'contour-pkg-test-ns-cluster-role'
    11:02:01AM: Creating cluster role binding 'contour-pkg-test-ns-cluster-rolebinding'
    11:02:01AM: Creating package install resource
    11:02:01AM: Waiting for PackageInstall reconciliation for 'contour-pkg'
    11:02:02AM: Fetch started (2s ago)
    11:02:04AM: Fetching
            | apiVersion: vendir.k14s.io/v1alpha1
            | directories:
            | - contents:
            |   - imgpkgBundle:
            |       image: projects-stg.registry.vmware.com/tkg/tkgextensions-dev/contour@sha256:d9be6b196fc35e354f7405cef5edc342858869e39b2f96c1f621721ef7020833
            |     path: .
            |   path: "0"
            | kind: LockConfig
            |
    11:02:04AM: Fetch succeeded
    11:02:05AM: Template succeeded
    11:02:05AM: Deploy started (2s ago)
    11:02:07AM: Deploying
            | Target cluster 'https://10.96.0.1:443' (nodes: minikube)
            | Changes
            | Namespace             Name                                         Kind                      Conds.  Age  Op      Op st.  Wait to    Rs  Ri

            ...

    11:04:14AM: App reconciled
    ```

9. Get information of an installed package

   Example 1: Get information of an installed package

   ```sh
   >>> tanzu package installed get contour-pkg --namespace test-ns
   NAMESPACE:               test-ns
   NAME:                    contour-pkg
   PACKAGE-NAME:            contour.tanzu.vmware.com
   PACKAGE-VERSION:         1.15.1+vmware.1-tkg.1
   STATUS:                  Reconcile succeeded
   CONDITIONS:              - type: ReconcileSucceeded
     status: "True"
     reason: ""
     message: ""
   ```

   Example 2: Get data value secret of an installed package and save it to file (example: config.yaml)

   ```sh
   >>> tanzu package installed get fluent-bit --namespace test-ns --values-file-output config.yaml

   cat config.yaml
   fluent_bit:
     config:
       outputs: |
         [OUTPUT]
           Name     stdout
           Match    *
   ```

10. Pause and kick a package

    Example 1: Pause a package

    ```sh
    >>> tanzu package installed pause contour-pkg --namespace test-ns
    Pausing reconciliation for package install 'contour-pkg' in namespace 'test-ns'
    Continue? [yN]: y

    >>> tanzu package installed get contour-pkg --namespace test-ns
    NAMESPACE:          test-ns
    NAME:               contour-pkg
    PACKAGE-NAME:       contour.tanzu.vmware.com
    PACKAGE-VERSION:    1.15.1+vmware.1-tkg.1
    STATUS:             Paused
    CONDITIONS:         - type: ReconcileSucceeded
      status: "True"
      reason: ""
      message: ""
    ```

    Example 2: Kick a package

    ```sh
    >>> tanzu package installed kick contour-pkg --namespace test-ns
    Triggering reconciliation for package install 'contour-pkg' in namespace 'test-ns'
    Continue? [yN]: y
    2:40:55PM: Waiting for PackageInstall reconciliation for 'contour-pkg'
    2:40:55PM: Waiting for generation 3 to be observed
    2:41:00PM: Fetching
            | apiVersion: vendir.k14s.io/v1alpha1
            | directories:
            | - contents:
            |   - imgpkgBundle:
            |       image: projects-stg.registry.vmware.com/tkg/tkgextensions-dev/contour@sha256:d9be6b196fc35e354f7405cef5edc342858869e39b2f96c1f621721ef7020833
            |     path: .
            |   path: "0"
            | kind: LockConfig
            |
    2:41:00PM: Fetch succeeded
    2:41:00PM: Template succeeded
    2:41:00PM: Deploy started (2s ago)
    2:41:02PM: Deploying
        | Target cluster 'https://10.96.0.1:443' (nodes: minikube)
        | Changes
        | Namespace  Name  Kind  Conds.  Age  Op  Op st.  Wait to  Rs  Ri
        | Op:      0 create, 0 delete, 0 update, 0 noop, 0 exists
        | Wait to: 0 reconcile, 0 delete, 0 noop
        | Succeeded
    2:41:03PM: App reconciled

    >>> tanzu package installed get contour-pkg --namespace test-ns
    NAMESPACE:          test-ns
    NAME:               contour-pkg
    PACKAGE-NAME:       contour.tanzu.vmware.com
    PACKAGE-VERSION:    1.15.1+vmware.1-tkg.1
    STATUS:             Reconcile succeeded
    CONDITIONS:         - type: ReconcileSucceeded
      status: "True"
      reason: ""
      message: ""
    ```

11. Update a package

    Example 1: Update a package with different version

    ```sh
    >>> tanzu package installed update carvel-test-pkg --version 2.0.0 --namespace test-ns
    2:50:58PM: Updating package install for 'test' in namespace 'default'
    2:50:58PM: Waiting for PackageInstall reconciliation for 'test'
    2:50:58PM: Waiting for generation 6 to be observed
    2:50:59PM: Fetching
            | apiVersion: vendir.k14s.io/v1alpha1
            | directories:
            | - contents:
            |   - imgpkgBundle:
            |       image: index.docker.io/k8slt/kctrl-example-pkg@sha256:73713d922b5f561c0db2a7ea5f4f6384f7d2d6289886f8400a8aaf5e8fdf134a
            |     path: .
            |   path: "0"
            | kind: LockConfig
            |
    2:50:59PM: Fetch succeeded
    2:50:59PM: Template succeeded
    2:50:59PM: Deploy started (2s ago)
    2:51:01PM: Deploying
            | Target cluster 'https://10.96.0.1:443' (nodes: minikube)
            | Changes

    ...

            | Succeeded
    2:51:03PM: App reconciled
    ```

    Example 2: Update an installed package with providing values.yaml file

    ```sh
    >>> tanzu package installed update fluent-bit --version 1.7.5+vmware.1-tkg.1 --namespace test-ns --values-file values.yaml
    Getting package install for 'fluent-bit'
    Creating secret 'fluent-bit-test-ns-values'
    Updating package install for 'fluent-bit'
    Waiting for PackageInstall reconciliation for 'fluent-bit'
    11:02:02AM: Fetch started (2s ago)
    11:02:04AM: Fetching

    ...

    11:03:04AM: App reconciled
    ```

    An example values.yaml is as follows:

    ```yaml
    fluent_bit:
       config:
         outputs: |
           [OUTPUT]
             Name     stdout
             Match    /
    ```

12. Uninstall a package

    ```sh
    >>> tanzu package installed delete contour-pkg --namespace test-ns
    Delete package install 'contour-pkg' from namespace 'test-ns'
    Continue? [yN]: y
    2:53:15PM: Deleting package install 'contour-pkg' from namespace 'test-ns'
    2:53:15PM: Waiting for deletion of package install 'contour-pkg' from namespace 'test-ns'
    2:53:15PM: Waiting for generation 4 to be observed
    2:53:15PM: Delete started (2s ago)
    2:53:17PM: Deleting
            | Target cluster 'https://10.96.0.1:443' (nodes: minikube)
            | Changes
            | Namespace             Name                                         Kind                      Conds.  Age  Op      Op st.  Wait to  Rs       Ri
            | (cluster)

    ...

    2:55:16PM: App 'contour-pkg' in namespace 'test-ns' deleted
    2:55:16PM: packageinstall/contour-pkg (packaging.carvel.dev/v1alpha1) namespace: test-ns: DeletionSucceeded
    2:55:16PM: Deleting 'ClusterRoleBinding': contour-pkg-test-ns-cluster-rolebinding
    2:55:16PM: Deleting 'ServiceAccount': contour-pkg-test-ns-sa
    2:55:16PM: Deleting 'ClusterRole': contour-pkg-test-ns-cluster-role
    ```

13. List the packages

    ```sh
    #List installed packages in the default namespace
    >>> tanzu package installed list
    NAME  PACKAGE-NAME         PACKAGE-VERSION  STATUS
    test  pkg.test.carvel.dev  2.0.0            Reconcile succeeded

    #List installed packages across all namespaces
    >>> tanzu package installed list -A
    NAMESPACE  NAME        PACKAGE-NAME                   PACKAGE-VERSION       STATUS
    default    test        pkg.test.carvel.dev            2.0.0                 Reconcile succeeded
    test-ns    cert-mng    cert-manager.tanzu.vmware.com  1.5.3+vmware.2-tkg.1  Reconcile succeeded
    test-ns    fluent-bit  fluent-bit.tanzu.vmware.com    1.7.5+vmware.1-tkg.1  Reconcile succeeded


    #List installed packages in user provided namespace
    >>> tanzu package installed list --namespace test-ns
    NAMESPACE  NAME        PACKAGE-NAME                   PACKAGE-VERSION       STATUS
    test-ns    cert-mng    cert-manager.tanzu.vmware.com  1.5.3+vmware.2-tkg.1  Reconcile succeeded
    test-ns    fluent-bit  fluent-bit.tanzu.vmware.com    1.7.5+vmware.1-tkg.1  Reconcile succeeded

    #List all available package CRs in default namespace
    >>> tanzu package available list
    NAME                 DISPLAY-NAME
    pkg.test.carvel.dev  Carvel Test Package

    #List all available package CRs across all namespace
    >>> tanzu package available list -A
    NAMESPACE  NAME                           DISPLAY-NAME
    default    pkg.test.carvel.dev            Carvel Test Package
    test-ns    cert-manager.tanzu.vmware.com  cert-manager
    test-ns    contour.tanzu.vmware.com       contour
    test-ns    external-dns.tanzu.vmware.com  external-dns
    test-ns    fluent-bit.tanzu.vmware.com    fluent-bit
    test-ns    grafana.tanzu.vmware.com       grafana
    test-ns    harbor.tanzu.vmware.com        harbor
    test-ns    multus-cni.tanzu.vmware.com    multus-cni
    test-ns    prometheus.tanzu.vmware.com    prometheus

    #List all available packages for package name
    >>> tanzu package available list contour.tanzu.vmware.com -A
    NAMESPACE  NAME                      VERSION                RELEASED-AT
    test-ns    contour.tanzu.vmware.com  1.15.1+vmware.1-tkg.1  0001-01-01 00:00:00 +0000 UTC
    ```

14. Delete the repository

    ```sh
    >>> tanzu package repository delete standard-repo --namespace test-ns
    Deleting package repository 'standard-repo' in namespace 'test-ns'
    Continue? [yN]: y
    Waiting for deletion to be completed...
    11:47:25PM: packagerepository/standard-repo (packaging.carvel.dev/v1alpha1) namespace: test-ns: Deleting
    11:47:38PM: packagerepository/standard-repo (packaging.carvel.dev/v1alpha1) namespace: test-ns: DeletionSucceeded
    ```

All the above commands are equipped with --kubeconfig flag to perform the package and repository operations on the desired cluster.

Example:

```sh
>>> tanzu package installed list -A --kubeconfig wc-kc-alpha8
 NAMESPACE  NAME        PACKAGE-NAME                   PACKAGE-VERSION       STATUS
 test-ns    cert-mng    cert-manager.tanzu.vmware.com  1.5.3+vmware.2-tkg.1  Reconcile succeeded
 test-ns    fluent-bit  fluent-bit.tanzu.vmware.com    1.7.5+vmware.1-tkg.1  Reconcile succeeded
