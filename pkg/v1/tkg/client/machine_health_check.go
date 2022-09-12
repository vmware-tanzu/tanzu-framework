// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// machinehealthcheck related default and constants
const (
	DefaulUnhealthyConditionTimeout = "5m"
	DefaultNodeStartupTimeout       = "20m"
	DefaultNodePoolSuffix           = "worker-pool"
	NodePoolKey                     = "node-pool"
)

// UnhealthyConditions unhealthy condition map
var UnhealthyConditions = map[string]bool{
	string(corev1.NodeReady):              true,
	string(corev1.NodeMemoryPressure):     true,
	string(corev1.NodeDiskPressure):       true,
	string(corev1.NodePIDPressure):        true,
	string(corev1.NodeNetworkUnavailable): true,
}

// ConditionStatus condition status map
var ConditionStatus = map[string]bool{
	string(corev1.ConditionTrue):    true,
	string(corev1.ConditionFalse):   true,
	string(corev1.ConditionUnknown): true,
}

// MachineHealthCheckOptions machinehealthcheck options
type MachineHealthCheckOptions struct {
	ClusterName            string
	MachineHealthCheckName string
	Namespace              string
	MatchLabel             string
}

// SetMachineHealthCheckOptions machinehealthcheck setter options
type SetMachineHealthCheckOptions struct {
	ClusterName            string
	Namespace              string
	MachineHealthCheckName string
	MatchLables            []string
	UnhealthyConditions    []string
	NodeStartupTimeout     string
}

// MachineHealthCheck object
type MachineHealthCheck struct {
	Name      string                        `json:"name"`
	Namespace string                        `json:"namespace"`
	Spec      capi.MachineHealthCheckSpec   `json:"spec"`
	Status    capi.MachineHealthCheckStatus `json:"status"`
}

// DeleteMachineHealthCheck delete machinehealthcheck
func (c *TkgClient) DeleteMachineHealthCheck(options MachineHealthCheckOptions) error {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "cannot get current management cluster context")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client")
	}

	cluster, err := findCluster(clusterClient, options.ClusterName, options.Namespace)
	if err != nil {
		return err
	}

	if cluster.Spec.Topology != nil {
		return errors.New("deleting machine health checks on clusterclass based clusters is not supported")
	}

	if options.Namespace == "" {
		options.Namespace = cluster.Namespace
		log.Infof("use %s as default namespace", options.Namespace)
	}

	return c.DeleteMachineHealthCheckWithClusterClient(clusterClient, options)
}

// DeleteMachineHealthCheckWithClusterClient delete machinehealthcheck with client
func (c *TkgClient) DeleteMachineHealthCheckWithClusterClient(clusterClient clusterclient.Client, options MachineHealthCheckOptions) error {
	candidates, err := getMachineHealthCheckCandidates(clusterClient, options.ClusterName, options.Namespace, options.MachineHealthCheckName, options.MatchLabel)
	if err != nil {
		return errors.Wrap(err, "unable to get the MachineHealthCheck object")
	}

	if len(candidates) == 0 {
		return errors.Errorf("MachineHealthCheck not found for cluster %s in namespace %s", options.ClusterName, options.Namespace)
	}

	if len(candidates) > 1 {
		return errors.Errorf("multiple MachineHealthCheck found for cluster %s in namespace %s", options.ClusterName, options.Namespace)
	}

	mhc := &capi.MachineHealthCheck{}
	mhc.Name = candidates[0].Name
	mhc.Namespace = candidates[0].Namespace
	return clusterClient.DeleteResource(mhc)
}

// GetMachineHealthChecks gets machinehealthcheck
func (c *TkgClient) GetMachineHealthChecks(options MachineHealthCheckOptions) ([]MachineHealthCheck, error) {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return []MachineHealthCheck{}, errors.Wrap(err, "cannot get current management cluster context")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return []MachineHealthCheck{}, errors.Wrap(err, "unable to get cluster client while getting machine health checks")
	}

	_, err = findCluster(clusterClient, options.ClusterName, options.Namespace)
	if err != nil {
		return []MachineHealthCheck{}, err
	}

	return c.GetMachineHealthChecksWithClusterClient(clusterClient, options)
}

