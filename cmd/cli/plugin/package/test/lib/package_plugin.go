// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	clitest "github.com/vmware-tanzu/tanzu-framework/cli/runtime/test"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

type PackagePluginBase interface {
	AddRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult
	GetRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult
	UpdateRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult
	DeleteRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult
	ListRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult

	GetAvailablePackage(packageName string, o *packagedatamodel.PackageAvailableOptions) PackagePluginResult
	ListAvailablePackage(packageName string, o *packagedatamodel.PackageAvailableOptions) PackagePluginResult

	CreateInstalledPackage(o *packagedatamodel.PackageOptions) PackagePluginResult
	GetInstalledPackage(o *packagedatamodel.PackageOptions) PackagePluginResult
	UpdateInstalledPackage(o *packagedatamodel.PackageOptions) PackagePluginResult
	DeleteInstalledPackage(o *packagedatamodel.PackageOptions) PackagePluginResult
	ListInstalledPackage(o *packagedatamodel.PackageOptions) PackagePluginResult
}

type PackagePluginHelpers interface {
	AddOrUpdateRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult
	CheckRepositoryAvailable(o *packagedatamodel.RepositoryOptions) PackagePluginResult
	CheckAndDeleteRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult
	CheckRepositoryDeleted(o *packagedatamodel.RepositoryOptions) PackagePluginResult
	CheckPackageAvailable(packageName string, o *packagedatamodel.PackageAvailableOptions) PackagePluginResult
	CheckAndInstallPackage(o *packagedatamodel.PackageOptions) PackagePluginResult
	CheckPackageInstalled(o *packagedatamodel.PackageOptions) PackagePluginResult
	CheckAndUninstallPackage(o *packagedatamodel.PackageOptions) PackagePluginResult
	CheckPackageDeleted(o *packagedatamodel.PackageOptions) PackagePluginResult
}

type PackagePlugin interface {
	// Base commands are implemented by the below interface
	PackagePluginBase
	// Extra helper commands will be implemented by other interface
	PackagePluginHelpers
}

type PackagePluginResult struct {
	Pass   bool
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
	Error  error
}

type packagePlugin struct {
	kubeconfigPath string
	interval       time.Duration
	timeout        time.Duration
	outputFormat   string
	logFile        string
	verbose        int32
}

func NewPackagePlugin(kubeconfigPath string,
	interval time.Duration,
	timeout time.Duration,
	outputFormat string,
	logFile string,
	verbose int32) PackagePlugin {
	return &packagePlugin{
		kubeconfigPath: kubeconfigPath,
		interval:       interval,
		timeout:        timeout,
		outputFormat:   outputFormat,
		logFile:        logFile,
		verbose:        verbose,
	}
}

func (p *packagePlugin) addKubeconfig(cmd string) string {
	if cmd != "" && p.kubeconfigPath != "" {
		cmd += fmt.Sprintf(" --kubeconfig %s", p.kubeconfigPath)
	}
	return cmd
}

func (p *packagePlugin) addOutputFormat(cmd string) string {
	if cmd != "" && p.outputFormat != "" {
		cmd += fmt.Sprintf(" --output %s", p.outputFormat)
	}
	return cmd
}

func (p *packagePlugin) addLogFile(cmd string) string {
	if cmd != "" && p.logFile != "" {
		cmd += fmt.Sprintf(" --log-file %s", p.logFile)
	}
	return cmd
}

func (p *packagePlugin) addVerbose(cmd string) string {
	if cmd != "" && p.verbose != 0 {
		cmd += fmt.Sprintf(" --verbose %d", p.verbose)
	}
	return cmd
}

func (p *packagePlugin) addGlobalOptions(cmd string) string {
	cmd = p.addLogFile(cmd)
	cmd = p.addVerbose(cmd)
	return cmd
}

