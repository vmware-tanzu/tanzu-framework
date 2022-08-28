// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackageclient"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

const (
	testClusterRoleName        = "test-pkg-test-ns-cluster-role"
	testClusterRoleBindingName = "test-pkg-test-ns-cluster-rolebinding"
	testSecretValuesName       = "test-pkg-test-ns-values" //nolint:gosec
	testServiceAccountName     = "test-pkg-test-ns-sa"
	testNamespaceName          = "test-ns"
	testPkgInstallName         = "test-pkg"
	testPkgName                = "test-pkg.com"
	testPkgVersion             = "1.0.0"
	testPollInterval           = 100 * time.Millisecond
	testPollTimeout            = 1 * time.Minute
	testFailureMsg             = "should return an error"
	testSuccessMsg             = "should not return an error"
	testUsefulErrMsg           = "some failure happened"
	testValuesFile             = "value-file"
)

var (
	testPkgInstall = &kappipkg.PackageInstall{
		TypeMeta:   metav1.TypeMeta{Kind: tkgpackagedatamodel.KindPackageInstall},
		ObjectMeta: metav1.ObjectMeta{Name: testPkgInstallName, Namespace: testNamespaceName, Generation: 1},
		Spec: kappipkg.PackageInstallSpec{
			ServiceAccountName: testServiceAccountName,
			PackageRef: &kappipkg.PackageRef{
				RefName:          testPkgName,
				VersionSelection: testVersionSelection,
			},
		},
		Status: kappipkg.PackageInstallStatus{
			GenericStatus: kappctrl.GenericStatus{
				Conditions:         []kappctrl.AppCondition{{Type: kappctrl.Reconciling, Status: corev1.ConditionTrue}, {Type: kappctrl.ReconcileSucceeded, Status: corev1.ConditionTrue}},
				UsefulErrorMessage: "",
				ObservedGeneration: 1,
			},
		},
	}

	testPkgVersionList = &kapppkg.PackageList{
		TypeMeta: metav1.TypeMeta{Kind: "PackageVersionList"},
		Items: []kapppkg.Package{
			{TypeMeta: metav1.TypeMeta{
				Kind: "PackageVersion"},
				ObjectMeta: metav1.ObjectMeta{Name: testPkgInstallName, Namespace: testNamespaceName},
				Spec:       kapppkg.PackageSpec{RefName: testPkgInstallName, Version: testPkgVersion},
			},
		},
	}

	testVersionSelection = &versions.VersionSelectionSemver{Constraints: "1.0.0"}
)

