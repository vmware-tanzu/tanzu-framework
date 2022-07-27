// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,stylecheck,nolintlint
package shared

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/registry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kappcontrollerv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/registry"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/tkr-source/compatibility"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/sets"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

type TKRCompatibilityValidationSpecInput struct {
	E2EConfig    *framework.E2EConfig
	OtherConfigs map[string]string
}

const (
	tkrSourceControllerValuesSecretName      = "tkr-source-controller-values"
	tkrSourceControllerValuesSecretNamespace = "tkg-system"
)

func TKRCompatibilityValidationSpec(ctx context.Context, inputGetter func() TKRCompatibilityValidationSpecInput) { //nolint:funlen
	var (
		err           error
		input         TKRCompatibilityValidationSpecInput
		mcProxy       *framework.ClusterProxy
		mcContextName string
		tkrs          []*runv1.TanzuKubernetesRelease
		tkrRegistry   registry.Registry
	)
	const (
		waitTimeout     = time.Minute * 15
		pollingInterval = time.Second * 10
	)
	BeforeEach(func() { //nolint:dupl
		input = inputGetter()
		mcClusterName := input.E2EConfig.ManagementClusterName
		mcContextName = mcClusterName + "-admin@" + mcClusterName
		mcProxy = framework.NewClusterProxy(mcClusterName, "", mcContextName)
		registryOps := &ctlimg.Opts{
			Anon: true,
		}
		tkrRegistry, err = registry.New(registryOps)
		Expect(err).To(BeNil(), "failed to initialize the TKR image registry")
	})

	It("Should validate the compatible TKRs and it's related resources(OSImages, ClusterBootStrapTemplate) are downloaded and available on management cluster  ", func() {
		tkrCompatibility := &compatibility.Compatibility{
			Client: mcProxy.GetClient(),
			Config: compatibility.Config{
				TKRNamespace: "tkg-system",
			},
			Log: logr.Discard(),
		}

		compatibleSet, err := tkrCompatibility.CompatibleVersions(context.Background())
		fmt.Printf("CompatibleSet is :%+v \n", compatibleSet)
		Expect(err).ToNot(HaveOccurred())
		tkrRepoImagePath, resNamespace, err := GetTKRRepoImagePathAndNamespaceFromSecret(ctx, mcProxy.GetClient())
		Expect(err).To(BeNil(), "failed to TKR repository Image path from secret")
		Expect(tkrRepoImagePath).ToNot(BeEmpty(), "TKR repo Image path cannot be empty")
		tkrImageTags, err := tkrRegistry.ListImageTags(tkrRepoImagePath)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to list TKR image tags for the repository url %s", tkrRepoImagePath))
		tkrNames := TKRNamesFromTags(tkrImageTags)
		expectedTKRs := filterTKRNames(tkrNames, func(tkr string) bool {
			if compatibleSet.Has(version.FromLabel(tkr)) {
				return true
			}
			return false
		})

		By("Validating the TKRs are available on management cluster")
		Eventually(func() bool {
			tkrs = mcProxy.GetTKRs(ctx)
			err = VerifyTKRsAreAvailable(expectedTKRs, tkrs)
			return err == nil
		}, waitTimeout, pollingInterval).Should(BeTrue(), fmt.Sprintf("TKRs availability validation failed: %v", err))

		ValidateTKRsRelatedObjectsAvailability(ctx, mcProxy, expectedTKRs, resNamespace)

		By("Test successful !")
	})

	It("Should validate the compatible status is correctly calculated for all TKRs", func() {
		tkrCompatibility := &compatibility.Compatibility{
			Client: mcProxy.GetClient(),
			Config: compatibility.Config{
				TKRNamespace: "tkg-system",
			},
			Log: logr.Discard(),
		}
		compatibleSet, err := tkrCompatibility.CompatibleVersions(context.Background())
		fmt.Printf("CompatibleSet is :%+v \n", compatibleSet)
		Expect(err).ToNot(HaveOccurred())
		By("Waiting for the compatible TKRs to be available on management cluster")
		compatibleTKRNamesSet := compatibleSet.Map(func(s string) string {
			return strings.ReplaceAll(s, "+", "---")
		})
		Eventually(func() bool {
			tkrs = mcProxy.GetTKRs(ctx)
			err = VerifyTKRsAreAvailable(compatibleTKRNamesSet.Slice(), tkrs)
			return err == nil
		}, waitTimeout, pollingInterval).Should(BeTrue(), fmt.Sprintf("failed validating the Compatible TKRs availability on management cluster: %v", err))

		By("Validating all TKRs compatibility status condition is updated correctly")
		for i := range tkrs {
			fmt.Printf("Validating the compatibility status condition for TKR '%s'\n", tkrs[i].Name)
			if compatibleSet.Has(tkrs[i].Spec.Version) {
				Expect(conditions.IsTrue(tkrs[i], runv1.ConditionCompatible)).To(BeTrue(),
					fmt.Sprintf("TKR '%s' is expected to have Compatible condition to be true", tkrs[i].Name))
			} else {
				Expect(conditions.IsFalse(tkrs[i], runv1.ConditionCompatible)).To(BeTrue(),
					fmt.Sprintf("TKR '%s' is expected to have Compatible condition to be false", tkrs[i].Name))
				Expect(*conditions.GetSeverity(tkrs[i], runv1.ConditionCompatible)).To(Equal(clusterv1.ConditionSeverityWarning),
					fmt.Sprintf("TKR '%s' Compatible condition's severity is expected to be 'Warning' if condition status is False", tkrs[i].Name))
			}
		}
		By("Test successful !")
	})
}
func ValidateTKRsRelatedObjectsAvailability(ctx context.Context, mcProxy *framework.ClusterProxy, expectedTKRs []string, resourceNS string) {
	crClient := mcProxy.GetClient()
	for i := range expectedTKRs {
		fmt.Printf("Validating the availability of objectes related to TKR: %s \n", expectedTKRs[i])
		tkr := &runv1.TanzuKubernetesRelease{}
		err := crClient.Get(ctx, client.ObjectKey{Name: expectedTKRs[i]}, tkr)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to get TKR :%s", expectedTKRs[i]))
		cbt := getClusterBootstrapTemplate(ctx, crClient, tkr.Name)
		Expect(cbt).ToNot(BeNil(), fmt.Sprintf("failed to get CBT :%s", expectedTKRs[i]))
		ValidateOSImagesOfTKR(ctx, mcProxy, tkr)
		ValidatePackagesOfTKR(ctx, mcProxy, tkr, resourceNS)
	}
}

