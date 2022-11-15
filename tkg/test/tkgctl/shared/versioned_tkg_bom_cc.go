// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,stylecheck,nolintlint
package shared

import (
	"context"

	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
)

// verifyCCClusterVersionedTKGBOM verify the versioned tkg bom configmap
func verifyCCClusterVersionedTKGBOM(ctx context.Context, c client.Client, tkgBomVersion string) {
	// verifying tkg bom
	bomConfigMap := &corev1.ConfigMap{}

	Eventually(func() error {
		return c.Get(ctx, client.ObjectKey{Namespace: "tkg-system-public", Name: "tkg-bom" + "-" + tkgBomVersion}, bomConfigMap)
	}, getResourceTimeout, pollingInterval).Should(Succeed())
	bomYaml := bomConfigMap.Data["bom.yaml"]
	Expect(bomYaml).ToNot(Equal(""))
	bom := &tkgconfigbom.BOMConfiguration{}
	err := yaml.Unmarshal([]byte(bomYaml), bom)
	Expect(err).ToNot(HaveOccurred())
}
