// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"

	"github.com/vmware-tanzu-private/core/pkg/v1/codegen/generators/feature"
)

var (
	// GenerateCmd generates the Tanzu API extension resources and code.
	GenerateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate Tanzu API extension resources and code.",
		Long:  "Generate Tanzu API extension resources and code.",
		RunE:  runGenerate,
	}

	// allGenerators maintains the list of all known generators
	allGenerators = map[string]genall.Generator{
		"feature": feature.Generator{},
	}

	// allOutputRules defines the list of all known output rules
	allOutputRules = map[string]genall.OutputRule{
		"dir":       genall.OutputToDirectory(""),
		"none":      genall.OutputToNothing,
		"stdout":    genall.OutputToStdout,
		"artifacts": genall.OutputArtifacts{},
	}

	// optionsRegistry contains all the marker definitions used to process command line options
	optionsRegistry = &markers.Registry{}
)

func registerOptionsMarkers() error {
	for genName, gen := range allGenerators {
		// make the generator options marker itself
		defn := markers.Must(markers.MakeDefinition(genName, markers.DescribesPackage, gen))
		if err := optionsRegistry.Register(defn); err != nil {
			return err
		}
		if helpGiver, hasHelp := gen.(genall.HasHelp); hasHelp {
			if help := helpGiver.Help(); help != nil {
				optionsRegistry.AddHelp(defn, help)
			}
		}

		// make per-generation output rule markers
		for ruleName, rule := range allOutputRules {
			ruleMarker := markers.Must(markers.MakeDefinition(fmt.Sprintf("output:%s:%s", genName, ruleName), markers.DescribesPackage, rule))
			if err := optionsRegistry.Register(ruleMarker); err != nil {
				return err
			}
			if helpGiver, hasHelp := rule.(genall.HasHelp); hasHelp {
				if help := helpGiver.Help(); help != nil {
					optionsRegistry.AddHelp(ruleMarker, help)
				}
			}
		}
	}

	// make "default output" output rule markers
	for ruleName, rule := range allOutputRules {
		ruleMarker := markers.Must(markers.MakeDefinition("output:"+ruleName, markers.DescribesPackage, rule))
		if err := optionsRegistry.Register(ruleMarker); err != nil {
			return err
		}
		if helpGiver, hasHelp := rule.(genall.HasHelp); hasHelp {
			if help := helpGiver.Help(); help != nil {
				optionsRegistry.AddHelp(ruleMarker, help)
			}
		}
	}

	// add in the common options markers
	if err := genall.RegisterOptionsMarkers(optionsRegistry); err != nil {
		return err
	}
	return nil
}

func runGenerate(cmd *cobra.Command, args []string) error {
	if err := registerOptionsMarkers(); err != nil {
		return err
	}
	runtime, err := genall.FromOptions(optionsRegistry, args)
	if err != nil {
		return err
	}
	if len(runtime.Generators) == 0 {
		return fmt.Errorf("no generators specified")
	}

	if hadErrs := runtime.Run(); hadErrs {
		return fmt.Errorf("not all generators ran successfully")
	}
	return nil
}
