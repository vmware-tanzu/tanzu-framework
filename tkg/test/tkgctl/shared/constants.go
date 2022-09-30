// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package shared

// customAntreaConfigAndCBResource declares resources required to deploy a custom ClusterBootstrap and AntreaConfig objects; name and namespace values need to be substituted prior to usage
const customAntreaConfigAndCBResource = `
apiVersion: csi.tanzu.vmware.com/v1alpha1
kind: AwsEbsCSIConfig
metadata:
  name: %s
  namespace: %s
spec:
  awsEBSCSIDriver:
    namespace: %s
    deploymentReplicas: 1
---
apiVersion: cni.tanzu.vmware.com/v1alpha1
kind: AntreaConfig
metadata:
  name: %s
  namespace: %s
spec:
  antrea:
    config:
      disableUdpTunnelOffload: false
      featureGates:
        AntreaPolicy: true
        AntreaProxy: true
        AntreaTraceflow: false
        Egress: true
        EndpointSlice: true
        FlowExporter: false
        NodePortLocal: true
      noSNAT: false
      trafficEncapMode: encap
---
apiVersion: run.tanzu.vmware.com/v1alpha3
kind: ClusterBootstrap
metadata:
  annotations:
    tkg.tanzu.vmware.com/add-missing-fields-from-tkr: %s
  name: %s
  namespace: %s
spec:
  additionalPackages:
    - refName: metrics-server*
    - refName: secretgen-controller*
    - refName: pinniped*
  cni:
    refName: antrea*
    valuesFrom:
      providerRef:
        apiGroup: cni.tanzu.vmware.com
        kind: AntreaConfig
        name: %s
  csi:
    refName: aws-ebs-csi*
    valuesFrom:
      providerRef:
        apiGroup: csi.tanzu.vmware.com
        kind: AwsEbsCSIConfig
        name: %s
  kapp:
    refName: kapp-controller*
---
`
