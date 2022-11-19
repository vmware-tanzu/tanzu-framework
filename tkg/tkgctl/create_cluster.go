// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfighelper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/tkg/yamlprocessor"
)

// CreateClusterOptions options to create the cluster
type CreateClusterOptions struct {
	IsWindowsWorkloadCluster bool
	GenerateOnly             bool
	SkipPrompt               bool

	ClusterConfigFile           string
	ClusterName                 string
	Plan                        string
	InfrastructureProvider      string
	Namespace                   string
	TkrVersion                  string
	Size                        string
	ControlPlaneSize            string
	WorkerSize                  string
	CniType                     string
	EnableClusterOptions        string
	VsphereControlPlaneEndpoint string
	ControlPlaneMachineCount    int
	WorkerMachineCount          int
	Timeout                     time.Duration
	// Tanzu edition (either tce or tkg)
	Edition string
}

// CreateCluster create tkg cluster
//
//nolint:gocritic,gocyclo,revive
func (t *tkgctl) CreateCluster(cc CreateClusterOptions) error {
	isTKGSCluster, err := t.tkgClient.IsPacificManagementCluster()
	if err != nil {
		return err
	}
	isInputFileClusterClassBased, err := t.processWorkloadClusterInputFile(&cc, isTKGSCluster)
	if err != nil {
		return err
	}
	if cc.GenerateOnly && isInputFileClusterClassBased {
		if config, err := os.ReadFile(cc.ClusterConfigFile); err == nil {
			_, err1 := os.Stdout.Write(config)
			return err1
		} else {
			return err
		}
	}
	if cc.GenerateOnly {
		return t.ConfigCluster(cc)
	}

	cc.ClusterConfigFile, err = t.ensureClusterConfigFile(cc.ClusterConfigFile)
	if err != nil {
		return err
	}

	// configures missing create cluster options from config file variables
	if err := t.configureCreateClusterOptionsFromConfigFile(&cc); err != nil {
		return err
	}

	isClusterExists := false
	isClusterExists, err = t.IsClusterExists(cc.ClusterName, cc.Namespace)
	if err != nil {
		return errors.Wrap(err, constants.ErrorMsgClusterListError)
	} else if isClusterExists {
		return fmt.Errorf(constants.ErrorMsgClusterExistsAlready, cc.ClusterName)
	}

	if logPath, err := t.getAuditLogPath(cc.ClusterName); err == nil {
		log.SetAuditLog(logPath)
	}

	// Always do blocking cluster create
	waitForCluster := true

	// create cluster requires minimum 15 minutes timeout
	minTimeoutReq := 15 * time.Minute
	if cc.Timeout < minTimeoutReq {
		log.V(6).Infof("timeout duration of at least 15 minutes is required, using default timeout %v", constants.DefaultLongRunningOperationTimeout)
		cc.Timeout = constants.DefaultLongRunningOperationTimeout
	}

	defer t.restoreAfterSettingTimeout(cc.Timeout)()

	options, err := t.getCreateClusterOptions(cc.ClusterName, &cc, isInputFileClusterClassBased)
	if err != nil {
		return err
	}

	if isTKGSCluster {
		// For TKGS kubernetesVersion will be same as TkrVersion
		options.KubernetesVersion = cc.TkrVersion
		options.TKRVersion = cc.TkrVersion
		err = confirmPacificKubernetesVersion(cc.SkipPrompt, options.KubernetesVersion)
		if err != nil {
			return errors.Wrap(err, "unable to determine the kubernetes version for the cluster to be created on vSphere with Tanzu")
		}
	} else {
		options.TKRVersion, options.KubernetesVersion, err = t.getAndDownloadTkrIfNeeded(cc.TkrVersion)
		if err != nil {
			return errors.Wrapf(err, "unable to determine the TKr version and kubernetes version based on '%v'", cc.TkrVersion)
		}
	}

	attempedClusterCreation, err := t.tkgClient.CreateCluster(&options, waitForCluster)
	if err != nil {
		return err
	}

	// if cluster creation was not attempted return
	if !attempedClusterCreation {
		return nil
	}

	if waitForCluster {
		log.Infof("\nWorkload cluster '%s' created\n\n", options.ClusterConfigOptions.ClusterName)
	} else {
		log.Infof("\nWorkload cluster '%s' is being created\n\n", options.ClusterConfigOptions.ClusterName)
	}

	return nil
}

