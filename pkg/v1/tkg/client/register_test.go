// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

var _ = Describe("TMC Registration Tests", func() {
	var (
		tkgClient          *client.TkgClient
		err                error
		tkgConfigPath      string
		clusterName        string
		tmcRegistrationURL string

		clusterClient        *fakes.ClusterClient
		clusterClientFactory *fakes.ClusterClientFactory
	)

	BeforeEach(func() {
		tmcRegistrationURL = "https://example.com"
		clusterName = "regional-cluster-1"
	})

	JustBeforeEach(func() {
		optMutator := func(options client.Options) client.Options {
			options.ClusterClientFactory = clusterClientFactory
			return options
		}
		tkgClient, err = CreateTKGClientOpts(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second, optMutator)
		Expect(err).ToNot(HaveOccurred())
		err = tkgClient.RegisterManagementClusterToTmc(clusterName, tmcRegistrationURL)
	})

	When("TKG client config does not include region", func() {
		BeforeEach(func() {
			tkgConfigPath = "../fakes/config/config.yaml"
		})

		It("TMC registration should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("TKG client config includes region", func() {
		BeforeEach(func() {
			tkgConfigPath = "../fakes/config/config2.yaml"
			clusterClient = &fakes.ClusterClient{}
			clusterClientFactory = &fakes.ClusterClientFactory{}
		})

		When("ClusterClientFactory cannot create a clusterClient", func() {
			BeforeEach(func() {
				clusterClientFactory.NewClientReturns(nil, errors.New("fail"))
			})

			It("Should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		When("Cluster is a PacificRegionalCluster", func() {
			BeforeEach(func() {
				clusterClient.IsPacificRegionalClusterReturns(true, nil)
				clusterClientFactory.NewClientReturns(clusterClient, nil)
			})
			It("Should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		When("Cluster is not a Pacific management cluster", func() {
			BeforeEach(func() {
				clusterClient.IsPacificRegionalClusterReturns(false, nil)
				clusterClientFactory.NewClientReturns(clusterClient, nil)
			})

			It("Should apply the TMC url", func() {
				Expect(clusterClient.ApplyFileCallCount()).To(Equal(1))
				Expect(clusterClient.ApplyFileArgsForCall(0)).To(Equal(tmcRegistrationURL))
			})
		})
	})
})
