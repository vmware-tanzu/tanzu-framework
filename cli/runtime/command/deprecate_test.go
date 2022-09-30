// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Command Suite")
}

var _ = Describe("Test suite for the deprecation module", func() {
	var testCmd = &cobra.Command{
		Use:          "test",
		Short:        "test",
		SilenceUsage: true,
	}

	flag := "foo"
	removalVersion := "2.0.0"
	alternative := "bar"

	var cmdVariable string
	testCmd.Flags().StringVarP(&cmdVariable, flag, "", "", "")

	Context("Test flag deprecation", func() {
		It("test flag deprecation without alternative", func() {
			DeprecateFlag(testCmd, flag, removalVersion)
			Expect(testCmd.Flag(flag).Deprecated).To(Equal(fmt.Sprintf("will be removed in version %q.", removalVersion)))
			Expect(testCmd.Flag(flag).Hidden).To(BeTrue())
		})

		It("test flag deprecation with alternative", func() {
			DeprecateFlagWithAlternative(testCmd, flag, removalVersion, alternative)
			Expect(testCmd.Flag(flag).Deprecated).To(Equal(fmt.Sprintf("will be removed in version %q. Use %q instead.", removalVersion, alternative)))
			Expect(testCmd.Flag(flag).Hidden).To(BeTrue())
		})
	})

	Context("Test command deprecation", func() {
		It("Test command deprecation without alternative", func() {
			DeprecateCommand(testCmd, removalVersion)
			Expect(testCmd.Deprecated).To(Equal(fmt.Sprintf("will be removed in version %q.", removalVersion)))
		})

		It("Test command deprecation with alternative", func() {
			DeprecateCommandWithAlternative(testCmd, removalVersion, alternative)
			Expect(testCmd.Deprecated).To(Equal(fmt.Sprintf("will be removed in version %q. Use %q instead.", removalVersion, alternative)))
		})
	})
})
