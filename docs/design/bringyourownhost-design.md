# Bring your own host enhancement design



## Abstract

The Bring Your Own Host aka BYOH aims at enabling Kubernetes cluster life cycle management on user provided hosts (most of the case are bare metal hosts) with TKGm/TMC

But currently only 1 infrastructure provider is supported in tanzu-framework,  this is to enhance current tanzu-framework so that users can use tanzu CLI to create BYOH workload cluster.

## Background

Consider the multiple provider work covers multiple components in Tanzu product family and from BYOH team's point of view we are searching to enable customer to be able to create BYOH cluster with Tanzu cli as a start point. We shape the project into 3 phases and firstly we intend to implement the "crawl phase":

Crawl Phase

1.	Enhance the tanzu cluster create cli so customer can use it be able to create BYOH cluster.
2.	Document on patch the existing Management cluster with BYOH provider

Walk Phase
1.	Implement the tanzu management-cluster provider cli so customer can use it patch management cluster with new provider
•	Support patching BYOH provider
2.	Support install BYOH provider when create management cluster with Tanzu management-cluster create cli or bootstrap UI.
3.	Support list/view BYOH workload cluster in TMC UI

Run Phase
1.	Support create BYOH cluster in TMC UI
2.	Support lifecycle management of AWS/Azure/Any other provider with tanzu management-cluster provider cli
3.	Support create Workload cluster with selected provider with multiple providers installed management cluster in TMC
4.	Support creates management cluster directly on BYOH hosts.

## Goals

* Define the configurations used for BYOH to create BYOH workload cluster.
* Use the existing Tanzu cluster create command to create BYOH workload cluster using the configuration file.
* Use the existing Tanzu cluster list/delete/scale command to list/delete/update BYOH workload cluster.

## Non Goals

* Any changes for TMC UI to support bringyourownhost provider

## High-Level Design

This is to descirbe what will be targeted for the first phase.

Add the “**infrastructure-byoh**” folder which contains the YTT templates to generate the YAML files, so that CAPI can provision BYOH workload clusters using these YAML files.

Add the configuration item to let users define control plane endpoint for BYOH workload cluster.

Update BOM files to support BYOH.

Update the logic to validate the configurations for provisioning BYOH clusters.



## Detailed Design

Update "pkg/v1/providers/config.yaml" and create new YTT under folder " pkg/v1/providers/infrastructure-byoh/v0.1.0/" to support BYOH as a new type of provider.

Update "pkg/v1/providers/config_default.yaml" and tanzu-framework go code, so that new configurations can be read to create BYOH workload cluster.

Configurations to provision BYOH workload clusters should be validated in case of invalid configuration input.

