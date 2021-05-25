// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgcli

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/framework/exec"
)

type Client struct {
	TkgCliPath    string
	TkgConfigPath string
	LogLocation   string
	LogLevel      int32
	Env           []string
}

type InitOptions struct {
	Infrastructure              string
	InfrastructureVersion       string
	Plan                        string
	UseExistingBootstrapCluster bool
	Name                        string
	Timeout                     string
	Size                        string
	ControlPlaneSize            string
	WorkerSize                  string
	CeipOptIn                   bool
	DeployTKGonVsphere7         bool
	EnableTKGSOnVsphere7        bool
	VsphereControlPlaneEndpoint string

	// hidden flags
	Cni string
}

type CreateClusterOptions struct {
	Name                        string
	Plan                        string
	KubernetesVersion           string
	ControlPlaneMachineCount    int
	WorkerMachineCount          int
	DryRun                      bool
	Namespace                   string
	Timeout                     string
	Manifest                    string
	Cni                         string
	Size                        string
	ControlPlaneSize            string
	WorkerSize                  string
	ClusterOptions              string
	VsphereControlPlaneEndpoint string
}

type Mhc struct {
	MatchLabels         string
	MhcName             string
	Namespace           string
	NodeStartupTimeout  string
	UnhealthyConditions string
}

func NewClient(cliPath string, tkgConfigPath string, logLocation string, logLevel int32) *Client {
	return &Client{
		TkgCliPath:    cliPath,
		TkgConfigPath: tkgConfigPath,
		LogLocation:   logLocation,
		LogLevel:      logLevel,
	}
}

func (c *Client) SetEnv(env ...string) {
	c.Env = env
}

func (c *Client) NewTkgCliCommand(opts ...Option) *Command {
	// introducing random sleep to decrease errors due to parallel test execution
	rand.Seed(time.Now().UnixNano())
	rand := rand.Intn(10)
	time.Sleep(time.Duration(rand) * time.Second)

	cmd := NewCommand(opts...)
	cmd.Env = c.Env
	cmd.Cmd = c.TkgCliPath

	cmd.Args = append(cmd.Args, "--config", c.TkgConfigPath)
	if c.LogLevel != 0 {
		cmd.Args = append(cmd.Args, "-v", fmt.Sprint(c.LogLevel))
	}
	return cmd
}

func (c *Client) Init(ctx context.Context, options InitOptions) error {
	initArgs := buildArgsFromInitOptions(options)

	logLocation := filepath.Join(c.LogLocation, options.Name+".log")
	initArgs = append(initArgs, "--log_file", logLocation)

	initCmd := c.NewTkgCliCommand(
		WithArgs(initArgs...),
		WithStdout(GinkgoWriter),
	)

	err := initCmd.RunAndRedirectOutput(ctx)
	return err
}

func (c *Client) GetManagementClusters(ctx context.Context) error {
	args := []string{"get", "mc"}

	cmd := c.NewTkgCliCommand(
		WithArgs(args...),
	)

	err := cmd.RunAndRedirectOutput(ctx)
	return err
}

func (c *Client) GenerateClusterConfiguration(ctx context.Context, options CreateClusterOptions) error {
	args := buildArgsFromCreateClusterOptions(options)
	args = append([]string{"config", "cluster", options.Name}, args...)

	cmd := c.NewTkgCliCommand(
		WithArgs(args...),
	)

	_, _, err := cmd.Run(ctx)
	return err
}

func (c *Client) CreateWorkloadCluster(ctx context.Context, options CreateClusterOptions) error {
	args := buildArgsFromCreateClusterOptions(options)
	args = append([]string{"create", "cluster", options.Name}, args...)

	logLocation := filepath.Join(c.LogLocation, options.Name+".log")
	args = append(args, "--log_file", logLocation, "--yes")

	cmd := c.NewTkgCliCommand(
		WithArgs(args...),
		WithStdout(GinkgoWriter),
	)

	err := cmd.RunAndRedirectOutput(ctx)
	return err
}

