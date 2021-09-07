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
- add a secret
- list secrets
- delete a secret

1. Add an image pull secret
   The "add" command creates a v1/Secret of type kubernetes.io/dockerconfigjson. 
   In case of specifying the --export-to-all-namespaces flag, a SecretExport resource with the same name will be created, which makes the secret available across all namespaces in the cluster.

   ```sh
   >>> tanzu imagepullsecret add tanzu-net -n test-ns --registry registry.pivotal.io --username test-user --password-file pass-file --export-to-all-namespaces
   **/** Adding image pull secret 'test-secret'...
      Added image pull secret 'test-secret' into namespace 'test-ns'
   ```

2. Delete an image pull secret
   The "delete" command deletes a v1/Secret of type kubernetes.io/dockerconfigjson from the specified namespace. If no namespace is specified, the secret will be deleted from the default namespace (if existing).
   In case a SecretExport resource with the same name exists, it will be deleted from the namespace as well.

   ```sh
   >>> tanzu imagepullsecret delete test-secret -n test-ns
   Deleting image pull secret 'test-secret' from namespace 'test-ns'. Are you sure? [y/N]: y
   **\** Deleting image pull secret 'test-secret'...
      Deleted image pull secret 'test-secret' from namespace 'test-ns'
   ```
