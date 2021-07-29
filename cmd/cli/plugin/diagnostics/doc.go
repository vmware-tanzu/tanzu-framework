// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Plugin diagnostics implements a Crashd plugin for Tanzu Framework.
// By default, the plugin will run its built in crashd scripts for diagnostics.
// Users can also provide the path to a crashd script for specific troubleshooting steps.
//
// Examples:
//   # Collect API object, pod logs, and nodes diagnostics from standalone docker cluster my-cluster
//   tanzu diagnostics collect --cluster-name=my-cluster
//
//   # Collect API object and pod logs from managed my-vsphere-cluster
//   tanzu diagnostics collect --cluster-name=my-vsphere-cluster --cluster-infra=vsphere --cluster-type=managed
//
//   # Collect API object and pod logs from managed my-vsphere-cluster using custom kubeconfig
//   tanzu diagnostics collect --cluster-name=my-vsphere-cluster --cluster-infra=vsphere --cluster-type=managed --kubeconfog=/my/kubeconfig
//
//   # Collect API objects, pod logs, and node diagnostics from standalone my-vsphere-cluster
//   tanzu diagnostics collect --cluster-name=my-vsphere-cluster --cluster-infra=vsphere --ssh_user=myssh --ssh_pk_path=/path/to/private_key

package main
