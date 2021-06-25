// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var _ = Describe("Delete Repository", func() {
	var (
		ctl     *pkgClient
		kappCtl *fakes.KappClient
		err     error
		opts    = tkgpackagedatamodel.RepositoryOptions{
			RepositoryName: testRepoName,
			IsForceDelete:  false,
		}
		options = opts
		found   bool
	)

	JustBeforeEach(func() {
		ctl = &pkgClient{kappClient: kappCtl}
		found, err = ctl.DeleteRepository(&options)
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

	Context("not being able to get the package repository, no error should be returned", func() {
		BeforeEach(func() {
			kappCtl = &fakes.KappClient{}
			kappCtl.GetPackageRepositoryReturns(nil, errors.New("failure in GetPackageRepository"))
		})
		It(testSuccessMsg, func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
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
			Expect(found).To(BeTrue())
		})
		AfterEach(func() { options = opts })
	})
})
