// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	admissionregistrationv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	controlplanev1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	controllerruntimescheme "sigs.k8s.io/controller-runtime/pkg/scheme"

	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kappdatapkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"

	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
)

const (
	getResourceTimeout          = time.Minute * 1
	waitForReadyTimeout         = time.Minute * 10
	waitTimeout                 = time.Minute * 10
	resourceDeletionWaitTimeout = time.Minute * 20
	pollingInterval             = time.Second * 30

	AddonFinalizer                     = "tkg.tanzu.vmware.com/addon"
	PreTerminateAddonsAnnotationPrefix = clusterapiv1beta1.PreTerminateDeleteHookAnnotationPrefix + "/tkg.tanzu.vmware.com"
	PreTerminateAddonsAnnotationValue  = "tkg.tanzu.vmware.com/addons"
)

func getRestConfig(exportFile string, scheme *runtime.Scheme) (*restclient.Config, error) {
	config, err := clientcmd.LoadFromFile(exportFile)
	Expect(err).ToNot(HaveOccurred(), "Failed to load cluster Kubeconfig file from %q", exportFile)

	rawConfig, err := clientcmd.Write(*config)
	Expect(err).ToNot(HaveOccurred(), "Failed to create raw config ")

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(rawConfig)
	Expect(err).ToNot(HaveOccurred(), "Failed to create rest config ")

	return restConfig, nil
}

// create cluster client from kubeconfig
func createClientFromKubeconfig(exportFile string, scheme *runtime.Scheme) (client.Client, error) {
	restConfig, err := getRestConfig(exportFile, scheme)
	Expect(err).NotTo(HaveOccurred(), "Failed to get rest config")

	client, err := client.New(restConfig, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred(), "Failed to create a cluster client")

	return client, nil
}

// GetClients gets the various kubernetes clients
func GetClients(ctx context.Context, exportFile string) (k8sClient client.Client, dynamicClient dynamic.Interface, aggregatedAPIResourcesClient client.Client, discoveryClient discovery.DiscoveryInterface, admissionRegistrationClient admissionregistrationv1.AdmissionregistrationV1Interface, err error) {
	scheme := runtime.NewScheme()

	_ = clientgoscheme.AddToScheme(scheme)
	_ = kappctrl.AddToScheme(scheme)
	_ = kapppkg.AddToScheme(scheme)
	_ = kappdatapkg.AddToScheme(scheme)
	_ = runtanzuv1alpha1.AddToScheme(scheme)
	_ = clusterapiv1beta1.AddToScheme(scheme)
	_ = controlplanev1beta1.AddToScheme(scheme)
	_ = runtanzuv1alpha3.AddToScheme(scheme)
	// The below schemes are added directly to avoid referring to apis/cni, apis/cpi in tanzu-framework that introduces some build issues
	_ = (&controllerruntimescheme.Builder{GroupVersion: schema.GroupVersion{Group: "cni.tanzu.vmware.com", Version: "v1alpha1"}}).AddToScheme(scheme)
	_ = (&controllerruntimescheme.Builder{GroupVersion: schema.GroupVersion{Group: "cpi.tanzu.vmware.com", Version: "v1alpha1"}}).AddToScheme(scheme)
	_ = (&controllerruntimescheme.Builder{GroupVersion: schema.GroupVersion{Group: "csi.tanzu.vmware.com", Version: "v1alpha1"}}).AddToScheme(scheme)
	_ = capvv1beta1.AddToScheme(scheme)
	_ = capvvmwarev1beta1.AddToScheme(scheme)

	k8sClient, err = createClientFromKubeconfig(exportFile, scheme)
	Expect(err).ToNot(HaveOccurred(), "Failed to create management cluster client")

	restConfig, err := getRestConfig(exportFile, scheme)
	Expect(err).NotTo(HaveOccurred())

	dynamicClient, err = dynamic.NewForConfig(restConfig)
	Expect(err).NotTo(HaveOccurred())

	clientset := kubernetes.NewForConfigOrDie(restConfig)
	discoveryClient = clientset.DiscoveryClient
	admissionRegistrationClient = clientset.AdmissionregistrationV1()

	aggregatedAPIResourcesClient, err = client.New(restConfig, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())

	return
}