var _ = Describe("Install Package", func() {
	var (
		ctl     TKGPackageClient
		crtCtl  *fakes.CRTClusterClient
		kappCtl *fakes.KappClient
		err     error
		opts    = tkgpackagedatamodel.PackageOptions{
			PkgInstallName:  testPkgInstallName,
			Namespace:       testNamespaceName,
			PackageName:     testPkgName,
			Version:         testPkgVersion,
			PollInterval:    testPollInterval,
			PollTimeout:     testPollTimeout,
			CreateNamespace: true,
		}
		options  = opts
		progress *tkgpackagedatamodel.PackageProgress
	)

	JustBeforeEach(func() {
		progress = &tkgpackagedatamodel.PackageProgress{
			ProgressMsg: make(chan string, 10),
			Err:         make(chan error),
			Done:        make(chan struct{}),
		}
		ctl, err = NewTKGPackageClientWithKappClient(kappCtl)
		Expect(err).NotTo(HaveOccurred())
		go ctl.InstallPackage(&options, progress, tkgpackagedatamodel.OperationTypeInstall)
		err = testReceive(progress)
	})

	Context("failure in getting installed package due to GetPackageInstall API error", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			kappCtl.GetPackageInstallReturnsOnCall(0, nil, errors.New("failure in GetPackageInstall"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in GetPackageInstall"))
		})
		AfterEach(func() { options = opts })
	})

	Context("falling back to update when trying to install an existing package install (with reconciliation failure)", func() {
		BeforeEach(func() {
			options.Wait = true
			options.ValuesFile = testValuesFile
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			kappCtl.GetPackageInstallReturns(testPkgInstall, nil)
			err = os.WriteFile(testValuesFile, []byte("test"), 0644)
			Expect(err).ToNot(HaveOccurred())
			Expect(testPkgInstall.Status.ObservedGeneration).To(Equal(testPkgInstall.Generation))
			Expect(len(testPkgInstall.Status.Conditions)).To(BeNumerically("==", 2))
			testPkgInstall.Status.Conditions[1] = kappctrl.AppCondition{Type: kappctrl.ReconcileFailed, Status: corev1.ConditionTrue}
			testPkgInstall.Status.UsefulErrorMessage = testUsefulErrMsg
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(testUsefulErrMsg))
		})
		AfterEach(func() {
			options = opts
			testPkgInstall.Status.Conditions[1].Type = kappctrl.ReconcileSucceeded
			err = os.Remove(testValuesFile)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("falling back to update when trying to install an existing package install (throwing non-critical error)", func() {
		BeforeEach(func() {
			options.Wait = true
			options.PackageName = testPkgName
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			Expect(options.PackageName).To(Equal(testPkgInstall.Spec.PackageRef.RefName))
			kappCtl.GetPackageInstallReturns(testPkgInstall, nil)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(tkgpackagedatamodel.ErrPackageAlreadyExists))
		})
		AfterEach(func() {
			options = opts
		})
	})

	Context("failure in listing package versions due to ListPackages API error (in GetPackage())", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(nil, errors.New("failure in ListPackages"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in ListPackages"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in finding the provided package version", func() {
		BeforeEach(func() {
			options.Version = "2.0.0"
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
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
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			kappCtl.GetPackageMetadataByNameReturns(nil, apierrors.NewNotFound(schema.GroupResource{Resource: "PackageMetadata"}, "test-pkg.org"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("PackageMetadata \"test-pkg.org\" not found"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in namespace creation due to namespace Get API error", func() {
		BeforeEach(func() {
			options.CreateNamespace = true
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			crtCtl.GetReturns(errors.New("failure in Get namespace"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in Get namespace"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in namespace creation due to namespace Create API error", func() {
		BeforeEach(func() {
			options.CreateNamespace = true
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: tkgpackagedatamodel.KindNamespace}, testNamespaceName))
			crtCtl.CreateReturns(errors.New("failure in Create namespace"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in Create namespace"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in getting an existing namespace (namespace NotFound error)", func() {
		BeforeEach(func() {
			options.CreateNamespace = false
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: tkgpackagedatamodel.KindNamespace}, testNamespaceName))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Namespace \"%s\" not found", testNamespaceName)))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in creating service account", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			crtCtl.CreateReturnsOnCall(0, errors.New("failure in Create ServiceAccount"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in Create ServiceAccount"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in updating service account", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			crtCtl.CreateReturnsOnCall(0, apierrors.NewAlreadyExists(schema.GroupResource{Resource: tkgpackagedatamodel.KindServiceAccount}, testServiceAccountName))
			crtCtl.UpdateReturnsOnCall(0, errors.New("failure in Update ServiceAccount"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in Update ServiceAccount"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in creating cluster admin role", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			crtCtl.CreateReturnsOnCall(0, nil)
			crtCtl.CreateReturnsOnCall(1, errors.New("failure in Create ClusterRole"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in Create ClusterRole"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in updating cluster admin role", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			crtCtl.CreateReturnsOnCall(0, apierrors.NewAlreadyExists(schema.GroupResource{Resource: tkgpackagedatamodel.KindServiceAccount}, testServiceAccountName))
			crtCtl.UpdateReturnsOnCall(0, nil)
			crtCtl.CreateReturnsOnCall(1, apierrors.NewAlreadyExists(schema.GroupResource{Resource: tkgpackagedatamodel.KindClusterRole}, testClusterRoleName))
			crtCtl.UpdateReturnsOnCall(1, errors.New("failure in Update ClusterRole"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in Update ClusterRole"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in creating cluster role binding", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			crtCtl.CreateReturnsOnCall(0, nil)
			crtCtl.CreateReturnsOnCall(1, nil)
			crtCtl.CreateReturnsOnCall(2, errors.New("failure in Create ClusterRoleBinding"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in Create ClusterRoleBinding"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in updating cluster role binding", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			crtCtl.CreateReturnsOnCall(0, apierrors.NewAlreadyExists(schema.GroupResource{Resource: tkgpackagedatamodel.KindServiceAccount}, testServiceAccountName))
			crtCtl.UpdateReturnsOnCall(0, nil)
			crtCtl.CreateReturnsOnCall(1, apierrors.NewAlreadyExists(schema.GroupResource{Resource: tkgpackagedatamodel.KindClusterRole}, testClusterRoleName))
			crtCtl.UpdateReturnsOnCall(1, nil)
			crtCtl.CreateReturnsOnCall(2, apierrors.NewAlreadyExists(schema.GroupResource{Resource: tkgpackagedatamodel.KindClusterRoleBinding}, testClusterRoleBindingName))
			crtCtl.UpdateReturnsOnCall(2, errors.New("failure in Update ClusterRoleBinding"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in Update ClusterRoleBinding"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in finding the provided service account", func() {
		BeforeEach(func() {
			options.ServiceAccountName = testServiceAccountName
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: tkgpackagedatamodel.KindServiceAccount}, testServiceAccountName))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ServiceAccount \"test-pkg-test-ns-sa\" not found"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in finding the provided secret value file", func() {
		BeforeEach(func() {
			options.ValuesFile = testValuesFile
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("open value-file: no such file or directory"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in creating secret", func() {
		BeforeEach(func() {
			options.ValuesFile = testValuesFile
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			err = os.WriteFile(testValuesFile, []byte("test"), 0644)
			Expect(err).ToNot(HaveOccurred())
			crtCtl.CreateReturnsOnCall(0, nil)
			crtCtl.CreateReturnsOnCall(1, nil)
			crtCtl.CreateReturnsOnCall(2, nil)
			crtCtl.CreateReturnsOnCall(3, errors.New("failure in Create Secret"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in Create Secret"))
		})
		AfterEach(func() {
			options = opts
			err = os.Remove(testValuesFile)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("failure in updating secret", func() {
		BeforeEach(func() {
			options.ValuesFile = testValuesFile
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			err = os.WriteFile(testValuesFile, []byte("test"), 0644)
			Expect(err).ToNot(HaveOccurred())
			crtCtl.CreateReturnsOnCall(0, apierrors.NewAlreadyExists(schema.GroupResource{Resource: tkgpackagedatamodel.KindServiceAccount}, testServiceAccountName))
			crtCtl.UpdateReturnsOnCall(0, nil)
			crtCtl.CreateReturnsOnCall(1, apierrors.NewAlreadyExists(schema.GroupResource{Resource: tkgpackagedatamodel.KindClusterRole}, testClusterRoleName))
			crtCtl.UpdateReturnsOnCall(1, nil)
			crtCtl.CreateReturnsOnCall(2, apierrors.NewAlreadyExists(schema.GroupResource{Resource: tkgpackagedatamodel.KindClusterRoleBinding}, testClusterRoleBindingName))
			crtCtl.UpdateReturnsOnCall(2, nil)
			crtCtl.CreateReturnsOnCall(3, apierrors.NewAlreadyExists(schema.GroupResource{Resource: tkgpackagedatamodel.KindSecret}, testSecretName))
			crtCtl.UpdateReturnsOnCall(3, errors.New("failure in Update Secret"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in Update Secret"))
		})
		AfterEach(func() {
			options = opts
			err = os.Remove(testValuesFile)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("failure in creating package install due to CreatePackageInstall API error", func() {
		BeforeEach(func() {
			options.ValuesFile = testValuesFile
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			err = os.WriteFile(testValuesFile, []byte("test"), 0644)
			Expect(err).ToNot(HaveOccurred())
			crtCtl.CreateReturnsOnCall(0, nil)
			crtCtl.CreateReturnsOnCall(1, nil)
			crtCtl.CreateReturnsOnCall(2, nil)
			crtCtl.CreateReturnsOnCall(3, nil)
			kappCtl.CreatePackageInstallReturns(errors.New("failure in CreatePackageInstall"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in CreatePackageInstall"))
		})
		AfterEach(func() {
			options = opts
			err = os.Remove(testValuesFile)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("failure when trying to install a package install (with reconciliation failure)", func() {
		BeforeEach(func() {
			options.Wait = true
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			kappCtl.GetPackageInstallReturnsOnCall(0, nil, apierrors.NewNotFound(schema.GroupResource{Resource: tkgpackagedatamodel.KindPackageInstall}, testPkgInstallName))
			kappCtl.GetPackageInstallReturnsOnCall(1, testPkgInstall, nil)
			Expect(testPkgInstall.Status.ObservedGeneration).To(Equal(testPkgInstall.Generation))
			Expect(len(testPkgInstall.Status.Conditions)).To(BeNumerically("==", 2))
			testPkgInstall.Status.Conditions[1] = kappctrl.AppCondition{Type: kappctrl.ReconcileFailed, Status: corev1.ConditionTrue}
			testPkgInstall.Status.UsefulErrorMessage = testUsefulErrMsg
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(testUsefulErrMsg))
		})
		AfterEach(func() {
			options = opts
			testPkgInstall.Status.Conditions[1].Type = kappctrl.ReconcileSucceeded
		})
	})

	Context("success in installing the package in not previously existing 'test-ns' namespace", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: tkgpackagedatamodel.KindNamespace}, testNamespaceName))
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
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
		})
		It(testSuccessMsg, func() {
			Expect(err).ToNot(HaveOccurred())
			expectedCreatedResourceNames := []string{testServiceAccountName, testClusterRoleName, testClusterRoleBindingName}
			testPackageInstallPostValidation(crtCtl, kappCtl, expectedCreatedResourceNames)
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in installing the package due to GetPackageInstall API error in waitForResourceInstallation", func() {
		BeforeEach(func() {
			options.Wait = true
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			kappCtl.GetPackageInstallReturnsOnCall(0, nil, nil)
			kappCtl.GetPackageInstallReturnsOnCall(1, nil, errors.New("failure in GetPackageInstall"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in GetPackageInstall"))
		})
		AfterEach(func() { options = opts })
	})

	Context("success in installing the package with a successful reconciliation (Wait flag being set)", func() {
		BeforeEach(func() {
			options.Wait = true
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			kappCtl.GetPackageInstallReturnsOnCall(0, nil, apierrors.NewNotFound(schema.GroupResource{Resource: tkgpackagedatamodel.KindPackageInstall}, testPkgInstallName))
			kappCtl.GetPackageInstallReturnsOnCall(1, testPkgInstall, nil)
			Expect(testPkgInstall.Status.ObservedGeneration).To(Equal(testPkgInstall.Generation))
		})
		It(testSuccessMsg, func() {
			Expect(err).ToNot(HaveOccurred())
			expectedCreatedResourceNames := []string{testServiceAccountName, testClusterRoleName, testClusterRoleBindingName}
			testPackageInstallPostValidation(crtCtl, kappCtl, expectedCreatedResourceNames)
		})
		AfterEach(func() {
			options = opts
		})
	})

	Context("success in installing the package with secret value file specified", func() {
		BeforeEach(func() {
			options.ValuesFile = testValuesFile
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.ListPackagesReturns(testPkgVersionList, nil)
			err = os.WriteFile(testValuesFile, []byte("test"), 0644)
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

	Context("failure when trying to update an existing package install, but providing a different o.PackageName than Spec.PackageRef.RefName", func() {
		BeforeEach(func() {
			options.Wait = true
			options.PackageName = "some-other-package"
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			Expect(options.PackageName).NotTo(Equal(testPkgInstall.Spec.PackageRef.RefName))
			kappCtl.GetPackageInstallReturns(testPkgInstall, nil)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("installed package '%s' is already associated with package '%s'", options.PkgInstallName, testPkgInstall.Spec.PackageRef.RefName)))
		})
		AfterEach(func() {
			options = opts
			options.PackageName = testPkgName
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

	kappCreateCallCnt := kappCtl.CreatePackageInstallCallCount()
	Expect(kappCreateCallCnt).To(BeNumerically("==", 1))
	installed, _ := kappCtl.CreatePackageInstallArgsForCall(0)
	Expect(installed.Name).Should(Equal(testPkgInstallName))
}

func testGetObjectName(o interface{}) string {
	accessor, err := meta.Accessor(o)
	Expect(err).ToNot(HaveOccurred())
	return accessor.GetName()
}

func testReceive(progress *tkgpackagedatamodel.PackageProgress) error {
	for {
		select {
		case err := <-progress.Err:
			return err
		case <-progress.ProgressMsg:
			continue
		case <-progress.Done:
			return nil
		}
	}
}
