// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package imgpkgutil

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.FailFast = true
	RunSpecs(t, "Fetcher Unit Tests", suiteConfig)
}

const imagesLockYAML = `
---
apiVersion: imgpkg.carvel.dev/v1alpha1
images:
- annotations:
    kbld.carvel.dev/id: projects-stg.registry.vmware.com/tkg/tkr-vsphere-nonparavirt:v1.24.9_vmware.1-tkg.1-zshippable
    kbld.carvel.dev/origins: |
      - resolved:
          tag: v1.24.9_vmware.1-tkg.1-zshippable
          url: projects-stg.registry.vmware.com/tkg/tkr-vsphere-nonparavirt:v1.24.9_vmware.1-tkg.1-zshippable
  image: projects-stg.registry.vmware.com/tkg/tkr-vsphere-nonparavirt@sha256:b56a4c11a3eef1d3fef51c66b1571f92d45f17cf11de8f89e7706dbbb9b6a287
kind: ImagesLock
`

var _ = Describe("parseImagesLock", func() {
	var (
		imagesLockBytes  []byte
		expectedImageMap map[string]string
		bundleImage      string
	)
	BeforeEach(func() {
		imagesLockBytes = nil
		expectedImageMap = nil
		bundleImage = ""
	})
	Context("valid input", func() {
		BeforeEach(func() {
			imagesLockBytes = []byte(imagesLockYAML)
			bundleImage = "10.92.174.209:8443/library/tkg/tkr-repository-vsphere-nonparavirt:v1.24.9_vmware.1-tkg.1-zshippable"
			expectedImageMap = map[string]string{
				"projects-stg.registry.vmware.com/tkg/tkr-vsphere-nonparavirt:v1.24.9_vmware.1-tkg.1-zshippable": "10.92.174.209:8443/library/tkg/tkr-repository-vsphere-nonparavirt@sha256:b56a4c11a3eef1d3fef51c66b1571f92d45f17cf11de8f89e7706dbbb9b6a287",
			}
		})
		It("should return the expected map of images", func() {
			imageMap, err := ParseImagesLock(bundleImage, imagesLockBytes)
			Expect(imageMap).To(Equal(expectedImageMap))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("nil input", func() {
		BeforeEach(func() {
			imagesLockBytes = nil
		})
		It("should return nil", func() {
			Expect(ParseImagesLock(bundleImage, imagesLockBytes)).To(BeNil())
		})
	})

	Context("invalid yaml", func() {
		BeforeEach(func() {
			imagesLockBytes = []byte(`%%%%`)
		})
		It("should return nil", func() {
			imageMap, err := ParseImagesLock(bundleImage, imagesLockBytes)
			Expect(imageMap).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})
})

const origFile = `kind: Package
apiVersion: data.packaging.carvel.dev/v1alpha1
metadata:
  name: tkr-vsphere-nonparavirt.tanzu.vmware.com.1.24.9+vmware.1-tkg.1-zshippable
  labels:
    run.tanzu.vmware.com/tkr-package: ""
spec:
  refName: tkr-vsphere-nonparavirt.tanzu.vmware.com
  version: 1.24.9+vmware.1-tkg.1-zshippable
  licenses:
  - 'VMware’s End User License Agreement (Underlying OSS license: Apache License 2.0)'
  releasedAt: "2023-01-07T14:52:02Z"
  releaseNotes: tkr release
  template:
    spec:
      fetch:
      - imgpkgBundle:
          image: projects-stg.registry.vmware.com/tkg/tkr-vsphere-nonparavirt:v1.24.9_vmware.1-tkg.1-zshippable
      template:
      - ytt:
          ignoreUnknownComments: true
          paths:
          - config/
          - packages/
      - kbld:
          paths:
          - '-'
          - .imgpkg/images.yml
      deploy:
      - kapp:
          intoNs: tkg-system
`

const resolvedFile = `kind: Package
apiVersion: data.packaging.carvel.dev/v1alpha1
metadata:
  name: tkr-vsphere-nonparavirt.tanzu.vmware.com.1.24.9+vmware.1-tkg.1-zshippable
  labels:
    run.tanzu.vmware.com/tkr-package: ""
spec:
  refName: tkr-vsphere-nonparavirt.tanzu.vmware.com
  version: 1.24.9+vmware.1-tkg.1-zshippable
  licenses:
  - 'VMware’s End User License Agreement (Underlying OSS license: Apache License 2.0)'
  releasedAt: "2023-01-07T14:52:02Z"
  releaseNotes: tkr release
  template:
    spec:
      fetch:
      - imgpkgBundle:
          image: 10.92.174.209:8443/library/tkg/tkr-repository-vsphere-nonparavirt@sha256:b56a4c11a3eef1d3fef51c66b1571f92d45f17cf11de8f89e7706dbbb9b6a287
      template:
      - ytt:
          ignoreUnknownComments: true
          paths:
          - config/
          - packages/
      - kbld:
          paths:
          - '-'
          - .imgpkg/images.yml
      deploy:
      - kapp:
          intoNs: tkg-system
`

var _ = Describe("resolveImages", func() {
	var (
		imageMap   map[string]string
		bundle     map[string][]byte
		wantBundle map[string][]byte
	)
	BeforeEach(func() {
		imageMap = map[string]string{
			"projects-stg.registry.vmware.com/tkg/tkr-vsphere-nonparavirt:v1.24.9_vmware.1-tkg.1-zshippable": "10.92.174.209:8443/library/tkg/tkr-repository-vsphere-nonparavirt@sha256:b56a4c11a3eef1d3fef51c66b1571f92d45f17cf11de8f89e7706dbbb9b6a287",
		}
		bundle = map[string][]byte{
			"path1": []byte(origFile),
			"path2": []byte("image2"),
		}
		wantBundle = map[string][]byte{
			"path1": []byte(resolvedFile),
			"path2": []byte("image2"),
		}
	})

	Context("matches", func() {
		It("should replace the original images with target images in the bundle", func() {
			ResolveImages(imageMap, bundle)
			Expect(bundle).To(Equal(wantBundle))
		})
	})

	Context("no matches", func() {
		BeforeEach(func() {
			imageMap = map[string]string{
				"image3": "image3:v1",
			}
			wantBundle = map[string][]byte{
				"path1": []byte(origFile),
				"path2": []byte("image2"),
			}
		})
		It("should not change the bundle", func() {
			ResolveImages(imageMap, bundle)
			Expect(bundle).To(Equal(wantBundle))
		})
	})
})