// getPackagesFromCB gets the list of packages from CB that are supposed to be installed on a cluster
func getPackagesFromCB(ctx context.Context, clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, mccl, wccl client.Client, mcClusterName, mcClusterNamespace, wcClusterName, wcClusterNamespace, infrastructureName string, isManagementCluster bool) ([]kapppkgv1alpha1.Package, error) {
	var packages []kapppkgv1alpha1.Package

	systemNamespace := constants.TkgNamespace
	if infrastructureName == "tkgs" {
		systemNamespace = constants.TKGSClusterClassNamespace
	}

	// verify cni package is installed on the workload cluster
	cniPkgShortName, cniPkgName, cniPkgVersion := getPackageDetailsFromCBS(clusterBootstrap.Spec.CNI.RefName)
	packages = append(packages, kapppkgv1alpha1.Package{ObjectMeta: metav1.ObjectMeta{Name: cniPkgShortName, Namespace: systemNamespace},
		Spec: kapppkgv1alpha1.PackageSpec{RefName: cniPkgName, Version: cniPkgVersion}})

	if !isManagementCluster {
		kappPkgShortName, kappPkgName, kappPkgVersion := getPackageDetailsFromCBS(clusterBootstrap.Spec.Kapp.RefName)
		packages = append(packages, kapppkgv1alpha1.Package{ObjectMeta: metav1.ObjectMeta{Name: kappPkgShortName, Namespace: wcClusterNamespace},
			Spec: kapppkgv1alpha1.PackageSpec{RefName: kappPkgName, Version: kappPkgVersion}})
	}

	if infrastructureName == "vsphere" || infrastructureName == "tkgs" {
		csiPkgShortName, csiPkgName, csiPkgVersion := getPackageDetailsFromCBS(clusterBootstrap.Spec.CSI.RefName)
		packages = append(packages, kapppkgv1alpha1.Package{ObjectMeta: metav1.ObjectMeta{Name: csiPkgShortName, Namespace: systemNamespace},
			Spec: kapppkgv1alpha1.PackageSpec{RefName: csiPkgName, Version: csiPkgVersion}})

		cpiPkgShortName, cpiPkgName, cpiPkgVersion := getPackageDetailsFromCBS(clusterBootstrap.Spec.CPI.RefName)
		packages = append(packages, kapppkgv1alpha1.Package{ObjectMeta: metav1.ObjectMeta{Name: cpiPkgShortName, Namespace: systemNamespace},
			Spec: kapppkgv1alpha1.PackageSpec{RefName: cpiPkgName, Version: cpiPkgVersion}})
	}

	if clusterBootstrap.Spec.AdditionalPackages != nil {
		// validate additional packages
		for _, additionalPkg := range clusterBootstrap.Spec.AdditionalPackages {
			pkgShortName, pkgName, pkgVersion := getPackageDetailsFromCBS(additionalPkg.RefName)

			// TODO: temporarily skip verifying tkg-storageclass due to install failure issue.
			//		 this should be removed once the issue is resolved.
			if pkgShortName == "tkg-storageclass" {
				continue
			}
			packages = append(packages, kapppkgv1alpha1.Package{ObjectMeta: metav1.ObjectMeta{Name: pkgShortName, Namespace: systemNamespace},
				Spec: kapppkgv1alpha1.PackageSpec{RefName: pkgName, Version: pkgVersion}})
		}
	}

	return packages, nil
}

