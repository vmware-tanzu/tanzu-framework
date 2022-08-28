// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/testdata"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

const (
	k8s1_20_1 = "v1.20.1+vmware.1"
	k8s1_20_2 = "v1.20.2+vmware.1"
	k8s1_21_1 = "v1.21.1+vmware.1"
	k8s1_21_3 = "v1.21.3+vmware.1"
	k8s1_22_0 = "v1.22.0+vmware.1"
)

var k8sVersions = []string{k8s1_20_1, k8s1_20_2, k8s1_21_1, k8s1_21_3, k8s1_22_0}

var _ = Describe("os get", func() {
	var (
		tkrName       string
		err           error
		clusterClient *fakes.ClusterClient
		tkr           *runv1alpha3.TanzuKubernetesRelease
		osImages      data.OSImages
		tkrs          data.TKRs
		cmdOptions    *getOSOptions
		gotOSInfoMap  map[string]runv1alpha3.OSInfo
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		tkrName = ""
		cmdOptions = &getOSOptions{}
		osImages = testdata.GenOSImages(k8sVersions, 3)
		tkrs = testdata.GenTKRs(2, testdata.SortOSImagesByK8sVersion(osImages))
		tkr = testdata.ChooseTKR(tkrs)
	})

	JustBeforeEach(func() {
		gotOSInfoMap, err = osInfoByTKR(clusterClient, tkrName, cmdOptions)
	})

	Context("When the get OSImages return error", func() {
		BeforeEach(func() {
			clusterClient.ListResourcesReturns(errors.New("fake get OSImages error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake get OSImages error"))
			Expect(gotOSInfoMap).To(BeNil())
		})
	})
	Context("When the get TKR return error", func() {
		BeforeEach(func() {
			clusterClient.ListResourcesCalls(func(o interface{}, option ...client.ListOption) error {
				imgl := o.(*runv1alpha3.OSImageList)
				imgl.Items = append(imgl.Items, getOSImagesList(osImages)...)
				return nil
			})
			clusterClient.GetResourceReturns(errors.New("fake get TKR error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake get TKR error"))
			Expect(gotOSInfoMap).To(BeNil())
		})
	})
	Context("When the OSImages and TKR are available", func() {
		BeforeEach(func() {
			clusterClient.ListResourcesCalls(func(o interface{}, option ...client.ListOption) error {
				osImagelist := o.(*runv1alpha3.OSImageList)
				osImagelist.Items = append(osImagelist.Items, getOSImagesList(osImages)...)
				return nil
			})
			clusterClient.GetResourceCalls(func(o interface{}, tkrName, ns string, pv clusterclient.PostVerifyrFunc, poll *clusterclient.PollOptions) error {
				*o.(*runv1alpha3.TanzuKubernetesRelease) = *tkr
				return nil
			})

		})
		It("should not return error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(gotOSInfoMap).ToNot(BeNil())
			Expect(len(gotOSInfoMap)).ToNot(BeZero())
			Expect(validateOSInfoResults(gotOSInfoMap, tkr, osImages, "")).To(BeNil())
		})
	})
	Context("When the OSImages and TKR are available and get cluster infrastructure returns error", func() {
		BeforeEach(func() {
			clusterClient.ListResourcesCalls(func(o interface{}, option ...client.ListOption) error {
				osImagelist := o.(*runv1alpha3.OSImageList)
				osImagelist.Items = append(osImagelist.Items, getOSImagesList(osImages)...)
				return nil
			})
			clusterClient.GetResourceCalls(func(o interface{}, tkrName, ns string, pv clusterclient.PostVerifyrFunc, poll *clusterclient.PollOptions) error {
				*o.(*runv1alpha3.TanzuKubernetesRelease) = *tkr
				return nil
			})
			clusterClient.GetClusterInfrastructureReturns("", errors.New("fake getclusterInfrastructure error"))

		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake getclusterInfrastructure error"))
			Expect(gotOSInfoMap).To(BeNil())
		})
	})
	Context("When user provides region and the OSImages and TKR are available and get cluster infrastructure returns success", func() {
		BeforeEach(func() {
			clusterClient.ListResourcesCalls(func(o interface{}, option ...client.ListOption) error {
				osImagelist := o.(*runv1alpha3.OSImageList)
				osImagelist.Items = append(osImagelist.Items, getOSImagesList(osImages)...)
				return nil
			})
			clusterClient.GetResourceCalls(func(o interface{}, tkrName, ns string, pv clusterclient.PostVerifyrFunc, poll *clusterclient.PollOptions) error {
				*o.(*runv1alpha3.TanzuKubernetesRelease) = *tkr
				return nil
			})
			cmdOptions = &getOSOptions{region: chooseRegionFromTKR(tkr, osImages)}
			clusterClient.GetClusterInfrastructureReturns(constants.InfrastructureRefAWS, nil)

		})
		It("should not return error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(gotOSInfoMap).ToNot(BeNil())
			Expect(validateOSInfoResults(gotOSInfoMap, tkr, osImages, cmdOptions.region)).To(BeNil())
		})
	})
})

func getOSImagesList(osImages data.OSImages) []runv1alpha3.OSImage {
	var result []runv1alpha3.OSImage
	for _, image := range osImages {
		result = append(result, *image)
	}
	return result
}

func chooseRegionFromTKR(tkr *runv1alpha3.TanzuKubernetesRelease, osImages data.OSImages) string {
	osImageName := tkr.Spec.OSImages[0].Name
	return osImages[osImageName].Spec.Image.Ref["region"].(string)
}

func validateOSInfoResults(gotOSInfoMap map[string]runv1alpha3.OSInfo,
	tkr *runv1alpha3.TanzuKubernetesRelease, osImages data.OSImages, region string) error {

	osImageNamesInTKR := []string{}
	for _, o := range tkr.Spec.OSImages {
		osImageNamesInTKR = append(osImageNamesInTKR, o.Name)
	}
	for _, name := range osImageNamesInTKR {
		osImage, exists := osImages[name]
		if !exists {
			return errors.Errorf("unable to find the TKR's osImage '%s' in the given osImages", name)
		}
		if region != "" && osImage.Spec.Image.Ref["region"] != region {
			continue
		}
		osInfo := osImage.Spec.OS
		osInfoInString := OsInfoString(osInfo)
		oi, exists := gotOSInfoMap[osInfoInString]
		if !exists {
			return errors.Errorf("unable to find the osInfo of OsImage '%s'", name)
		}
		if oi != osInfo {
			return errors.Errorf("osInfo of OsImage '%s' doesn't match with osInfo returned ", name)
		}
	}

	return nil
}
