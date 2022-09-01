// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	. "github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackageclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

var _ = Describe("Delete Repository", func() {
	var (
		ctl     TKGPackageClient
		kappCtl *fakes.KappClient
		err     error
		opts    = tkgpackagedatamodel.RepositoryOptions{
			RepositoryName: testRepoName,
			IsForceDelete:  false,
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
		go ctl.DeleteRepository(&options, progress)
		err = testReceive(progress)
	})

	Context("failure in deleting the package repository due to DeletePackageRepository API error", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageRepositoryReturns(testRepository, nil)
			kappCtl.DeletePackageRepositoryReturns(errors.New("failure in DeletePackageRepository"))
		})
		It(testFailureMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in DeletePackageRepository"))
		})
		AfterEach(func() { options = opts })
	})

	Context("not being able to get the package repository due to failure in GetPackageRepository", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageRepositoryReturns(nil, errors.New("failure in GetPackageRepository"))
		})
		It(testSuccessMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failure in GetPackageRepository"))
		})
		AfterEach(func() { options = opts })
	})

	Context("not being able to get the package repository due to non existent repository", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageRepositoryReturns(nil, apierrors.NewNotFound(schema.GroupResource{Resource: tkgpackagedatamodel.KindPackageRepository}, testRepoName))
		})
		It(testSuccessMsg, func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(tkgpackagedatamodel.ErrRepoNotExists))
		})
		AfterEach(func() { options = opts })
	})

	Context("success in deleting the package repository", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageRepositoryReturns(testRepository, nil)
			kappCtl.DeletePackageRepositoryReturns(nil)

		})
		It(testSuccessMsg, func() {
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() { options = opts })
	})
})
