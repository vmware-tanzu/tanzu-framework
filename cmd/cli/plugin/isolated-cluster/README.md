# isolated-cluster

## Summary

Enable Tanzu CLI users to copy TKG images from one registry to another that does not have access to the internet.
This is done by following the next steps
1.Copy all associated TKG images to tars on a local disk using this plugin
2.Copy these tar files to the air-gapped network (using a USB drive or other mechanism)
3.Upload all the tar files to the air-gapped registry using this plugin

## Usage

### download-bundle

#### help output

```shell
$ tanzu isolated-cluster download-bundle --help
Download images/bundle into local disk as TAR

Usage:
  tanzu isolated-cluster download-bundle [flags]

Examples:
    # copy image from projects.registry.vmware.com/tkg to /tmp folder
    tanzu isolated-cluster download-bundle --source-repo mirror-registry.test/tkg --tkg-version v2.1.0


Flags:
      ----source-repo         OCI repo where TKG bundles or images are hosted (required)
      --tkg-version           TKG version (required)
      --insecure       Trusts the server certificate without validating it
      --ca-certificate The private repository’s CA certificate
      -h, --help              help for image pull
```

### upload-bundle

#### help output

```shell
$ tanzu isolated-cluster upload-bundle --help
Upload images to private repository

Usage:
  tanzu isolated-cluster upload-bundle [flags]

Examples:
    # push images from tar file to private repository
    tanzu isolated-cluster upload-bundle --source-directory /tmp --ca-certificate cacert

Flags:
      --source-directory            Path to the directory that contains the TAR file  (required)
      --ca-certificate              The private repository’s CA certificate  (optional)
      --destination-repo            Private OCI repository where the images should be hosted in air-gapped (required)
      --insecure        Trusts the server certificate without validating it (optional)
      -h, --help            help for image push
```
