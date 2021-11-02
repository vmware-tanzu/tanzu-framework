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
title: external-dns.community.tanzu.vmware.com.0.8.0 values schema
examples:
  - namespace: tanzu-system-service-discovery
    deployment:
      args:
        - --source=service
        - --txt-owner-id=k8s
        - --domain-filter=k8s.example.org
        - --namespace=tanzu-system-service-discovery
        - --provider=rfc2136
        - --rfc2136-host=100.69.97.77
        - --rfc2136-port=53
        - --rfc2136-zone=k8s.example.org
        - --rfc2136-tsig-secret=MTlQs3NNU=
        - --rfc2136-tsig-secret-alg=hmac-sha256
        - --rfc2136-tsig-keyname=externaldns-key
        - --rfc2136-tsig-axfr
      env: []
      securityContext: {}
      volumeMounts: []
      volumes: []
properties:
  namespace:
    type: string
    description: The namespace in which to deploy ExternalDNS.
    default: external-dns
    examples:
      - external-dns
  deployment:
    type: object
    description: Deployment related configuration
    properties:
      args:
        type: array
        description: |
          List of arguments passed via command-line to external-dns.  For
          more guidance on configuration options for your desired DNS
          provider, consult the ExternalDNS docs at
          https://github.com/kubernetes-sigs/external-dns#running-externaldns
        items:
          type: string
      env:
        type: array
        description: "List of environment variables to set in the external-dns container."
        items:
          $ref: "#/definitions/io.k8s.api.core.v1.EnvVar"
      securityContext:
        description: "SecurityContext defines the security options the external-dns container should be run with. More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/"
        $ref: "#/definitions/io.k8s.api.core.v1.SecurityContext"
      volumeMounts:
        type: array
        description: "Pod volumes to mount into the external-dns container's filesystem."
        items:
          $ref: "#/definitions/io.k8s.api.core.v1.VolumeMount"
      volumes:
        type: array
        description: "List of volumes that can be mounted by containers belonging to the external-dns pod. More info: https://kubernetes.io/docs/concepts/storage/volumes"
        items:
          $ref: "#/definitions/io.k8s.api.core.v1.Volume"
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

		It("should return error if schema is empty", func() {
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
			Expect(err).To(HaveOccurred())
		})

		It("should return error if schema is empty", func() {
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