func (c *Client) ScaleCluster(ctx context.Context, clusterName string, namespace string, controlplaneCount int, workerCount int) error {
	args := []string{"scale", "cluster", clusterName}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	if controlplaneCount > 0 {
		args = append(args, "-c", strconv.Itoa(controlplaneCount))
	}

	if workerCount > 0 {
		args = append(args, "-w", strconv.Itoa(workerCount))
	}

	cmd := c.NewTkgCliCommand(
		WithArgs(args...),
		WithStdout(GinkgoWriter),
	)

	err := cmd.RunAndRedirectOutput(ctx)
	return err
}

func (c *Client) GetClusterCredentials(ctx context.Context, name string, namespace string) error {
	args := []string{"get", "credentials", name}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	cmd := c.NewTkgCliCommand(
		WithArgs(args...),
		WithStdout(GinkgoWriter),
	)

	err := cmd.RunAndRedirectOutput(ctx)
	return err
}

func (c *Client) GetClusterCredentialsInFile(ctx context.Context, name string, namespace string, filePath string) error {
	err := c.GetClusterCredentials(ctx, name, namespace)
	if err != nil {
		return err
	}

	writer, err := os.Create(filePath)
	if err != nil {
		return err
	}

	contextName := name + "-admin@" + name
	args := []string{"config", "view", "--minify", "--flatten", "--context=" + contextName}
	cmd := exec.CommandContext(context.TODO(), "kubectl", args...) //nolint:gosec
	cmd.Stdout = writer
	err = cmd.Run()

	return err
}

func (c *Client) UpgradeManagementCluster(ctx context.Context, clusterName string) error {
	args := []string{"upgrade", "mc", clusterName, "--yes"}

	logLocation := filepath.Join(c.LogLocation, clusterName+"-upgrade"+".log")
	args = append(args, "--log_file", logLocation)

	upgradeCmd := c.NewTkgCliCommand(
		WithArgs(args...),
		WithStdout(GinkgoWriter),
	)

	err := upgradeCmd.RunAndRedirectOutput(ctx)
	return err
}

func (c *Client) UpgradeWorkloadCluster(ctx context.Context, clusterName string, namespace string, k8sVersion string) error {
	args := []string{"upgrade", "cluster", clusterName, "--yes"}
	if namespace != "" {
		args = append(args, namespace)
	}

	if k8sVersion != "" {
		args = append(args, "--kubernetes-version", k8sVersion)
	}

	logLocation := filepath.Join(c.LogLocation, clusterName+"-upgrade"+".log")
	args = append(args, "--log_file", logLocation)

	upgradeCmd := c.NewTkgCliCommand(
		WithArgs(args...),
		WithStdout(GinkgoWriter),
	)

	err := upgradeCmd.RunAndRedirectOutput(ctx)
	return err
}

func (c *Client) DeleteManagementCluster(ctx context.Context, name string, force bool) error {
	args := []string{"delete", "mc", name, "--yes"}
	if force {
		args = append(args, "--force")
	}

	deleteCmd := c.NewTkgCliCommand(
		WithArgs(args...),
		WithStdout(GinkgoWriter),
	)

	err := deleteCmd.RunAndRedirectOutput(ctx)
	return err
}

func (c *Client) DeleteWorkloadCluster(ctx context.Context, name string, namespace string) error {
	args := []string{"delete", "cluster", name, "--yes"}
	if namespace != "" {
		args = append(args, namespace)
	}

	deleteCmd := c.NewTkgCliCommand(
		WithArgs(args...),
		WithStdout(GinkgoWriter),
	)

	err := deleteCmd.RunAndRedirectOutput(ctx)
	return err
}

