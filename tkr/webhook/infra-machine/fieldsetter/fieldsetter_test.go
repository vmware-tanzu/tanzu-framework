// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fieldsetter

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	jsonpatch "gomodules.xyz/jsonpatch/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Infra Machine webhook test")
}

var _ = Describe("Infra Machine webhook", func() {
	var (
		fieldPathMap           map[string]string
		infraMachineObj        *unstructured.Unstructured
		infraMachineAnnotation map[string]string

		fs  *FieldSetter
		err error
	)
	BeforeEach(func() {
		fs = &FieldSetter{
			Log:          ctrllog.Log,
			FieldPathMap: map[string]string{},
		}
		decoder, err := admission.NewDecoder(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())
		Expect(decoder).NotTo(BeNil())
		err = fs.InjectDecoder(decoder)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("fieldSetter Handle tests", func() {
		var req admission.Request
		var resp admission.Response
		BeforeEach(func() {
			infraMachineObj = &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"image": map[string]interface{}{
							"id": "image-id-to-be-replaced",
						},
					},
				},
			}
			Expect(err).ToNot(HaveOccurred())
		})
		JustBeforeEach(func() {
			infraMachineObj.SetAnnotations(infraMachineAnnotation)
			req.Object.Raw, err = json.Marshal(infraMachineObj)

			resp = fs.Handle(context.Background(), req)
		})
		Context("when the Infra machine annotations is not set", func() {
			BeforeEach(func() {
				infraMachineAnnotation = nil
			})
			It("should admit the request and should not set any fields(zero patches)", func() {
				Expect(resp.Allowed).To(Equal(true))
				Expect(len(resp.Patches)).To(Equal(0))
				Expect(len(resp.Result.Reason)).To(Equal(0))
			})
		})
		Context("when the Infra machine annotations doesn't have os Image reference annotation", func() {
			BeforeEach(func() {
				infraMachineAnnotation = map[string]string{
					"fake-key": "fake-value",
				}
			})
			It("should admit the request and should not set any fields(zero patches)", func() {
				Expect(resp.Allowed).To(Equal(true))
				Expect(len(resp.Patches)).To(Equal(0))
				Expect(len(string(resp.Result.Reason))).To(Equal(0))
			})
		})
		Context("when os Image reference annotation is not a valid yaml", func() {
			BeforeEach(func() {
				infraMachineAnnotation = map[string]string{
					osImageRefAnnotationKey: "invalid_yaml",
				}
			})
			It("should not admit the request with the message set and should not set any fields(zero patches)", func() {
				Expect(resp.Allowed).To(Equal(false))
				Expect(len(resp.Patches)).To(Equal(0))
				Expect(resp.Result.Message).To(ContainSubstring("failed to unmarshal 'run.tanzu.vmware.com/os-image-ref' annotation"))
			})
		})
		Context("when the os image reference annotation value's field doesn't have corresponding entry in the fieldPathMap", func() {
			BeforeEach(func() {
				infraMachineAnnotation = map[string]string{
					osImageRefAnnotationKey: "id: image-value",
				}
				fieldPathMap = map[string]string{
					"invalidID": "spec.image.id",
				}
				fs.FieldPathMap = fieldPathMap
			})
			It("should admit the request and should not set any fields(zero patches)", func() {
				Expect(resp.Allowed).To(Equal(true))
				Expect(len(resp.Patches)).To(Equal(0))
			})
		})
		Context("when the fieldPathMap value has incorrect path", func() {
			BeforeEach(func() {
				infraMachineAnnotation = map[string]string{
					osImageRefAnnotationKey: "id: image-value",
				}
				fieldPathMap = map[string]string{
					"id": "wrong.path.id",
				}
				fs.FieldPathMap = fieldPathMap
			})
			It("should not admit the request with the message set and and should not set any fields(zero patches)", func() {
				Expect(resp.Allowed).To(Equal(false))
				Expect(len(resp.Patches)).To(Equal(0))
				Expect(resp.Result.Message).To(ContainSubstring(`field path "wrong.path.id" doesn't exists in the request object`))
			})
		})
		Context("when os image reference annotation value's field matches the fieldPathMap's field and has correct path", func() {
			BeforeEach(func() {
				infraMachineAnnotation = map[string]string{
					osImageRefAnnotationKey: "id: image-value-updated",
				}
				fieldPathMap = map[string]string{
					"id": "spec.image.id",
				}
				fs.FieldPathMap = fieldPathMap
			})
			It("should admit the request with patch to update the value ", func() {
				Expect(resp.Allowed).To(Equal(true))
				Expect(len(resp.Patches)).To(Equal(1))
				Expect(resp.Patches[0]).To(Equal(jsonpatch.Operation{
					Operation: "replace",
					Path:      "/spec/image/id",
					Value:     "image-value-updated"}))
			})
		})
	})
})
