// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package client_test

import (
	"reflect"
	"testing"

	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/aws"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/azure"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/providersupgradeclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/vc"
)

func Test_DeleteMachineHealthCheck_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DeleteMachineHealthCheck",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.MachineHealthCheckOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DeleteMachineHealthCheckWithClusterClient_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DeleteMachineHealthCheckWithClusterClient",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(client.MachineHealthCheckOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetMachineHealthChecks_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetMachineHealthChecks",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.MachineHealthCheckOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]client.MachineHealthCheck{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetMachineHealthChecksWithClusterClient_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetMachineHealthChecksWithClusterClient",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(client.MachineHealthCheckOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]client.MachineHealthCheck{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetMachineHealthCheck_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetMachineHealthCheck",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.SetMachineHealthCheckOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_CreateMachineHealthCheck_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "CreateMachineHealthCheck",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(&client.SetMachineHealthCheckOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_UpdateMachineHealthCheck_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "UpdateMachineHealthCheck",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&capi.MachineHealthCheck{}),
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(&client.SetMachineHealthCheckOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_VerifyRegion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "VerifyRegion",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(region.RegionContext{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_AddRegionContext_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "AddRegionContext",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(region.RegionContext{}),
			reflect.TypeOf(false),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetRegionContexts_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetRegionContexts",
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

func Test_SetRegionContext_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetRegionContext",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetCurrentRegionContext_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetCurrentRegionContext",
		ParamTypes: []reflect.Type{},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(region.RegionContext{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_RegisterManagementClusterToTmc_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "RegisterManagementClusterToTmc",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ScaleCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ScaleCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.ScaleClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ScalePacificCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ScalePacificCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.ScaleClusterOptions{}),
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_UpdateCredentialsRegion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "UpdateCredentialsRegion",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.UpdateCredentialsOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_UpdateCredentialsCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "UpdateCredentialsCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.UpdateCredentialsOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_UpdateVSphereClusterCredentials_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "UpdateVSphereClusterCredentials",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(&client.UpdateCredentialsOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_UpgradeAddon_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "UpgradeAddon",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.UpgradeAddonOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DoUpgradeAddon_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DoUpgradeAddon",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(&client.UpgradeAddonOptions{}),
			reflect.TypeOf(func(*client.CreateClusterOptions) ([]byte, error) { return nil, nil }),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_UpgradeCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "UpgradeCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.UpgradeClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DoPacificClusterUpgrade_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DoPacificClusterUpgrade",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(&client.UpgradeClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DoClusterUpgrade_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DoClusterUpgrade",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(&client.UpgradeClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_UpgradeManagementCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "UpgradeManagementCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.UpgradeClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DoProvidersUpgrade_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DoProvidersUpgrade",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(""),
			reflect.TypeOf((*providersupgradeclient.Client)(nil)).Elem(),
			reflect.TypeOf(&client.UpgradeClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_WaitForAddonsDeployments_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "WaitForAddonsDeployments",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_WaitForPackages_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "WaitForPackages",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
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

func Test_DownloadBomFile_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DownloadBomFile",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateTkrVersion_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateTkrVersion",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(""),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAzureVMImage_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAzureVMImage",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateAzureConfig_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateAzureConfig",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(client.NodeSizeOptions{}),
			reflect.TypeOf(false),
			reflect.TypeOf(false),
			reflect.TypeOf(int64(0)),
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ValidateAzurePublicSSHKey_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ValidateAzurePublicSSHKey",
		ParamTypes: []reflect.Type{},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateDockerConfig_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateDockerConfig",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(client.NodeSizeOptions{}),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateWindowsVsphereConfig_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateWindowsVsphereConfig",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(client.NodeSizeOptions{}),
			reflect.TypeOf(""),
			reflect.TypeOf(false),
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*client.ValidationError)(nil)),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateAwsConfig_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateAwsConfig",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(false),
			reflect.TypeOf(false),
			reflect.TypeOf(int64(0)),
			reflect.TypeOf(false),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateAWSConfig_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateAWSConfig",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(client.NodeSizeOptions{}),
			reflect.TypeOf(false),
			reflect.TypeOf(false),
			reflect.TypeOf(int64(0)),
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_TrimVsphereSSHKey_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:      tkgClientVal,
		MethodName:  "TrimVsphereSSHKey",
		ParamTypes:  []reflect.Type{},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateVSphereTemplate_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateVSphereTemplate",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*vc.Client)(nil)).Elem(),
			reflect.TypeOf(""),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetVSphereEndpoint_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetVSphereEndpoint",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*vc.Client)(nil)).Elem(),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateManagementClusterConfiguration_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateManagementClusterConfiguration",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(&client.InitRegionOptions{}),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*client.ValidationError)(nil)),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ValidateVsphereControlPlaneEndpointIP_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ValidateVsphereControlPlaneEndpointIP",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*client.ValidationError)(nil)),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateVsphereConfig_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateVsphereConfig",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(client.NodeSizeOptions{}),
			reflect.TypeOf(""),
			reflect.TypeOf(false),
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*client.ValidationError)(nil)),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ValidateVsphereVipWorkloadCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ValidateVsphereVipWorkloadCluster",
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

func Test_ValidateVsphereResources_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ValidateVsphereResources",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*vc.Client)(nil)).Elem(),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetAndValidateDefaultAWSVPCConfiguration_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetAndValidateDefaultAWSVPCConfiguration",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(false),
			reflect.TypeOf((*aws.Client)(nil)).Elem(),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(false),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ValidateVsphereNodeSize_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ValidateVsphereNodeSize",
		ParamTypes: []reflect.Type{},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetVsphereNodeSize_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:      tkgClientVal,
		MethodName:  "SetVsphereNodeSize",
		ParamTypes:  []reflect.Type{},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_OverrideAzureNodeSizeWithOptions_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "OverrideAzureNodeSizeWithOptions",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*azure.Client)(nil)).Elem(),
			reflect.TypeOf(client.NodeSizeOptions{}),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_OverrideAWSNodeSizeWithOptions_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "OverrideAWSNodeSizeWithOptions",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.NodeSizeOptions{}),
			reflect.TypeOf((*aws.Client)(nil)).Elem(),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_OverrideVsphereNodeSizeWithOptions_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "OverrideVsphereNodeSizeWithOptions",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.NodeSizeOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetTKGClusterRole_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetTKGClusterRole",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*client.TKGClusterType)(nil)).Elem(),
		},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_EncodeAzureCredentialsAndGetClient_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "EncodeAzureCredentialsAndGetClient",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*azure.Client)(nil)).Elem(),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateCNIType_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateCNIType",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DistributeMachineDeploymentWorkers_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DistributeMachineDeploymentWorkers",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(int64(0)),
			reflect.TypeOf(false),
			reflect.TypeOf(false),
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf([]int{}),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetMachineDeploymentWorkerCounts_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "SetMachineDeploymentWorkerCounts",
		ParamTypes: []reflect.Type{
			reflect.TypeOf([]int{}),
			reflect.TypeOf(int64(0)),
			reflect.TypeOf(false),
		},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_EncodeAWSCredentialsAndGetClient_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "EncodeAWSCredentialsAndGetClient",
		ParamTypes: []reflect.Type{
			reflect.TypeOf((*clusterclient.Client)(nil)).Elem(),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*aws.Client)(nil)).Elem(),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_ConfigureAndValidateHTTPProxyConfiguration_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "ConfigureAndValidateHTTPProxyConfiguration",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(""),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_SetDefaultProxySettings_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:      tkgClientVal,
		MethodName:  "SetDefaultProxySettings",
		ParamTypes:  []reflect.Type{},
		ReturnTypes: []reflect.Type{},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_GetWorkloadClusterCredentials_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "GetWorkloadClusterCredentials",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.GetWorkloadClusterCredentialsOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf(""),
			reflect.TypeOf(""),
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}

func Test_DeleteWorkloadCluster_Signature(t *testing.T) {
	tkgClientVal := reflect.ValueOf(&client.TkgClient{})
	enforce := EnforceMethodParams{
		Target:     tkgClientVal,
		MethodName: "DeleteWorkloadCluster",
		ParamTypes: []reflect.Type{
			reflect.TypeOf(client.DeleteWorkloadClusterOptions{}),
		},
		ReturnTypes: []reflect.Type{
			reflect.TypeOf((*error)(nil)).Elem(),
		},
	}
	enforceMethodSignature(&enforce, t)
}
