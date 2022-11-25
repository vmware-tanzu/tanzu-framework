# Get Available TKr

## Get available Tanzu Kubernetes releases

```sh
tanzu kubernetes-release get -h
Usage:
  tanzu kubernetes-release get TKR_NAME [flags]

Flags:
  -h, --help   help for get
```

### Sample command and output

```sh
tanzu kubernetes-release get
  NAME                VERSION                         COMPATIBLE     UPGRADEAVAILABLE
  v1.16.3---vmware.2  v1.16.3+vmware.2-tkg.1            False             True
  v1.17.13---vmware.1 v1.17.13+vmware.1-tkg.1           True              True
  v1.18.2---vmware.1  v1.18.2+vmware.1-tkg.1            True              True
  v1.18.6---vmware.1  v1.18.6+vmware.1-tkg.1            True              True
  v1.19.3---vmware.1  v1.19.3+vmware.1-tkg.1            True              True
  v1.19.3---vmware.2  v1.19.3+vmware.2-tkg.1            True              False
```

## Get available upgrades for a Tanzu Kubernetes release

```sh
tanzu kubernetes-release  available-upgrades get -h
Usage:
  tanzu kubernetes-release available-upgrades get TKR_NAME [flags]

Flags:
  -h, --help   help for get
```

## Import your own os image to your cloud infrastructure

Run `tanzu kubernetes-release osimage` to import your own Kubernetes node image to specific cloud infrastructure, and patched TKR
to consume the imported image. For example,

```sh
tanzu kubernetes-release osimage oracle populate \
  --image https://objectstorage.us-sanjose-1.oraclecloud.com/n/axxxxxxxxxx8/b/exported-node-images/o/ubuntu-2004 \
  --tkr-path gcr.io/my-project-1527816345739/tkg/tkr/tkr-oci:latest \
  --compartment ocid1.compartment.oc1..aaaaaaaawgxbth6afwfzkxxxxxxxxxxxxxxmrf2ouxqa6ifrfa
```

### Sample command and output

```sh
tanzu kubernetes-release  available-upgrades get v1.18.6---vmware.1
 NAME                   VERSION
 v1.19.3---vmware.1     v1.19.3+vmware.1-tkg.1
 v1.19.3---vmware.2     v1.19.3+vmware.2-tkg.1
```

## Get supported OS info of a Tanzu Kubernetes release

```sh
tanzu kubernetes-release os get -h
Usage:
tanzu kubernetes-release os get TKR_NAME [flags]

Flags:
-h, --help help for get
--region string The AWS region where AMIs are available
```

### Sample command and output

```sh
tanzu kubernetes-release os get v1.18.6---vmware.1
NAME    VERSION   ARCH
photonos 1.1       amd64
```
