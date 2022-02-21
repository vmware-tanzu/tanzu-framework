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

	k8sVersionPrefix := "1.22"
	tkrSelector, _ := labels.Parse("!deprecated")
	osImageSelector, _ := labels.Parse("os-name=ubuntu,ami-region=us-west-2")

	query := data.Query{
		ControlPlane: data.OSImageQuery{
			K8sVersionPrefix: k8sVersionPrefix,
			TKRSelector:      tkrSelector,
			OSImageSelector:  osImageSelector,
		},
		MachineDeployments: map[string]data.OSImageQuery{
			"np1": {
				K8sVersionPrefix: k8sVersionPrefix,
				TKRSelector:      tkrSelector,
				OSImageSelector:  osImageSelector,
			},
		},
	}

	result := tkrResolver.Resolve(query)

	println("KubernetesVersion", result.ControlPlane.K8sVersion)
}