// GetMachineHealthChecksWithClusterClient gets machinehealthcheck with client
func (c *TkgClient) GetMachineHealthChecksWithClusterClient(clusterClient clusterclient.Client, options MachineHealthCheckOptions) ([]MachineHealthCheck, error) {
	res := []MachineHealthCheck{}

	candidates, err := getMachineHealthCheckCandidates(clusterClient, options.ClusterName, options.Namespace, options.MachineHealthCheckName, options.MatchLabel)
	if err != nil {
		return res, nil
	}

	for i := range candidates {
		obj := MachineHealthCheck{
			Name:      candidates[i].ObjectMeta.Name,
			Namespace: candidates[i].ObjectMeta.Namespace,
			Spec:      candidates[i].Spec,
			Status:    candidates[i].Status,
		}
		res = append(res, obj)
	}

	return res, nil
}

func getMachineHealthCheckCandidates(clusterClient clusterclient.Client, clusterName, namespace, mhcName, matchLabel string) ([]capi.MachineHealthCheck, error) {
	mhcList := &capi.MachineHealthCheckList{}
	err := clusterClient.ListResources(mhcList, &crtclient.ListOptions{Namespace: namespace})
	if err != nil {
		return []capi.MachineHealthCheck{}, err
	}

	candidates := []capi.MachineHealthCheck{}

	for i := range mhcList.Items {
		if mhcList.Items[i].Spec.ClusterName == clusterName &&
			(namespace == "" || namespace == mhcList.Items[i].Namespace) &&
			(mhcName == "" || mhcName == mhcList.Items[i].Name) {
			_, ok := mhcList.Items[i].Spec.Selector.MatchLabels[matchLabel]
			if matchLabel == "" || ok {
				candidates = append(candidates, mhcList.Items[i])
			}
		}
	}

	return candidates, nil
}

func findCluster(clusterClient clusterclient.Client, clusterName, namespace string) (capi.Cluster, error) {
	var result capi.Cluster
	clusters, err := clusterClient.ListClusters(namespace)
	if err != nil {
		return result, err
	}

	count := 0
	for i := range clusters {
		if clusters[i].Name == clusterName {
			count++
			result = clusters[i]
		}
	}

	if count > 1 {
		return result, errors.Errorf("found multiple clusters with name %s, please specify a namespace", clusterName)
	}
	if count == 0 {
		return result, errors.Errorf("cluster %s not found", clusterName)
	}

	return result, nil
}

// SetMachineHealthCheck sets machinehealthcheck
func (c *TkgClient) SetMachineHealthCheck(options *SetMachineHealthCheckOptions) error {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "cannot get current management cluster context")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client")
	}

	cluster, err := findCluster(clusterClient, options.ClusterName, options.Namespace)
	if err != nil {
		return err
	}

	if cluster.Spec.Topology != nil {
		return errors.New("setting machine health checks on clusterclass based clusters is not supported")
	}

	if options.Namespace == "" {
		options.Namespace = cluster.Namespace
		log.Infof("use %s as default namespace", options.Namespace)
	}

	candidates, err := getMachineHealthCheckCandidates(clusterClient, options.ClusterName, options.Namespace, options.MachineHealthCheckName, "")
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return c.CreateMachineHealthCheck(clusterClient, options)
	}

	if len(candidates) > 1 {
		return errors.Errorf("multiple MachineHealthCheck found for cluster %s in namespace %s", options.ClusterName, options.Namespace)
	}
	return c.UpdateMachineHealthCheck(&candidates[0], clusterClient, options)
}

// CreateMachineHealthCheck creates machinehealthchecks
func (c *TkgClient) CreateMachineHealthCheck(clusterClient clusterclient.Client, options *SetMachineHealthCheckOptions) error {
	var mhc capi.MachineHealthCheck
	err := setDefaultSetMachineHealthCheckOptions(options)
	if err != nil {
		return err
	}

	mhc.Name = options.MachineHealthCheckName
	mhc.Namespace = options.Namespace
	mhc.Spec.ClusterName = options.ClusterName

	err = setNodeStartupTimeout(&mhc, options.NodeStartupTimeout)
	if err != nil {
		return err
	}

	for _, uhc := range options.UnhealthyConditions {
		condition, err := getUnhealthConditions(uhc)
		if err != nil {
			return err
		}
		item := condition
		mhc.Spec.UnhealthyConditions = append(mhc.Spec.UnhealthyConditions, item)
	}

	err = setMatchLabels(&mhc, options.MatchLables)
	if err != nil {
		return err
	}

	return clusterClient.CreateResource(&mhc, mhc.Name, mhc.Namespace)
}

