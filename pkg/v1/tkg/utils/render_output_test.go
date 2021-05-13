// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	originalStdOut *os.File
	TempOutputFile string
)

var jsonRenderType = "json"

var _ = Describe("Render Output Tests", func() {
	var (
		data       interface{}
		renderType string
		bytes      []byte
		err        error
		file       *os.File
	)

	spaceReplacer := strings.NewReplacer(" ", "", "\n", "", "\t", "", "\r", "")

	BeforeEach((func() {
		originalStdOut = os.Stdout
		TempOutputFile = filepath.Join(os.TempDir(), "test-output")
		file, err = os.OpenFile(TempOutputFile, os.O_RDWR|os.O_CREATE, 0o644)
		Expect(err).NotTo(HaveOccurred())
		os.Stdout = file
	}))

	AfterEach((func() {
		os.Stdout = originalStdOut
		Expect(file.Close()).NotTo(HaveOccurred())
		Expect(os.Remove(TempOutputFile)).NotTo(HaveOccurred())
	}))

	JustBeforeEach(func() {
		err = RenderOutput(data, renderType)
	})

	Context("When output is array and array is not empty", func() {
		BeforeEach(func() {
			listObjects := []testStruct{}
			listObjects = append(listObjects, testStruct{Name: "fake-cluster-name", Namespace: "fake-namespace"})
			data = listObjects
			renderType = jsonRenderType
		})
		It("returns an error", func() {
			Expect(err).NotTo(HaveOccurred())
			bytes, err = ioutil.ReadFile(TempOutputFile)
			Expect(err).NotTo(HaveOccurred())
			output := spaceReplacer.Replace(string(bytes))
			Expect(output).To(Equal(`[{"name":"fake-cluster-name","namespace":"fake-namespace"}]`))
		})
	})

	Context("When output is array and array is empty", func() {
		BeforeEach(func() {
			data = []testStruct{}
			renderType = jsonRenderType
		})
		It("returns an error", func() {
			Expect(err).NotTo(HaveOccurred())
			bytes, err = ioutil.ReadFile(TempOutputFile)
			Expect(err).NotTo(HaveOccurred())
			output := spaceReplacer.Replace(string(bytes))
			Expect(output).To(Equal(`[]`))
		})
	})

	Context("When output is object and object is not empty", func() {
		BeforeEach(func() {
			data = testStruct{Name: "fake-cluster-name", Namespace: "fake-namespace"}
			renderType = jsonRenderType
		})
		It("returns an error", func() {
			Expect(err).NotTo(HaveOccurred())
			bytes, err = ioutil.ReadFile(TempOutputFile)
			Expect(err).NotTo(HaveOccurred())
			output := spaceReplacer.Replace(string(bytes))
			Expect(output).To(Equal(`{"name":"fake-cluster-name","namespace":"fake-namespace"}`))
		})
	})

	Context("When output is object and object is empty", func() {
		BeforeEach(func() {
			data = testStruct{}
			renderType = jsonRenderType
		})
		It("returns an error", func() {
			Expect(err).NotTo(HaveOccurred())
			bytes, err = ioutil.ReadFile(TempOutputFile)
			Expect(err).NotTo(HaveOccurred())
			output := spaceReplacer.Replace(string(bytes))
			Expect(output).To(Equal(`{}`))
		})
	})
})

type testStruct struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}
