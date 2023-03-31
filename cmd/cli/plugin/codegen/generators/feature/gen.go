// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package feature

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

const markerName = "tanzu:feature"

var (
	// RuleDefinition is a marker for defining Feature rules.
	RuleDefinition = markers.Must(markers.MakeDefinition(markerName, markers.DescribesType, Rule{}))
)

// Generator is feature generator that registers feature markers and produces output artifacts
type Generator struct{}

// Rule is the output type of the marker value
type Rule struct {
	// Name of the feature.
	Name string
	// Description of the feature.
	Description string `marker:",optional"`
	// Stability indicates stability level of this feature.
	Stability corev1alpha2.StabilityLevel
}

// Generate generates artifacts produced by feature marker.
func (g Generator) Generate(ctx *genall.GenerationContext) error {
	objs := generateFeatures(ctx)

	if len(objs) == 0 {
		return nil
	}
	for _, obj := range objs {
		if err := ctx.WriteYAML(obj.(corev1alpha2.Feature).Name+".yaml", obj); err != nil {
			return err
		}
	}
	return nil
}

// RegisterMarkers registers all markers needed by this Generator
func (Generator) RegisterMarkers(reg *markers.Registry) error {
	if err := reg.Register(RuleDefinition); err != nil {
		return err
	}
	reg.AddHelp(RuleDefinition, &markers.DefinitionHelp{
		Category: "feature",
		DetailedHelp: markers.DetailedHelp{
			Summary: "is the output type of the marker value",
			Details: "",
		},
		FieldHelp: map[string]markers.DetailedHelp{
			"Name": {
				Summary: "specifies name of the feature.",
				Details: "",
			},
			"Description": {
				Summary: "specifies description of the feature.",
				Details: "",
			},
			"Stability": {
				Summary: "indicates stability level of this feature.",
				Details: "",
			},
		},
	})
	return nil
}

func generateFeatures(ctx *genall.GenerationContext) []interface{} {
	var objs []interface{}
	for _, root := range ctx.Roots {
		if err := markers.EachType(ctx.Collector, root, func(info *markers.TypeInfo) {
			markerValues := getMarkerValues(markerName, info.Markers)
			for _, markerValue := range markerValues {
				val := markerValue.(Rule)
				objs = append(objs, corev1alpha2.Feature{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Feature",
						APIVersion: corev1alpha2.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: val.Name,
					},
					Spec: corev1alpha2.FeatureSpec{
						Description: val.Description,
						Stability:   val.Stability,
					},
					Status: corev1alpha2.FeatureStatus{},
				})
			}
		}); err != nil {
			root.AddError(err)
		}
	}
	return objs
}

func getMarkerValues(name string, markerValues map[string][]interface{}) []interface{} {
	values := markerValues[name]
	if len(values) == 0 {
		return nil
	}
	return values
}