func setDefaultSetMachineHealthCheckOptions(options *SetMachineHealthCheckOptions) error {
	if options.MachineHealthCheckName == "" {
		options.MachineHealthCheckName = options.ClusterName
	}

	if len(options.MatchLables) == 0 {
		options.MatchLables = append(options.MatchLables, fmt.Sprintf("%s:%s-%s", NodePoolKey, options.ClusterName, DefaultNodePoolSuffix))
	}

	if options.NodeStartupTimeout == "" {
		options.NodeStartupTimeout = DefaultNodeStartupTimeout
	}

	if len(options.UnhealthyConditions) == 0 {
		options.UnhealthyConditions = append(options.UnhealthyConditions, fmt.Sprintf("%s:%s:%s", string(corev1.NodeReady), string(corev1.ConditionFalse), DefaulUnhealthyConditionTimeout), fmt.Sprintf("%s:%s:%s", string(corev1.NodeReady), string(corev1.ConditionUnknown), DefaulUnhealthyConditionTimeout))
	}

	return nil
}

// UpdateMachineHealthCheck updates machinehealthcheck
func (c *TkgClient) UpdateMachineHealthCheck(candidate *capi.MachineHealthCheck, clusterClient clusterclient.Client, options *SetMachineHealthCheckOptions) error {
	oldMHC := candidate
	newMHC := capi.MachineHealthCheck{}
	newMHC.Spec = *oldMHC.Spec.DeepCopy()

	err := setMatchLabels(&newMHC, options.MatchLables)
	if err != nil {
		return err
	}

	for _, newCondition := range options.UnhealthyConditions {
		condition, err := getUnhealthConditions(newCondition)
		if err != nil {
			return err
		}

		exist := false
		for i, oldCondition := range oldMHC.Spec.UnhealthyConditions {
			if oldCondition.Type == condition.Type && oldCondition.Status == condition.Status {
				newMHC.Spec.UnhealthyConditions[i].Timeout = condition.Timeout
				exist = true
			}
		}

		if !exist {
			item := condition
			newMHC.Spec.UnhealthyConditions = append(newMHC.Spec.UnhealthyConditions, item)
		}
	}

	if options.NodeStartupTimeout != "" {
		err = setNodeStartupTimeout(&newMHC, options.NodeStartupTimeout)
		if err != nil {
			return err
		}
	}

	bytes, err := json.Marshal(newMHC)
	if err != nil {
		return err
	}

	return clusterClient.PatchResource(oldMHC, oldMHC.Name, oldMHC.Namespace, string(bytes), types.MergePatchType, nil)
}

func getUnhealthConditions(condition string) (capi.UnhealthyCondition, error) {
	strs := strings.Split(condition, ":")
	if len(strs) != 3 {
		return capi.UnhealthyCondition{}, errors.New("please specify the unhealthConditions using the following format:  NodeConditionType:ConditionStatus:Timeout")
	}

	if _, ok := UnhealthyConditions[strs[0]]; !ok {
		return capi.UnhealthyCondition{}, errors.Errorf("NodeConditionType %s is not supported", strs[0])
	}

	if _, ok := ConditionStatus[strs[1]]; !ok {
		return capi.UnhealthyCondition{}, errors.Errorf("ConditionStatus %s is not supported", strs[1])
	}

	d, err := time.ParseDuration(strs[2])
	if err != nil {
		return capi.UnhealthyCondition{}, errors.Wrap(err, "cannot parse the timeout value")
	}

	return capi.UnhealthyCondition{
		Type:    corev1.NodeConditionType(strings.TrimSpace(strs[0])),
		Status:  corev1.ConditionStatus(strings.TrimSpace(strs[1])),
		Timeout: metav1.Duration{Duration: d},
	}, nil
}

func setMatchLabels(mhc *capi.MachineHealthCheck, labels []string) error {
	if mhc.Spec.Selector.MatchLabels == nil {
		mhc.Spec.Selector.MatchLabels = make(map[string]string)
	}
	for _, label := range labels {
		strs := strings.Split(label, ":")
		if len(strs) != 2 {
			return errors.New("please specify the matchLabels using the following format: label-key:label-value")
		}
		mhc.Spec.Selector.MatchLabels[strings.TrimSpace(strs[0])] = strings.TrimSpace(strs[1])
	}
	return nil
}

func setNodeStartupTimeout(mhc *capi.MachineHealthCheck, timeout string) error {
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return errors.Wrap(err, "cannot parse the timeout value")
	}
	mhc.Spec.NodeStartupTimeout = &metav1.Duration{Duration: d}
	return nil
}