func (t *tkgctl) processManagementClusterInputFile(ir *InitRegionOptions) (bool, error) {
	var clusterobj unstructured.Unstructured
	var err error
	isInputFileClusterClassBased := false

	if t.tkgClient.IsFeatureActivated(constants.FeatureFlagPackageBasedCC) {
		isInputFileClusterClassBased, clusterobj, err = CheckIfInputFileIsClusterClassBased(ir.ClusterConfigFile)
		if err != nil {
			return isInputFileClusterClassBased, err
		}
		if isInputFileClusterClassBased {
			err = t.processClusterObjectForConfigurationVariables(clusterobj, ir.ClusterConfigFile)
			if err != nil {
				return isInputFileClusterClassBased, err
			}
			t.overrideManagementClusterOptionsWithLatestEnvironmentConfigurationValues(ir)
		}
	}
	return isInputFileClusterClassBased, nil
}

func (t *tkgctl) processWorkloadClusterInputFile(cc *CreateClusterOptions, isTKGSCluster bool) (bool, error) {
	isInputFileClusterClassBased, clusterobj, err := CheckIfInputFileIsClusterClassBased(cc.ClusterConfigFile)
	if err != nil {
		return isInputFileClusterClassBased, err
	}
	if isInputFileClusterClassBased {
		if !isTKGSCluster && t.tkgClient.IsFeatureActivated(constants.FeatureFlagAllowLegacyCluster) {
			return isInputFileClusterClassBased, fmt.Errorf(constants.ErrorMsgCClassInputFeatureFlagEnabled, constants.FeatureFlagAllowLegacyCluster)
		}
		if isTKGSCluster {
			t.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterName, clusterobj.GetName())
			t.TKGConfigReaderWriter().Set(constants.ConfigVariableNamespace, clusterobj.GetNamespace())
		} else {
			err = t.processClusterObjectForConfigurationVariables(clusterobj, cc.ClusterConfigFile)
			if err != nil {
				return isInputFileClusterClassBased, err
			}
		}
		t.overrideClusterOptionsWithLatestEnvironmentConfigurationValues(cc)
	}
	if isTKGSCluster {
		err = t.validateTKGSFeatureGateStatus(isInputFileClusterClassBased)
		if err != nil {
			return isInputFileClusterClassBased, err
		}
	}
	return isInputFileClusterClassBased, nil
}

// validateTKGSFeatureGateStatus validates the TKGS featuregate status for a specific feature
func (t *tkgctl) validateTKGSFeatureGateStatus(isInputFileClusterClassBased bool) error {
	if isInputFileClusterClassBased {
		isClusterClassFeatureActivated, err := t.featureGateHelper.FeatureActivatedInNamespace(context.Background(), constants.ClusterClassFeature, constants.TKGSClusterClassNamespace)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf(constants.ErrorMsgFeatureGateStatus, constants.ClusterClassFeature, constants.TKGSClusterClassNamespace))
		}
		if !isClusterClassFeatureActivated {
			return fmt.Errorf(constants.ErrorMsgFeatureGateNotActivated, constants.ClusterClassFeature, constants.TKGSClusterClassNamespace)
		}
	} else {
		isTKCFeatureActivated, err := t.featureGateHelper.FeatureActivatedInNamespace(context.Background(), constants.TKCAPIFeature, constants.TKGSTKCAPINamespace)
		if err != nil {
			// ignore the error, because the supervisor could be vSphere7 based (which does not support tkc-api featuregate)
			return nil
		}
		if !isTKCFeatureActivated {
			return fmt.Errorf(constants.ErrorMsgFeatureGateNotActivated, constants.TKCAPIFeature, constants.TKGSTKCAPINamespace)
		}
	}
	return nil
}

