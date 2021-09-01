// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package client_test

import (
	"reflect"
	"testing"
	"time"

	"sigs.k8s.io/cluster-api-provider-aws/cmd/clusterawsadm/credentials"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/tree"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
)

func Test_ActivateTanzuKubernetesReleases_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ActivateTanzuKubernetesReleases",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DeactivateTanzuKubernetesReleases_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DeactivateTanzuKubernetesReleases",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetCEIPParticipation_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetCEIPParticipation",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(false),
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DoSetCEIPParticipation_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DoSetCEIPParticipation",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(region.RegionContext{}),
			reflect.TypeOf(false),
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetCEIPParticipation_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetCEIPParticipation",
		ParamTypes: []reflect.Type{},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(client.ClusterCeipInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DoGetCEIPParticipation_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DoGetCEIPParticipation",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(client.ClusterCeipInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_CreateCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "CreateCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.CreateClusterOptions{}),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DoCreateCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DoCreateCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(""),
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_WaitForClusterInitializedAndGetKubeConfig_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "WaitForClusterInitializedAndGetKubeConfig",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]byte{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_WaitForClusterReadyForMove_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "WaitForClusterReadyForMove",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_WaitForClusterReadyAfterCreate_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "WaitForClusterReadyAfterCreate",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_WaitForClusterReadyAfterReverseMove_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "WaitForClusterReadyAfterReverseMove",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateWorkloadClusterConfiguration_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateWorkloadClusterConfiguration",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.CreateClusterOptions{}),
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ValidateSupportOfK8sVersionForManagmentCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ValidateSupportOfK8sVersionForManagmentCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(""),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_IsPacificManagementCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "IsPacificManagementCluster",
		ParamTypes: []reflect.Type{},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(false),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ValidateAndConfigureClusterOptions_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ValidateAndConfigureClusterOptions",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.CreateClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ValidateManagementClusterVersionWithCLI_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ValidateManagementClusterVersionWithCLI",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ValidatePrerequisites_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ValidatePrerequisites",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(false),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureTimeout_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureTimeout",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*time.Duration)(nil)).Elem(),
		},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_CreateAWSCloudFormationStack_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "CreateAWSCloudFormationStack",
		ParamTypes: []reflect.Type{},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetAWSCreds_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetAWSCreds",
		ParamTypes: []reflect.Type{},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(&credentials.AWSCredentials{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetClusterConfiguration_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetClusterConfiguration",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.CreateClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]byte{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetPlan_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetPlan",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetProviderType_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetProviderType",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetVsphereVersion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetVsphereVersion",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetBuildEdition_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetBuildEdition",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetTKGVersion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:      tkgClientVal,
		MethodName:  "SetTKGVersion",
		ParamTypes:  []reflect.Type{},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetPinnipedConfigForWorkloadCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetPinnipedConfigForWorkloadCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DeleteRegion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DeleteRegion",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.DeleteRegionOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_IsManagementClusterAKindCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "IsManagementClusterAKindCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(false),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DeRegisterManagementClusterFromTmc_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DeRegisterManagementClusterFromTmc",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DescribeCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DescribeCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.DescribeTKGClustersOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(&tree.ObjectTree{}),
			reflect.TypeOf(&capi.Cluster{}),
			reflect.TypeOf(&clusterctlv1.ProviderList{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DescribeProvider_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DescribeProvider",
		ParamTypes: []reflect.Type{},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(&clusterctlv1.ProviderList{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetClusterPinnipedInfo_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetClusterPinnipedInfo",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.GetClusterPinnipedInfoOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(&client.ClusterPinnipedInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetWCClusterPinnipedInfo_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetWCClusterPinnipedInfo",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(region.RegionContext{}),
			reflect.TypeOf(client.GetClusterPinnipedInfoOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(&client.ClusterPinnipedInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetMCClusterPinnipedInfo_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetMCClusterPinnipedInfo",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(region.RegionContext{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(&client.ClusterPinnipedInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ListTKGClusters_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ListTKGClusters",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.ListTKGClustersOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]client.ClusterInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetClusterObjects_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetClusterObjects",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(&crtclient.ListOptions{}),
			reflect.TypeOf(""),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]client.ClusterInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetClusterObjectsForPacific_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetClusterObjectsForPacific",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(""),
			reflect.TypeOf(&crtclient.ListOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]client.ClusterInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetKubernetesVersions_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetKubernetesVersions",
		ParamTypes: []reflect.Type{},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(&client.KubernetesVersionsInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DoGetTanzuKubernetesReleases_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DoGetTanzuKubernetesReleases",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(&client.KubernetesVersionsInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetTanzuKubernetesReleases_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetTanzuKubernetesReleases",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]runv1alpha1.TanzuKubernetesRelease{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_InitRegionDryRun_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "InitRegionDryRun",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.InitRegionOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]byte{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_InitRegion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "InitRegion",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.InitRegionOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_PatchClusterInitOperations_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "PatchClusterInitOperations",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(&client.InitRegionOptions{}),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_MoveObjects_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "MoveObjects",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_InitializeProviders_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "InitializeProviders",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.InitRegionOptions{}),
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_BuildRegionalClusterConfiguration_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "BuildRegionalClusterConfiguration",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.InitRegionOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]byte{}),
			reflect.TypeOf(""),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ParseHiddenArgsAsFeatureFlags_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ParseHiddenArgsAsFeatureFlags",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.InitRegionOptions{}),
		},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SaveFeatureFlags_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SaveFeatureFlags",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(map[string]string{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetMachineDeployment_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetMachineDeployment",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.SetMachineDeploymentOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DeleteMachineDeployment_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DeleteMachineDeployment",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.DeleteMachineDeploymentOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetMachineDeployments_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetMachineDeployments",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.GetMachineDeploymentOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]capi.MachineDeployment{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

type EnforceMethodParams struct {
	Target      reflect.Value
	MethodName  string
	ParamTypes  []reflect.Type
	ReturnTypes []reflect.Type
}

func enforceMethodSignature(en *EnforceMethodParams, t *testing.T) {
	methodVal := en.Target.MethodByName(en.MethodName)
	if methodVal.IsZero() {
		t.Fatalf("target value does not contain method %s", en.MethodName)
	}

	if numParams := methodVal.Type().NumIn(); numParams != len(en.ParamTypes) {
		t.Fatalf("Expected %d parameters on method %s, got %d", len(en.ParamTypes), en.MethodName, numParams)
	}

	for i, paramType := range en.ParamTypes {
		if !methodVal.Type().In(i).AssignableTo(paramType) {
			t.Errorf("param at index %d is not of type %s", i, paramType.String())
		}
	}

	if numReturns := methodVal.Type().NumOut(); numReturns != len(en.ReturnTypes) {
		t.Fatalf("Expected %d parameters on method %s, got %d", len(en.ReturnTypes), en.MethodName, numReturns)
	}

	for i, returnType := range en.ReturnTypes {
		if !methodVal.Type().Out(i).AssignableTo(returnType) {
			t.Errorf("return value at index %d is not of type %s", i, returnType.String())
		}
	}
}
