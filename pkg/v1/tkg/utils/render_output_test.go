/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
