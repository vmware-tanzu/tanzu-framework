// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,stylecheck,nolintlint
package shared

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	secretutil "sigs.k8s.io/cluster-api/util/secret"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	addonutil "github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

type E2ECommonSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
	Plan            string
	Namespace       string
	OtherConfigs    map[string]string
}

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
		err = checkUtkgAddons(context, client, scheme, "antrea")
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

func getPackageDetailsFromCBS(CBSRefName string) (string, string, string, error) {
	pkgShortName := strings.Split(CBSRefName, ".")[0]

	pkgName := strings.Join(strings.Split(CBSRefName, ".")[0:4], ".")

	pkgVersion := strings.Join(strings.Split(CBSRefName, ".")[4:], ".")

	return pkgShortName, pkgName, pkgVersion, nil
}

func checkUtkgAddons(ctx context.Context, cl client.Client, scheme *runtime.Scheme, testPkgName string) error {
	var (
		err          error
		pkgShortName string
		pkgName      string
		pkgVersion   string
	)
	mngCluster := &clusterapiv1beta1.Cluster{}

	// create remoteClient for workload cluster
	remoteClient, err := addonutil.GetClusterClient(ctx, cl, scheme, util.ObjectKey(mngCluster))
	if err != nil {
	    return err
	}

	key := client.ObjectKey{Namespace: constants.TkgNamespace, Name: secretutil.Name(mngCluster.Name, secretutil.Kubeconfig)}
	clusterKubeConfigSecret := &corev1.Secret{}
	err = remoteClient.Get(ctx, key, clusterKubeConfigSecret)
	if err != nil {
		return err
	}

	//clusterBootstrap := getClusterBootstrap(client.ObjectKeyFromObject(mngCluster))

	// Get ClusterBootstrap and return error if not found
	clusterBootstrap := getClusterBootstrap(ctx, cl, constants.TkgNamespace, mngCluster.Name)

	//wlcCluster := &clusterapiv1beta1.Cluster{}

	// packageInstall name for for both management and workload clusters should follow the <cluster name>-<addon short name>
	// packageInstall name and version should match info in clusterBootstrap for all packages, format is <package name>.<package version>
	switch {
	case testPkgName == "CNI":
		pkgShortName, pkgName, pkgVersion, err = getPackageDetailsFromCBS(clusterBootstrap.Spec.CNI.RefName)
	case testPkgName == "CSI":
		pkgShortName, pkgName, pkgVersion, err = getPackageDetailsFromCBS(clusterBootstrap.Spec.CSI.RefName)
	case testPkgName == "CPI":
		pkgShortName, pkgName, pkgVersion, err = getPackageDetailsFromCBS(clusterBootstrap.Spec.CPI.RefName)
	case testPkgName == "Kapp":
		pkgShortName, pkgName, pkgVersion, err = getPackageDetailsFromCBS(clusterBootstrap.Spec.Kapp.RefName)
	case testPkgName == "metrics-server":
		pkgShortName, pkgName, pkgVersion, err = getPackageDetailsFromCBS(clusterBootstrap.Spec.AdditionalPackages[0].RefName)
	case testPkgName == "secretgen-controller":
		pkgShortName, pkgName, pkgVersion, err = getPackageDetailsFromCBS(clusterBootstrap.Spec.AdditionalPackages[1].RefName)
	case testPkgName == "pinniped":
		pkgShortName, pkgName, pkgVersion, err = getPackageDetailsFromCBS(clusterBootstrap.Spec.AdditionalPackages[2].RefName)
	}

	pkgiName := addonutil.GeneratePackageInstallName(mngCluster.Name, pkgShortName)
	pkgi := getPackageInstall(ctx, remoteClient, constants.TkgNamespace, pkgiName)
	// check package install reconcile status is succeed
	Expect(pkgi.Status.GenericStatus.Conditions[1].Type).Should(Equal(kappctrl.ReconcileSucceeded))
	Expect(pkgi.Status.GenericStatus.Conditions[1].Status).Should(Equal(corev1.ConditionTrue))

	// Verify package name match between clusterBootstrap and packageInstall
	Expect(pkgName).Should(Equal(pkgi.Spec.PackageRef.RefName))

	// Verify package version match between clusterBootstrap and packageInstall
	Expect(pkgVersion).Should(Equal(pkgi.Spec.PackageRef.VersionSelection.Constraints))

	return nil
}
