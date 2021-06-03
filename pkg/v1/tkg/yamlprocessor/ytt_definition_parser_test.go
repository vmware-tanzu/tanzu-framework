// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package yamlprocessor_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/yamlprocessor"
)

var _ = Describe("YttDefinitionParser", func() {
	Context("ParsePath", func() {
		It("returns error if any path cannot be parsed", func() {
			// create a fake home dir
			homeDir, err := os.MkdirTemp("", "tkg-cli")
			setHomeDirectoryEnvVariable(homeDir)
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(homeDir)
			// make our fake config dir in home dir
			tkgDir := filepath.Join(homeDir, ".tkg")
			dirRelPath := filepath.Join("providers", "aws", "v0.1", "ytt", "plan1")
			dirAbsPath := filepath.Join(tkgDir, dirRelPath)
			Expect(os.MkdirAll(dirAbsPath, os.ModePerm)).To(Succeed())

			ydp := yamlprocessor.NewYttDefinitionParser()
			templateDef := []byte(fmt.Sprintf(`
apiVersion: run.tanzu.vmware.com/v1alpha1
kind: TemplateDefinition
spec:
  paths:
  - path: http://192.168.0.31:8080/
  - path: '%s'`, dirRelPath))
			paths, err := ydp.ParsePath(templateDef)
			Expect(err).To(HaveOccurred())
			Expect(paths).To(BeEmpty())
		})

		It("returns the path specified in the template definition", func() {
			// create a fake home dir
			homeDir, err := os.MkdirTemp("", "tkg-cli")
			setHomeDirectoryEnvVariable(homeDir)
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(homeDir)
			// make our fake config dir in home dir
			tkgDir := filepath.Join(homeDir, ".tkg")
			dirRelPath := filepath.Join("providers", "aws", "v0.1", "ytt", "plan1")
			dirAbsPath := filepath.Join(tkgDir, dirRelPath)

			dirRelPath2 := filepath.Join("providers", "aws", "ytt", "addons")
			dirAbsPath2 := filepath.Join(tkgDir, dirRelPath2)

			Expect(os.MkdirAll(dirAbsPath, os.ModePerm)).To(Succeed())
			Expect(os.MkdirAll(dirAbsPath2, os.ModePerm)).To(Succeed())

			ydp := yamlprocessor.NewYttDefinitionParser(yamlprocessor.InjectTKGDir(tkgDir))

			templateDef := []byte(fmt.Sprintf(`
apiVersion: run.tanzu.vmware.com/v1alpha1
kind: TemplateDefinition
spec:
  paths:
  - path: '%s'
  - path: '%s'
    filemark: text-plain`, dirRelPath, dirRelPath2))
			paths, err := ydp.ParsePath(templateDef)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(paths)).To(Equal(2))
			Expect(paths[0].Path).To(Equal(dirAbsPath))
			Expect(paths[1].Path).To(Equal(dirAbsPath2))
			Expect(paths[1].FileMark).To(Equal("text-plain"))
		})

		It("returns an error if the path specified within the template definition is outside the config dir", func() {
			// create a fake home dir
			homeDir, err := os.MkdirTemp("", "tkg-cli")
			setHomeDirectoryEnvVariable(homeDir)
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(homeDir)
			// make another fake dir outside the config directory so we don't
			// run into the directory doesn't exist error
			dir2 := filepath.Join(homeDir, "tmp", "bad", "yamldir")
			Expect(os.MkdirAll(dir2, os.ModePerm)).To(Succeed())

			ydp := yamlprocessor.NewYttDefinitionParser(yamlprocessor.InjectTKGDir(homeDir))

			dirPathOutsideConfigDir := filepath.Join("providers", "..", "..", "tmp", "bad", "yamldir")

			templateDef := []byte(fmt.Sprintf(`
apiVersion: run.tanzu.vmware.com/v1alpha1
kind: TemplateDefinition
spec:
  paths:
  - path: '%s'`, dirPathOutsideConfigDir))

			paths, err := ydp.ParsePath(templateDef)
			Expect(err).To(HaveOccurred())
			Expect(paths).To(BeEmpty())
		})

		DescribeTable("returns error regarding the template definition format",
			func(templateDef string) {
				ydp := yamlprocessor.NewYttDefinitionParser(yamlprocessor.InjectTKGDir("/Users/foo"))

				path, err := ydp.ParsePath([]byte(templateDef))
				Expect(err).To(HaveOccurred())
				Expect(path).To(BeEmpty())
			},
			Entry("if unable to unmarshal the template definition", `stuff:=1231,asdasa`),

			Entry("if unable to parse the template definition", `
kind: TemplateDefinition
spec:
  paths:
  - path: ~/.tkg/providers/infrastructure-aws/v0.5.3/ytt`),

			Entry("if it doesn't recognize the apiVersion of the template definition", `
apiVersion: some-version
kind: TemplateDefinition
spec:
  paths:
  - path: ~/.tkg/providers/infrastructure-aws/v0.5.3/ytt`),

			Entry("if it doesn't recognize the kind of the template definition", `
apiVersion: run.tanzu.vmware.com/v1alpha1
kind: SomeDefinition
spec:
  paths:
  - path: ~/.tkg/providers/infrastructure-aws/v0.5.3/ytt`),

			Entry("if the path is not absolute", `
apiVersion: run.tanzu.vmware.com/v1alpha1
kind: TemplateDefinition
spec:
  paths:
  - path: ../plan1/ytt`),

			Entry("if the path doesn't exist", `
apiVersion: run.tanzu.vmware.com/v1alpha1
kind: TemplateDefinition
spec:
  paths:
  - path: /Users/foo/.tkg/providers/nothing-here`),

			Entry("if the path is not within the expected basepath of $HOME/.tkg/providers", `
apiVersion: run.tanzu.vmware.com/v1alpha1
kind: TemplateDefinition
spec:
  paths:
  - path: /Users/foo/.tkg/providers/../../aws/v0.2.1/file`),
		)
	})
})

func setHomeDirectoryEnvVariable(homeDir string) {
	os.Setenv("HOME", homeDir)
	// Set USERPROFILE for windows
	os.Setenv("USERPROFILE", homeDir)
}
