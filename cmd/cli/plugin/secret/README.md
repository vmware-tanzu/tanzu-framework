# secret

Manage secret operations.

## secret registry

Manage registry secret operations. Registry secrets enable the package and package repository consumers to authenticate to and pull images from private registries.

### Usage

tanzu secret registry [command]

```sh
>>> tanzu secret registry --help
Manage registry secret operations. Registry secrets enable the package and package repository consumers to authenticate to and pull images from private registries.

Usage:
  tanzu secret registry [command]

Available Commands:
  add           Creates a v1/Secret resource of type kubernetes.io/dockerconfigjson
  delete        Deletes v1/Secret  resource of type kubernetes.io/dockerconfigjson and the associated SecretExport from the cluster
  list          Lists all v1/Secret of type kubernetes.io/dockerconfigjson and checks for the associated SecretExport with the same name

Flags:
  -h, --help              help for registry secret
      --log-file string   Log file path
      --verbose int32     Number for the log level verbosity(0-9)

Use "tanzu secret registry [command] --help" for more information about a command.
```

### Supported commands

the "secret registry" can be used to:

* add a registry secret
* list registry secrets
* delete a registry secret
* update a registry secret

#### Add a registry secret

   The "add" command creates a v1/Secret of type kubernetes.io/dockerconfigjson.
   In case of specifying the --export-to-all-namespaces flag, a SecretExport resource with the same name will be created, which makes the secret available across all namespaces in the cluster.

   ```sh
   >>> tanzu secret registry add tanzu-net -n test-ns --server registry.pivotal.io --username test-user --password-file pass-file --export-to-all-namespaces
   / Adding registry secret 'test-secret'...
     Added registry secret 'test-secret' into namespace 'test-ns'
   ```

#### List registry secrets

   The "list" command lists a v1/Secret of type kubernetes.io/dockerconfigjson for a specified namespace. If namespace flag is not specified, it lists the secrets from `default` namespace.
   It also supports -A flag to list the secrets across all namespaces.
   In case a SecretExport resource exists with the same name as the Secret, it will check if it is exported `to all namespaces` or `to some namespaces` and display that in EXPORTED column. If not, it will display `not exported` in the EXPORTED column.

##### List registry secrets from specified namespace

   ```sh
   >>> tanzu secret registry list -n test-ns
   / Retrieving registry secrets...
     NAME         REGISTRY                 EXPORTED           AGE
     pkg-dev-reg  registry.pivotal.io      to all namespaces  15d
   ```

##### List registry secrets across all namespaces

   ```sh
   >>> tanzu secret registry list -A
   \ Retrieving registry secrets...
     NAME                          REGISTRY             EXPORTED           AGE  NAMESPACE
     pkg-dev-reg                   registry.pivotal.io  to all namespaces  15d  test-ns
     tanzu-standard-fetch-0        registry.pivotal.io  not exported       15d  tanzu-package-repo-global
     private-repo-fetch-0          registry.pivotal.io  not exported       15d  test-ns
     antrea-fetch-0                registry.pivotal.io  not exported       15d  tkg-system
     metrics-server-fetch-0        registry.pivotal.io  not exported       15d  tkg-system
     tanzu-addons-manager-fetch-0  registry.pivotal.io  not exported       15d  tkg-system
     tanzu-core-fetch-0            registry.pivotal.io  not exported       15d  tkg-system
   ```

##### List registry secrets in json output format

   ```sh
   >>> tanzu secret registry list -n kapp-controller-packaging-global -o json
   [
     {
       "age": "15d",
       "exported": "to all namespaces",
       "name": "pkg-dev-reg",
       "registry": "us-east4-docker.pkg.dev"
     }
   ]
   ```

##### List registry secrets in yaml output format

   ```sh
   >>> tanzu secret registry list -n kapp-controller-packaging-global -o yaml
   - age: 15d
     exported: to all namespaces
     name: pkg-dev-reg
     registry: us-east4-docker.pkg.dev
   ```

##### List registry secrets in table output format

   ```sh
   >>> tanzu secret registry list -n kapp-controller-packaging-global -o table
   / Retrieving registry secrets...
     NAME         REGISTRY                 EXPORTED           AGE
     pkg-dev-reg  us-east4-docker.pkg.dev  to all namespaces  15d
   ```

#### Delete a registry secret

   The "delete" command deletes a v1/Secret of type kubernetes.io/dockerconfigjson from the specified namespace. If no namespace is specified, the secret will be deleted from the default namespace (if existing).
   In case a SecretExport resource with the same name exists, it will be deleted from the namespace as well.

   ```sh
   >>> tanzu secret registry delete test-secret -n test-ns
   Deleting registry secret 'test-secret' from namespace 'test-ns'. Are you sure? [y/N]: y
   \ Deleting registry secret 'test-secret'...
     Deleted registry secret 'test-secret' from namespace 'test-ns'
   ```

