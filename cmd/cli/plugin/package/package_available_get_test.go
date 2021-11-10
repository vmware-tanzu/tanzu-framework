// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

// This is a dummy package schema copied from external-dns
var rawValidDummyPackageSchema = `
additionalProperties: false
properties:
  accelerator:
    additionalProperties: true
    description: Accelerator configuration
    properties: {}
    type: object
  api_portal:
    additionalProperties: true
    description: API Portal configuration
    properties: {}
    type: object
  appliveview:
    additionalProperties: true
    description: App Live View configuration
    properties: {}
    type: object
  buildservice:
    description: Build Service configuration
    properties:
      ca_cert_data:
        description: tbs registry ca certificate (used for self signed registry)
        type: string
      kp_default_repository:
        description: docker repository (required)
        examples:
        - registry.io/build-service
        type: string
      kp_default_repository_password:
        description: registry password (required)
        examples:
        - password
        type: string
      kp_default_repository_username:
        description: registry username (required)
        examples:
        - janedoe@vmware.com
        type: string
      tanzunet_password:
        description: tanzunet registry password (required for dependency updater
          feature)
        examples:
        - password
        type: string
      tanzunet_username:
        description: tanzunet registry username (required for dependency updater
          feature)
        examples:
        - janedoe@vmware.com
        type: string
    type: object
  cnrs:
    additionalProperties: true
    description: Cloud Native Runtimes configuration
    properties: {}
    type: object
  image_policy_webhook:
    additionalProperties: true
    description: Image Policy Webhook configuration
    properties: {}
    type: object
  install_cert_manager:
    default: true
    description: Install cert-manager
    type: boolean
  install_flux:
    default: true
    description: Install FluxCD source-controller
    type: boolean
  install_tekton:
    default: true
    description: Install Tekton
    type: boolean
  learningcenter:
    additionalProperties: true
    description: Learning Center configuration
    properties: {}
    type: object
  ootb_supply_chain_basic:
    additionalProperties: true
    description: OOTB Supply Chain Basic configuration
    properties: {}
    type: object
  ootb_supply_chain_testing:
    additionalProperties: true
    description: OOTB Supply Chain Testing configuration
    properties: {}
    type: object
  ootb_supply_chain_testing_scanning:
    additionalProperties: true
    description: OOTB Supply Chain Testing Scanning configuration
    type: object
    properties:
      dummyProperty-1:
        default: dummyDefault
        description: this is for testing purposes
        type: string
      dummyProperty-2:
        description: this is for testing purposes
        type: object
        properties: {}
  profile:
    default: full
    description: 'Profile to install. Valid values: full, dev-light'
    type: string
  tap_gui:
    additionalProperties: true
    description: TAP GUI configuration
    type: object
type: object
`

var _ = Describe("PackageAvailableGet", func() {
	var (
		kappClient   *fakes.KappClient
		pkgNamespace = "default"
		pkgName      = "external-dns.tanzu.vmware.com"
		pkgVersion   = "1.2.1+vmware.1-tkg.2-zshippable"
	)

	Context("getValuesSchemaForPackage()", func() {
		BeforeEach(func() {
			kappClient = &fakes.KappClient{}
		})

		It("should return error if version is not specified", func() {
			var buf bytes.Buffer
			err := getValuesSchemaForPackage(pkgNamespace, pkgName, "", kappClient, &buf)
			Expect(err).To(HaveOccurred())
		})

		It("should return error if package does not exist", func() {
			kappClient.GetPackageReturns(nil,
				apierrors.NewNotFound(kapppkg.Resource("package"), "dummyPackage"))

			var buf bytes.Buffer
			err := getValuesSchemaForPackage(pkgNamespace, pkgName, pkgVersion, kappClient, &buf)
			Expect(err).To(HaveOccurred())
		})

		It("should not return error if schema is empty", func() {
			kappClient.GetPackageReturns(&kapppkg.Package{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pkgName,
					Namespace: pkgNamespace,
				},
				Spec: kapppkg.PackageSpec{
					RefName: pkgName,
					Version: pkgVersion,
					ValuesSchema: kapppkg.ValuesSchema{
						OpenAPIv3: runtime.RawExtension{
							Raw: []byte{},
						},
					},
				},
			}, nil)

			var buf bytes.Buffer
			err := getValuesSchemaForPackage(pkgNamespace, pkgName, pkgVersion, kappClient, &buf)
			Expect(err).To(BeNil())
		})

		It("should return the values schema without any errors", func() {
			kappClient.GetPackageReturns(&kapppkg.Package{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pkgName,
					Namespace: pkgNamespace,
				},
				Spec: kapppkg.PackageSpec{
					RefName: pkgName,
					Version: pkgVersion,
					ValuesSchema: kapppkg.ValuesSchema{
						OpenAPIv3: runtime.RawExtension{
							Raw: []byte(rawValidDummyPackageSchema),
						},
					},
				},
			}, nil)

			var buf bytes.Buffer
			err := getValuesSchemaForPackage(pkgNamespace, pkgName, pkgVersion, kappClient, &buf)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println(buf.String()) // Print out the output in terminal to help debugging if nothing printed out
			Expect(buf.String()).NotTo(BeEmpty())
		})

	})
})
