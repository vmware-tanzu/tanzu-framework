// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package tkgctl

import (
	"reflect"
	"testing"

	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
)

func Test_AddRegion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "AddRegion",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(AddRegionOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(CreateClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_CreateAWSCloudFormationStack_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "CreateAWSCloudFormationStack",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_CreateCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "CreateCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(CreateClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DeleteCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DeleteCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(DeleteClustersOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DeleteMachineHealthCheck_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DeleteMachineHealthCheck",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(DeleteMachineHealthCheckOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DeleteRegion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DeleteRegion",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(DeleteRegionOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DeregisterFromTmc_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DeregisterFromTmc",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(DeregisterFromTMCOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetCEIP_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetCEIP",
		ParamTypes: []reflect.Type{},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(client.ClusterCeipInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetClusters_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetClusters",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(ListTKGClustersOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]client.ClusterInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DescribeCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DescribeCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(DescribeTKGClustersOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(DescribeClusterResult{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DescribeProviders_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DescribeProviders",
		ParamTypes: []reflect.Type{},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(&clusterctlv1.ProviderList{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetCredentials_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetCredentials",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(GetWorkloadClusterCredentialsOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetKubernetesVersions_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
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

func Test_GetMachineHealthCheck_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetMachineHealthCheck",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(GetMachineHealthCheckOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]client.MachineHealthCheck{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetRegions_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetRegions",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]region.RegionContext{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_Init_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "Init",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(InitRegionOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_RegisterWithTmc_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "RegisterWithTmc",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(RegisterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ScaleCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ScaleCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(ScaleClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetCeip_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetCeip",
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

func Test_SetMachineHealthCheck_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetMachineHealthCheck",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(SetMachineHealthCheckOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetMachineDeployments_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
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

func Test_SetMachineDeployment_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
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
	tkgClientVal := reflect.ValueOf(&tkgctl{})
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

func Test_SetRegion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetRegion",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(SetRegionOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_UpgradeCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "UpgradeCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(UpgradeClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_UpgradeRegion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "UpgradeRegion",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(UpgradeRegionOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_UpdateCredentialsRegion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "UpdateCredentialsRegion",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(UpdateCredentialsRegionOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_UpdateCredentialsCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "UpdateCredentialsCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(UpdateCredentialsClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetClusterPinnipedInfo_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetClusterPinnipedInfo",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(GetClusterPinnipedInfoOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(&client.ClusterPinnipedInfo{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetTanzuKubernetesReleases_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
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
func Test_ActivateTanzuKubernetesReleases_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&tkgctl{})
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
	tkgClientVal := reflect.ValueOf(&tkgctl{})
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
