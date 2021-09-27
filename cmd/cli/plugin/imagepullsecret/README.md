# imagepullsecret

Manage image pull secret operations. Image pull secrets enable the package and package repository consumers to authenticate to private registries.

## Usage

tanzu imagepullsecret [command]

```sh
>>> tanzu imagepullsecret --help
Manage image pull secret operations. Image pull secrets enable the package and package repository consumers to authenticate to private registries.

Usage:
  tanzu imagepullsecret [command]

Available Commands:
  add           Creates a v1/Secret resource of type kubernetes.io/dockerconfigjson
  delete        Deletes v1/Secret  resource of type kubernetes.io/dockerconfigjson and the associated SecretExport from the cluster
  list          Lists all v1/Secret of type kubernetes.io/dockerconfigjson and checks for the associated SecretExport by the same name

Flags:
  -h, --help              help for imagepullsecret
      --log-file string   Log file path
      --verbose int32     Number for the log level verbosity(0-9)

Use "tanzu imagepullsecret [command] --help" for more information about a command.
```

## Supported commands

imagepullsecret plugin can be used to:

* add a secret
* list secrets
* delete a secret

1. Add an image pull secret
   The "add" command creates a v1/Secret of type kubernetes.io/dockerconfigjson.
   In case of specifying the --export-to-all-namespaces flag, a SecretExport resource with the same name will be created, which makes the secret available across all namespaces in the cluster.

   ```sh
   >>> tanzu imagepullsecret add tanzu-net -n test-ns --registry registry.pivotal.io --username test-user --password-file pass-file --export-to-all-namespaces
   **/** Adding image pull secret 'test-secret'...
      Added image pull secret 'test-secret' into namespace 'test-ns'
   ```

2. List an image pull secret

   The "list" command lists a v1/Secret of type kubernetes.io/dockerconfigjson for a specified namespace. If namespace flag is not specified, it lists the secrets from `default` namespace.
   It also supports -A flag to list the secrets across all namespaces.
   In case a SecretExport resource exists with the same name as the Secret, it will check if it is exported `to all namespaces` or `to some namespaces` and display that in EXPORTED column. If not, it will display `not exported` in the EXPORTED column.

   ```sh
   # List image pull secrets from specified namespace
   >>> tanzu imagepullsecret list -n test-ns
   **/** Retrieving image pull secrets...
     NAME         REGISTRY                 EXPORTED           AGE
     pkg-dev-reg  registry.pivotal.io      to all namespaces  15d

   # List image pull secrets across all namespaces
   >>> tanzu imagepullsecret list -A
   \ Retrieving image pull secrets...
     NAME                          REGISTRY             EXPORTED           AGE  NAMESPACE
     pkg-dev-reg                   registry.pivotal.io  to all namespaces  15d  test-ns
     tanzu-standard-fetch-0        registry.pivotal.io  not exported       15d  tanzu-package-repo-global
     private-repo-fetch-0          registry.pivotal.io  not exported       15d  test-ns
     antrea-fetch-0                registry.pivotal.io  not exported       15d  tkg-system
     metrics-server-fetch-0        registry.pivotal.io  not exported       15d  tkg-system
     tanzu-addons-manager-fetch-0  registry.pivotal.io  not exported       15d  tkg-system
     tanzu-core-fetch-0            registry.pivotal.io  not exported       15d  tkg-system

   # List image pull secrets in json output format
   >>> tanzu imagepullsecret list -n kapp-controller-packaging-global -o json
   [
     {
       "age": "15d",
       "exported": "to all namespaces",
       "name": "pkg-dev-reg",
       "registry": "us-east4-docker.pkg.dev"
     }
   ]

   # List image pull secrets in json output format
   >>> tanzu imagepullsecret list -n kapp-controller-packaging-global -o yaml
   - age: 15d
     exported: to all namespaces
     name: pkg-dev-reg
     registry: us-east4-docker.pkg.dev

   # List image pull secrets in json output format
   >>> tanzu imagepullsecret list -n kapp-controller-packaging-global -o table
   / Retrieving image pull secrets...
     NAME         REGISTRY                 EXPORTED           AGE
     pkg-dev-reg  us-east4-docker.pkg.dev  to all namespaces  15d
   ```

3. Delete an image pull secret

   The "delete" command deletes a v1/Secret of type kubernetes.io/dockerconfigjson from the specified namespace. If no namespace is specified, the secret will be deleted from the default namespace (if existing).
   In case a SecretExport resource with the same name exists, it will be deleted from the namespace as well.

   ```sh
   >>> tanzu imagepullsecret delete test-secret -n test-ns
   Deleting image pull secret 'test-secret' from namespace 'test-ns'. Are you sure? [y/N]: y
   **\** Deleting image pull secret 'test-secret'...
      Deleted image pull secret 'test-secret' from namespace 'test-ns'
   ```