// CheckClusterCB checks if clusterbootstrap resource is created correctly and packages are reconciled successfully on cluster
func CheckClusterCB(ctx context.Context, mccl, wccl client.Client, mcClusterName, mcClusterNamespace, wcClusterName, wcClusterNamespace, infrastructureName string, isManagementCluster, isCustomCB bool) error {
	log.Infof("Verify addons on workload cluster %s with management cluster %s", wcClusterName, mcClusterName)

	var clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap

	systemNamespace := constants.TkgNamespace
	if infrastructureName == "tkgs" {
		systemNamespace = constants.TKGSClusterClassNamespace
	}

	if isManagementCluster {
		clusterBootstrap = getClusterBootstrap(ctx, mccl, mcClusterNamespace, mcClusterName)
	} else {
		clusterBootstrap = getClusterBootstrap(ctx, mccl, wcClusterNamespace, wcClusterName)
	}

	By(fmt.Sprintf("Verify clusterbootstrap matches clusterbootstraptemplate"))
	verifyClusterBootstrap(ctx, mccl, clusterBootstrap, clusterBootstrap.Status.ResolvedTKR, systemNamespace, isManagementCluster, isCustomCB)

	packages, err := getPackagesFromCB(ctx, clusterBootstrap, mccl, wccl, mcClusterName, mcClusterNamespace, wcClusterName, wcClusterNamespace, infrastructureName, isManagementCluster)
	Expect(err).NotTo(HaveOccurred())

	var (
		client                   client.Client
		clusterName, clusterType string
	)

	if isManagementCluster {
		client = mccl
		clusterName = mcClusterName
		clusterType = "management"
	} else {
		client = wccl
		clusterName = wcClusterName
		clusterType = "workload"
	}

	for _, pkg := range packages {
		if strings.Contains(pkg.Name, "kapp-controller") {
			verifyPackageInstall(ctx, mccl, pkg.Namespace, GeneratePackageInstallName(clusterName, pkg.Name), pkg.Spec.RefName, pkg.Spec.Version)
		} else {
			// TODO: Verify the capabilities package install when it is fixed to reconcile
			if infrastructureName != "tkgs" || pkg.Name != "capabilities" {
				verifyPackageInstall(ctx, client, pkg.Namespace, GeneratePackageInstallName(clusterName, pkg.Name), pkg.Spec.RefName, pkg.Spec.Version)
			}
		}
	}

	By(fmt.Sprintf("Verify addon packages on %q cluster %q status is reflected correctly in clusterBootstrap status", clusterType, clusterName))
	verifyPackageInstallStatusinClusterBootstrapStatus(ctx, mccl, mcClusterName, mcClusterNamespace, wcClusterName, wcClusterNamespace, infrastructureName, isManagementCluster, packages)

	// For Management cluster we dont expect the finalizers and hooks to be present. So check only for workload cluster
	if !isManagementCluster {
		By(fmt.Sprintf("Verify addon finalizers and machine termination hooks for %q cluster %q is set correctly", clusterType, clusterName))
		verifyFinalizersAndMachineHooks(ctx, mccl, wcClusterName, wcClusterNamespace)
	}

	return nil
}

// verifyClusterBootstrap checks if cluster bootstrap is created as expected i.e. it is cloned correctly from ClusterBootstrapTemplate
func verifyClusterBootstrap(ctx context.Context, c client.Client, clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, tkrName string, systemNamespace string, isManagementCluster, isCustomCB bool) {
	resolvedTKr := clusterBootstrap.Status.ResolvedTKR
	Expect(resolvedTKr).NotTo(BeEmpty())

	clusterBootstrapTemplate := getClusterBootstrapTemplate(ctx, c, resolvedTKr, systemNamespace)

	expectedClusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	expectedClusterBootstrap.Spec = clusterBootstrapTemplate.Spec.DeepCopy()

	var packages []*runtanzuv1alpha3.ClusterBootstrapPackage
	packages = append(packages, expectedClusterBootstrap.Spec.Kapp, expectedClusterBootstrap.Spec.CNI)

	if expectedClusterBootstrap.Spec.CPI != nil {
		packages = append(packages, expectedClusterBootstrap.Spec.CPI)
	}

	if expectedClusterBootstrap.Spec.CSI != nil {
		packages = append(packages, expectedClusterBootstrap.Spec.CSI)
	}

	packages = append(packages, expectedClusterBootstrap.Spec.AdditionalPackages...)

	for _, pkg := range packages {
		pkgShortName, _, _ := getPackageDetailsFromCBS(pkg.RefName)
		if pkg.ValuesFrom != nil {
			if pkg.ValuesFrom.SecretRef != "" {
				pkg.ValuesFrom.SecretRef = GeneratePackageSecretName(clusterBootstrap.Name, pkgShortName)
			} else if pkg.ValuesFrom.ProviderRef != nil {
				pkg.ValuesFrom.ProviderRef.Name = GeneratePackageSecretName(clusterBootstrap.Name, pkgShortName)
			}
		}

		// guest-cluster-auth-service needs to be handled explicitly as they are different for CBT and custom CB
		if strings.Contains(pkg.RefName, "guest-cluster-auth-service") {
			if pkg.ValuesFrom == nil {
				pkg.ValuesFrom = &runtanzuv1alpha3.ValuesFrom{
					SecretRef: GenerateDataValueSecretName(clusterBootstrap.Name, pkgShortName),
				}
			}
		}

		// custom CB needs to have package names the same as the cluster name, so this will not match the expected providerRef name
		// any package specified in the custom CB manifest will have to be special-cased
		// so, handling this currently for antrea for the custom CB test case
		if !isManagementCluster && isCustomCB && (strings.Contains(pkg.RefName, "antrea") || strings.Contains(pkg.RefName, "aws-ebs-csi-driver")) {
			pkg.ValuesFrom.ProviderRef.Name = clusterBootstrap.Name
		}

		// storageclass needs special handling
		if strings.Contains(pkg.RefName, "tkg-storageclass") {
			for _, bootstrapPkg := range clusterBootstrap.Spec.AdditionalPackages {
				if strings.Contains(bootstrapPkg.RefName, "tkg-storageclass") {
					pkg.ValuesFrom = bootstrapPkg.ValuesFrom
				}
			}
		}
	}

	// Calico needs special handling
	if strings.Contains(clusterBootstrap.Spec.CNI.RefName, "calico") {
		tkr := getTKr(ctx, c, tkrName)
		expectedClusterBootstrap.Spec.CNI.RefName = ""
		for _, bootstrapPkg := range tkr.Spec.BootstrapPackages {
			if strings.Contains(bootstrapPkg.Name, "calico") {
				expectedClusterBootstrap.Spec.CNI.RefName = bootstrapPkg.Name
				break
			}
		}
		Expect(expectedClusterBootstrap.Spec.CNI.RefName).NotTo(BeEmpty())
		apiGroup := "cni.tanzu.vmware.com"
		expectedClusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef = &corev1.TypedLocalObjectReference{
			APIGroup: &apiGroup,
			Kind:     "CalicoConfig",
			Name:     clusterBootstrap.Name,
		}
	}

	clusterBootstrapSpecJson, _ := json.MarshalIndent(clusterBootstrap.Spec, "", "\t")
	expectedClusterBootstrapSpecJson, _ := json.MarshalIndent(expectedClusterBootstrap.Spec, "", "\t")
	log.Infof("Cluster bootstrap: %v", string(clusterBootstrapSpecJson))
	log.Infof("Expected cluster bootstrap: %v", string(expectedClusterBootstrapSpecJson))

	Expect(clusterBootstrap.Spec).To(BeEquivalentTo(expectedClusterBootstrap.Spec), "Clusterbootstrap should match clusterbootstraptemplate")
}

