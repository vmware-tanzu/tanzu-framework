// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
)

var _ = Describe("Validate", func() {
	var (
		tkgClient       *TkgClient
		nodeSizeOptions NodeSizeOptions
		err             error
	)

	BeforeEach(func() {
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())

		nodeSizeOptions = NodeSizeOptions{
			Size:             "medium",
			ControlPlaneSize: "medium",
			WorkerSize:       "medium",
		}
	})

	Context("When vCenter IP and vSphere Control Plane Endpoint are different", func() {
		It("Should validate successfully", func() {
			vip := "10.10.10.11"
			err = tkgClient.ConfigureAndValidateVsphereConfig("", nodeSizeOptions, vip, true, nil)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When vCenter IP and vSphere Control Plane Endpoint are the same", func() {
		It("Should throw a validation error", func() {
			vip := "10.10.10.10"
			err = tkgClient.ConfigureAndValidateVsphereConfig("", nodeSizeOptions, vip, true, nil)
			Expect(err).To(HaveOccurred())
		})
	})
})
