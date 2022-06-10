// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,stylecheck,nolintlint
package shared

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tj/assert"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"

	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

type E2ECommonSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
	Plan            string
	Namespace       string
	OtherConfigs    map[string]string
}

const (
	clusterNameMng = "ggao-aws-june"
)

func E2ECommonSpec(context context.Context, inputGetter func() E2ECommonSpecInput) { //nolint:funlen
	var (
		err          error
		input        E2ECommonSpecInput
		tkgCtlClient tkgctl.TKGClient
		client       client.Client
		logsDir      string
		clusterName  string
		namespace    string
	)

	BeforeEach(func() { //nolint:dupl
		namespace = constants.DefaultNamespace
		input = inputGetter()
		if input.Namespace != "" {
			namespace = input.Namespace
		}

		logsDir = filepath.Join(input.ArtifactsFolder, "logs")

		rand.Seed(time.Now().UnixNano())
		clusterName = input.E2EConfig.ClusterPrefix + "wc-" + util.RandomString(4) // nolint:gomnd

		tkgCtlClient, err = tkgctl.New(tkgctl.Options{
			ConfigDir: input.E2EConfig.TkgConfigDir,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, clusterName+".log"),
				Verbosity: input.E2EConfig.TkgCliLogLevel,
			},
		})

		Expect(err).To(BeNil())
	})

	It("Should verify basic cluster lifecycle operations", func() {
		By(fmt.Sprintf("Generating workload cluster configuration for cluster %q", clusterName))
		options := framework.CreateClusterOptions{
			ClusterName:  clusterName,
			Namespace:    namespace,
			Plan:         "dev",
			CniType:      input.Cni,
			OtherConfigs: input.OtherConfigs,
		}

		if input.Plan != "" {
			options.Plan = input.Plan
		}

		if input.E2EConfig.InfrastructureName == "vsphere" {
			if input.Cni == "antrea" {
				if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_ANTREA"); ok {
					options.VsphereControlPlaneEndpoint = clusterIP
				}
			}
			if input.Cni == "calico" {
				if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_CALICO"); ok {
					options.VsphereControlPlaneEndpoint = clusterIP
				}
			}
		}
		clusterConfigFile, err := framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, &options)
		Expect(err).To(BeNil())

		defer os.Remove(clusterConfigFile)
		err = tkgCtlClient.ConfigCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
			Edition:           "tkg",
			Namespace:         namespace,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Creating a workload cluster %q", clusterName))

		options = framework.CreateClusterOptions{
			ClusterName:  clusterName,
			Namespace:    namespace,
			Plan:         "dev",
			CniType:      input.Cni,
			OtherConfigs: input.OtherConfigs,
		}
		if input.Plan != "" {
			options.Plan = input.Plan
		}

		if input.E2EConfig.InfrastructureName == "vsphere" {
			if input.Cni == "antrea" {
				if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_ANTREA"); ok {
					options.VsphereControlPlaneEndpoint = clusterIP
				}
			}
			if input.Cni == "calico" {
				if clusterIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_CALICO"); ok {
					options.VsphereControlPlaneEndpoint = clusterIP
				}
			}
		}

		clusterConfigFile, err = framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, &options)
		Expect(err).To(BeNil())

		defer os.Remove(clusterConfigFile)
		err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
			Edition:           "tkg",
			Namespace:         namespace,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Generating credentials for workload cluster %q", clusterName))
		kubeConfigFileName := clusterName + ".kubeconfig"
		tempFilePath := filepath.Join(os.TempDir(), kubeConfigFileName)
		err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			ExportFile:  tempFilePath,
		})
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Waiting for workload cluster %q nodes to be up and running", clusterName))
		framework.WaitForNodes(framework.NewClusterProxy(clusterName, tempFilePath, ""), 2)

		By(fmt.Sprintf("Verify addon packages on workload cluster %q matches clusterBootstrap info on management cluster", clusterName))
		err = checkUtkgAddons(context, client)
		Expect(err).To(BeNil())

		By(fmt.Sprintf("Deleting workload cluster %q", clusterName))
		err = tkgCtlClient.DeleteCluster(tkgctl.DeleteClustersOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
			SkipPrompt:  true,
		})
		Expect(err).To(BeNil())

		By("Test successful !")
	})
}

// getClusterBootstrap gets ClusterBootstrap resource with the provided object key
func getClusterBootstrap(ctx context.Context, k8sClient client.Client, namespace, clusterName string) *runtanzuv1alpha3.ClusterBootstrap {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	objKey := client.ObjectKey{Namespace: namespace, Name: clusterName}

	Eventually(func() bool {
		err := k8sClient.Get(ctx, objKey, clusterBootstrap)
		return err == nil
	}, waitTimeout, pollingInterval).Should(BeTrue())

	Expect(clusterBootstrap).ShouldNot(BeNil())
	return clusterBootstrap
}