// verifyPackageInstall verifies if package is reconciled successfully on the cluster
func verifyPackageInstall(ctx context.Context, c client.Client, namespace, pkgInstallName, pkgName, pkgVersion string) {
	log.Infof("Check PackageInstall %s for package %s of version %s", pkgInstallName, pkgName, pkgVersion)

	// verify the package is successfully deployed and its name and version match with the clusterBootstrap CR
	pkgInstall := &kapppkgiv1alpha1.PackageInstall{}
	objKey := client.ObjectKey{Namespace: namespace, Name: pkgInstallName}
	Eventually(func() bool {
		if err := c.Get(ctx, objKey, pkgInstall); err != nil {
			return false
		}
		if len(pkgInstall.Status.GenericStatus.Conditions) == 0 {
			return false
		}
		if pkgInstall.Status.GenericStatus.Conditions[0].Type != kappctrl.ReconcileSucceeded {
			return false
		}
		if pkgInstall.Status.GenericStatus.Conditions[0].Status != corev1.ConditionTrue {
			return false
		}
		if pkgInstall.Spec.PackageRef.RefName != pkgName {
			return false
		}
		if pkgInstall.Spec.PackageRef.VersionSelection.Constraints != pkgVersion {
			return false
		}
		return true
	}, waitForReadyTimeout, pollingInterval).Should(BeTrue(), "Package %s is not deployed successfully", pkgName)
}

