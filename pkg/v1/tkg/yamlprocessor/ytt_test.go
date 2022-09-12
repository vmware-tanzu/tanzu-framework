// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package yamlprocessor_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/apis/providers/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/yamlprocessor"
)

var _ = Describe("YttProcessor", func() {
	Context("NewYttProcessor", func() {
		It("sets a default definition parser", func() {
			yp := yamlprocessor.NewYttProcessor()

			Expect(func() {
				_, _ = yp.GetVariables([]byte("doesn't matter"))
			}).ToNot(Panic())
		})
	})

	Context("GetTemplateName", func() {
		It("returns the name of the template definition file", func() {
			yp := yamlprocessor.NewYttProcessor()

			Expect(yp.GetTemplateName("version", "plan1")).To(Equal("cluster-template-definition-plan1.yaml"))
		})
	})

	Context("GetClusterClassTemplateName", func() {
		It("returns the name of the clusterclass definition file", func() {
			yp := yamlprocessor.NewYttProcessor()

			Expect(yp.GetClusterClassTemplateName("version", "foo")).To(Equal("clusterclass-foo.yaml"))
		})
	})

	Context("GetVariables", func() {
		It("returns the variables defined in the data yaml", func() {
			dir, err := os.MkdirTemp("", "tkg-cli")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(dir)
			templateDataFile := filepath.Join(dir, "test-template-data.yaml")
			Expect(os.WriteFile(templateDataFile, []byte(templateDataYaml), 0o600)).To(Succeed())

			path := v1alpha1.PathInfo{Path: dir}
			dp := &fakeDefinitionParser{
				parsePath: []v1alpha1.PathInfo{path},
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			variables, err := yp.GetVariables([]byte("doesn't matter"))

			Expect(err).ToNot(HaveOccurred())
			Expect(variables).To(ConsistOf("cluster_name", "count", "secretValue"))
		})

		It("returns the variables defined in all data yamls", func() {
			// See https://get-ytt.io/#example:example-multiple-data-values
			// for more information.
			dir, err := os.MkdirTemp("", "tkg-cli")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(dir)
			templateFile := filepath.Join(dir, "test-template.yaml")
			templateDataFile := filepath.Join(dir, "test-template-data.yaml")
			templateDataFile2 := filepath.Join(dir, "test-template-data2.yaml")

			Expect(os.WriteFile(templateFile, []byte(templateYaml), 0o600)).To(Succeed())
			Expect(os.WriteFile(templateDataFile, []byte(templateDataYaml), 0o600)).To(Succeed())
			Expect(os.WriteFile(templateDataFile2, []byte(templateDataYaml2), 0o600)).To(Succeed())

			path := v1alpha1.PathInfo{Path: dir}
			dp := &fakeDefinitionParser{
				parsePath: []v1alpha1.PathInfo{path},
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			variables, err := yp.GetVariables([]byte("doesn't matter"))
			Expect(err).ToNot(HaveOccurred())

			Expect(variables).To(ConsistOf("cluster_name", "count", "secretValue", "AWS_REGION"))
		})

		It("doesn't return any variables since no external variables are defined", func() {
			dir, err := os.MkdirTemp("", "tkg-cli")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(dir)
			templateFile := filepath.Join(dir, "test-template.yaml")
			Expect(os.WriteFile(templateFile, []byte(simpleYttYaml), 0o600)).To(Succeed())

			path := v1alpha1.PathInfo{Path: dir}
			dp := &fakeDefinitionParser{
				parsePath: []v1alpha1.PathInfo{path},
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			actual, err := yp.GetVariables([]byte("doesn't matter"))

			Expect(err).ToNot(HaveOccurred())
			// Technically there are no external ytt data values.
			// That is no data.values.foo
			Expect(actual).To(BeEmpty())
		})

		It("returns error if unable to parse template definition", func() {
			dp := &fakeDefinitionParser{
				parsePathErr: errors.New("some err"),
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			actual, err := yp.GetVariables([]byte("doesn't matter"))

			Expect(err).To(HaveOccurred())
			Expect(actual).To(BeEmpty())
		})

		It("returns error if unable to retrieve ytt files", func() {
			path := v1alpha1.PathInfo{Path: "some-path-that-doesn't-exist"}
			dp := &fakeDefinitionParser{
				parsePath: []v1alpha1.PathInfo{path},
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			actual, err := yp.GetVariables([]byte("doesn't matter"))

			Expect(err).To(HaveOccurred())
			Expect(actual).To(BeEmpty())
		})

		It("returns error if data values is malformed", func() {
			dir, err := os.MkdirTemp("", "tkg-cli")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(dir)
			// missing document separator between ytt annotation and document
			badDataFile := `
#@data/values
cluster_name: default
count: 1`

			templateDataFile := filepath.Join(dir, "test-template-data.yaml")
			Expect(os.WriteFile(templateDataFile, []byte(badDataFile), 0o600)).To(Succeed())

			path := v1alpha1.PathInfo{Path: dir}
			dp := &fakeDefinitionParser{
				parsePath: []v1alpha1.PathInfo{path},
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			_, err = yp.GetVariables([]byte("doesn't matter"))

			Expect(err).To(HaveOccurred())
		})
	})

	Context("Process", func() {
		const (
			defaultClusterDef = `name: default-cluster
replicas: 1
region: us-west-1a
`
		)
		It("returns the final processed yaml with multiple data values files and override values", func() {
			// See https://get-ytt.io/#example:example-multiple-data-values
			// for more information.
			dir, err := os.MkdirTemp("", "tkg-cli")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(dir)
			templateFile := filepath.Join(dir, "test-template.yaml")
			templateDataFile := filepath.Join(dir, "test-template-data.yaml")
			templateDataFile2 := filepath.Join(dir, "test-template-data2.yaml")

			Expect(os.WriteFile(templateFile, []byte(templateYaml), 0o600)).To(Succeed())
			Expect(os.WriteFile(templateDataFile, []byte(templateDataYaml), 0o600)).To(Succeed())
			Expect(os.WriteFile(templateDataFile2, []byte(templateDataYaml2), 0o600)).To(Succeed())

			configClient := NewFakeVariableClient().WithVar("cluster_name", "foo").WithVar("count", "22").WithVar("AWS_REGION", "us-east-1a")
			path := v1alpha1.PathInfo{Path: dir}
			dp := &fakeDefinitionParser{
				parsePath: []v1alpha1.PathInfo{path},
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			actual, err := yp.Process([]byte("doesn't matter"), configClient.Get)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).ToNot(BeNil())
			expectedYaml := `name: foo-cluster
replicas: 22
region: us-east-1a
`
			Expect(string(actual)).To(Equal(expectedYaml))
		})

		It("returns the final processed yaml with default values", func() {
			dir, err := os.MkdirTemp("", "tkg-cli")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(dir)
			templateFile := filepath.Join(dir, "test-template.yaml")
			templateDataFile := filepath.Join(dir, "test-template-data.yaml")
			templateDataFile2 := filepath.Join(dir, "test-template-data2.yaml")

			Expect(os.WriteFile(templateFile, []byte(templateYaml), 0o600)).To(Succeed())
			Expect(os.WriteFile(templateDataFile, []byte(templateDataYaml), 0o600)).To(Succeed())
			Expect(os.WriteFile(templateDataFile2, []byte(templateDataYaml2), 0o600)).To(Succeed())

			configClient := NewFakeVariableClient()
			path := v1alpha1.PathInfo{Path: dir}
			dp := &fakeDefinitionParser{
				parsePath: []v1alpha1.PathInfo{path},
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			actual, err := yp.Process([]byte("doesn't matter"), configClient.Get)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).ToNot(BeNil())
			expectedYaml := defaultClusterDef
			Expect(string(actual)).To(Equal(expectedYaml))
		})

		It("is able to processed ytt files that are nested in subdirectories", func() {
			dir, err := os.MkdirTemp("", "tkg-cli")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(dir)
			// This nested dir starts with "z" because of ytt file ordering.
			// The second data values file is nested in this directory but ytt
			// reads the filename along with the directory name. So in this
			// case the order of the files would be ./test-template-data.yaml
			// and then ./znested-dir<random>/test-template-data2.yaml. If the
			// "z" was omitted it would've been reversed.
			nestedDir, err := os.MkdirTemp(dir, "znested-dir")
			Expect(err).NotTo(HaveOccurred())
			// keep template file on top level dir
			templateFile := filepath.Join(dir, "test-template.yaml")
			templateDataFile := filepath.Join(dir, "test-template-data.yaml")
			templateDataFile2 := filepath.Join(nestedDir, "test-template-data2.yaml")
			Expect(os.WriteFile(templateFile, []byte(templateYaml), 0o600)).To(Succeed())
			Expect(os.WriteFile(templateDataFile, []byte(templateDataYaml), 0o600)).To(Succeed())
			Expect(os.WriteFile(templateDataFile2, []byte(templateDataYaml2), 0o600)).To(Succeed())

			configClient := NewFakeVariableClient()
			path := v1alpha1.PathInfo{Path: dir}
			dp := &fakeDefinitionParser{
				parsePath: []v1alpha1.PathInfo{path},
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			actual, err := yp.Process([]byte("doesn't matter"), configClient.Get)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).ToNot(BeNil())
			expectedYaml := defaultClusterDef
			Expect(string(actual)).To(Equal(expectedYaml))
		})

		It("returns error if unable to parse template definition", func() {
			dp := &fakeDefinitionParser{
				parsePathErr: errors.New("some err"),
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			actual, err := yp.Process([]byte("doesn't matter"), NewFakeVariableClient().Get)

			Expect(err).To(HaveOccurred())
			Expect(actual).To(BeEmpty())
		})

		It("returns error if unable to retrieve ytt files", func() {
			path := v1alpha1.PathInfo{Path: "some-path-that-doesn't-exist"}
			dp := &fakeDefinitionParser{
				parsePath: []v1alpha1.PathInfo{path},
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			actual, err := yp.Process([]byte("doesn't matter"), NewFakeVariableClient().Get)

			Expect(err).To(HaveOccurred())
			Expect(actual).To(BeEmpty())
		})

		It("returns an error since assertion failed", func() {
			dir, err := os.MkdirTemp("", "tkg-cli")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(dir)
			templateFile := filepath.Join(dir, "test-template.yaml")
			Expect(os.WriteFile(templateFile, []byte(assertYttYaml), 0o600)).To(Succeed())

			configClient := NewFakeVariableClient()
			path := v1alpha1.PathInfo{Path: dir}
			dp := &fakeDefinitionParser{
				parsePath: []v1alpha1.PathInfo{path},
			}
			yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

			actual, err := yp.Process([]byte("doesn't matter"), configClient.Get)

			Expect(err).To(HaveOccurred())
			Expect(actual).To(BeNil())
		})

		Context("yaml scalar value tests", func() {
			var testTemplate = `
#@ load("@ytt:data", "data")
---
decimal: #@ data.values.decimalInt
octal: #@ data.values.octalInt
hexadecimal: #@ data.values.hexadecimalInt
boolean: #@ data.values.booleanVal
float: #@ data.values.floatVal
`
			var dataTemplate = `
#@data/values
---
decimalInt:
octalInt:
hexadecimalInt:
booleanVal:
floatVal:
`
			When("data values are yaml scalars", func() {
				It("Should not quote the value", func() {
					dir, err := os.MkdirTemp("", "tkg-cli")
					Expect(err).NotTo(HaveOccurred())
					defer os.RemoveAll(dir)
					templateFile := filepath.Join(dir, "test-template.yaml")
					dataFile := filepath.Join(dir, "data-template.yaml")

					Expect(os.WriteFile(templateFile, []byte(testTemplate), 0o600)).To(Succeed())
					Expect(os.WriteFile(dataFile, []byte(dataTemplate), 0o600)).To(Succeed())

					configClient := NewFakeVariableClient().
						WithVar("decimalInt", "9").
						WithVar("octalInt", "0o600").
						WithVar("hexadecimalInt", "0xF").
						WithVar("booleanVal", "true").
						WithVar("floatVal", "3.14")
					path := v1alpha1.PathInfo{Path: dir}
					dp := &fakeDefinitionParser{
						parsePath: []v1alpha1.PathInfo{path},
					}
					yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

					actual, err := yp.Process([]byte("doesn't matter"), configClient.Get)
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).ToNot(BeNil())
					expectedYaml := `decimal: 9
octal: 384
hexadecimal: 15
boolean: true
float: 3.14
`
					Expect(string(actual)).To(Equal(expectedYaml))
				})
			})
			When("data values are non scalars", func() {
				var testTemplate = `
#@ load("@ytt:data", "data")			
---
password1: #@ data.values.password1
password2: #@ data.values.password2
password3: #@ data.values.password3
password4: #@ data.values.password4
password5: #@ data.values.password5
multiline: #@ data.values.multiline
structured: #@ data.values.structured
`
				var dataTemplate = `
#@data/values
---
password1:
password2:
password3:
password4:
password5:
multiline:
structured:
`
				It("Should quote the value", func() {
					dir, err := os.MkdirTemp("", "tkg-cli")
					Expect(err).NotTo(HaveOccurred())
					defer os.RemoveAll(dir)
					templateFile := filepath.Join(dir, "test-template.yaml")
					dataFile := filepath.Join(dir, "data-template.yaml")

					Expect(os.WriteFile(templateFile, []byte(testTemplate), 0o600)).To(Succeed())
					Expect(os.WriteFile(dataFile, []byte(dataTemplate), 0o600)).To(Succeed())

					configClient := NewFakeVariableClient().
						WithVar("password1", "%&%VMware!").
						WithVar("password2", "!VMware!").
						WithVar("password3", "#VMware!").
						WithVar("password4", "&VMware!").
						WithVar("password5", "*VMware!").
						WithVar("multiline", "multi-line\nstring").
						WithVar("structured", "{\n\t\"key\": \"value\"}")
					path := v1alpha1.PathInfo{Path: dir}
					dp := &fakeDefinitionParser{
						parsePath: []v1alpha1.PathInfo{path},
					}
					yp := yamlprocessor.NewYttProcessor(yamlprocessor.InjectDefinitionParser(dp))

					actual, err := yp.Process([]byte("doesn't matter"), configClient.Get)
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).ToNot(BeNil())
					expectedYaml := `password1: '%&%VMware!'
password2: '!VMware!'
password3: '#VMware!'
password4: '&VMware!'
password5: '*VMware!'
multiline: multi-line string
structured:
  key: value
`
					Expect(string(actual)).To(Equal(expectedYaml))
				})
			})
		})
	})
})

var simpleYttYaml = `
#@ clusterName = "foobar"
#@ count = 4
#@ namespace = "foobar-ns"
---
name: #@ clusterName + "-cluster"
namespace: #@ namespace
replicas: #@ count
---`

var assertYttYaml = `
#@ load("@ytt:assert", "assert")
#@ val = 123
key: #@ val if val > 130 else assert.fail("val is too small")
`

var templateYaml = `
#@ load("@ytt:data", "data")
#@ clusterName = data.values.cluster_name
#@ count = data.values.count
#@ region = data.values.AWS_REGION
---
name: #@ clusterName + "-cluster"
replicas: #@ count
region: #@ region
---`

var templateDataYaml = `
#@data/values
---
cluster_name: default
count: 1
secretValue:
`

// calling this templateDataYaml2 to keep consistent with ytt file order
// naming conventions.
var templateDataYaml2 = `
#@data/values
---
#@overlay/match missing_ok=True
AWS_REGION: us-west-1a
`

// FakeVariableClient provides a VariableClient backed by a map
type FakeVariableClient struct {
	variables map[string]string
}

func NewFakeVariableClient() *FakeVariableClient {
	return &FakeVariableClient{
		variables: map[string]string{},
	}
}

func (f FakeVariableClient) Get(key string) (string, error) {
	if val, ok := f.variables[key]; ok {
		return val, nil
	}
	return "", errors.Errorf("value for variable %q is not set", key)
}

func (f FakeVariableClient) Set(key, value string) {
	f.variables[key] = value
}

func (f *FakeVariableClient) WithVar(key, value string) *FakeVariableClient {
	f.variables[key] = value
	return f
}

type fakeDefinitionParser struct {
	parsePath    []v1alpha1.PathInfo
	parsePathErr error
}

var _ yamlprocessor.DefinitionParser = &fakeDefinitionParser{}

func (f *fakeDefinitionParser) ParsePath(_ []byte) ([]v1alpha1.PathInfo, error) {
	return f.parsePath, f.parsePathErr
}
