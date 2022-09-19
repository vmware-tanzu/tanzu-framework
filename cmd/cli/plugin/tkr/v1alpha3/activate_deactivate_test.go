// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
)

var _ = Describe("deactivate/activate TKR", func() {
	var (
		tkrName       string
		err           error
		clusterClient *fakes.ClusterClient
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		tkrName = "fakeTKRName"
	})

	Context("deactivating TKR", func() {
		JustBeforeEach(func() {
			err = deactivateKubernetesReleases(clusterClient, tkrName)
		})
		Context("When the patching TKR return error", func() {
			BeforeEach(func() {
				clusterClient.PatchResourceReturns(errors.New("fake TKR patch error"))
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake TKR patch error"))
			})
		})
		Context("When the patching TKR return success", func() {
			BeforeEach(func() {
				clusterClient.PatchResourceReturns(nil)
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
				_, getTKRName, _, gotPatch, gotPatchType, _ := clusterClient.PatchResourceArgsForCall(0)
				Expect(getTKRName).To(Equal("fakeTKRName"))
				Expect(gotPatchType).To(Equal(types.MergePatchType))
				expectedPatchLabel := `"deactivated": ""`
				Expect(gotPatch).To(ContainSubstring(expectedPatchLabel))
			})
		})
	})

	Context("activating TKR", func() {
		JustBeforeEach(func() {
			err = activateKubernetesReleases(clusterClient, tkrName)
		})
		Context("When the patching TKR return error", func() {
			BeforeEach(func() {
				clusterClient.PatchResourceReturns(errors.New("fake TKR patch error"))
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake TKR patch error"))
			})
		})
		Context("When the patching TKR return success", func() {
			BeforeEach(func() {
				clusterClient.PatchResourceReturns(nil)
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
				_, getTKRName, _, gotPatch, gotPatchType, _ := clusterClient.PatchResourceArgsForCall(0)
				Expect(getTKRName).To(Equal("fakeTKRName"))
				Expect(gotPatchType).To(Equal(types.MergePatchType))
				expectedPatchLabel := `"deactivated": null`
				Expect(gotPatch).To(ContainSubstring(expectedPatchLabel))
			})
		})
	})

})
