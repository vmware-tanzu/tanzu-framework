// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

// Warningvsphere7WithoutPacific indicates that a vSphere7 environment has been detected without vSphere with Tanzu
var Warningvsphere7WithoutPacific = `
vSphere 7.0 Environment Detected.

You have connected to a vSphere 7.0 environment which does not have vSphere with Tanzu enabled. vSphere with Tanzu includes
an integrated Tanzu Kubernetes Grid Service which turns a vSphere cluster into a platform for running Kubernetes workloads in dedicated
resource pools. Configuring Tanzu Kubernetes Grid Service is done through vSphere HTML5 client.

Tanzu Kubernetes Grid Service is the preferred way to consume Tanzu Kubernetes Grid in vSphere 7.0 environments. Alternatively you may
deploy a non-integrated Tanzu Kubernetes Grid instance on vSphere 7.0.`

// Warningvsphere7WithPacific indicates that a vSphere7 environment has been detected with vSphere with Tanzu
var Warningvsphere7WithPacific = `
vSphere 7.0 with Tanzu Detected.

You have connected to a vSphere 7.0 with Tanzu environment that includes an integrated Tanzu Kubernetes Grid Service which
turns a vSphere cluster into a platform for running Kubernetes workloads in dedicated resource pools. Configuring Tanzu
Kubernetes Grid Service is done through the vSphere HTML5 Client.

Tanzu Kubernetes Grid Service is the preferred way to consume Tanzu Kubernetes Grid in vSphere 7.0 environments. Alternatively you may
deploy a non-integrated Tanzu Kubernetes Grid instance on vSphere 7.0.`
