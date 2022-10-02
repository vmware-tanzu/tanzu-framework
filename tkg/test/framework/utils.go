// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo" // nolint:stylecheck
	"github.com/pkg/errors"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

// CreateClusterOptions represent options to create a TKG cluster
type CreateClusterOptions struct {
	GenerateOnly                bool
	SkipPrompt                  bool
	ControlPlaneMachineCount    int
	WorkerMachineCount          int
	ClusterName                 string
	Plan                        string
	InfrastructureProvider      string
	Namespace                   string
	KubernetesVersion           string
	Size                        string
	ControlPlaneSize            string
	WorkerSize                  string
	CniType                     string
	EnableClusterOptions        string
	VsphereControlPlaneEndpoint string
	Timeout                     time.Duration

	OtherConfigs map[string]string
}

// WaitForNodes waits for desiredCount number of nodes to be ready
func WaitForNodes(proxy *ClusterProxy, desiredCount int) {
	const timeout = 10 * time.Minute

	start := time.Now()
	for time.Since(start) < timeout {
		count := len(proxy.GetClusterNodes())
		_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("Node count for cluster %q: %d\n", proxy.name, count)))
		if count == desiredCount {
			return
		}

		time.Sleep(30 * time.Second) // nolint:gomnd
	}

	Fail(fmt.Sprintf("Timed out waiting for nodes count to reach %q", desiredCount))
}

// GetClusterClass returns ClusterClass used by the Cluster
func GetClusterClass(proxy *ClusterProxy, clusterName string, namespace string) (string, error) {
	var err error
	c := proxy.GetClient()
	cluster := &clusterv1.Cluster{}

	err = c.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: clusterName}, cluster)
	if err != nil {
		return "", err
	}

	if cluster.Spec.Topology == nil {
		return "", nil
	}

	return cluster.Spec.Topology.Class, nil
}

// GetTempClusterConfigFile gets temporary config file
func GetTempClusterConfigFile(clusterConfigFile string, options *CreateClusterOptions) (string, error) { // nolint:gocyclo
	clusterOptions := map[string]string{}

	_, err := os.Stat(clusterConfigFile)
	if err == nil {
		yamlFile, err := os.ReadFile(clusterConfigFile)
		if err != nil {
			return "", err
		}

		err = yaml.Unmarshal(yamlFile, clusterOptions)
		if err != nil {
			return "", err
		}
	}

	if options.ClusterName != "" {
		clusterOptions["CLUSTER_NAME"] = options.ClusterName
	}

	if options.InfrastructureProvider != "" {
		clusterOptions["INFRASTRUCTURE_PROVIDER"] = options.InfrastructureProvider
	}

	if options.KubernetesVersion != "" {
		clusterOptions["KUBERNETES_VERSION"] = options.KubernetesVersion
	}

	if options.Size != "" {
		clusterOptions["SIZE"] = options.Size
	}

	if options.ControlPlaneSize != "" {
		clusterOptions["CONTROLPLANE_SIZE"] = options.ControlPlaneSize
	}

	if options.WorkerSize != "" {
		clusterOptions["WORKER_SIZE"] = options.WorkerSize
	}

	if options.CniType != "" {
		clusterOptions["CNI"] = options.CniType
	}

	if options.Plan != "" {
		clusterOptions["CLUSTER_PLAN"] = options.Plan
	}

	if options.Namespace != "" {
		clusterOptions["NAMESPACE"] = options.Namespace
	}

	if options.EnableClusterOptions != "" {
		clusterOptions["ENABLE_CLUSTER_OPTIONS"] = options.EnableClusterOptions
	}

	if options.ControlPlaneMachineCount != 0 {
		clusterOptions["CONTROL_PLANE_MACHINE_COUNT"] = strconv.Itoa(options.ControlPlaneMachineCount)
	}

	if options.WorkerMachineCount != 0 {
		clusterOptions["WORKER_MACHINE_COUNT"] = strconv.Itoa(options.WorkerMachineCount)
	}

	if options.VsphereControlPlaneEndpoint != "" {
		clusterOptions["VSPHERE_CONTROL_PLANE_ENDPOINT"] = options.VsphereControlPlaneEndpoint
	}

	if options.OtherConfigs != nil {
		for k, v := range options.OtherConfigs {
			clusterOptions[k] = v
		}
	}

	out, err := yaml.Marshal(clusterOptions)
	if err != nil {
		return "", err
	}

	f, err := os.CreateTemp("", "temp_cluster_config_"+util.RandomString(4)+".yaml") // nolint:gomnd
	if err != nil {
		return "", err
	}

	if _, err := f.Write(out); err != nil {
		return "", err
	}

	configFilePath, err := filepath.Abs(f.Name())
	if err != nil {
		return "", err
	}

	return configFilePath, nil
}

// SetCliConfigFlag sets cli flag
func SetCliConfigFlag(flagName string, value string) error {
	// Acquire tanzu config lock
	config.AcquireTanzuConfigLock()
	defer config.ReleaseTanzuConfigLock()

	cfg, err := config.GetClientConfigNoLock()
	if err != nil {
		return err
	}
	paramArray := strings.Split(flagName, ".")
	err = SetFeatures(cfg, paramArray, value)
	if err != nil {
		return err
	}
	return config.StoreClientConfig(cfg)
}

// SetFeatures sets flags in given ClientConfig
func SetFeatures(cfg *configapi.ClientConfig, paramArray []string, value string) error {
	if len(paramArray) != 3 {
		return errors.New("unable to parse config path parameter into three parts [" + strings.Join(paramArray, ".") + "]  (was expecting 'features.<plugin>.<feature>'")
	}
	plugin := paramArray[1]
	featureName := paramArray[2]

	if cfg.ClientOptions == nil {
		cfg.ClientOptions = &configapi.ClientOptions{}
	}
	if cfg.ClientOptions.Features == nil {
		cfg.ClientOptions.Features = make(map[string]configapi.FeatureMap)
	}
	if cfg.ClientOptions.Features[plugin] == nil {
		cfg.ClientOptions.Features[plugin] = configapi.FeatureMap{}
	}
	cfg.ClientOptions.Features[plugin][featureName] = value
	return nil
}
