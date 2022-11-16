// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
)

var _ = Describe("Unit tests for core package repository cleanup", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		tkgClient             *TkgClient
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Millisecond)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("When cleanup core package repository", func() {
		JustBeforeEach(func() {
			err = tkgClient.CleanupCorePackageRepo(regionalClusterClient)
		})
		Context("When the core package repository is not found", func() {
			BeforeEach(func() {
				regionalClusterClient.GetResourceReturns(apierrors.NewNotFound(
					schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
					"fakeGroupResource"))
			})
			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("When unable to find the core package repository", func() {
			BeforeEach(func() {
				regionalClusterClient.GetResourceReturns(errors.New("unable to find the resource"))
			})
			It("should not return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to find the resource"))
			})
		})
		Context("When unable to delete the core package repository", func() {
			BeforeEach(func() {
				regionalClusterClient.GetResourceReturns(nil)
				regionalClusterClient.DeleteResourceReturns(errors.New("unable to delete the resource"))
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to delete the resource"))
			})
		})
		Context("When the core package repository is not found for deleting", func() {
			BeforeEach(func() {
				regionalClusterClient.GetResourceReturns(nil)
				regionalClusterClient.DeleteResourceReturns(apierrors.NewNotFound(
					schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
					"fakeGroupResource"))
			})
			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("When the core package repository is deleted successfully", func() {
			BeforeEach(func() {
				regionalClusterClient.GetResourceReturns(nil)
				regionalClusterClient.DeleteResourceReturns(nil)
			})
			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})
})
