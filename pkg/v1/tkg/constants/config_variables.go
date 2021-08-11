// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

// Configuration variable name constants
const (
	ConfigVariableDefaultBomFile                     = "TKG_DEFAULT_BOM"
	ConfigVariableCustomImageRepository              = "TKG_CUSTOM_IMAGE_REPOSITORY"
	ConfigVariableDevImageRepository                 = "TKG_DEV_IMAGE_REPOSITORY"
	ConfigVariableCompatibilityCustomImagePath       = "TKG_CUSTOM_COMPATIBILITY_IMAGE_PATH"
	ConfigVariableCustomImageRepositorySkipTLSVerify = "TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY"
	ConfigVariableCustomImageRepositoryCaCertificate = "TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE"

	ConfigVariableAWSRegion          = "AWS_REGION"
	ConfigVariableAWSSecretAccessKey = "AWS_SECRET_ACCESS_KEY" //nolint:gosec
	ConfigVariableAWSAccessKeyID     = "AWS_ACCESS_KEY_ID"     //nolint:gosec
	ConfigVariableAWSSessionToken    = "AWS_SESSION_TOKEN"     //nolint:gosec
	ConfigVariableAWSProfile         = "AWS_PROFILE"
	ConfigVariableAWSB64Credentials  = "AWS_B64ENCODED_CREDENTIALS"
	ConfigVariableAWSVPCID           = "AWS_VPC_ID"

	ConfigVariableAWSPublicNodeCIDR   = "AWS_PUBLIC_NODE_CIDR"
	ConfigVariableAWSPrivateNodeCIDR  = "AWS_PRIVATE_NODE_CIDR"
	ConfigVariableAWSPublicNodeCIDR1  = "AWS_PUBLIC_NODE_CIDR_1"
	ConfigVariableAWSPrivateNodeCIDR1 = "AWS_PRIVATE_NODE_CIDR_1"
	ConfigVariableAWSPublicNodeCIDR2  = "AWS_PUBLIC_NODE_CIDR_2"
	ConfigVariableAWSPrivateNodeCIDR2 = "AWS_PRIVATE_NODE_CIDR_2"
	ConfigVariableAWSPublicSubnetID   = "AWS_PUBLIC_SUBNET_ID"
	ConfigVariableAWSPrivateSubnetID  = "AWS_PRIVATE_SUBNET_ID"
	ConfigVariableAWSPublicSubnetID1  = "AWS_PUBLIC_SUBNET_ID_1"
	ConfigVariableAWSPrivateSubnetID1 = "AWS_PRIVATE_SUBNET_ID_1"
	ConfigVariableAWSPublicSubnetID2  = "AWS_PUBLIC_SUBNET_ID_2"
	ConfigVariableAWSPrivateSubnetID2 = "AWS_PRIVATE_SUBNET_ID_2"
	ConfigVariableAWSVPCCIDR          = "AWS_VPC_CIDR"
	ConfigVariableAWSNodeAz           = "AWS_NODE_AZ"
	ConfigVariableAWSNodeAz1          = "AWS_NODE_AZ_1"
	ConfigVariableAWSAMIID            = "AWS_AMI_ID"

	ConfigVariableVsphereControlPlaneEndpoint = "VSPHERE_CONTROL_PLANE_ENDPOINT"
	ConfigVariableVsphereServer               = "VSPHERE_SERVER"
	ConfigVariableVsphereUsername             = "VSPHERE_USERNAME"
	ConfigVariableVspherePassword             = "VSPHERE_PASSWORD"
	ConfigVariableVsphereTLSThumbprint        = "VSPHERE_TLS_THUMBPRINT"
	ConfigVariableVsphereSSHAuthorizedKey     = "VSPHERE_SSH_AUTHORIZED_KEY"
	ConfigVariableVsphereTemplate             = "VSPHERE_TEMPLATE"
	ConfigVariableVsphereDatacenter           = "VSPHERE_DATACENTER"
	ConfigVariableVsphereResourcePool         = "VSPHERE_RESOURCE_POOL"
	ConfigVariableVsphereDatastore            = "VSPHERE_DATASTORE"
	ConfigVariableVsphereFolder               = "VSPHERE_FOLDER"
	ConfigVariableVsphereNumCpus              = "VSPHERE_NUM_CPUS"
	ConfigVariableVsphereMemMib               = "VSPHERE_MEM_MIB"
	ConfigVariableVsphereDiskGib              = "VSPHERE_DISK_GIB"
	ConfigVariableVsphereWorkerNumCpus        = "VSPHERE_WORKER_NUM_CPUS"
	ConfigVariableVsphereWorkerMemMib         = "VSPHERE_WORKER_MEM_MIB"
	ConfigVariableVsphereWorkerDiskGib        = "VSPHERE_WORKER_DISK_GIB"
	ConfigVariableVsphereCPNumCpus            = "VSPHERE_CONTROL_PLANE_NUM_CPUS"
	ConfigVariableVsphereCPMemMib             = "VSPHERE_CONTROL_PLANE_MEM_MIB"
	ConfigVariableVsphereCPDiskGib            = "VSPHERE_CONTROL_PLANE_DISK_GIB"
	ConfigVariableVsphereInsecure             = "VSPHERE_INSECURE" // VCInsecure decides if the vc connection will skip the ssl validation or not.
	ConfigVariableVsphereVersion              = "VSPHERE_VERSION"
	ConfigVariableVsphereNetwork              = "VSPHERE_NETWORK"
	ConfigVariableVsphereHaProvider           = "AVI_CONTROL_PLANE_HA_PROVIDER"

	ConfigVariableAzureLocation               = "AZURE_LOCATION"
	ConfigVariableAzureImageID                = "AZURE_IMAGE_ID"
	ConfigVariableAzureImagePublisher         = "AZURE_IMAGE_PUBLISHER"
	ConfigVariableAzureImageOffer             = "AZURE_IMAGE_OFFER"
	ConfigVariableAzureImageSku               = "AZURE_IMAGE_SKU"
	ConfigVariableAzureImageVersion           = "AZURE_IMAGE_VERSION"
	ConfigVariableAzureImageThirdParty        = "AZURE_IMAGE_THIRD_PARTY"
	ConfigVariableAzureImageResourceGroup     = "AZURE_IMAGE_RESOURCE_GROUP"
	ConfigVariableAzureImageName              = "AZURE_IMAGE_NAME"
	ConfigVariableAzureImageSubscriptionID    = "AZURE_IMAGE_SUBSCRIPTION_ID"
	ConfigVariableAzureImageGallery           = "AZURE_IMAGE_GALLERY"
	ConfigVariableAzureSubscriptionIDB64      = "AZURE_SUBSCRIPTION_ID_B64"
	ConfigVariableAzureTenantIDB64            = "AZURE_TENANT_ID_B64"
	ConfigVariableAzureClientSecretB64        = "AZURE_CLIENT_SECRET_B64" //nolint:gosec
	ConfigVariableAzureClientIDB64            = "AZURE_CLIENT_ID_B64"
	ConfigVariableAzureSubscriptionID         = "AZURE_SUBSCRIPTION_ID"
	ConfigVariableAzureTenantID               = "AZURE_TENANT_ID"
	ConfigVariableAzureClientSecret           = "AZURE_CLIENT_SECRET" //nolint:gosec
	ConfigVariableAzureClientID               = "AZURE_CLIENT_ID"
	ConfigVariableAzureResourceGroup          = "AZURE_RESOURCE_GROUP"
	ConfigVariableAzureVnetName               = "AZURE_VNET_NAME"
	ConfigVariableAzureVnetResourceGroup      = "AZURE_VNET_RESOURCE_GROUP"
	ConfigVariableAzureVnetCidr               = "AZURE_VNET_CIDR"
	ConfigVariableAzureControlPlaneSubnet     = "AZURE_CONTROL_PLANE_SUBNET_NAME"
	ConfigVariableAzureWorkerSubnet           = "AZURE_NODE_SUBNET_NAME"
	ConfigVariableAzureControlPlaneSubnetCidr = "AZURE_CONTROL_PLANE_SUBNET_CIDR"
	ConfigVariableAzureWorkerNodeSubnetCidr   = "AZURE_NODE_SUBNET_CIDR"
	ConfigVariableAzureSSHPublicKeyB64        = "AZURE_SSH_PUBLIC_KEY_B64"
	ConfigVariableAzureCPMachineType          = "AZURE_CONTROL_PLANE_MACHINE_TYPE"
	ConfigVariableAzureNodeMachineType        = "AZURE_NODE_MACHINE_TYPE"
	ConfigVariableAzureEnvironment            = "AZURE_ENVIRONMENT"

	ConfigVariableDockerMachineTemplateImage = "DOCKER_MACHINE_TEMPLATE_IMAGE"

	ConfigVariablePinnipedSupervisorIssuerURL          = "SUPERVISOR_ISSUER_URL"
	ConfigVariablePinnipedSupervisorIssuerCABundleData = "SUPERVISOR_ISSUER_CA_BUNDLE_DATA_B64"

	ConfigVariableClusterRole            = "TKG_CLUSTER_ROLE"
	ConfigVariableForceRole              = "_TKG_CLUSTER_FORCE_ROLE"
	ConfigVariableProviderType           = "PROVIDER_TYPE"
	ConfigVariableTKGVersion             = "TKG_VERSION"
	ConfigVariableBuildEdition           = "BUILD_EDITION"
	ConfigVariableFilterByAddonType      = "FILTER_BY_ADDON_TYPE"
	ConfigVaraibleDisableCRSForAddonType = "DISABLE_CRS_FOR_ADDON_TYPE"
	ConfigVariableEnableAutoscaler       = "ENABLE_AUTOSCALER"

	ConfigVariableControlPlaneMachineCount = "CONTROL_PLANE_MACHINE_COUNT"
	ConfigVariableWorkerMachineCount       = "WORKER_MACHINE_COUNT"
	ConfigVariableWorkerMachineCount0      = "WORKER_MACHINE_COUNT_0"
	ConfigVariableWorkerMachineCount1      = "WORKER_MACHINE_COUNT_1"
	ConfigVariableWorkerMachineCount2      = "WORKER_MACHINE_COUNT_2"
	ConfigVariableNodeMachineType          = "NODE_MACHINE_TYPE"
	ConfigVariableCPMachineType            = "CONTROL_PLANE_MACHINE_TYPE"

	ConfigVariableNamespace            = "NAMESPACE"
	ConfigVariableEnableClusterOptions = "ENABLE_CLUSTER_OPTIONS"

	TKGHTTPProxy        = "TKG_HTTP_PROXY"
	TKGHTTPSProxy       = "TKG_HTTPS_PROXY"
	TKGHTTPProxyEnabled = "TKG_HTTP_PROXY_ENABLED"
	TKGNoProxy          = "TKG_NO_PROXY"

	ConfigVariableOSName    = "OS_NAME"
	ConfigVariableOSVersion = "OS_VERSION"
	ConfigVariableOSArch    = "OS_ARCH"

	ConfigVariableClusterCIDR = "CLUSTER_CIDR"
	ConfigVariableServiceCIDR = "SERVICE_CIDR"

	ConfigVariableIPFamily = "TKG_IP_FAMILY"

	// Below config variables are added based on init and create command flags

	ConfigVariableClusterPlan             = "CLUSTER_PLAN"
	ConfigVariableClusterName             = "CLUSTER_NAME"
	ConfigVariableInfraProvider           = "INFRASTRUCTURE_PROVIDER"
	ConfigVariableTkrName                 = "KUBERNETES_RELEASE"
	ConfigVariableKubernetesVersion       = "KUBERNETES_VERSION"
	ConfigVariableCNI                     = "CNI"
	ConfigVariableEnableCEIPParticipation = "ENABLE_CEIP_PARTICIPATION"
	ConfigVariableDeployTKGOnVsphere7     = "DEPLOY_TKG_ON_VSPHERE7"
	ConfigVariableEnableTKGSonVsphere7    = "ENABLE_TKGS_ON_VSPHERE7"
	ConfigVariableTMCRegistrationURL      = "TMC_REGISTRATION_URL"
	ConfigVariableSize                    = "SIZE"
	ConfigVariableControlPlaneSize        = "CONTROLPLANE_SIZE"
	ConfigVariableWorkerSize              = "WORKER_SIZE"

	// Config variable for passwords and secrets

	ConfigVariableNsxtPassword                     = "NSXT_PASSWORD"
	ConfigVariableAviPassword                      = "AVI_PASSWORD"
	ConfigVariableLDAPBindPassword                 = "LDAP_BIND_PASSWORD"                   //nolint:gosec
	ConfigVariableOIDCIdentiryProviderClientSecret = "OIDC_IDENTITY_PROVIDER_CLIENT_SECRET" //nolint:gosec
)
