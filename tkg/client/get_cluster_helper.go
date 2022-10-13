// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"

	"github.com/pkg/errors"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

const (
	// TKGPlanAnnotation plan annotation
	TKGPlanAnnotation     = "tkg/plan"
	LegacyClusterTKRLabel = "tanzuKubernetesRelease"
)

type clusterObjects struct {
	cluster  capi.Cluster
	kcp      controlplanev1.KubeadmControlPlane
	mds      []capi.MachineDeployment
	machines []capi.Machine
}

type clusterObjectsForPacific struct {
	cluster  tkgsv1alpha2.TanzuKubernetesCluster
	mds      []capiv1alpha3.MachineDeployment
	machines []capiv1alpha3.Machine
}

// ################### Helpers for Pacific ##################

func getRunningCPMachineCountForPacific(clusterInfo *clusterObjectsForPacific) int {
	cpMachineCount := 0
	for i := range clusterInfo.machines {
		if _, labelExists := clusterInfo.machines[i].GetLabels()[capiv1alpha3.MachineControlPlaneLabelName]; labelExists && strings.EqualFold(clusterInfo.machines[i].Status.Phase, "running") {
			cpMachineCount++
		}
	}
	return cpMachineCount
}

func getClusterControlPlaneCountForPacific(clusterInfo *clusterObjectsForPacific) string {
	if clusterInfo.cluster.Spec.Topology.ControlPlane.Replicas == nil {
		return ""
	}
	cpReplicas := fmt.Sprintf("%v/%v", getRunningCPMachineCountForPacific(clusterInfo), *clusterInfo.cluster.Spec.Topology.ControlPlane.Replicas)
	return cpReplicas
}

func getClusterObjectsMapForPacific(clusterClient clusterclient.Client, listOptions *crtclient.ListOptions) (map[string]*clusterObjectsForPacific, error) {
	var tkcObjList tkgsv1alpha2.TanzuKubernetesClusterList
	err := clusterClient.ListResources(&tkcObjList, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get list of clusters")
	}

	var mdList capiv1alpha3.MachineDeploymentList
	err = clusterClient.ListResources(&mdList, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get list of MachineDeployment objects")
	}

	var machineList capiv1alpha3.MachineList
	err = clusterClient.ListResources(&machineList, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get list of Machine objects")
	}

	clusterInfoMap := make(map[string]*clusterObjectsForPacific)

	for idx := range tkcObjList.Items {
		cl := tkcObjList.Items[idx]
		clusterObjectCombined := clusterObjectsForPacific{cluster: cl}
		key := cl.Name + "-" + cl.Namespace
		clusterInfoMap[key] = &clusterObjectCombined
	}

	for i := range mdList.Items {
		clusterName, labelExists := mdList.Items[i].GetLabels()[capi.ClusterLabelName]
		if !labelExists {
			continue
		}
		key := clusterName + "-" + mdList.Items[i].Namespace
		clusterObjectCombined, clusterExists := clusterInfoMap[key]
		if !clusterExists {
			continue
		}
		clusterObjectCombined.mds = append(clusterObjectCombined.mds, mdList.Items[i])
	}

	for i := range machineList.Items {
		clusterName, labelExists := machineList.Items[i].GetLabels()[capi.ClusterLabelName]
		if !labelExists {
			continue
		}
		key := clusterName + "-" + machineList.Items[i].Namespace
		clusterObjectCombined, clusterExists := clusterInfoMap[key]
		if !clusterExists {
			continue
		}
		clusterObjectCombined.machines = append(clusterObjectCombined.machines, machineList.Items[i])
	}

	return clusterInfoMap, nil
}

func getClusterWorkerCountForPacific(clusterInfo *clusterObjectsForPacific) string {
	var desiredReplicasCount int32
	nodePools := clusterInfo.cluster.Spec.Topology.NodePools
	for idx := range nodePools {
		if nodePools[idx].Replicas != nil {
			desiredReplicasCount += *nodePools[idx].Replicas
		}
	}

	var readyReplicasCount int32
	for idx := range clusterInfo.mds {
		readyReplicasCount += clusterInfo.mds[idx].Status.ReadyReplicas
	}
	return fmt.Sprintf("%v/%v", readyReplicasCount, desiredReplicasCount)
}

func getClusterPlanForPacific(clusterInfo *clusterObjectsForPacific) string {
	plan, exists := clusterInfo.cluster.GetAnnotations()[TKGPlanAnnotation]
	if !exists {
		return ""
	}
	return plan
}

