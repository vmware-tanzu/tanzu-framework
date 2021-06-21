# Get Available TKr


## Get available Tanzu Kubernetes releases
```
tanzu kubernetes-release get -h 
Usage:
  tanzu kubernetes-release get TKR_NAME [flags]

Flags:
  -h, --help   help for get
```

### Sample command and output
```
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
```
Tanzu kubernetes-release  available-upgrades get -h
Usage:
  tanzu kubernetes-release available-upgrades get TKR_NAME [flags]

Flags:
  -h, --help   help for get
```
### Sample command and output
```
tanzu kubernetes-release  available-upgrades get v1.18.6---vmware.1
 NAME                   VERSION         
 v1.19.3---vmware.1     v1.19.3+vmware.1-tkg.1          
 v1.19.3---vmware.2     v1.19.3+vmware.2-tkg.1      
```

## Get supported OS info of a Tanzu Kubernetes release
```
tanzu kubernetes-release os get -h
Usage:
tanzu kubernetes-release os get TKR_NAME [flags]

Flags:
-h, --help help for get
--region string The AWS region where AMIs are available
```
### Sample command and output
```
tanzu kubernetes-release os get v1.18.6---vmware.1
NAME    VERSION   ARCH
photonos 1.1       amd64
```