// verifyPackageInstallStatusinClusterBootstrapStatus verifies if package install status is synced correctly to ClusterBootstrap status
func verifyPackageInstallStatusinClusterBootstrapStatus(ctx context.Context, mccl client.Client, mcClusterName, mcClusterNamespace, wcClusterName, wcClusterNamespace string, infrastructureName string, isManagementCluster bool, packages []kapppkgv1alpha1.Package) {
	Eventually(func() bool {
		var (
			clusterBootstrap    *runtanzuv1alpha3.ClusterBootstrap
			pkgConditionSuccess int
		)
		if isManagementCluster {
			clusterBootstrap = getClusterBootstrap(ctx, mccl, mcClusterNamespace, mcClusterName)
		} else {
			clusterBootstrap = getClusterBootstrap(ctx, mccl, wcClusterNamespace, wcClusterName)
		}

		for _, pkg := range packages {
			log.Infof("Check Package %q status in clusterbootstrap status", pkg.Name)
			// TODO: temporarily skip verifying tkg-storageclass due to install failure issue.
			//		 this should be removed once the issue is resolved.
			if pkg.Name == "tkg-storageclass" {
				pkgConditionSuccess = pkgConditionSuccess + 1
				continue
			}
			// TODO: Temporarily skip verifying capabilities package install for tkgs cluster
			//       this needs to removed once the issue is resolved
			if infrastructureName == "tkgs" && pkg.Name == "capabilities" {
				pkgConditionSuccess = pkgConditionSuccess + 1
				continue
			}

			var conditionFound bool
			for _, condition := range clusterBootstrap.GetConditions() {
				if strings.Contains(string(condition.Type), cases.Title(language.Und).String(pkg.Name)) {
					if string(condition.Type) == fmt.Sprintf("%s-%s", cases.Title(language.Und).String(pkg.Name), kappctrl.ReconcileSucceeded) &&
						condition.Status == corev1.ConditionTrue {
						pkgConditionSuccess = pkgConditionSuccess + 1
					} else {
						log.Infof("Package %q condition failed. condition type: %q, status: %q, message: %q, reason: %q", pkg.Name, condition.Type, condition.Status, condition.Message, condition.Reason)
					}
					conditionFound = true
					break
				}
			}
			if !conditionFound {
				log.Infof("Package %q is not found in clusterbootstrap condition", pkg.Name)
			}
		}

		if len(packages) == pkgConditionSuccess {
			return true
		}

		return false
	}, waitForReadyTimeout, pollingInterval).Should(BeTrue(), "Packages install status failed in clusterbootstrap status")
}

// verifyFinalizersAndMachineHooks verifies if package install status is synced correctly to ClusterBootstrap status
func verifyFinalizersAndMachineHooks(ctx context.Context, mccl client.Client, wcClusterName, wcClusterNamespace string) {
	clusterName := wcClusterName
	clusterNamespace := wcClusterNamespace

	cb := getClusterBootstrap(ctx, mccl, clusterNamespace, clusterName)
	Expect(cb.Finalizers).To(ContainElement(AddonFinalizer))

	cluster := getCluster(ctx, mccl, clusterNamespace, clusterName)
	Expect(cluster.Finalizers).To(ContainElement(AddonFinalizer))

	clusterKubeconfig := getClusterKubeconfig(ctx, mccl, clusterNamespace, clusterName)
	Expect(clusterKubeconfig.Finalizers).To(ContainElement(AddonFinalizer))

	machines := getClusterMachines(ctx, mccl, clusterNamespace, clusterName)
	for _, machine := range machines.Items {
		Expect(machine.Annotations).Should(HaveKeyWithValue(PreTerminateAddonsAnnotationPrefix, PreTerminateAddonsAnnotationValue))
	}
}

func getPackageDetailsFromCBS(CBSRefName string) (pkgShortName, pkgName, pkgVersion string) {
	pkgShortName = strings.Split(CBSRefName, ".")[0]
	pkgName = strings.Join(strings.Split(CBSRefName, ".")[0:4], ".")
	pkgVersion = strings.Join(strings.Split(CBSRefName, ".")[4:], ".")
	return
}

func getClusterBootstrapTemplate(ctx context.Context, k8sClient client.Client, tkrName string, systemNamespace string) *runtanzuv1alpha3.ClusterBootstrapTemplate {
	var objKey client.ObjectKey

	clusterBootstrapTemplate := &runtanzuv1alpha3.ClusterBootstrapTemplate{}
	objKey = client.ObjectKey{Name: tkrName, Namespace: systemNamespace}

	Eventually(func() error {
		return k8sClient.Get(ctx, objKey, clusterBootstrapTemplate)
	}, getResourceTimeout, pollingInterval).Should(Succeed())

	Expect(clusterBootstrapTemplate).NotTo(BeNil())
	return clusterBootstrapTemplate
}

func getClusterBootstrap(ctx context.Context, k8sClient client.Client, namespace, clusterName string) *runtanzuv1alpha3.ClusterBootstrap {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	objKey := client.ObjectKey{Namespace: namespace, Name: clusterName}

	Eventually(func() error {
		return k8sClient.Get(ctx, objKey, clusterBootstrap)
	}, getResourceTimeout, pollingInterval).Should(Succeed())

	Expect(clusterBootstrap).ShouldNot(BeNil())
	return clusterBootstrap
}

