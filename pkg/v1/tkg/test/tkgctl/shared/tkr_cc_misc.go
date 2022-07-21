// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,stylecheck,nolintlint
package shared

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/tkr-source/compatibility"
)

type TKRCompatibilityValidationSpecInput struct {
	E2EConfig    *framework.E2EConfig
	OtherConfigs map[string]string
}

func TKRCompatibilityValidationSpec(ctx context.Context, inputGetter func() TKRCompatibilityValidationSpecInput) { //nolint:funlen
	var (
		input         TKRCompatibilityValidationSpecInput
		mcProxy       *framework.ClusterProxy
		mcContextName string
		tkrs          []*runv1.TanzuKubernetesRelease
	)

	BeforeEach(func() { //nolint:dupl
		input = inputGetter()
		mcClusterName := input.E2EConfig.ManagementClusterName
		mcContextName = mcClusterName + "-admin@" + mcClusterName
		mcProxy = framework.NewClusterProxy(mcClusterName, "", mcContextName)
	})

	It("Should validate the compatible status is correctly calculated for all TKRs", func() {
		tkrCompatibility := &compatibility.Compatibility{
			Client: mcProxy.GetClient(),
			Config: compatibility.Config{
				TKRNamespace: "tkg-system",
			},
			Log: logr.Discard(),
		}
		By("Validating all TKRs compatibility status condition is updated correctly")
		compatibleSet, err := tkrCompatibility.CompatibleVersions(context.Background())
		fmt.Printf("CompatibleSet is :%+v \n", compatibleSet)
		Expect(err).ToNot(HaveOccurred())
		tkrs = mcProxy.GetTKRs(ctx)
		for i := range tkrs {
			fmt.Printf("Validating the compatibility status condition for TKR '%s'\n", tkrs[i].Name)
			if compatibleSet.Has(tkrs[i].Spec.Version) {
				Expect(conditions.IsTrue(tkrs[i], runv1.ConditionCompatible)).To(BeTrue(),
					fmt.Sprintf("TKR '%s' is expected to have Compatible condition to be true", tkrs[i].Name))
			} else {
				Expect(conditions.IsFalse(tkrs[i], runv1.ConditionCompatible)).To(BeTrue(),
					fmt.Sprintf("TKR '%s' is expected to have Compatible condition to be false", tkrs[i].Name))
				Expect(*conditions.GetSeverity(tkrs[i], runv1.ConditionCompatible)).ToNot(Equal(clusterv1.ConditionSeverityWarning),
					fmt.Sprintf("TKR '%s' Compatible condition's severity is expected to be 'Warning' if condition status is False", tkrs[i].Name))
			}
		}
		By("Test successful !")
	})

}
