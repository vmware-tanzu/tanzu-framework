# Library components in tkgctl

 This library includes coresponding [providers](/providers) to provide the templates and ytt overlays required to generate cluster configuration YAML documents.

## [tkgctl library](/tkg/tkgctl)

This library provides an interface that can be used to invoke TKG related functionalities. This interface is currently used by the cluster and management-cluster plugin in tanzu cli to perform TKG related operations. TMC also consumes this library interface to do lifecycle operations on TKG clusters.

For more details about the client and interface: [/tkg/tkgctl/client.go](/tkg/tkgctl/client.go)

## clusterctl library

Library uses clusterctl library to perform many core tasks like

* installing cluster-api providers and cert-manager during management cluster creation (tanzu management-cluster create)
* moving cluster-api objects from bootstrap cluster to management cluster during management cluster creation (tanzu management-cluster create)
* for cluster template creation using YAML processor (tanzu management-cluster create and tanzu cluster create)
* moving cluster-api objects from management cluster to cleanup cluster during management cluster deletion (tanzu management-cluster delete)
* upgrading cluster-api providers and cert-manager during management cluster upgrade (tanzu management-cluster upgrade)

reference:
[https://pkg.go.dev/sigs.k8s.io/cluster-api/cmd/clusterctl/](https://pkg.go.dev/sigs.k8s.io/cluster-api/cmd/clusterctl/)

## [ytt processor](/tkg/yamlprocessor)

* Library has this YTT processor package
* Implements YAML processor interface that is used with clusterctl when creation cluster templates. [https://pkg.go.dev/sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor#Processor](https://pkg.go.dev/sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor#Processor)
* The responsibility of this package is to use the ytt library and implement functionality that returns cluster template bytes when requested by the clusterctl library when generating cluster templates.
  * reads ytt overlay files based on the template-definition-file provided and processes the overlay using the provided config variable generating cluster templates.
  * [https://github.com/vmware-tanzu/tkg-cli/tree/master/pkg/yamlprocessor](https://github.com/vmware-tanzu/tkg-cli/tree/master/pkg/yamlprocessor)

## [tkgconfigreaderwriter](/tkg/tkgconfigreaderwriter)

* This is also a very core package under the tkg library
* It is responsible for reading user inputs for config variables from different sources (descending order of precedence)
  * environment variables
  * cluster-config file (file provided with `tanzu management-cluster create` and `tanzu cluster create` commands)
  * tkg settings file (`$HOME/.config/tanzu/tkg/config.yaml`)
* It internally uses viper implementation to read this configuration
* This client is passed to many other clients that need to rely on tkg settings or user-provided configuration

## [tkgconfigupdater](/tkg/tkgconfigupdater)

* The main responsibility of this package is to ensure all the necessary configurations are present on the user's local file system or not, which includes
  * extracting and/or updating providers bundle to `$HOME/.config/tanzu/tkg/providers`
  * extracting and/or updating BoM files to `$HOME/.config/tanzu/tkg/bom`
  * creating and/or updating TKG settings file at `$HOME/.config/tanzu/tkg/config.yaml`
    * adding/updating provider map based on user's file system path
    * adding/updating images map based on BoM file's image repository or based on `TKG_CUSTOM_IMAGE_REPOSITORY` config variable
  * creating default(empty) cluster-config.yaml if it does not exist at `$HOME/.config/tanzu/tkg/cluster-config.yaml`

## [tkgconfigpaths](/tkg/tkgconfigpaths)

* Implemented functions to get different file and directory paths for TKG library

## [tkgconfigproviders](/tkg/tkgconfigproviders)

* Implements methods that convert UI provided configuration to config variables that can be saved to cluster-config files
* Also implements image getter functions for all providers

## [tkgconfigbom](/tkg/tkgconfigbom)

* Implements TKG and TKR BoM file loader methods
* Also implements methods to get default or specific TKR BoM file based on the TKR version

## AWS cloud-formation library

* tkgctl interface implements functions that use CAPA's cloud-formation library to create a cloud-formation stack on AWS

## [kind](/tkg/kind)

* Implements methods to create and delete kind clusters that get used to create and delete management cluster

## [aws](/tkg/aws)

* Implements AWS specific API using AWS SDK that can be used for some verification purpose as well as serves as resource retriever for the kick-start UI

## [azure](/tkg/azure)

* Implements Azure specific API using Azure SDK that can be used for some verification purpose as well as serves as resource retriever for the kick-start UI

## [vc](/tkg/vc)

* Implements vCenter specific API using govmomi library that can be used for some verification purpose as well as serves as resource retriever for the kick-start UI

## [Web UI and Server](/tkg/web)

* It implements web server which serves kick-start UI for management cluster creation purpose. (tanzu management-cluster create --ui)
* This uses bundled go-bindata file which contains kick-start UI bits
* API is developed using swagger specification which gets used by UI as well as backend

## [Providers](/providers)

* This package maintains a set of CAPI provider CRDs, cluster templates, and ytt overlays required by the library.
* This allows cluster configuration generation to be customizable potentially without modifying a single line of golang code.
* More information regarding providers are [here](/providers)
