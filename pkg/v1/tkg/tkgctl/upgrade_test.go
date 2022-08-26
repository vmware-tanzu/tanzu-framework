// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	fakeproviders "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/providers"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
)

var _ = Describe("Unit tests for upgrade management cluster", func() {
	var (
		ctl           *tkgctl
		tkgClient     = &fakes.Client{}
		updaterClient = &fakes.TKGConfigUpdaterClient{}
		bomClient     = &fakes.TKGConfigBomClient{}
		configDir     string
		err           error
		ops           UpgradeRegionOptions
		tkrBom        = tkgconfigbom.BOMConfiguration{
			Release:    &tkgconfigbom.ReleaseInfo{Version: "v1.21.4+vmware.1.tkg.1"},
			Components: map[string][]*tkgconfigbom.ComponentInfo{"kubernetes": {{Version: "v1.21.4"}}},
		}
	)
	JustBeforeEach(func() {
		configDir, err = os.MkdirTemp("", "test")
		err = os.MkdirAll(configDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
		prepareConfiDir(configDir)
		options := Options{
			ConfigDir:      configDir,
			ProviderGetter: fakeproviders.FakeProviderGetter(),
		}
		c, createErr := New(options)
		Expect(createErr).ToNot(HaveOccurred())
		ctl, _ = c.(*tkgctl)
		ctl.tkgClient = tkgClient
		ctl.tkgBomClient = bomClient
		ctl.tkgConfigUpdaterClient = updaterClient
		ops = UpgradeRegionOptions{
			ClusterName:         "my-mgmt-cluster",
			VSphereTemplateName: "ubuntu-ova",
			OSName:              "ubuntu",
			OSVersion:           "xxxx",
			OSArch:              "amd64",
			SkipPrompt:          true,
		}
		err = ctl.UpgradeRegion(ops)
	})

	Context("when cannot get default tkr bom", func() {
		BeforeEach(func() {
			bomClient.GetDefaultTkrBOMConfigurationReturns(nil, errors.New("unable to get default tkr bom"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When tkg bom is mal-formated", func() {
		BeforeEach(func() {
			bomClient.GetDefaultTkrBOMConfigurationReturns(&tkrBom, nil)
			bomClient.GetDefaultTkgBOMConfigurationReturns(nil, errors.New("failed to get tkg bom"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When mgmt cluster upgrading fails", func() {
		BeforeEach(func() {
			bomClient.GetDefaultTkrBOMConfigurationReturns(&tkrBom, nil)
			bomClient.GetDefaultTkgBOMConfigurationReturns(&tkgconfigbom.BOMConfiguration{Release: &tkgconfigbom.ReleaseInfo{Version: "v1.3.1"}}, nil)
			tkgClient.UpgradeManagementClusterReturns(errors.New("failed to upgrade management cluster"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When mgmt cluster upgrading succeeds", func() {
		BeforeEach(func() {
			bomClient.GetDefaultTkrBOMConfigurationReturns(&tkrBom, nil)
			bomClient.GetDefaultTkgBOMConfigurationReturns(&tkgconfigbom.BOMConfiguration{Release: &tkgconfigbom.ReleaseInfo{Version: "v1.3.1"}}, nil)
			tkgClient.UpgradeManagementClusterReturns(nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	AfterEach(func() {
		os.Remove(configDir)
	})
})

var _ = Describe("Unit tests for upgrade cluster", func() {
	var (
		ctl       *tkgctl
		tkgClient = &fakes.Client{}
		bomClient = &fakes.TKGConfigBomClient{}
		configDir string
		err       error
		ops       UpgradeClusterOptions
	)
	JustBeforeEach(func() {
		configDir, err = os.MkdirTemp("", "test")
		err = os.MkdirAll(configDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
		prepareConfiDir(configDir)
		options := Options{
			ConfigDir:      configDir,
			ProviderGetter: fakeproviders.FakeProviderGetter(),
		}
		c, createErr := New(options)
		Expect(createErr).ToNot(HaveOccurred())
		ctl, _ = c.(*tkgctl)
		ctl.tkgClient = tkgClient
		ctl.tkgBomClient = bomClient
		ops = UpgradeClusterOptions{
			ClusterName:         "my-mgmt-cluster",
			VSphereTemplateName: "ubuntu-ova",
			OSName:              "ubuntu",
			OSVersion:           "xxxx",
			OSArch:              "amd64",
			SkipPrompt:          true,
			TkrVersion:          "v1.21.1+vmware.1.tkg.1",
			Timeout:             time.Minute * 30,
		}
		err = ctl.UpgradeCluster(ops)
	})

	Context("it cannot determine if this is a pacific cluster", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(true, errors.New("failed to determine if this is a pacific cluster"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("upgrading a Pacific cluster succeeds", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(true, nil)
			tkgClient.UpgradeClusterReturns(nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("upgrade a tkgm cluster succeeds", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(false, nil)
			tkgClient.UpgradeClusterReturns(nil)
			bomClient.GetBOMConfigurationFromTkrVersionReturns(nil, nil)
			bomClient.GetK8sVersionFromTkrVersionReturns("1.19.0", nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	AfterEach(func() {
		os.Remove(configDir)
	})
})