func getCluster(ctx context.Context, k8sClient client.Client, namespace, clusterName string) *clusterapiv1beta1.Cluster {
	cluster := &clusterapiv1beta1.Cluster{}
	objKey := client.ObjectKey{Namespace: namespace, Name: clusterName}

	Eventually(func() error {
		return k8sClient.Get(ctx, objKey, cluster)
	}, getResourceTimeout, pollingInterval).Should(Succeed())

	Expect(cluster).ShouldNot(BeNil())
	return cluster
}

func getClusterKubeconfig(ctx context.Context, k8sClient client.Client, namespace, clusterName string) *corev1.Secret {
	secret := &corev1.Secret{}
	clusterKubeconfigName := fmt.Sprintf("%s-kubeconfig", clusterName)
	objKey := client.ObjectKey{Namespace: namespace, Name: clusterKubeconfigName}

	Eventually(func() error {
		return k8sClient.Get(ctx, objKey, secret)
	}, getResourceTimeout, pollingInterval).Should(Succeed())

	Expect(secret).ShouldNot(BeNil())
	return secret
}

func getClusterMachines(ctx context.Context, k8sClient client.Client, namespace, clusterName string) clusterapiv1beta1.MachineList {
	var machines clusterapiv1beta1.MachineList
	listOptions := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{clusterapiv1beta1.ClusterLabelName: clusterName}),
	}

	Eventually(func() error {
		return k8sClient.List(ctx, &machines, listOptions...)
	}, getResourceTimeout, pollingInterval).Should(Succeed())

	Expect(len(machines.Items)).To(BeNumerically(">", 0))
	return machines
}

func getTKr(ctx context.Context, k8sClient client.Client, tkrName string) *runtanzuv1alpha3.TanzuKubernetesRelease {
	tkr := &runtanzuv1alpha3.TanzuKubernetesRelease{}
	objKey := client.ObjectKey{Namespace: constants.TkgNamespace, Name: tkrName}

	Eventually(func() error {
		return k8sClient.Get(ctx, objKey, tkr)
	}, getResourceTimeout, pollingInterval).Should(Succeed())

	Expect(tkr).NotTo(BeNil())
	return tkr
}

func objectExists(ctx context.Context, k8sClient client.Client, namespace, name string, obj client.Object) bool {
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, obj)
	if apierrors.IsNotFound(err) {
		return false
	}

	if err != nil {
		log.Infof("Error getting object name:%q namespace: %q error: %q", name, namespace, err)
	}
	return true
}

type ClusterResource struct {
	name      string
	namespace string
	obj       client.Object
}

// clusterResourcesDeleted checks if all the cluster resources are deleted or not
func clusterResourcesDeleted(ctx context.Context, k8sClient client.Client, clusterResources []ClusterResource) bool {
	for _, r := range clusterResources {
		log.Infof("Check if cluster resource of type %q kind %q name %q is deleted from namespace %q", reflect.TypeOf(r.obj), r.obj.GetObjectKind().GroupVersionKind().Kind, r.name, r.namespace)
		if objectExists(ctx, k8sClient, r.namespace, r.name, r.obj) {
			return false
		}
	}
	return true
}

/* GetManagementClusterResources gets all the resources thats created by addons-manager plus
 * all the resources on which finalizer is added by addons-manager during a cluster creation.
 */