// ################### Helpers for Non-Pacific cluster info ##################

func getClusterObjectsMap(clusterClient clusterclient.Client, listOptions *crtclient.ListOptions) (map[string]*clusterObjects, error) { // nolint:funlen
	var clusters capi.ClusterList
	err := clusterClient.ListResources(&clusters, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get list of clusters")
	}

	var kcpList controlplanev1.KubeadmControlPlaneList
	err = clusterClient.ListResources(&kcpList, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get list of KubeadmControlPlane objects")
	}

	var mdList capi.MachineDeploymentList
	err = clusterClient.ListResources(&mdList, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get list of MachineDeployment objects")
	}

	var machineList capi.MachineList
	err = clusterClient.ListResources(&machineList, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get list of Machine objects")
	}

	clusterInfoMap := make(map[string]*clusterObjects)

	for i := range clusters.Items {
		clusterObjectCombined := clusterObjects{cluster: clusters.Items[i]}
		key := clusters.Items[i].Name + "-" + clusters.Items[i].Namespace
		clusterInfoMap[key] = &clusterObjectCombined
	}

	for i := range kcpList.Items {
		clusterName, labelExists := kcpList.Items[i].GetLabels()[capi.ClusterLabelName]
		if !labelExists {
			continue
		}
		key := clusterName + "-" + kcpList.Items[i].Namespace
		clusterObjectCombined, clusterExists := clusterInfoMap[key]
		if !clusterExists {
			log.V(3).Infof("unable to find cluster object '%s' for KubeadmControlPlane object '%s'", clusterName, kcpList.Items[i].Name)
			continue
		}
		clusterObjectCombined.kcp = kcpList.Items[i]
	}

	for i := range mdList.Items {
		clusterName, labelExists := mdList.Items[i].GetLabels()[capi.ClusterLabelName]
		if !labelExists {
			continue
		}
		key := clusterName + "-" + mdList.Items[i].Namespace
		clusterObjectCombined, clusterExists := clusterInfoMap[key]
		if !clusterExists {
			log.V(3).Infof("unable to find cluster object '%s' for MachineDeployment object '%s'", clusterName, mdList.Items[i].Name)
			continue
		}
		clusterObjectCombined.mds = append(clusterObjectCombined.mds, mdList.Items[i])
	}

	for i := range machineList.Items {
		clusterName, labelExists := machineList.Items[i].GetLabels()[capi.ClusterLabelName]
		if !labelExists {
			continue
		}
		key := clusterName + "-" + machineList.Items[i].Namespace
		clusterObjectCombined, clusterExists := clusterInfoMap[key]
		if !clusterExists {
			log.V(3).Infof("unable to find cluster object '%s' for Machine object '%s'", clusterName, machineList.Items[i].Name)
			continue
		}
		clusterObjectCombined.machines = append(clusterObjectCombined.machines, machineList.Items[i])
	}

	return clusterInfoMap, nil
}

func getClusterPlan(clusterInfo *clusterObjects) string {
	plan, exists := clusterInfo.cluster.GetAnnotations()[TKGPlanAnnotation]
	if !exists {
		return ""
	}
	return plan
}

func getClusterControlPlaneCount(clusterInfo *clusterObjects) string {
	if clusterInfo.kcp.Spec.Replicas == nil {
		return ""
	}
	cpReplicas := fmt.Sprintf("%v/%v", clusterInfo.kcp.Status.ReadyReplicas, *clusterInfo.kcp.Spec.Replicas)
	return cpReplicas
}

func getClusterWorkerCount(clusterInfo *clusterObjects) string {
	readyReplicas, specReplicas, _, _ := getClusterReplicas(clusterInfo.mds)
	// return empty string if specReplicas are not yet set
	if specReplicas == 0 {
		return ""
	}
	cpReplicas := fmt.Sprintf("%v/%v", readyReplicas, specReplicas)
	return cpReplicas
}

func getClusterReplicas(mds []capi.MachineDeployment) (int32, int32, int32, int32) {
	var readyReplicas int32
	var specReplicas int32
	var replicas int32
	var updatedReplicas int32
	for i := range mds {
		readyReplicas += mds[i].Status.ReadyReplicas
		replicas += mds[i].Status.Replicas
		updatedReplicas += mds[i].Status.UpdatedReplicas
		if mds[i].Spec.Replicas != nil {
			specReplicas += *mds[i].Spec.Replicas
		}
	}
	return readyReplicas, specReplicas, replicas, updatedReplicas
}