func (p *packagePlugin) AddRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package repository add %s --url %s", o.RepositoryName, o.RepositoryURL)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.CreateNamespace {
		cmd += fmt.Sprintf(" --create-namespace")
	}
	if o.Wait {
		cmd += fmt.Sprintf(" --wait=true")
	}
	if o.PollInterval != 0 {
		cmd += fmt.Sprintf(" --poll-interval %s", o.PollInterval)
	}
	if o.PollTimeout != 0 {
		cmd += fmt.Sprintf(" --poll-timeout %s", o.PollTimeout)
	}
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) GetRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package repository get %s", o.RepositoryName)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) UpdateRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package repository update %s --url %s", o.RepositoryName, o.RepositoryURL)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.CreateNamespace {
		cmd += fmt.Sprintf(" --create-namespace")
	}
	if o.CreateRepository {
		cmd += fmt.Sprintf(" --create")
	}
	if o.Wait {
		cmd += fmt.Sprintf(" --wait=true")
	}
	if o.PollInterval != 0 {
		cmd += fmt.Sprintf(" --poll-interval %s", o.PollInterval)
	}
	if o.PollTimeout != 0 {
		cmd += fmt.Sprintf(" --poll-timeout %s", o.PollTimeout)
	}
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) DeleteRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package repository delete %s", o.RepositoryName)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.IsForceDelete {
		cmd += fmt.Sprintf(" --force")
	}
	if o.Wait {
		cmd += fmt.Sprintf(" --wait=true")
	}
	if o.PollInterval != 0 {
		cmd += fmt.Sprintf(" --poll-interval %s", o.PollInterval)
	}
	if o.PollTimeout != 0 {
		cmd += fmt.Sprintf(" --poll-timeout %s", o.PollTimeout)
	}
	if o.SkipPrompt {
		cmd += fmt.Sprintf(" -y")
	}
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) ListRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package repository list")
	if o.AllNamespaces {
		cmd += fmt.Sprintf(" --all-namespaces")
	}
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) GetAvailablePackage(packageName string, o *packagedatamodel.PackageAvailableOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package available get %s", packageName)

	if o.ValuesSchema {
		cmd += fmt.Sprintf(" --values-schema")
	}
	if o.AllNamespaces {
		cmd += fmt.Sprintf(" --all-namespaces")
	}
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) ListAvailablePackage(packageName string, o *packagedatamodel.PackageAvailableOptions) PackagePluginResult {
	var (
		cmd    string
		result PackagePluginResult
	)
	if packageName == "" {
		cmd = fmt.Sprintf("tanzu package available list")
	} else {
		cmd = fmt.Sprintf("tanzu package available list %s", packageName)
	}
	if o.AllNamespaces {
		cmd += fmt.Sprintf(" --all-namespaces")
	}
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) CreateInstalledPackage(o *packagedatamodel.PackageOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package installed create %s --package-name %s", o.PkgInstallName, o.PackageName)
	if o.Version != "" {
		cmd += fmt.Sprintf(" --version %s", o.Version)
	}
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.CreateNamespace {
		cmd += fmt.Sprintf(" --create-namespace")
	}
	if o.ValuesFile != "" {
		cmd += fmt.Sprintf(" --values-file %s", o.ValuesFile)
	}
	if o.ServiceAccountName != "" {
		cmd += fmt.Sprintf(" --service-account-name %s", o.ServiceAccountName)
	}
	if o.Wait {
		cmd += fmt.Sprintf(" --wait=true")
	}
	if o.PollInterval != 0 {
		cmd += fmt.Sprintf(" --poll-interval %s", o.PollInterval)
	}
	if o.PollTimeout != 0 {
		cmd += fmt.Sprintf(" --poll-timeout %s", o.PollTimeout)
	}
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) GetInstalledPackage(o *packagedatamodel.PackageOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package installed get %s", o.PkgInstallName)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) UpdateInstalledPackage(o *packagedatamodel.PackageOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package installed update %s --package-name %s", o.PkgInstallName, o.PackageName)
	if o.Version != "" {
		cmd += fmt.Sprintf(" --version %s", o.Version)
	}
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.Install {
		cmd += fmt.Sprintf(" --install")
	}
	if o.ValuesFile != "" {
		cmd += fmt.Sprintf(" --values-file %s", o.ValuesFile)
	}
	if o.Wait {
		cmd += fmt.Sprintf(" --wait=true")
	}
	if o.PollInterval != 0 {
		cmd += fmt.Sprintf(" --poll-interval %s", o.PollInterval)
	}
	if o.PollTimeout != 0 {
		cmd += fmt.Sprintf(" --poll-timeout %s", o.PollTimeout)
	}
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) DeleteInstalledPackage(o *packagedatamodel.PackageOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package installed delete %s", o.PkgInstallName)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.PollInterval != 0 {
		cmd += fmt.Sprintf(" --poll-interval %s", o.PollInterval)
	}
	if o.PollTimeout != 0 {
		cmd += fmt.Sprintf(" --poll-timeout %s", o.PollTimeout)
	}
	if o.SkipPrompt {
		cmd += fmt.Sprintf(" -y")
	}
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) ListInstalledPackage(o *packagedatamodel.PackageOptions) PackagePluginResult {
	var result PackagePluginResult
	cmd := fmt.Sprintf("tanzu package installed list")
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.AllNamespaces {
		cmd += fmt.Sprintf(" -A")
	}
	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeconfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *packagePlugin) AddOrUpdateRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	getResult := p.GetRepository(&packagedatamodel.RepositoryOptions{
		RepositoryName: o.RepositoryName,
		Namespace:      o.Namespace,
	})
	if getResult.Error != nil {
		return p.AddRepository(o)
	} else {
		if !strings.Contains(getResult.Stdout.String(), o.RepositoryURL) {
			return p.UpdateRepository(o)
		}
	}
	return result
}

