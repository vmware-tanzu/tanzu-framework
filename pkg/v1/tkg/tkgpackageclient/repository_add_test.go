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

const (
	testRepoName       = "test-repo"
	testRepoURL        = "test.registry.vmware.com/test-repo"
	testSecondRepoName = "test-repo-2"
	testSecondRepoURL  = "test.registry.vmware.com/test-repo-2"
)

var testRepository = &kappipkg.PackageRepository{
	TypeMeta:   metav1.TypeMeta{Kind: tkgpackagedatamodel.KindPackageRepository},
	ObjectMeta: metav1.ObjectMeta{Name: testRepoName, Namespace: testNamespaceName},
	Spec: kappipkg.PackageRepositorySpec{
		Fetch: &kappipkg.PackageRepositoryFetch{
			ImgpkgBundle: &kappctrl.AppFetchImgpkgBundle{Image: testRepoURL},
		},
	}}

var _ = Describe("Add Repository", func() {
	var (
		ctl     *pkgClient
		crtCtl  *fakes.CRTClusterClient
		kappCtl *fakes.KappClient
		err     error
		opts    = tkgpackagedatamodel.RepositoryOptions{
			RepositoryName:   testRepoName,
			RepositoryURL:    testRepoURL,
			Namespace:        testNamespaceName,
			CreateRepository: false,
			CreateNamespace:  false,
		}
		options           = opts
		pkgRepositoryList = &kappipkg.PackageRepositoryList{
			Items: []kappipkg.PackageRepository{*testRepository},
		}
	)

	JustBeforeEach(func() {
		ctl = &pkgClient{kappClient: kappCtl}
		err = ctl.AddRepository(&options)
	})

	Context("failure in listing package repositories", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageRepositoriesReturns(nil, errors.New("failure in ListPackageRepositories"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in ListPackageRepositories"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in validating repository name", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageRepositoriesReturns(pkgRepositoryList, nil)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("repository with the same name already exists"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in validating repository OCI registry URL", func() {
		BeforeEach(func() {
			options.RepositoryName = testSecondRepoName
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageRepositoriesReturns(pkgRepositoryList, nil)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("repository with the same OCI registry URL already exists"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in namespace creation", func() {
		BeforeEach(func() {
			options.CreateNamespace = true
			options.RepositoryName = testSecondRepoName
			options.RepositoryURL = testSecondRepoURL
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			crtCtl.GetReturns(errors.New("failure in Get namespace"))
			kappCtl.ListPackageRepositoriesReturns(pkgRepositoryList, nil)
			kappCtl.CreatePackageRepositoryReturns(nil)
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("failure in Get namespace"))
		})
		AfterEach(func() { options = opts })
	})

	Context("failure in creating package repository", func() {
		BeforeEach(func() {
			options.RepositoryName = testSecondRepoName
			options.RepositoryURL = testSecondRepoURL
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageRepositoriesReturns(pkgRepositoryList, nil)
			kappCtl.CreatePackageRepositoryReturns(errors.New("failure in CreatePackageRepository"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in CreatePackageRepository"))
		})
		AfterEach(func() { options = opts })
	})

	Context("success in creating the package repository in not previously existing 'test-ns' namespace", func() {
		BeforeEach(func() {
			options.CreateNamespace = true
			options.RepositoryName = testSecondRepoName
			options.RepositoryURL = testSecondRepoURL
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			crtCtl.GetReturns(apierrors.NewNotFound(schema.GroupResource{Resource: tkgpackagedatamodel.KindNamespace}, testNamespaceName))
			kappCtl.ListPackageRepositoriesReturns(pkgRepositoryList, nil)
			kappCtl.CreatePackageRepositoryReturns(nil)
		})
		It(testSuccessMsg, func() {
			Expect(err).NotTo(HaveOccurred())
			testRepositoryAddPostValidation(kappCtl, &options)
		})
		AfterEach(func() { options = opts })
	})

	Context("success in  creating the package repository in the already existing 'test-ns' namespace", func() {
		BeforeEach(func() {
			options.CreateNamespace = true
			options.RepositoryName = testSecondRepoName
			options.RepositoryURL = testSecondRepoURL
			kappCtl = &fakes.KappClient{}
			crtCtl = &fakes.CRTClusterClient{}
			kappCtl.GetClientReturns(crtCtl)
			crtCtl.GetReturns(nil)
			kappCtl.ListPackageRepositoriesReturns(pkgRepositoryList, nil)
			kappCtl.CreatePackageRepositoryReturns(nil)
		})
		It(testSuccessMsg, func() {
			Expect(err).NotTo(HaveOccurred())
			testRepositoryAddPostValidation(kappCtl, &options)
		})
		AfterEach(func() { options = opts })
	})

	Context("success in creating package repository", func() {
		BeforeEach(func() {
			options.RepositoryName = testSecondRepoName
			options.RepositoryURL = testSecondRepoURL
			kappCtl = &fakes.KappClient{}
			kappCtl.ListPackageRepositoriesReturns(pkgRepositoryList, nil)
			kappCtl.CreatePackageRepositoryReturns(nil)
		})
		It(testSuccessMsg, func() {
			Expect(err).ToNot(HaveOccurred())
			testRepositoryAddPostValidation(kappCtl, &options)
		})
		AfterEach(func() { options = opts })
	})
})

func testRepositoryAddPostValidation(kappCtl *fakes.KappClient, options *tkgpackagedatamodel.RepositoryOptions) {
	createRepoCallCnt := kappCtl.CreatePackageRepositoryCallCount()
	Expect(createRepoCallCnt).To(BeNumerically("==", 1))
	pkgRepo := kappCtl.CreatePackageRepositoryArgsForCall(0)
	Expect(pkgRepo.Name).Should(Equal(options.RepositoryName))
	Expect(pkgRepo.Namespace).Should(Equal(options.Namespace))
	Expect(pkgRepo.Spec.Fetch.ImgpkgBundle.Image).Should(Equal(options.RepositoryURL))
}