func GetManagementClusterResources(ctx context.Context, mccl client.Client, dynamicClient dynamic.Interface, aggregatedAPIResourcesClient client.Client, discoveryClient discovery.DiscoveryInterface, clusterNamespace, clusterName, infrastructureName string) ([]ClusterResource, error) {
	// get ClusterBootstrap and return error if not found
	clusterResources := []ClusterResource{
		{namespace: clusterNamespace, name: clusterName, obj: &clusterapiv1beta1.Cluster{}},
		{namespace: clusterNamespace, name: clusterName, obj: &runtanzuv1alpha3.ClusterBootstrap{}},
		{namespace: clusterNamespace, name: clusterName + "-kubeconfig", obj: &corev1.Secret{}},
		{namespace: clusterNamespace, name: clusterName + "-kapp-controller", obj: &kapppkgiv1alpha1.PackageInstall{}},
		{namespace: clusterNamespace, name: clusterName + "-kapp-controller-data-values", obj: &corev1.Secret{}},
	}

	clusterBootstrap := getClusterBootstrap(ctx, mccl, clusterNamespace, clusterName)

	var packages []*runtanzuv1alpha3.ClusterBootstrapPackage
	packages = append(packages, clusterBootstrap.Spec.Kapp, clusterBootstrap.Spec.CNI)

	if clusterBootstrap.Spec.CPI != nil {
		packages = append(packages, clusterBootstrap.Spec.CPI)
	}

	if clusterBootstrap.Spec.CSI != nil {
		packages = append(packages, clusterBootstrap.Spec.CSI)
	}

	packages = append(packages, clusterBootstrap.Spec.AdditionalPackages...)

	for _, pkg := range packages {
		if pkg.ValuesFrom != nil {
			if pkg.ValuesFrom.Inline != nil {
				packageRefName, _, err := GetPackageMetadata(ctx, aggregatedAPIResourcesClient, pkg.RefName, clusterNamespace)
				if err != nil {
					return nil, err
				}
				packageSecretName := GeneratePackageSecretName(clusterName, packageRefName)
				clusterResources = append(clusterResources, ClusterResource{name: packageSecretName, namespace: clusterNamespace, obj: &corev1.Secret{}})
			}
			if pkg.ValuesFrom.SecretRef != "" {
				clusterResources = append(clusterResources, ClusterResource{name: pkg.ValuesFrom.SecretRef, namespace: clusterNamespace, obj: &corev1.Secret{}})
			}
			if pkg.ValuesFrom.ProviderRef != nil {
				gvr, err := gvrForGroupKind(discoveryClient, schema.GroupKind{Group: *pkg.ValuesFrom.ProviderRef.APIGroup, Kind: pkg.ValuesFrom.ProviderRef.Kind})
				if err != nil {
					return nil, err
				}
				provider, err := dynamicClient.Resource(*gvr).Namespace(clusterNamespace).Get(ctx, pkg.ValuesFrom.ProviderRef.Name, metav1.GetOptions{}, "status")
				if err != nil {
					return nil, err
				}
				secretName, found, err := unstructured.NestedString(provider.UnstructuredContent(), "status", "secretRef")
				if err != nil {
					return nil, err
				}
				if !found || secretName == "" {
					// We expect the secretRef to be present under status subresource and its value gets updated by
					// the corresponding controller after cluster is created. If it is not exist, throw an error.
					return nil, fmt.Errorf("the secret is not found for the provider %s/%s", provider.GroupVersionKind().Kind, provider.GetName())
				}
				clusterResources = append(clusterResources, ClusterResource{name: provider.GetName(), namespace: clusterNamespace, obj: provider})
				clusterResources = append(clusterResources, ClusterResource{name: secretName, namespace: clusterNamespace, obj: &corev1.Secret{}})
			}
		} else {
			// In tkgs case a secret could be created to add nodeSelector, deployment/daemonset updateStrategies. Hence need to add that secret
			if infrastructureName == "tkgs" {
				packageRefName, _, err := GetPackageMetadata(ctx, aggregatedAPIResourcesClient, pkg.RefName, clusterNamespace)
				if err != nil {
					return nil, err
				}
				packageSecretName := GeneratePackageSecretName(clusterName, packageRefName)
				clusterResources = append(clusterResources, ClusterResource{name: packageSecretName, namespace: clusterNamespace, obj: &corev1.Secret{}})
			}
		}
	}

	for _, r := range clusterResources {
		log.Infof("Cluster resource of type %q kind %q name %q exists in namespace %q", reflect.TypeOf(r.obj), r.obj.GetObjectKind().GroupVersionKind().Kind, r.name, r.namespace)
	}

	return clusterResources, nil
}