func getClusterK8sVersion(clusterInfo *clusterObjects) string {
	return clusterInfo.kcp.Spec.Version
}

func getClusterRoles(clusterLabels map[string]string) []string {
	clusterRoles := make([]string, 0)
	for labelKey := range clusterLabels {
		if !strings.HasPrefix(labelKey, TkgLabelClusterRolePrefix) {
			continue
		}

		clusterRole := strings.TrimPrefix(labelKey, TkgLabelClusterRolePrefix)
		clusterRoles = append(clusterRoles, clusterRole)
	}

	return clusterRoles
}

func getClusterTKR(clusterLabels map[string]string) string {
	var clusterTKR string
	if tkrName, exists := clusterLabels[LegacyClusterTKRLabel]; exists {
		clusterTKR = tkrName
	}
	if tkrName, exists := clusterLabels[runv1.LabelTKR]; exists {
		clusterTKR = tkrName
	}
	return clusterTKR
}

// ################### Helpers for determining Cluster Status ##################

func getClusterStatus(clusterInfo *clusterObjects) TKGClusterPhase {
	if strings.EqualFold(clusterInfo.cluster.Status.Phase, string(TKGClusterPhaseDeleting)) {
		return TKGClusterPhaseDeleting
	}

	readyReplicas, specReplicas, replicas, updatedReplicas := getClusterReplicas(clusterInfo.mds)

	creationCompleteCondition := clusterInfo.cluster.Status.InfrastructureReady &&
		clusterInfo.cluster.Status.ControlPlaneReady &&
		clusterInfo.kcp.Status.ReadyReplicas > 0 &&
		readyReplicas >= 0

	if !creationCompleteCondition {
		return getClusterStatusWhileCreating(clusterInfo)
	}

	runningCondition := clusterInfo.cluster.Status.ControlPlaneReady &&
		*clusterInfo.kcp.Spec.Replicas == clusterInfo.kcp.Status.Replicas &&
		*clusterInfo.kcp.Spec.Replicas == clusterInfo.kcp.Status.ReadyReplicas &&
		*clusterInfo.kcp.Spec.Replicas == clusterInfo.kcp.Status.UpdatedReplicas &&
		specReplicas == replicas &&
		specReplicas == readyReplicas &&
		specReplicas == updatedReplicas &&
		!isUpgradeInProgress(clusterInfo)

	if runningCondition {
		return TKGClusterPhaseRunning
	}

	return getClusterStatusWhileUpdating(clusterInfo)
}

func getClusterStatusWhileCreating(clusterInfo *clusterObjects) TKGClusterPhase {
	if isOperationStalled(clusterclient.OperationTypeCreate, clusterInfo) {
		return TKGClusterPhaseCreationStalled
	}
	return TKGClusterPhaseCreating
}

func isUpgradeInProgress(clusterInfo *clusterObjects) bool {
	// check k8s version of all machine objects version with kcp.Spec.Version
	// to determine this is scaling operation or upgrade operation
	for i := range clusterInfo.machines {
		if clusterInfo.machines[i].Spec.Version != nil && *clusterInfo.machines[i].Spec.Version != clusterInfo.kcp.Spec.Version {
			return true
		}
	}
	return false
}

func getClusterStatusWhileUpdating(clusterInfo *clusterObjects) TKGClusterPhase {
	// if not upgrade opeation then it has to be scaling operation return updating status directly
	if !isUpgradeInProgress(clusterInfo) {
		return TKGClusterPhaseUpdating
	}

	if isOperationStalled(clusterclient.OperationTypeUpgrade, clusterInfo) {
		return TKGClusterPhaseUpdateStalled
	}

	return TKGClusterPhaseUpdating
}

func isOperationStalled(operationType string, clusterInfo *clusterObjects) bool {
	operationString, exists := clusterInfo.cluster.Annotations[clusterclient.TKGOperationInfoKey]
	if !exists {
		return false
	}

	lastObservedTimestamp, exists := clusterInfo.cluster.Annotations[clusterclient.TKGOperationLastObservedTimestampKey]
	if !exists {
		return false
	}

	operationString = strings.ReplaceAll(operationString, "\\\"", "\"")
	operationStatusObject := &clusterclient.OperationStatus{}
	err := json.Unmarshal([]byte(operationString), operationStatusObject)
	if err != nil {
		return false
	}

	if operationStatusObject.Operation != operationType {
		return false
	}

	t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", lastObservedTimestamp)
	if err != nil {
		return false
	}

	return time.Since(t).Seconds() > float64(operationStatusObject.OperationTimeout)
}