func (t *tkgctl) getCreateClusterOptions(name string, cc *CreateClusterOptions, isInputFileClusterClassBased bool) (client.CreateClusterOptions, error) {
	providerRepositorySource := &clusterctl.ProviderRepositorySourceOptions{
		InfrastructureProvider: cc.InfrastructureProvider,
		Flavor:                 cc.Plan,
	}
	if !isInputFileClusterClassBased && cc.Plan == "" {
		return client.CreateClusterOptions{}, errors.New("required config variable 'CLUSTER_PLAN' not set")
	}

	definitionParser := yamlprocessor.InjectDefinitionParser(yamlprocessor.NewYttDefinitionParser(yamlprocessor.InjectTKGDir(t.configDir)))

	configOptions := client.ClusterConfigOptions{
		ClusterName:              name,
		ProviderRepositorySource: providerRepositorySource,
		ControlPlaneMachineCount: swag.Int64(int64(cc.ControlPlaneMachineCount)),
		WorkerMachineCount:       swag.Int64(int64(cc.WorkerMachineCount)),
		TargetNamespace:          cc.Namespace,
		Kubeconfig:               clusterctl.Kubeconfig{Path: t.kubeconfig},
		YamlProcessor:            yamlprocessor.NewYttProcessor(definitionParser),
	}

	nodeSizeOptions := client.NodeSizeOptions{
		Size:             cc.Size,
		ControlPlaneSize: cc.ControlPlaneSize,
		WorkerSize:       cc.WorkerSize,
	}

	clusterOptionsEnableList := []string{}
	if cc.EnableClusterOptions != "" {
		clusterOptionsEnableList = strings.Split(cc.EnableClusterOptions, ",")
	}

	return client.CreateClusterOptions{
		ClusterConfigOptions:         configOptions,
		NodeSizeOptions:              nodeSizeOptions,
		CniType:                      cc.CniType,
		VsphereControlPlaneEndpoint:  cc.VsphereControlPlaneEndpoint,
		ClusterOptionsEnableList:     clusterOptionsEnableList,
		Edition:                      cc.Edition,
		IsWindowsWorkloadCluster:     cc.IsWindowsWorkloadCluster,
		ClusterConfigFile:            cc.ClusterConfigFile,
		IsInputFileClusterClassBased: isInputFileClusterClassBased,
	}, nil
}

func confirmPacificKubernetesVersion(shouldSkipPrompt bool, kubernetesVersion string) error {
	if !shouldSkipPrompt {
		log.Warningf("You are trying to create a cluster with kubernetes version '%s' on vSphere with Tanzu, Please make sure virtual machine image for the same is available in the cluster content library.", kubernetesVersion)
		err := askForConfirmation("Do you want to continue?")
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *tkgctl) configureCreateClusterOptionsFromConfigFile(cc *CreateClusterOptions) error { // nolint
	// set ClusterName from config variable
	if cc.ClusterName == "" {
		clusterName, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
		if err == nil {
			cc.ClusterName = clusterName
		} else {
			return errors.Errorf("cluster name is required, please provide cluster name")
		}
	}

	// set InfrastructureProvider from config variable
	if cc.InfrastructureProvider == "" {
		infraProvider, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableInfraProvider)
		if err == nil {
			cc.InfrastructureProvider = infraProvider
		}
	}

	// set Size variable from config File
	if cc.Size == "" {
		size, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableSize)
		if err == nil {
			cc.Size = size
		}
	}

	// set ControlPlaneSize variable from config File
	if cc.ControlPlaneSize == "" {
		controlPlaneSize, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneSize)
		if err == nil {
			cc.ControlPlaneSize = controlPlaneSize
		}
	}

	// set WorkerSize variable from config File
	if cc.WorkerSize == "" {
		workerSize, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerSize)
		if err == nil {
			cc.WorkerSize = workerSize
		}
	}

	// set CniType from config variable
	if cc.CniType == "" {
		cniType, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI)
		if err == nil {
			cc.CniType = cniType
		} else {
			cc.CniType = constants.DefaultCNIType
		}
	}

	// set IsWindowsWorkloadCluster from config variable
	if !cc.IsWindowsWorkloadCluster {
		strIWC, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableIsWindowsWorkloadCluster)
		// error on reading this parameter is a no-op, since its probably ephemeral and will be replaced w/ multitenant/multiworkload node-pools eventually
		if err == nil {
			isWindowsWorkloadCluster, err := strconv.ParseBool(strIWC)
			if err == nil {
				cc.IsWindowsWorkloadCluster = isWindowsWorkloadCluster
			} else {
				// if no value, set to the default, which should be false since most clusters are linux.
				cc.IsWindowsWorkloadCluster = constants.DefaultIsWindowsWorkloadCluster
			}
			// log this since its generally a less common use case, and windows support is relatively new.
			if cc.IsWindowsWorkloadCluster {
				log.Infof("\n Creating a windows workload cluster %v\n\n", cc.ClusterName)
			}
		}
	}

	// set Plan from config variable
	if cc.Plan == "" {
		plan, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterPlan)
		if err == nil {
			cc.Plan = plan
		}
	}

	// set vSphereControlPlaneEndpoint from config variable
	if cc.VsphereControlPlaneEndpoint == "" {
		vSphereControlPlaneEndpoint, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereControlPlaneEndpoint)
		if err == nil {
			cc.VsphereControlPlaneEndpoint = vSphereControlPlaneEndpoint
		}
	}

	if cc.ControlPlaneMachineCount == 0 {
		cpmc, err := tkgconfighelper.GetIntegerVariableFromConfig(constants.ConfigVariableControlPlaneMachineCount, t.TKGConfigReaderWriter())
		if err == nil {
			cc.ControlPlaneMachineCount = cpmc
		} else {
			cc.ControlPlaneMachineCount = constants.DefaultDevControlPlaneMachineCount
			if client.IsProdPlan(cc.Plan) {
				cc.ControlPlaneMachineCount = constants.DefaultProdControlPlaneMachineCount
			}
		}
	}

	if cc.ControlPlaneMachineCount%2 == 0 {
		return errors.Errorf("The number of control plane machines should be an odd number but provided %v", cc.ControlPlaneMachineCount)
	}

	if cc.WorkerMachineCount == 0 {
		wmc, err := tkgconfighelper.GetIntegerVariableFromConfig(constants.ConfigVariableWorkerMachineCount, t.TKGConfigReaderWriter())
		if err == nil {
			cc.WorkerMachineCount = wmc
		} else {
			cc.WorkerMachineCount = constants.DefaultDevWorkerMachineCount
			if client.IsProdPlan(cc.Plan) {
				cc.WorkerMachineCount = constants.DefaultProdWorkerMachineCount
			}
		}
	}

	if cc.Namespace == "" {
		namespace, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
		if err == nil {
			cc.Namespace = namespace
			log.V(1).Infof("Using namespace from config: %s", cc.Namespace)
		} else {
			cc.Namespace = constants.DefaultNamespace
			log.V(1).Infof("Using namespace: %s", cc.Namespace)
		}
	}

	// set EnableClusterOptions from config variable
	if cc.EnableClusterOptions == "" {
		enableClusterOptions, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableEnableClusterOptions)
		if err == nil {
			cc.EnableClusterOptions = enableClusterOptions
		}
	}

	// set BuildEdition from config variable
	if cc.Edition == "" {
		edition, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableBuildEdition)
		if err == nil {
			cc.Edition = edition
		}
	}

	return nil
}