func (c *Client) GetMhc(ctx context.Context, name string, namespace string) error {
	args := []string{"get", "mhc", name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	cmd := c.NewTkgCliCommand(
		WithArgs(args...),
		WithStdout(GinkgoWriter),
	)

	err := cmd.RunAndRedirectOutput(ctx)
	return err
}

func (c *Client) SetMhc(ctx context.Context, clusterName string, mhc Mhc) error {
	args := []string{"set", "mhc", clusterName}
	if mhc.Namespace != "" {
		args = append(args, "--namespace", mhc.Namespace)
	}

	if mhc.MatchLabels != "" {
		args = append(args, "--match-labels", mhc.MatchLabels)
	}

	if mhc.MhcName != "" {
		args = append(args, "--mhc-name", mhc.MhcName)
	}

	if mhc.NodeStartupTimeout != "" {
		args = append(args, "--node-startup-timeout", mhc.NodeStartupTimeout)
	}

	if mhc.UnhealthyConditions != "" {
		args = append(args, "--unhealthy-conditions", mhc.UnhealthyConditions)
	}

	cmd := c.NewTkgCliCommand(
		WithArgs(args...),
		WithStdout(GinkgoWriter),
	)

	err := cmd.RunAndRedirectOutput(ctx)
	return err
}

func (c *Client) DeleteMhc(ctx context.Context, clusterName string, mhcName string, namespace string) error {
	args := []string{"delete", "mhc", clusterName, "-y"}
	if mhcName != "" {
		args = append(args, "--mhc-name", mhcName)
	}

	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	cmd := c.NewTkgCliCommand(
		WithArgs(args...),
		WithStdout(GinkgoWriter),
	)

	err := cmd.RunAndRedirectOutput(ctx)
	return err
}

func buildArgsFromCreateClusterOptions(options CreateClusterOptions) []string {
	args := []string{}
	args = append(args, "-p", options.Plan)

	if options.KubernetesVersion != "" {
		args = append(args, "--kubernetes-version", options.KubernetesVersion)
	}

	if options.Timeout != "" {
		args = append(args, "-t", options.Timeout)
	}

	if options.Size != "" {
		args = append(args, "--size", options.Size)
	}

	if options.ControlPlaneSize != "" {
		args = append(args, "--controlplane-size", options.ControlPlaneSize)
	}

	if options.WorkerSize != "" {
		args = append(args, "--worker-size", options.WorkerSize)
	}

	if options.ControlPlaneMachineCount != 0 {
		args = append(args, "--controlplane-machine-count", strconv.Itoa(options.ControlPlaneMachineCount))
	}

	if options.WorkerMachineCount != 0 {
		args = append(args, "--worker-machine-count", strconv.Itoa(options.WorkerMachineCount))
	}

	if options.DryRun {
		args = append(args, "--dry-run", "true")
	}

	if options.Namespace != "" {
		args = append(args, "--namespace", options.Namespace)
	}

	if options.Cni != "" {
		args = append(args, "--cni", options.Cni)
	}

	if options.VsphereControlPlaneEndpoint != "" {
		args = append(args, "--vsphere-controlplane-endpoint", options.VsphereControlPlaneEndpoint)
	}

	if options.ClusterOptions != "" {
		args = append(args, "--enable-cluster-options", options.ClusterOptions)
	}

	return args
}

func buildArgsFromInitOptions(options InitOptions) []string {
	args := []string{"init"}
	if options.InfrastructureVersion != "" {
		options.Infrastructure = options.Infrastructure + ":" + options.InfrastructureVersion
	}
	args = append(args, "-i", options.Infrastructure)
	args = append(args, "-p", options.Plan)

	if options.UseExistingBootstrapCluster {
		args = append(args, "--use-existing-bootstrap-cluster", "true")
	}

	if options.Name != "" {
		args = append(args, "--name", options.Name)
	}

	if options.Timeout != "" {
		args = append(args, "-t", options.Timeout)
	}

	if options.Size != "" {
		args = append(args, "--size", options.Size)
	}

	if options.ControlPlaneSize != "" {
		args = append(args, "--controlplane-size", options.ControlPlaneSize)
	}

	if options.WorkerSize != "" {
		args = append(args, "--worker-size", options.WorkerSize)
	}

	if options.CeipOptIn {
		args = append(args, "--ceip-participation", "true")
	}

	if options.DeployTKGonVsphere7 {
		args = append(args, "--deploy-tkg-on-vSphere7", "true")
	}

	if options.EnableTKGSOnVsphere7 {
		args = append(args, "--enable-tkgs-on-vSphere7", "true")
	}

	if options.VsphereControlPlaneEndpoint != "" {
		args = append(args, "--vsphere-controlplane-endpoint", options.VsphereControlPlaneEndpoint)
	}

	if options.Cni != "" {
		args = append(args, "--cni", options.Cni)
	}

	return args
}
