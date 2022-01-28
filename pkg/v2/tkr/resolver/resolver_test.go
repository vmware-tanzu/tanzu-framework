// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
)

func TestResolverNew(t *testing.T) {
	t.Run("resolver.New() returns a non-nil tkr.Resolver", func(t *testing.T) {
		assert.NotNil(t, resolver.New())
	})
}

func Example() {
	tkrResolver := resolver.New()

	tkrSelector := labels.SelectorFromSet(labels.Set{})
	osImageSelector := labels.SelectorFromSet(labels.Set{})

	query := data.Query{
		ControlPlane: data.OSImageQuery{
			K8sVersionPrefix: "1.22",
			TKRSelector:      tkrSelector,
			OSImageSelector:  osImageSelector,
		},
		MachineDeployments: map[string]data.OSImageQuery{
			"np1": {
				K8sVersionPrefix: "1.22",
				TKRSelector:      tkrSelector,
				OSImageSelector:  osImageSelector,
			},
		},
	}

	result := tkrResolver.Resolve(query)

	println("KubernetesVersion", result.ControlPlane.K8sVersion)
}