func (t *tkgctl) getAndDownloadTkrIfNeeded(tkrVersion string) (string, string, error) {
	var k8sVersion string
	var err error
	var tkrBoMConfig *tkgconfigbom.BOMConfiguration

	if tkrVersion == "" {
		tkrBoMConfig, err = t.tkgBomClient.GetDefaultTkrBOMConfiguration()
		if err != nil {
			return "", "", errors.Wrap(err, "unable to get default TKr BoM configuration")
		}

		tkrVersion = tkrBoMConfig.Release.Version
		k8sVersion, err = t.tkgBomClient.GetDefaultK8sVersion()
		if err != nil {
			return "", "", errors.Wrap(err, "unable to get default kubernetes version")
		}
		return tkrVersion, k8sVersion, nil
	}

	// BoM downloading should only be required if user are providing tkrVersion,
	// otherwise we should use default config which is always present on user's machine

	// download bom if not present locally for given TKr
	// Put a file lock here to prevent several processes from downloading BOM at the same time
	lock, err := utils.GetFileLockWithTimeOut(filepath.Join(t.configDir, constants.LocalTanzuFileLock), utils.DefaultLockTimeout)
	if err != nil {
		return "", "", errors.Wrap(err, "cannot acquire lock for ensuring the TKr BOM file")
	}

	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Warningf("cannot release lock for ensuring TKr BOM file, reason: %v", err)
		}
	}()

	_, err = t.tkgBomClient.GetBOMConfigurationFromTkrVersion(tkrVersion)
	if err != nil {
		_, ok := err.(tkgconfigbom.BomNotPresent)
		if ok { // bom not present locally
			if err := t.tkgClient.DownloadBomFile(utils.GetTkrNameFromTkrVersion(tkrVersion)); err != nil {
				return "", "", err
			}
		} else {
			return "", "", err
		}
	}

	k8sVersion, err = t.tkgBomClient.GetK8sVersionFromTkrVersion(tkrVersion)
	if err != nil {
		return "", "", err
	}

	// Set tkrName and k8sVersion to the tkg config
	t.TKGConfigReaderWriter().Set(constants.ConfigVariableKubernetesVersion, k8sVersion)
	t.TKGConfigReaderWriter().Set(constants.ConfigVariableTkrName, utils.GetTkrNameFromTkrVersion(tkrVersion))

	return tkrVersion, k8sVersion, nil
}

// # val 1 - cannot create single node cluster without feature toggle
//  - if feature toggle true -> allow cp to be 1 and wokrer 0
//  - if feature toggle false -> worker cannot be 0 - fail saying min worker count is 1
// # val 2 - control plane taint must be removed for single node cluster
// - if feature toggle true
