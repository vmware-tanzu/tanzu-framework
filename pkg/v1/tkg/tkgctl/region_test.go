// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
)

var _ = Describe("Unit tests for add region", func() {
	var (
		ctl       tkgctl
		tkgClient = &fakes.Client{}
		options   AddRegionOptions
		err       error
	)

	JustBeforeEach(func() {
		ctl = tkgctl{
			configDir: testingDir,
			tkgClient: tkgClient,
		}
		err = ctl.AddRegion(options)
	})

	Context("When the given cluster is not a management cluster", func() {
		BeforeEach(func() {
			options = AddRegionOptions{
				Overwrite:          true,
				UseDirectReference: true,
			}
			tkgClient.VerifyRegionReturns(region.RegionContext{}, errors.New("not a mgmt cluster"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("When the same mgmt cluster is already added", func() {
		BeforeEach(func() {
			options = AddRegionOptions{
				Overwrite:          true,
				UseDirectReference: true,
			}
			tkgClient.VerifyRegionReturns(region.RegionContext{}, nil)
			tkgClient.AddRegionContextReturns(errors.New("mgmt cluster already exists"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("When mgmt context is overwritten", func() {
		BeforeEach(func() {
			options = AddRegionOptions{
				Overwrite:          true,
				UseDirectReference: true,
			}
			tkgClient.VerifyRegionReturns(region.RegionContext{}, nil)
			tkgClient.AddRegionContextReturns(nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When new mgmt context is added", func() {
		BeforeEach(func() {
			options = AddRegionOptions{
				Overwrite:          false,
				UseDirectReference: true,
			}
			tkgClient.VerifyRegionReturns(region.RegionContext{}, nil)
			tkgClient.AddRegionContextReturns(nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("Unit test for set region", func() {
	var (
		ctl       tkgctl
		tkgClient = &fakes.Client{}
		options   SetRegionOptions
		err       error
	)

	JustBeforeEach(func() {
		ctl = tkgctl{
			configDir: testingDir,
			tkgClient: tkgClient,
		}
		options = SetRegionOptions{
			ClusterName: "my-cluster",
		}
		err = ctl.SetRegion(options)
	})
	Context("when region does not exist", func() {
		BeforeEach(func() {
			tkgClient.SetRegionContextReturns(errors.New("region not found"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("when region exists", func() {
		BeforeEach(func() {
			tkgClient.SetRegionContextReturns(nil)
		})
		It("should rnot eturn an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("Unit test for get region", func() {
	var (
		ctl       tkgctl
		tkgClient = &fakes.Client{}

		err error
	)

	JustBeforeEach(func() {
		ctl = tkgctl{
			configDir: testingDir,
			tkgClient: tkgClient,
		}
		_, err = ctl.GetRegions("mgmt-cluster-name")
	})
	Context("when failed to list the regions", func() {
		BeforeEach(func() {
			tkgClient.GetRegionContextsReturns(nil, errors.New("region not found"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("when region exists", func() {
		BeforeEach(func() {
			tkgClient.GetRegionContextsReturns(nil, nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("Unit test for delete region", func() {
	var (
		ctl       tkgctl
		tkgClient = &fakes.Client{}
		ops       = DeleteRegionOptions{
			ClusterName: "my-cluster",
			SkipPrompt:  true,
			Timeout:     time.Minute * 30,
		}
		err error
	)

	JustBeforeEach(func() {
		ctl = tkgctl{
			configDir:  testingDir,
			tkgClient:  tkgClient,
			kubeconfig: "./kube",
		}
		err = ctl.DeleteRegion(ops)
	})

	Context("when failed to delete region", func() {
		BeforeEach(func() {
			tkgClient.DeleteRegionReturns(errors.New("region not found"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("when it is able to delete region", func() {
		BeforeEach(func() {
			tkgClient.DeleteRegionReturns(nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
