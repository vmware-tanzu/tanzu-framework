// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

var _ = Describe("Unit tests for GetTanzuKubernetesReleases", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		tkgClient             *TkgClient
		tkrInfo               *KubernetesVersionsInfo
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("When management cluster is vSphere with Kubernetes", func() {
		BeforeEach(func() {
			regionalClusterClient.IsPacificRegionalClusterReturns(true, nil)
		})
		Context("When vSphere with Kubernetes TKC API version is not supported", func() {
			JustBeforeEach(func() {
				regionalClusterClient.GetPacificTanzuKubernetesReleasesReturns(nil, errors.New("fake-error"))
				regionalClusterClient.GetPacificTKCAPIVersionReturns("fake-api-versions", nil)
				tkrInfo, err = tkgClient.DoGetTanzuKubernetesReleases(regionalClusterClient)
			})
			/*
				TODO: (chandrareddyp) we need revisit below API Version validation, for now we disabled it.
				It("should return error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Only %q Tanzu Kubernetes Cluster API version is supported by current version of Tanzu CLI", constants.DefaultPacificClusterAPIVersion)))
				})
			*/
		})
		Context("When vSphere with Kubernetes does not support tanzukuberenetesrelease objects", func() {
			JustBeforeEach(func() {
				regionalClusterClient.GetPacificTanzuKubernetesReleasesReturns(nil, errors.New("fake-error"))
				regionalClusterClient.GetPacificTKCAPIVersionReturns(constants.DefaultPacificClusterAPIVersion, nil)
				tkrInfo, err = tkgClient.DoGetTanzuKubernetesReleases(regionalClusterClient)
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get supported kubernetes release versions for vSphere with Kubernetes clusters"))
			})
		})

		Context("When vSphere with Kubernetes supports tanzukuberenetesrelease objects", func() {
			versions := []string{"v1.17.0", "v1.18.1"}
			JustBeforeEach(func() {
				regionalClusterClient.GetPacificTanzuKubernetesReleasesReturns(versions, nil)
				regionalClusterClient.GetPacificTKCAPIVersionReturns(constants.DefaultPacificClusterAPIVersion, nil)
				tkrInfo, err = tkgClient.DoGetTanzuKubernetesReleases(regionalClusterClient)
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			It("should correct kubernetes vesions", func() {
				Expect(tkrInfo.Versions).To(Equal(versions))
			})
		})
	})
})

func copyAllBoMFilesToTestingDir(listBoMFiles []string, configDir string) {
	bomDir, err := tkgconfigpaths.New(configDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	err = os.RemoveAll(bomDir)
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}
	for _, bomFile := range listBoMFiles {
		destFile := filepath.Join(bomDir, filepath.Base(bomFile))
		err = utils.CopyFile(bomFile, destFile)
		Expect(err).ToNot(HaveOccurred())
	}
}
