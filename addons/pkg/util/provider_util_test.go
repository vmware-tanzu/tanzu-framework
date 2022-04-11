// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var validInput1 = `
apiVersion: cpi.tanzu.vmware.com/v1alpha1
kind: VSphereCPIConfig
metadata:
  name: vspherecpiconfig-sample
spec:
  vsphereCPI:
    mode: vsphereCPI
    tlsThumbprint: "test"
    server: 10.1.1.1
    datacenter: ds-test
    vSphereCredentialLocalObjRef:
      kind: Secret
      name: vsphere-credentials
    insecureFlag: False
    vmInternalNetwork: null
    vmExternalNetwork: null
    cloudProviderExtraArgs:
      tlsCipherSuites: TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
    nsxt:
      podRoutingEnabled: false
      routes:
        routerPath: ""
      credentialLocalObjRef:
        apiGroup: ""
        kind: Secret
        name: nsxt-credentials
      nestedForTesting:
        dummyLocalObjRef:
          apiGroup: ""
          kind: "ConfigMap"
          name: "dummy-configmap"
      host: "10.1.1.2"
      insecureFlag: false
      remoteAuth: false
      vmcAccessToken: ""
      vmcAuthHost: ""
      clientCertKeyData: ""
      clientCertData: ""
      rootCAData: ""
      secretName: "cloud-provider-vsphere-nsxt-credentials"
      secretNamespace: "kube-system"
`

var invalidInput1 = `
vsphereCPI:
  mode: vsphereCPI
  tlsThumbprint: "test"
  server: 10.1.1.1
  datacenter: ds-test
  vSphereCredentialLocalObjRef:
    foo:  bar
`

var invalidInput2 = `
vsphereCPI:
  mode: vsphereCPI
  tlsThumbprint: "test"
  server: 10.1.1.1
  datacenter: ds-test
`

var _ = Describe("Verifying the functionality of provider_util", func() {

	var testData map[string]interface{}
	JustBeforeEach(func() {
		testData = make(map[string]interface{})
	})

	Context("when the unstructuredContent passed into ExtractTypedLocalObjectRef() has valid embedded local object reference", func() {

		It("should return non-empty result", func() {
			Expect(yaml.Unmarshal([]byte(validInput1), testData)).Should(Succeed())
			result := util.ExtractTypedLocalObjectRef(testData, "LocalObjRef")
			Expect(len(result)).To(Equal(2))
			Expect(len(result[schema.GroupKind{Group: "", Kind: "Secret"}])).To(Equal(2))
			Expect(len(result[schema.GroupKind{Group: "", Kind: "ConfigMap"}])).To(Equal(1))
		})
	})

	Context("when the unstructuredContent passed into ExtractTypedLocalObjectRef() has invalid embedded local object reference", func() {
		It("should return empty result if the content of local object reference is invalid", func() {
			Expect(yaml.Unmarshal([]byte(invalidInput1), testData)).Should(Succeed())
			result := util.ExtractTypedLocalObjectRef(testData, "LocalObjRef")
			Expect(len(result)).To(Equal(0))
		})
		It("should return empty result if no embedded local object reference", func() {
			Expect(yaml.Unmarshal([]byte(invalidInput2), testData)).Should(Succeed())
			result := util.ExtractTypedLocalObjectRef(testData, "LocalObjRef")
			Expect(len(result)).To(Equal(0))
		})
	})
})