func ValidatePackagesOfTKR(ctx context.Context, mcProxy *framework.ClusterProxy, tkr *runv1.TanzuKubernetesRelease, namespace string) {
	var err error
	expectedPackages := tkr.Spec.BootstrapPackages
	Eventually(func() bool {
		packages := mcProxy.GetPackages(ctx, namespace)
		err = VerifyPackagesAreAvailable(expectedPackages, packages)
		return err == nil
	}, waitTimeout, pollingInterval).Should(BeTrue(), fmt.Sprintf("OSImages availability validation failed: %v", err))
}

func ValidateOSImagesOfTKR(ctx context.Context, mcProxy *framework.ClusterProxy, tkr *runv1.TanzuKubernetesRelease) {
	var err error
	expectedOSImages := tkr.Spec.OSImages
	Eventually(func() bool {
		osImages := mcProxy.GetOSImages(ctx)
		err = VerifyOSImagesAreAvailable(expectedOSImages, osImages)
		return err == nil
	}, waitTimeout, pollingInterval).Should(BeTrue(), fmt.Sprintf("OSImages availability validation failed: %v", err))
}

func GetTKRRepoImagePathAndNamespaceFromSecret(ctx context.Context, crClient client.Client) (string, string, error) {
	s := &v1.Secret{}
	if err := crClient.Get(ctx, client.ObjectKey{
		Namespace: tkrSourceControllerValuesSecretNamespace,
		Name:      tkrSourceControllerValuesSecretName,
	}, s); err != nil {
		return "", "", err
	}

	valuesMap := make(map[string]interface{})
	if err := yaml.Unmarshal(s.Data["values.yaml"], valuesMap); err != nil {
		return "", "", err
	}

	tkrRepoImagePath := valuesMap["tkrRepoImagePath"].(string)
	resourcesNamespace := valuesMap["namespace"].(string)
	return tkrRepoImagePath, resourcesNamespace, nil
}

func filterTKRNames(tkrNames []string, f func(string) bool) []string {
	var res []string
	for i := range tkrNames {
		if f(tkrNames[i]) {
			res = append(res, tkrNames[i])
		}
	}
	return res
}

func VerifyTKRsAreAvailable(expectedTKRs []string, tkrs []*runv1.TanzuKubernetesRelease) error {
	var missingTKRs []string
	tkrNames := getTKRNames(tkrs)
	for i := range expectedTKRs {
		if !tkrNames.Has(expectedTKRs[i]) {
			missingTKRs = append(missingTKRs, expectedTKRs[i])
		}
	}
	if len(missingTKRs) != 0 {
		return errors.Errorf("TKRs %v are not available on management cluster", missingTKRs)
	}
	return nil
}

func VerifyOSImagesAreAvailable(expectedOSImages []v1.LocalObjectReference, osImages []*runv1.OSImage) error {
	var missingOSImages []string
	osImageNames := getOSImagesNames(osImages)
	for i := range expectedOSImages {
		if !osImageNames.Has(expectedOSImages[i].Name) {
			missingOSImages = append(missingOSImages, expectedOSImages[i].Name)
		}
	}
	if len(missingOSImages) != 0 {
		return errors.Errorf("OSImages %v are not available on management cluster", missingOSImages)
	}
	return nil
}

func VerifyPackagesAreAvailable(expectedPackages []v1.LocalObjectReference, packages []*kappcontrollerv1alpha1.Package) error {
	var missingPackages []string
	osImageNames := getPackagesNames(packages)
	for i := range expectedPackages {
		if !osImageNames.Has(expectedPackages[i].Name) {
			missingPackages = append(missingPackages, expectedPackages[i].Name)
		}
	}
	if len(missingPackages) != 0 {
		return errors.Errorf("OSImages %v are not available on management cluster", missingPackages)
	}
	return nil
}

func TKRNamesFromTags(tkrImageTags []string) []string {
	var res []string
	for i := range tkrImageTags {
		tkrName := strings.ReplaceAll(tkrImageTags[i], "_", "---")
		res = append(res, tkrName)
	}
	return res
}

func getTKRNames(tkrs []*runv1.TanzuKubernetesRelease) sets.StringSet {
	tkrNames := sets.StringSet{}
	for i := range tkrs {
		tkrNames.Add(tkrs[i].Name)
	}
	return tkrNames
}

func getOSImagesNames(osImages []*runv1.OSImage) sets.StringSet {
	osImagesNames := sets.StringSet{}
	for i := range osImages {
		osImagesNames.Add(osImages[i].Name)
	}
	return osImagesNames
}
func getPackagesNames(packages []*kappcontrollerv1alpha1.Package) sets.StringSet {
	packagesNames := sets.StringSet{}
	for i := range packages {
		packagesNames.Add(packages[i].Name)
	}
	return packagesNames
}