#### Update a registry secret

   The "update" command updates the v1/Secret of type kubernetes.io/dockerconfigjson.
   In case of specifying the --export-to-all-namespaces flag, the SecretExport resource will also get updated. Otherwise, there will be no changes in the SecretExport resource.

##### Update a registry secret. There will be no changes in the associated SecretExport resource

   ```sh
   >>> tanzu secret registry update test-secret --username test-user -n test-ns --password-env-var PASSENV
   \ Updating registry secret 'test-secret'...
     Updated registry secret 'test-secret' in namespace 'test-ns'
   ```

##### Update a registry secret with 'export-to-all-namespaces' flag being set

   ```sh
   >>> tanzu secret registry update test-secret--username test-user -n test-ns --password-env-var PASSENV --export-to-all-namespaces=true
   Warning: By specifying --export-to-all-namespaces as true, given secret contents will be available to ALL users in ALL namespaces. Please ensure that included registry credentials allow only read-only access to the registry with minimal necessary scope.
   Are you sure you want to proceed? [y/N]: y

   \ Updating registry secret 'test-secret'...
     Updated registry secret 'test-secret' in namespace 'test-ns'
     Exported registry secret 'test-secret' to all namespaces
   ```

   ```sh
   >>> tanzu secret registry update test-secret --username test-user -n test-ns --password-env-var PASSENV --export-to-all-namespaces
   Warning: By specifying --export-to-all-namespaces as true, given secret contents will be available to ALL users in ALL namespaces. Please ensure that included registry credentials allow only read-only access to the registry with minimal necessary scope.
   Are you sure you want to proceed? [y/N]: y

   \ Updating registry secret 'test-secret'...
     Updated registry secret 'test-secret' in namespace 'test-ns'
     Exported registry secret 'test-secret' to all namespaces
   ```

##### Update a registry secret with 'export-to-all-namespaces' flag being clear. In this case, the associated SecretExport resource will get deleted

   ```sh
   >>> tanzu secret registry update test-secret --username test-user -n test-ns --password-env-var PASSENV --export-to-all-namespaces=false
   Warning: By specifying --export-to-all-namespaces as false, the secret contents will get unexported from ALL namespaces in which it was previously available to.
   Are you sure you want to proceed? [y/N]: y

   \ Updating registry secret 'test-secret'...
     Updated registry secret 'test-secret' in namespace 'test-ns'
     Unexported registry secret 'test-secret' from all namespaces
   ```

## Workflow for adding a private package repository and installation of a private package

You can add a private package repository and install a private package using the following procedure:

1. First, create the namespace in which the secret is getting added to:

   ```sh
   kubectl create namespace <NAMESPACE>
   ```

2. Before adding a private package repository, registry secret should be added to the cluster:

   ```sh
   tanzu secret registry add <SECRET-NAME> --server <PRIVATE-REGISTRY> --username <USERNAME> --namespace <SECRET-NAMESPACE> --password <PASSWORD> -y
   ```

3. In case you want to add the private repository in a different namespace than the namespace in which the secret was added to, you need o export the secret to all other namespaces. Please be aware that by doing so, the given secret contents will be available to ALL users in ALL namespaces. Please ensure that included registry credentials allow only read-only access to the registry with minimal necessary scope. You can export the secret to other namespaces by running:

   ```sh
   tanzu secret registry update <SECRET-NAME> --namespace <SECRET-NAMESPACE> --export-to-all-namespaces=true -y
   ```

4. Add the private package repository to the target namespace in which you want to install the private package by running:

   ```sh
   tanzu package repository add <REPOSITORY-NAME> --url <REPOSITORY-URL> --namespace <TARGET-NAMESPACE> --create-namespace
   ```

5. Verify that the private package repository has been successfully added to the target namespace by running:

   ```sh
   tanzu package repository get <REPOSITORY-NAME> --namespace <TARGET-NAMESPACE>
   ```

6. List the available packages by running:

   ```sh
   tanzu package available list --namespace <TARGET-NAMESPACE>
   ```

7. List version information for the package by running:

   ```sh
   tanzu package available list <PACKAGE-NAME> --namespace <TARGET-NAMESPACE>
   ```

8. Install the private package:

   ```sh
   tanzu package installed create <INSTALLED-PACKAGE-NAME> --package-name <PACKAGE-NAME> --version <PACKAGE-VERSION> --namespace <TARGET-NAMESPACE>
   ```

   Please follow the specific installation instructions for the package in case additional configuration parameters are needed

9. Verify the successful package installation by running:

   ```sh
   tanzu package installed get <INSTALLED-PACKAGE-NAME> --namespace <TARGET-NAMESPAC>
   ```
