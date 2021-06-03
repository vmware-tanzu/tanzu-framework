// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
)

var _ = Describe("Unit tests for delete machine health check", func() {
	var (
		ctl       *tkgctl
		tkgClient = &fakes.Client{}
		err       error
		configDir string
	)

	JustBeforeEach(func() {
		configDir, err = os.MkdirTemp("", "test")
		err = os.MkdirAll(testingDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
		prepareConfiDir(configDir)
		options := Options{
			ConfigDir: configDir,
		}
		c, createErr := New(options)
		Expect(createErr).ToNot(HaveOccurred())
		ctl, _ = c.(*tkgctl)
		ctl.tkgClient = tkgClient
	})
	Context("When deleting existing mhc", func() {
		It("should not return an error", func() {
			tkgClient.DeleteMachineHealthCheckReturns(nil)
			ops := DeleteMachineHealthCheckOptions{
				SkipPrompt:  true,
				ClusterName: "my-cluster",
			}
			err = ctl.DeleteMachineHealthCheck(ops)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When deleting non-existing mhc", func() {
		It("should return an error", func() {
			tkgClient.DeleteMachineHealthCheckReturns(errors.New("not found"))
			ops := DeleteMachineHealthCheckOptions{
				SkipPrompt:  true,
				ClusterName: "my-cluster",
			}
			err = ctl.DeleteMachineHealthCheck(ops)
			Expect(err).To(HaveOccurred())
		})
	})
	Context("When getting mhc objects", func() {
		It("should not return an error", func() {
			tkgClient.GetMachineHealthChecksReturns(nil, nil)
			ops := GetMachineHealthCheckOptions{
				ClusterName: "my-cluster",
			}
			_, err = ctl.GetMachineHealthCheck(ops)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When setting mhc object", func() {
		It("should not return an error", func() {
			tkgClient.SetMachineHealthCheckReturns(nil)
			ops := SetMachineHealthCheckOptions{
				ClusterName:            "my-cluster",
				MachineHealthCheckName: "my-mhc",
				Namespace:              "my-namespace",
				MatchLabels:            "label1:value1,label2:value2",
			}
			err = ctl.SetMachineHealthCheck(ops)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	AfterEach(func() {
		os.Remove(configDir)
	})
})