// getPackageInstall get PackageInstall resource with the provided object key
func getPackageInstall(ctx context.Context, k8sClient client.Client, namespace, pkgiName string) *kapppkgiv1alpha1.PackageInstall {
	pkgInstall := &kapppkgiv1alpha1.PackageInstall{}
	objKey := client.ObjectKey{Namespace: namespace, Name: pkgiName}

	Eventually(func() bool {
		err := k8sClient.Get(ctx, objKey, pkgInstall)
		return err == nil
	}, waitTimeout, pollingInterval).Should(BeTrue())

	Expect(pkgInstall).ShouldNot(BeNil())

	return pkgInstall
}

func generatePkgiNameWithVersion(packageName, PackageVersion string) string {
	return fmt.Sprintf("%s.%s", packageName, PackageVersion)
}

func getPackageDetailsFromCBS(CBSRefName string) (string, string, string, error) {
	pkgShortName, err := strings.Split(CBSRefName, ".")[0]
	if err != nil || pkgShortName == nil {
		return "", "", "", errors.Wrapf(err, "error fetching package short name from clusterBootstrap resource")
	}
	pkgName, err := strings.Join(strings.Split(addonName, ".")[0:4], ".")
	if err != nil || pkgName == nil {
		return "", "", "", errors.Wrapf(err, "error fetching package name from clusterBootstrap resource")
	}
	pkgVersion, err := strings.Join(strings.Split(addonName, ".")[4:], ".")
	if err != nil || pkgVersion == nil {
		return "", "", "", errors.Wrapf(err, "error fetching package version from clusterBootstrap resource")
	}
	return pkgShortName, pkgName, pkgVersion, nil
}

func checkUtkgAddons(ctx context.Context, cl client.Client, testPkgName string, t *testing.T) {
	/*
	   1. Kubeconfig for management cluster
	   2. use this client to get wlc kubeconfig secret / get kubeconfig from it
	   k8sclient.get(namespace, name), secret
	   get secret
	   3. create another client with this kubeconfig
	   4. wlc client get pkgi information
	*/
	var (
		err          error
		pkgShortName string
		pkgName      string
		pkgVersion   string
	)
	assert := assert.New(t)
	//mngCluster := &clusterapiv1beta1.Cluster{}
	//clusterBootstrap := getClusterBootstrap(client.ObjectKeyFromObject(mngCluster))

	//switch the context or kubeconfig
	// Get ClusterBootstrap and return error if not found
	clusterBootstrap := getClusterBootstrap(ctx, cl, constants.TkgNamespace, clusterNameMng)

	//wlcCluster := &clusterapiv1beta1.Cluster{}

	// packageInstall name for for both management and workload clusters should follow the <cluster name>-<addon short name>
	// packageInstall name and version should match info in clusterBootstrap for all packages, format is <package name>.<package version>
	switch {
	case testPkgName == "CNI":
		pkgShortName, pkgName, pkgVersion = getPackageDetailsFromCBS(clusterBootstrap.Spec.CNI.RefName)
	case testPkgName == "CSI":
		pkgShortName, pkgName, pkgVersion = getPackageDetailsFromCBS(clusterBootstrap.Spec.CSI.RefName)
	case testPkgName == "CPI":
		pkgShortName, pkgName, pkgVersion = getPackageDetailsFromCBS(clusterBootstrap.Spec.CPI.RefName)
	case testPkgName == "Kapp":
		pkgShortName, pkgName, pkgVersion = getPackageDetailsFromCBS(clusterBootstrap.Spec.Kapp.RefName)
	case testPkgName == "metrics-server":
		pkgShortName, pkgName, pkgVersion = getPackageDetailsFromCBS(clusterBootstrap.Spec.AdditionalPackages[0].RefName)
	case testPkgName == "secretgen-controller":
		pkgShortName, pkgName, pkgVersion = getPackageDetailsFromCBS(clusterBootstrap.Spec.AdditionalPackages[1].RefName)
	case testPkgName == "pinniped":
		pkgShortName, pkgName, pkgVersion = getPackageDetailsFromCBS(clusterBootstrap.Spec.AdditionalPackages[2].RefName)
	}

	pkgiName := util.GeneratePackageInstallName(clusterName, pkgShortName)
	pkgi := getPackageInstall(ctx, cl, constants.TkgNamespace, pkgiName)
	// Verify package name matches between clusterBootstrap and packageInstall
	assert.Equal(pkgName, pkgi.Spec.PackageRef.RefName)
	// Verify package version matches between clusterBootstrap and packageInstall
	assert.Equal(pkgVersion, pkgi.Spec.PackageRef.VersionSelection.Constraints)
}