func (p *packagePlugin) CheckRepositoryAvailable(o *packagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	if err := wait.PollImmediate(p.interval, p.timeout, func() (done bool, err error) {
		result := p.GetRepository(&packagedatamodel.RepositoryOptions{
			RepositoryName: o.RepositoryName,
			Namespace:      o.Namespace,
		})
		if result.Error != nil {
			return false, result.Error
		}
		if result.Stdout != nil && strings.Contains(result.Stdout.String(), "Reconcile succeeded") {
			return true, nil
		}
		return false, nil
	}); err != nil {
		if result.Error == nil {
			result.Error = err
		}
		return result
	}
	return result
}

func (p *packagePlugin) CheckAndDeleteRepository(o *packagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	getResult := p.GetRepository(&packagedatamodel.RepositoryOptions{
		RepositoryName: o.RepositoryName,
		Namespace:      o.Namespace,
	})
	if getResult.Error == nil {
		return p.DeleteRepository(o)
	}
	return result
}

func (p *packagePlugin) CheckRepositoryDeleted(o *packagedatamodel.RepositoryOptions) PackagePluginResult {
	var result PackagePluginResult
	if err := wait.PollImmediate(p.interval, p.timeout, func() (done bool, err error) {
		result := p.GetRepository(&packagedatamodel.RepositoryOptions{
			RepositoryName: o.RepositoryName,
			Namespace:      o.Namespace,
		})
		if result.Stderr != nil && strings.Contains(result.Stderr.String(), "does not exist") ||
			result.Stdout != nil && strings.Contains(result.Stdout.String(), "does not exist") {
			if result.Error != nil {
				// Setting result error to nil since there will be error on get after repository is deleted
				result.Error = nil
			}
			return true, nil
		}
		return false, result.Error
	}); err != nil {
		if result.Error == nil {
			result.Error = err
		}
		return result
	}
	return result
}

func (p *packagePlugin) CheckPackageAvailable(packageName string, o *packagedatamodel.PackageAvailableOptions) PackagePluginResult {
	var result PackagePluginResult
	if err := wait.PollImmediate(p.interval, p.timeout, func() (done bool, err error) {
		result = p.GetAvailablePackage(packageName, o)
		if result.Error == nil {
			return true, nil
		}
		return false, nil
	}); err != nil {
		if result.Error == nil {
			result.Error = err
		}
		return result
	}
	return result
}

func (p *packagePlugin) CheckAndInstallPackage(o *packagedatamodel.PackageOptions) PackagePluginResult {
	var result PackagePluginResult
	getResult := p.GetInstalledPackage(&packagedatamodel.PackageOptions{
		PkgInstallName: o.PkgInstallName,
		Namespace:      o.Namespace,
	})
	if getResult.Error != nil {
		return p.CreateInstalledPackage(o)
	}
	return result
}

func (p *packagePlugin) CheckPackageInstalled(o *packagedatamodel.PackageOptions) PackagePluginResult {
	var result PackagePluginResult
	if err := wait.PollImmediate(p.interval, p.timeout, func() (done bool, err error) {
		result = p.GetInstalledPackage(&packagedatamodel.PackageOptions{
			PkgInstallName: o.PkgInstallName,
			Namespace:      o.Namespace,
		})
		if result.Error != nil {
			return false, result.Error
		}
		if result.Stdout != nil && strings.Contains(result.Stdout.String(), "Reconcile succeeded") {
			return true, nil
		}
		return false, nil
	}); err != nil {
		if result.Error == nil {
			result.Error = err
		}
		return result
	}
	return result
}

func (p *packagePlugin) CheckAndUninstallPackage(o *packagedatamodel.PackageOptions) PackagePluginResult {
	var result PackagePluginResult
	getResult := p.GetInstalledPackage(&packagedatamodel.PackageOptions{
		PkgInstallName: o.PkgInstallName,
		Namespace:      o.Namespace,
	})
	if getResult.Error == nil {
		return p.DeleteInstalledPackage(o)
	}
	return result
}

func (p *packagePlugin) CheckPackageDeleted(o *packagedatamodel.PackageOptions) PackagePluginResult {
	var result PackagePluginResult
	if err := wait.PollImmediate(p.interval, p.timeout, func() (done bool, err error) {
		result = p.GetInstalledPackage(&packagedatamodel.PackageOptions{
			PkgInstallName: o.PkgInstallName,
			Namespace:      o.Namespace,
		})
		if result.Stderr != nil && strings.Contains(result.Stderr.String(), "does not exist") {
			if result.Error != nil {
				// Setting result error to nil since there will be error on get after package is deleted
				result.Error = nil
			}
			return true, nil
		}
		return false, result.Error
	}); err != nil {
		if result.Error == nil {
			result.Error = err
		}
		return result
	}
	return result
}
