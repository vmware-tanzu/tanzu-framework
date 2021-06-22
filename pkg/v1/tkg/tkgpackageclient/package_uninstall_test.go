// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var _ = Describe("Uninstall Package", func() {
	var (
		ctl     *pkgClient
		crtCtl  *fakes.CRTClusterClient
		kappCtl *fakes.KappClient
		err     error
		found   bool
		opts    = tkgpackagedatamodel.PackageUninstallOptions{
			PkgInstallName: testPkgInstallName,
			Namespace:      testNamespaceName,
			PollInterval:   testPollInterval,
			PollTimeout:    testPollTimeout,
		}
		options = opts
		app     = kappctrl.App{
			TypeMeta:   metav1.TypeMeta{Kind: "App"},
			ObjectMeta: metav1.ObjectMeta{Name: testPkgInstallName, Namespace: testNamespaceName},
			Status: kappctrl.AppStatus{
				GenericStatus: kappctrl.GenericStatus{
					Conditions: []kappctrl.AppCondition{
						{Type: kappctrl.Deleting},
					},
				},
			},
		}
		pkgInstall = kappipkg.PackageInstall{
			TypeMeta: metav1.TypeMeta{Kind: tkgpackagedatamodel.KindPackageInstall},
			ObjectMeta: metav1.ObjectMeta{
				Name:      testPkgInstallName,
				Namespace: testNamespaceName,
				Annotations: map[string]string{
					tkgpackagedatamodel.TanzuPkgPluginAnnotation + "-" + tkgpackagedatamodel.KindClusterRole:        "test-pkg-test-ns-cluster-role",
					tkgpackagedatamodel.TanzuPkgPluginAnnotation + "-" + tkgpackagedatamodel.KindClusterRoleBinding: "test-pkg-test-ns-cluster-rolebinding",
					tkgpackagedatamodel.TanzuPkgPluginAnnotation + "-" + tkgpackagedatamodel.KindServiceAccount:     "test-pkg-test-ns-sa",
					tkgpackagedatamodel.TanzuPkgPluginAnnotation + "-" + tkgpackagedatamodel.KindSecret:             "test-pkg-test-ns-values",
				}},
		}
	)

	JustBeforeEach(func() {
		ctl = &pkgClient{kappClient: kappCtl}
		found, err = ctl.UninstallPackage(&options)
	})

	Context("failure in getting installed packages", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageInstallReturns(nil, errors.New("failure in GetPackageInstall"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in GetPackageInstall"))
			Expect(found).To(BeFalse())
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in deleting installed package", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.GetPackageInstallReturns(&pkgInstall, nil)
			crtCtl.DeleteReturns(errors.New("failure in PackageInstall deletion"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in PackageInstall deletion"))
			Expect(found).To(BeTrue())
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in App CR deletion", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.GetPackageInstallReturns(&pkgInstall, nil)
			crtCtl.DeleteReturns(nil)
			app.Status.Conditions = append(app.Status.Conditions, kappctrl.AppCondition{
				Type: kappctrl.DeleteFailed,
			})
			app.Status.UsefulErrorMessage = testUsefulErrMsg
			kappCtl.GetAppCRReturns(&app, nil)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(testUsefulErrMsg))
			Expect(found).To(BeTrue())
		})
		AfterEach(func() { options = opts })
	})

	Context("success in uninstalling the package and associated resources", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			kappCtl.GetPackageInstallReturns(&pkgInstall, nil)
			crtCtl.DeleteReturns(nil)
			kappCtl.GetAppCRReturns(nil, apierrors.NewNotFound(schema.GroupResource{Resource: "App"}, testPkgInstallName))
			Expect(len(pkgInstall.GetAnnotations())).To(Equal(len(pkgInstall.Annotations)))
		})
		It(testSuccessMsg, func() {
			Expect(err).ToNot(HaveOccurred())
			expectedDeletedResourceNames := []string{testServiceAccountName, testClusterRoleName, testClusterRoleBindingName, testSecretValuesName, testPkgInstallName}
			deleteCallCnt := crtCtl.DeleteCallCount()
			Expect(deleteCallCnt).To(BeNumerically("==", len(expectedDeletedResourceNames)))
			deletedResourceNames := make([]string, deleteCallCnt)
			for i := 0; i < deleteCallCnt; i++ {
				_, obj, _ := crtCtl.DeleteArgsForCall(i)
				deletedResourceNames[i] = testGetObjectName(obj)
			}
			Expect(deletedResourceNames).Should(ConsistOf(expectedDeletedResourceNames))
			Expect(found).To(BeTrue())
		})
		AfterEach(func() { options = opts })
	})
})
