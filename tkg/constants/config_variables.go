// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

// Configuration variable name constants
const (
	ConfigVariableDefaultBomFile                      = "TKG_DEFAULT_BOM"
	ConfigVariableCustomImageRepository               = "TKG_CUSTOM_IMAGE_REPOSITORY"
	ConfigVariableDevImageRepository                  = "TKG_DEV_IMAGE_REPOSITORY"
	ConfigVariableCompatibilityCustomImagePath        = "TKG_CUSTOM_COMPATIBILITY_IMAGE_PATH"
	ConfigVariableCustomImageRepositorySkipTLSVerify  = "TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY"
	ConfigVariableCustomImageRepositoryCaCertificate  = "TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE"
	ConfigVariableDefaultStandaloneDiscoveryImagePath = "TKG_DEFAULT_STANDALONE_DISCOVERY_IMAGE_PATH"
	ConfigVariableDefaultStandaloneDiscoveryImageTag  = "TKG_DEFAULT_STANDALONE_DISCOVERY_IMAGE_TAG"
	ConfigVariableDefaultStandaloneDiscoveryType      = "TKG_DEFAULT_STANDALONE_DISCOVERY_TYPE"
	ConfigVariableDefaultStandaloneDiscoveryLocalPath = "TKG_DEFAULT_STANDALONE_DISCOVERY_LOCAL_PATH"
	ConfigVariableClusterAPIServerPort                = "CLUSTER_API_SERVER_PORT"
	ConfigVariableBastionHostEnabled                  = "BASTION_HOST_ENABLED"
	ConfigVariableVipNetworkInterface                 = "VIP_NETWORK_INTERFACE"

	ConfigVariableAWSRegion          = "AWS_REGION"
	ConfigVariableAWSSecretAccessKey = "AWS_SECRET_ACCESS_KEY" //nolint:gosec
	ConfigVariableAWSAccessKeyID     = "AWS_ACCESS_KEY_ID"     //nolint:gosec
	ConfigVariableAWSSessionToken    = "AWS_SESSION_TOKEN"     //nolint:gosec
	ConfigVariableAWSProfile         = "AWS_PROFILE"
	ConfigVariableAWSB64Credentials  = "AWS_B64ENCODED_CREDENTIALS" //nolint:gosec
	ConfigVariableAWSVPCID           = "AWS_VPC_ID"
	ConfigVariableAWSSSHKeyName      = "AWS_SSH_KEY_NAME"

	ConfigVariableAWSPublicNodeCIDR             = "AWS_PUBLIC_NODE_CIDR"
	ConfigVariableAWSPrivateNodeCIDR            = "AWS_PRIVATE_NODE_CIDR"
	ConfigVariableAWSPublicNodeCIDR1            = "AWS_PUBLIC_NODE_CIDR_1"
	ConfigVariableAWSPrivateNodeCIDR1           = "AWS_PRIVATE_NODE_CIDR_1"
	ConfigVariableAWSPublicNodeCIDR2            = "AWS_PUBLIC_NODE_CIDR_2"
	ConfigVariableAWSPrivateNodeCIDR2           = "AWS_PRIVATE_NODE_CIDR_2"
	ConfigVariableAWSPublicSubnetID             = "AWS_PUBLIC_SUBNET_ID"
	ConfigVariableAWSPrivateSubnetID            = "AWS_PRIVATE_SUBNET_ID"
	ConfigVariableAWSPublicSubnetID1            = "AWS_PUBLIC_SUBNET_ID_1"
	ConfigVariableAWSPrivateSubnetID1           = "AWS_PRIVATE_SUBNET_ID_1"
	ConfigVariableAWSPublicSubnetID2            = "AWS_PUBLIC_SUBNET_ID_2"
	ConfigVariableAWSPrivateSubnetID2           = "AWS_PRIVATE_SUBNET_ID_2"
	ConfigVariableAWSVPCCIDR                    = "AWS_VPC_CIDR"
	ConfigVariableAWSNodeAz                     = "AWS_NODE_AZ"
	ConfigVariableAWSNodeAz1                    = "AWS_NODE_AZ_1"
	ConfigVariableAWSNodeAz2                    = "AWS_NODE_AZ_2"
	ConfigVariableAWSAMIID                      = "AWS_AMI_ID"
	ConfigVariableAWSLoadBalancerSchemeInternal = "AWS_LOAD_BALANCER_SCHEME_INTERNAL"
	ConfigVariableAWSNodeOsDiskSizeGib          = "AWS_NODE_OS_DISK_SIZE_GIB"

	ConfigVariableAWSIdentityRefKind           = "AWS_IDENTITY_REF_KIND"
	ConfigVariableAWSIdentityRefName           = "AWS_IDENTITY_REF_NAME"
	ConfigVariableAWSSecurityGroupNode         = "AWS_SECURITY_GROUP_NODE"
	ConfigVariableAWSSecurityGroupApiserverLb  = "AWS_SECURITY_GROUP_APISERVER_LB"
	ConfigVariableAWSSecurityGroupBastion      = "AWS_SECURITY_GROUP_BASTION"
	ConfigVariableAWSSecurityGroupControlplane = "AWS_SECURITY_GROUP_CONTROLPLANE"
	ConfigVariableAWSSecurityGroupLb           = "AWS_SECURITY_GROUP_LB"
	ConfigVariableAWSControlplaneOsDiskSizeGib = "AWS_CONTROL_PLANE_OS_DISK_SIZE_GIB"

	ConfigVariableVsphereAz0                         = "VSPHERE_AZ_0"
	ConfigVariableVsphereAz1                         = "VSPHERE_AZ_1"
	ConfigVariableVsphereAz2                         = "VSPHERE_AZ_2"
	ConfigVariableVsphereCloneMode                   = "VSPHERE_CLONE_MODE"
	ConfigVariableVsphereControlPlaneEndpoint        = "VSPHERE_CONTROL_PLANE_ENDPOINT"
	ConfigVariableVsphereServer                      = "VSPHERE_SERVER"
	ConfigVariableVsphereUsername                    = "VSPHERE_USERNAME"
	ConfigVariableVspherePassword                    = "VSPHERE_PASSWORD"
	ConfigVariableVsphereTLSThumbprint               = "VSPHERE_TLS_THUMBPRINT"
	ConfigVariableVsphereSSHAuthorizedKey            = "VSPHERE_SSH_AUTHORIZED_KEY"
	ConfigVariableVsphereTemplate                    = "VSPHERE_TEMPLATE"
	ConfigVariableVsphereTemplateMoid                = "VSPHERE_TEMPLATE_MOID"
	ConfigVariableVsphereDatacenter                  = "VSPHERE_DATACENTER"
	ConfigVariableVsphereResourcePool                = "VSPHERE_RESOURCE_POOL"
	ConfigVariableVsphereStoragePolicyID             = "VSPHERE_STORAGE_POLICY_ID"
	ConfigVariableVsphereDatastore                   = "VSPHERE_DATASTORE"
	ConfigVariableVsphereFolder                      = "VSPHERE_FOLDER"
	ConfigVariableVsphereWorkerpciDevices            = "VSPHERE_WORKER_PCI_DEVICES"
	ConfigVariableVsphereControlPlanepciDevices      = "VSPHERE_CONTROL_PLANE_PCI_DEVICES"
	ConfigVariableVsphereControlPlaneCustomVMXKeys   = "VSPHERE_CONTROL_PLANE_CUSTOM_VMX_KEYS"
	ConfigVariableVsphereWorkerCustomVMXKeys         = "VSPHERE_WORKER_CUSTOM_VMX_KEYS"
	ConfigVariableVsphereIgnorepciDevicesAllowList   = "VSPHERE_IGNORE_PCI_DEVICES_ALLOW_LIST"
	ConfigVariableVsphereWorkerRolloutStrategy       = "WORKER_ROLLOUT_STRATEGY"
	ConfigVariableVsphereNumCpus                     = "VSPHERE_NUM_CPUS"
	ConfigVariableVsphereMemMib                      = "VSPHERE_MEM_MIB"
	ConfigVariableVsphereDiskGib                     = "VSPHERE_DISK_GIB"
	ConfigVariableVsphereWorkerNumCpus               = "VSPHERE_WORKER_NUM_CPUS"
	ConfigVariableVsphereWorkerMemMib                = "VSPHERE_WORKER_MEM_MIB"
	ConfigVariableVsphereWorkerDiskGib               = "VSPHERE_WORKER_DISK_GIB"
	ConfigVariableVsphereCPNumCpus                   = "VSPHERE_CONTROL_PLANE_NUM_CPUS"
	ConfigVariableVsphereCPMemMib                    = "VSPHERE_CONTROL_PLANE_MEM_MIB"
	ConfigVariableVsphereCPDiskGib                   = "VSPHERE_CONTROL_PLANE_DISK_GIB"
	ConfigVariableVsphereInsecure                    = "VSPHERE_INSECURE" // VCInsecure decides if the vc connection will skip the ssl validation or not.
	ConfigVariableVsphereVersion                     = "VSPHERE_VERSION"
	ConfigVariableVsphereNetwork                     = "VSPHERE_NETWORK"
	ConfigVariableVSphereControlPlaneHardwareVersion = "VSPHERE_CONTROL_PLANE_HARDWARE_VERSION"
	ConfigVariableVSphereWorkerHardwareVersion       = "VSPHERE_WORKER_HARDWARE_VERSION"
	ConfigVariableVsphereHaProvider                  = "AVI_CONTROL_PLANE_HA_PROVIDER"

	ConfigVariableAzureControlPlaneSubnet                    = "AZURE_CONTROL_PLANE_SUBNET_NAME"
	ConfigVariableAzureControlPlaneSubnetName                = "AZURE_CONTROL_PLANE_SUBNET_NAME"
	ConfigVariableAzureControlPlaneSubnetCidr                = "AZURE_CONTROL_PLANE_SUBNET_CIDR"
	ConfigVariableAzureCPMachineType                         = "AZURE_CONTROL_PLANE_MACHINE_TYPE"
	ConfigVariableAzureControlPlaneDataDiskSizeGib           = "AZURE_CONTROL_PLANE_DATA_DISK_SIZE_GIB"
	ConfigVariableAzureControlPlaneOsDiskStorageAccountType  = "AZURE_CONTROL_PLANE_OS_DISK_STORAGE_ACCOUNT_TYPE"
	ConfigVariableAzureControlPlaneOsDiskSizeGib             = "AZURE_CONTROL_PLANE_OS_DISK_SIZE_GIB"
	ConfigVariableAzureControlPlaneOutboundLbFrontendIPCount = "AZURE_CONTROL_PLANE_OUTBOUND_LB_FRONTEND_IP_COUNT"
	ConfigVariableAzureControlPlaneOutboundLb                = "AZURE_ENABLE_CONTROL_PLANE_OUTBOUND_LB"
	ConfigVariableAzureControlPlaneSubnetSecurityGroup       = "AZURE_CONTROL_PLANE_SUBNET_SECURITY_GROUP"

	ConfigVariableAzureCustomTags                  = "AZURE_CUSTOM_TAGS"
	ConfigVariableAzureEnableAcceleratedNetworking = "AZURE_ENABLE_ACCELERATED_NETWORKING"
	ConfigVariableAzureEnablePrivateCluster        = "AZURE_ENABLE_PRIVATE_CLUSTER"
	ConfigVariableAzureFrontendPrivateIP           = "AZURE_FRONTEND_PRIVATE_IP"
	ConfigVariableAzureLocation                    = "AZURE_LOCATION"
	ConfigVariableAzureIdentityName                = "AZURE_IDENTITY_NAME"
	ConfigVariableAzureIdentityNamespace           = "AZURE_IDENTITY_NAMESPACE"
	ConfigVariableAzureImageID                     = "AZURE_IMAGE_ID"
	ConfigVariableAzureImagePublisher              = "AZURE_IMAGE_PUBLISHER"
	ConfigVariableAzureImageOffer                  = "AZURE_IMAGE_OFFER"
	ConfigVariableAzureImageSku                    = "AZURE_IMAGE_SKU"
	ConfigVariableAzureImageVersion                = "AZURE_IMAGE_VERSION"
	ConfigVariableAzureImageThirdParty             = "AZURE_IMAGE_THIRD_PARTY"
	ConfigVariableAzureImageResourceGroup          = "AZURE_IMAGE_RESOURCE_GROUP"
	ConfigVariableAzureImageName                   = "AZURE_IMAGE_NAME"
	ConfigVariableAzureImageSubscriptionID         = "AZURE_IMAGE_SUBSCRIPTION_ID"
	ConfigVariableAzureImageGallery                = "AZURE_IMAGE_GALLERY"
	ConfigVariableAzureSubscriptionIDB64           = "AZURE_SUBSCRIPTION_ID_B64"
	ConfigVariableAzureTenantIDB64                 = "AZURE_TENANT_ID_B64"
	ConfigVariableAzureClientSecretB64             = "AZURE_CLIENT_SECRET_B64" //nolint:gosec
	ConfigVariableAzureClientIDB64                 = "AZURE_CLIENT_ID_B64"
	ConfigVariableAzureSubscriptionID              = "AZURE_SUBSCRIPTION_ID"
	ConfigVariableAzureTenantID                    = "AZURE_TENANT_ID"
	ConfigVariableAzureClientSecret                = "AZURE_CLIENT_SECRET" //nolint:gosec
	ConfigVariableAzureClientID                    = "AZURE_CLIENT_ID"
	ConfigVariableAzureResourceGroup               = "AZURE_RESOURCE_GROUP"
	ConfigVariableAzureVnetName                    = "AZURE_VNET_NAME"
	ConfigVariableAzureVnetResourceGroup           = "AZURE_VNET_RESOURCE_GROUP"
	ConfigVariableAzureVnetCidr                    = "AZURE_VNET_CIDR"

	ConfigVariableAzureWorkerSubnet                       = "AZURE_NODE_SUBNET_NAME"
	ConfigVariableAzureWorkerSubnetName                   = "AZURE_NODE_SUBNET_NAME"
	ConfigVariableAzureAZ                                 = "AZURE_NODE_AZ"
	ConfigVariableAzureAZ1                                = "AZURE_NODE_AZ_1"
	ConfigVariableAzureAZ2                                = "AZURE_NODE_AZ_2"
	ConfigVariableAzureNodeOsDiskSizeGib                  = "AZURE_NODE_OS_DISK_SIZE_GIB"
	ConfigVariableAzureNodeOsDiskStorageAccountType       = "AZURE_NODE_OS_DISK_STORAGE_ACCOUNT_TYPE"
	ConfigVariableAzureEnableNodeDataDisk                 = "AZURE_ENABLE_NODE_DATA_DISK"
	ConfigVariableAzureNodeDataDiskSizeGib                = "AZURE_NODE_DATA_DISK_SIZE_GIB"
	ConfigVariableAzureNodeSubnetSecurityGroup            = "AZURE_NODE_SUBNET_SECURITY_GROUP"
	ConfigVariableAzureEnableNodeOutboundLb               = "AZURE_ENABLE_NODE_OUTBOUND_LB"
	ConfigVariableAzureNodeOutboundLbFrontendIPCount      = "AZURE_NODE_OUTBOUND_LB_FRONTEND_IP_COUNT"
	ConfigVariableAzureNodeOutboundLbIdleTimeoutInMinutes = "AZURE_NODE_OUTBOUND_LB_IDLE_TIMEOUT_IN_MINUTES"
	ConfigVariableAzureWorkerNodeSubnetCidr               = "AZURE_NODE_SUBNET_CIDR"
	ConfigVariableAzureSSHPublicKeyB64                    = "AZURE_SSH_PUBLIC_KEY_B64"
	ConfigVariableAzureNodeMachineType                    = "AZURE_NODE_MACHINE_TYPE"
	ConfigVariableAzureEnvironment                        = "AZURE_ENVIRONMENT"

	ConfigVariableDockerMachineTemplateImage = "DOCKER_MACHINE_TEMPLATE_IMAGE"

	ConfigVariablePinnipedSupervisorIssuerURL          = "SUPERVISOR_ISSUER_URL"
	ConfigVariablePinnipedSupervisorIssuerCABundleData = "SUPERVISOR_ISSUER_CA_BUNDLE_DATA_B64"

	ConfigVariableClusterRole                = "TKG_CLUSTER_ROLE"
	ConfigVariableForceRole                  = "_TKG_CLUSTER_FORCE_ROLE"
	ConfigVariableProviderType               = "PROVIDER_TYPE"
	ConfigVariableTKGVersion                 = "TKG_VERSION"
	ConfigVariableBuildEdition               = "BUILD_EDITION"
	ConfigVariableFilterByAddonType          = "FILTER_BY_ADDON_TYPE"
	ConfigVaraibleDisableCRSForAddonType     = "DISABLE_CRS_FOR_ADDON_TYPE"
	ConfigVariableEnableAutoscaler           = "ENABLE_AUTOSCALER"
	ConfigVariableDisableTMCCloudPermissions = "DISABLE_TMC_CLOUD_PERMISSIONS"
	AutoscalerDeploymentNameSuffix           = "-cluster-autoscaler"

	ConfigVariableControlPlaneMachineCount = "CONTROL_PLANE_MACHINE_COUNT"
	ConfigVariableControlPlaneMachineType  = "CONTROL_PLANE_MACHINE_TYPE"

	ConfigVariableWorkerMachineCount  = "WORKER_MACHINE_COUNT"
	ConfigVariableWorkerMachineCount0 = "WORKER_MACHINE_COUNT_0"
	ConfigVariableWorkerMachineCount1 = "WORKER_MACHINE_COUNT_1"
	ConfigVariableWorkerMachineCount2 = "WORKER_MACHINE_COUNT_2"
	ConfigVariableNodeMachineType     = "NODE_MACHINE_TYPE"
	ConfigVariableNodeMachineType1    = "NODE_MACHINE_TYPE_1"
	ConfigVariableNodeMachineType2    = "NODE_MACHINE_TYPE_2"
	ConfigVariableCPMachineType       = "CONTROL_PLANE_MACHINE_TYPE"

	ConfigVariableNamespace            = "NAMESPACE"
	ConfigVariableEnableClusterOptions = "ENABLE_CLUSTER_OPTIONS"

	TKGHTTPProxy        = "TKG_HTTP_PROXY"
	TKGHTTPSProxy       = "TKG_HTTPS_PROXY"
	TKGHTTPProxyEnabled = "TKG_HTTP_PROXY_ENABLED"
	TKGNoProxy          = "TKG_NO_PROXY"
	TKGProxyCACert      = "TKG_PROXY_CA_CERT"
	EnableAuditLogging  = "ENABLE_AUDIT_LOGGING"
	TKGIPFamily         = "TKG_IP_FAMILY"

	PodSecurityStandardDeactivated = "POD_SECURITY_STANDARD_DEACTIVATED"
	PodSecurityStandardAudit       = "POD_SECURITY_STANDARD_AUDIT"
	PodSecurityStandardWarn        = "POD_SECURITY_STANDARD_WARN"
	PodSecurityStandardEnforce     = "POD_SECURITY_STANDARD_ENFORCE"

	ConfigVariableOSName    = "OS_NAME"
	ConfigVariableOSVersion = "OS_VERSION"
	ConfigVariableOSArch    = "OS_ARCH"

	ConfigVariableClusterCIDR = "CLUSTER_CIDR"
	ConfigVariableServiceCIDR = "SERVICE_CIDR"

	ConfigVariableCoreDNSIP = "CORE_DNS_IP"

	ConfigVariableIPFamily = "TKG_IP_FAMILY"
	TKGIPV6Primary         = "TKG_IPV6_PRIMARY"

	ConfigVariableNodeIPAMIPPoolName = "NODE_IPAM_IP_POOL_NAME"

	ConfigVariableControlPlaneNodeNameservers = "CONTROL_PLANE_NODE_NAMESERVERS"
	ConfigVariableWorkerNodeNameservers       = "WORKER_NODE_NAMESERVERS"

	// Below config variables are added based on init and create command flags

	ConfigVariableClusterPlan             = "CLUSTER_PLAN"
	ConfigVariableClusterName             = "CLUSTER_NAME"
	ConfigVariableClusterClass            = "CLUSTER_CLASS"
	ConfigVariableInfraProvider           = "INFRASTRUCTURE_PROVIDER"
	ConfigVariableTkrName                 = "KUBERNETES_RELEASE"
	ConfigVariableKubernetesVersion       = "KUBERNETES_VERSION"
	ConfigVariableCNI                     = "CNI"
	ConfigVariableEnableCEIPParticipation = "ENABLE_CEIP_PARTICIPATION"
	ConfigVariableDeployTKGOnVsphere7     = "DEPLOY_TKG_ON_VSPHERE7"
	ConfigVariableEnableTKGSonVsphere7    = "ENABLE_TKGS_ON_VSPHERE7"
	ConfigVariableSize                    = "SIZE"
	ConfigVariableControlPlaneSize        = "CONTROLPLANE_SIZE"
	ConfigVariableWorkerSize              = "WORKER_SIZE"

	// Config variable for passwords and secrets

	ConfigVariableNsxtPassword                     = "NSXT_PASSWORD"
	ConfigVariableAviPassword                      = "AVI_PASSWORD"
	ConfigVariableLDAPBindPassword                 = "LDAP_BIND_PASSWORD"                   //nolint:gosec
	ConfigVariableOIDCIdentiryProviderClientSecret = "OIDC_IDENTITY_PROVIDER_CLIENT_SECRET" //nolint:gosec

	// Config variables for image tags used for provider installation
	ConfigVariableInternalKubeRBACProxyImageTag             = "KUBE_RBAC_PROXY_IMAGE_TAG"
	ConfigVariableInternalCABPKControllerImageTag           = "CABPK_CONTROLLER_IMAGE_TAG"
	ConfigVariableInternalCAPIControllerImageTag            = "CAPI_CONTROLLER_IMAGE_TAG"
	ConfigVariableInternalKCPControllerImageTag             = "KCP_CONTROLLER_IMAGE_TAG"
	ConfigVariableInternalCAPDManagerImageTag               = "CAPD_CONTROLLER_IMAGE_TAG"
	ConfigVariableInternalCAPAManagerImageTag               = "CAPA_CONTROLLER_IMAGE_TAG"
	ConfigVariableInternalCAPVManagerImageTag               = "CAPV_CONTROLLER_IMAGE_TAG"
	ConfigVariableInternalCAPZManagerImageTag               = "CAPZ_CONTROLLER_IMAGE_TAG"
	ConfigVariableInternalCAPOCIManagerImageTag             = "CAPOCI_CONTROLLER_IMAGE_TAG"
	ConfigVariableInternalCAPIIPAMProviderInClusterImageTag = "CAPI_IPAM_PROVIDER_IN_CLUSTER_IMAGE_TAG"
	ConfigVariableInternalNMIImageTag                       = "NMI_IMAGE_TAG"

	// Other variables related to provider installation
	ConfigVariableClusterTopology = "CLUSTER_TOPOLOGY"

	ConfigVariablePackageInstallTimeout = "PACKAGE_INSTALL_TIMEOUT"

	// Windows specific variables
	ConfigVariableIsWindowsWorkloadCluster = "IS_WINDOWS_WORKLOAD_CLUSTER"

	// AVI aka. NSX Advanced Load Balancer specific variables
	ConfigVariableAviEnable = "AVI_ENABLE"

	ConfigVariableAviControllerAddress  = "AVI_CONTROLLER"
	ConfigVariableAviControllerVersion  = "AVI_CONTROLLER_VERSION"
	ConfigVariableAviControllerUsername = "AVI_USERNAME"
	ConfigVariableAviControllerPassword = "AVI_PASSWORD"
	ConfigVariableAviControllerCA       = "AVI_CA_DATA_B64"

	ConfigVariableAviCloudName                           = "AVI_CLOUD_NAME"
	ConfigVariableAviServiceEngineGroup                  = "AVI_SERVICE_ENGINE_GROUP"
	ConfigVariableAviManagementClusterServiceEngineGroup = "AVI_MANAGEMENT_CLUSTER_SERVICE_ENGINE_GROUP"

	ConfigVariableAviDataPlaneNetworkName    = "AVI_DATA_NETWORK"
	ConfigVariableAviDataPlaneNetworkCIDR    = "AVI_DATA_NETWORK_CIDR"
	ConfigVariableAviControlPlaneNetworkName = "AVI_CONTROL_PLANE_NETWORK"
	ConfigVariableAviControlPlaneNetworkCIDR = "AVI_CONTROL_PLANE_NETWORK_CIDR"

	ConfigVariableAviManagementClusterDataPlaneNetworkName       = "AVI_MANAGEMENT_CLUSTER_VIP_NETWORK_NAME"
	ConfigVariableAviManagementClusterDataPlaneNetworkCIDR       = "AVI_MANAGEMENT_CLUSTER_VIP_NETWORK_CIDR"
	ConfigVariableAviManagementClusterControlPlaneVipNetworkName = "AVI_MANAGEMENT_CLUSTER_CONTROL_PLANE_VIP_NETWORK_NAME"
	ConfigVariableAviManagementClusterControlPlaneVipNetworkCIDR = "AVI_MANAGEMENT_CLUSTER_CONTROL_PLANE_VIP_NETWORK_CIDR"

	ConfigVariableFeatureFlagPackageBasedLCM = "FEATURE_FLAG_PACKAGE_BASED_LCM"
)