func gvrForGroupKind(discoveryClient discovery.DiscoveryInterface, gk schema.GroupKind) (*schema.GroupVersionResource, error) {
	apiResourceList, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	for _, apiResource := range apiResourceList {
		gv, err := schema.ParseGroupVersion(apiResource.GroupVersion)
		if err != nil {
			return nil, err
		}
		if gv.Group == gk.Group {
			for i := 0; i < len(apiResource.APIResources); i++ {
				if apiResource.APIResources[i].Kind == gk.Kind {
					return &schema.GroupVersionResource{Group: gv.Group, Resource: apiResource.APIResources[i].Name, Version: gv.Version}, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("unable to find server preferred resource %s/%s", gk.Group, gk.Kind)
}

/*
 * The below functions are duplicated from addons/pkg/util in order to avoid circular dependency between tanzu-framework and addons modules.
 */
func GetPackageMetadata(ctx context.Context, c client.Client, carvelPkgName, carvelPkgNamespace string) (string, string, error) {
	pkg := &kapppkgv1alpha1.Package{}
	if err := c.Get(ctx, client.ObjectKey{Name: carvelPkgName, Namespace: carvelPkgNamespace}, pkg); err != nil {
		return "", "", err
	}
	return pkg.Spec.RefName, pkg.Spec.Version, nil
}

func packageShortName(pkgRefName string) string {
	nameTokens := strings.Split(pkgRefName, ".")
	if len(nameTokens) >= 1 {
		return nameTokens[0]
	}
	return pkgRefName
}

// GeneratePackageSecretName generates secret name for a package from the cluster and the package name
func GeneratePackageSecretName(clusterName, carvelPkgRefName string) string {
	return fmt.Sprintf("%s-%s-package", clusterName, packageShortName(carvelPkgRefName))
}

// GenerateDataValueSecretName generates data value secret name from the cluster and the package name
func GenerateDataValueSecretName(clusterName, carvelPkgRefName string) string {
	return fmt.Sprintf("%s-%s-data-values", clusterName, packageShortName(carvelPkgRefName))
}

// GeneratePackageInstallName is the util function to generate the PackageInstall CR name in a consistent manner.
// clusterName is the name of cluster within which all resources associated with this PackageInstall CR is installed.
// It does not necessarily
// mean the PackageInstall CR will be installed in that cluster. I.e., the kapp-controller PackageInstall CR is installed
// in the management cluster but is named after "<workload-cluster-name>-kapp-controller". It indicates that this kapp-controller
// PackageInstall is for reconciling resources in a cluster named "<workload-cluster-name>".
// addonName is the short name of a Tanzu addon with which the PackageInstall CR is associated.
func GeneratePackageInstallName(clusterName, addonName string) string {
	return fmt.Sprintf("%s-%s", clusterName, strings.Split(addonName, ".")[0])
}

func CheckTKGSAddons(ctx context.Context, tkgctlClient tkgctl.TKGClient, svClusterName, clusterName, namespace, KubeconfigPath, InfrastructureName string, isCustomeCB bool) error {
	By(fmt.Sprintf("Get k8s client for supervisor cluster %q", svClusterName))
	mngclient, _, _, _, _, err := GetClients(ctx, KubeconfigPath)
	Expect(err).NotTo(HaveOccurred())

	By(fmt.Sprintf("Generating credentials for workload cluster %q", clusterName))
	kubeConfigFileName := clusterName + ".kubeconfig"
	tempFilePath := filepath.Join(os.TempDir(), kubeConfigFileName)
	err = tkgctlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
		ClusterName: clusterName,
		Namespace:   namespace,
		ExportFile:  tempFilePath,
	})
	Expect(err).To(BeNil())

	By(fmt.Sprintf("Get k8s client for workload cluster %q", clusterName))
	wlcClient, _, _, _, _, err := GetClients(ctx, tempFilePath)
	Expect(err).NotTo(HaveOccurred())

	By(fmt.Sprintf("Verify addon packages on workload cluster %q matches clusterBootstrap info on supervisor cluster %q", clusterName, svClusterName))
	err = CheckClusterCB(ctx, mngclient, wlcClient, svClusterName, namespace, clusterName, namespace, InfrastructureName, false, isCustomeCB)
	Expect(err).To(BeNil())

	return nil
}

func ValidateClusterClassConfigFile(clusterclassConfigFilePath string) (string, string) {
	By(fmt.Sprintf("validating input cluster class based config file:%v", clusterclassConfigFilePath))
	cclusterFile, err := os.ReadFile(clusterclassConfigFilePath)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to read the input cluster class based config file from: %v", clusterclassConfigFilePath))
	Expect(cclusterFile).ToNot(BeEmpty(), fmt.Sprintf("the input cluster class based config file should not be empty, file path: %v", clusterclassConfigFilePath))
	isCC, ccObject, err := tkgctl.CheckIfInputFileIsClusterClassBased(clusterclassConfigFilePath)
	Expect(err).ShouldNot(HaveOccurred(), fmt.Sprintf("failed to process cluster class input config file, reason: %v", err))
	Expect(isCC).To(Equal(true), fmt.Sprintf("input cluster class based config file is not cluster class based, does not have Cluster object."))
	return ccObject.GetName(), ccObject.GetNamespace()
}
