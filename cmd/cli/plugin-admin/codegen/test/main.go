// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
	clitest "github.com/vmware-tanzu/tanzu-framework/pkg/v1/test/cli"
)

var descriptor = cli.NewTestFor("codegen")

var featureGenerationTest *clitest.Test

func main() {
	p, err := plugin.NewPlugin(descriptor)
	if err != nil {
		log.Fatal(err)
	}

	p.Cmd.RunE = test
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func initialize() error {
	tempDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return err
	}
	featureGenerationCommand := fmt.Sprintf("codegen generate paths=./cmd/cli/plugin-admin/codegen/test/fakeData/... feature output:feature:artifacts:config=%s", tempDir)
	featureGenerationTest = clitest.NewTest("feature generation", featureGenerationCommand, func(t *clitest.Test) error {
		defer os.Remove(tempDir)
		if err := t.Exec(); err != nil {
			return err
		}

		fooFeatureCRBytes, err := os.ReadFile(filepath.Join(tempDir, "foo.yaml"))
		if err != nil {
			return err
		}

		var actualFooFeature configv1alpha1.Feature
		if err := yaml.Unmarshal(fooFeatureCRBytes, &actualFooFeature); err != nil {
			return err
		}

		if actualFooFeature.Name != "foo" ||
			actualFooFeature.Spec.Activated ||
			actualFooFeature.Spec.Discoverable ||
			actualFooFeature.Spec.Immutable ||
			actualFooFeature.Spec.Maturity != "dev" {
			return fmt.Errorf("feature generation was not successful")
		}
		return nil
	})
	return nil
}

func test(c *cobra.Command, _ []string) error {
	m := clitest.NewMain("codegen", c, Cleanup)
	defer m.Finish()

	if err := initialize(); err != nil {
		return err
	}

	m.AddTest(featureGenerationTest)
	if err := featureGenerationTest.Run(); err != nil {
		return err
	}

	return nil
}

// Cleanup the test.
func Cleanup() error {
	return nil
}
