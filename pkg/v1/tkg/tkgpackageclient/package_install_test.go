// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/installpackage/v1alpha1"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	v1alpha1b "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/packages/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

const (
	testClusterRoleName        = "test-pkg-test-ns-cluster-role"
	testClusterRoleBindingName = "test-pkg-test-ns-cluster-rolebinding"
	testSecretValuesName       = "test-pkg-test-ns-values" //nolint:gosec
	testServiceAccountName     = "test-pkg-test-ns-sa"
	testNamespaceName          = "test-ns"
	testInstalledPkgName       = "test-pkg"
	testPollInterval           = 100 * time.Millisecond
	testPollTimeout            = 1 * time.Minute
	testFailureMsg             = "should return an error"
	testSuccessMsg             = "should not return an error"
	testUsefulErrMsg           = "some failure happened"
	testValuesFile             = "value-file"
)

var _ = Describe("Install Package", func() {
	var (
		ctl     *pkgClient
		crtCtl  *fakes.CRTClusterClient
		kappCtl *fakes.KappClient
		err     error
		opts    = tkgpackagedatamodel.PackageOptions{
			InstalledPkgName: testInstalledPkgName,
			Namespace:        testNamespaceName,
			PackageName:      "test-pkg.com",
			Version:          "1.0.0",
			PollInterval:     testPollInterval,
			PollTimeout:      testPollTimeout,
			CreateNamespace:  true,
		}
		options      = opts
		installedPkg = v1alpha1.InstalledPackage{
			TypeMeta:   metav1.TypeMeta{Kind: "InstalledPackage"},
			ObjectMeta: metav1.ObjectMeta{Name: testInstalledPkgName, Namespace: testNamespaceName},
			Status: v1alpha1.InstalledPackageStatus{
				GenericStatus: kappctrl.GenericStatus{
					Conditions: []kappctrl.AppCondition{
						{Type: kappctrl.Reconciling},
						{Type: kappctrl.ReconcileSucceeded},
					},
					UsefulErrorMessage: "",
				},
			},
		}
		pkgVersionList = &v1alpha1b.PackageVersionList{
			Items: []v1alpha1b.PackageVersion{
				{TypeMeta: metav1.TypeMeta{
					Kind: "PackageVersion"},
					ObjectMeta: metav1.ObjectMeta{Name: testInstalledPkgName, Namespace: testNamespaceName},
					Spec:       v1alpha1b.PackageVersionSpec{PackageName: testInstalledPkgName, Version: "1.0.0"},
				},
			},
		}
	)

	JustBeforeEach(func() {
		ctl = &pkgClient{kappClient: kappCtl}
		err = ctl.InstallPackage(&options)
	})

	Context("failure in listing package versions", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageVersionsReturns(nil, errors.New("failure in ListPackageVersions"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in ListPackageVersions"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in finding the provided service account", func() {
		BeforeEach(func() {
			options.ServiceAccountName = testServiceAccountName
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackageVersionsReturns(pkgVersionList, nil)
			crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: "ServiceAccount"}, testServiceAccountName))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ServiceAccount \"test-pkg-test-ns-sa\" not found"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in finding the provided package version", func() {
		BeforeEach(func() {
			options.Version = "2.0.0"
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackageVersionsReturns(pkgVersionList, nil)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to resolve package 'test-pkg.com' with version '2.0.0'"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in finding the provided package name", func() {
		BeforeEach(func() {
			options.PackageName = "test-pkg.org"
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackageVersionsReturns(pkgVersionList, nil)
			kappCtl.GetPackageByNameReturns(nil, apierrors.NewNotFound(schema.GroupResource{Resource: "Package"}, "test-pkg.org"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Package \"test-pkg.org\" not found"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in finding the provided secret value file", func() {
		BeforeEach(func() {
			options.ValuesFile = testValuesFile
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackageVersionsReturns(pkgVersionList, nil)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("open value-file: no such file or directory"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in getting installed package", func() {
		BeforeEach(func() {
			options.Wait = true
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackageVersionsReturns(pkgVersionList, nil)
			kappCtl.GetInstalledPackageReturns(nil, errors.New("failure in GetInstalledPackage"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in GetInstalledPackage"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in installed package reconciliation", func() {
		BeforeEach(func() {
			options.Wait = true
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackageVersionsReturns(pkgVersionList, nil)
			kappCtl.GetInstalledPackageReturns(&installedPkg, nil)
			Expect(len(installedPkg.Status.Conditions)).To(BeNumerically("==", 2))
			installedPkg.Status.Conditions[1] = kappctrl.AppCondition{Type: kappctrl.ReconcileFailed}
			installedPkg.Status.UsefulErrorMessage = testUsefulErrMsg
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(testUsefulErrMsg))
		})
		AfterEach(func() {
			options = opts
			installedPkg.Status.Conditions[1].Type = kappctrl.ReconcileSucceeded
		})
	})

	Context("success in installing the package in not previously existing 'test-ns' namespace", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackageVersionsReturns(pkgVersionList, nil)
			crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: "Namespace"}, testNamespaceName))
		})
		It(testSuccessMsg, func() {
			Expect(err).ToNot(HaveOccurred())
			expectedCreatedResourceNames := []string{testServiceAccountName, testClusterRoleName, testClusterRoleBindingName, testNamespaceName}
			testPackageInstallPostValidation(crtCtl, kappCtl, expectedCreatedResourceNames)
		})
		AfterEach(func() { options = opts })
	})

	Context("success in installing the package in the already existing 'test-ns' namespace", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackageVersionsReturns(pkgVersionList, nil)
		})
		It(testSuccessMsg, func() {
			Expect(err).ToNot(HaveOccurred())
			expectedCreatedResourceNames := []string{testServiceAccountName, testClusterRoleName, testClusterRoleBindingName}
			testPackageInstallPostValidation(crtCtl, kappCtl, expectedCreatedResourceNames)
		})
		AfterEach(func() { options = opts })
	})

	Context("success in installing the package with wait flag being set", func() {
		BeforeEach(func() {
			options.Wait = true
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackageVersionsReturns(pkgVersionList, nil)
			kappCtl.GetInstalledPackageReturns(&installedPkg, nil)
		})
		It(testSuccessMsg, func() {
			Expect(err).ToNot(HaveOccurred())
			expectedCreatedResourceNames := []string{testServiceAccountName, testClusterRoleName, testClusterRoleBindingName}
			testPackageInstallPostValidation(crtCtl, kappCtl, expectedCreatedResourceNames)
		})
		AfterEach(func() { options = opts })
	})

	Context("success in installing the package with secret value file specified", func() {
		BeforeEach(func() {
			options.ValuesFile = testValuesFile
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackageVersionsReturns(pkgVersionList, nil)
			err = ioutil.WriteFile(testValuesFile, []byte("test"), 0644)
			Expect(err).ToNot(HaveOccurred())
		})
		It(testSuccessMsg, func() {
			Expect(err).ToNot(HaveOccurred())
			expectedCreatedResourceNames := []string{testServiceAccountName, testClusterRoleName, testClusterRoleBindingName, testSecretValuesName}
			testPackageInstallPostValidation(crtCtl, kappCtl, expectedCreatedResourceNames)
		})
		AfterEach(func() {
			options = opts
			err = os.Remove(testValuesFile)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

func testPackageInstallPostValidation(crtCtl *fakes.CRTClusterClient, kappCtl *fakes.KappClient, expCreatedResourceNames []string) {
	createCallCnt := crtCtl.CreateCallCount()
	Expect(createCallCnt).To(BeNumerically("==", len(expCreatedResourceNames)))
	createdResourceNames := make([]string, createCallCnt)
	for i := 0; i < createCallCnt; i++ {
		_, obj, _ := crtCtl.CreateArgsForCall(i)
		createdResourceNames[i] = testGetObjectName(obj)
	}
	Expect(createdResourceNames).Should(ConsistOf(expCreatedResourceNames))

	kappCreateCallCnt := kappCtl.CreateInstalledPackageCallCount()
	Expect(kappCreateCallCnt).To(BeNumerically("==", 1))
	installed, _, _ := kappCtl.CreateInstalledPackageArgsForCall(0)
	Expect(installed.Name).Should(Equal(testInstalledPkgName))
}

func testGetObjectName(o interface{}) string {
	switch o := o.(type) {
	case *corev1.Namespace:
		return o.Name
	case *corev1.Secret:
		return o.Name
	case *corev1.ServiceAccount:
		return o.Name
	case *rbacv1.ClusterRole:
		return o.Name
	case *rbacv1.ClusterRoleBinding:
		return o.Name
	case *v1alpha1.InstalledPackage:
		return o.Name
	default:
		return ""
	}
}
