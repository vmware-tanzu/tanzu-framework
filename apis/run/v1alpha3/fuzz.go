// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	fuzz "github.com/google/gofuzz"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

// FuzzTKRSpec fuzzes the passed TanzuKubernetesReleaseSpec.
func FuzzTKRSpec(spec *TanzuKubernetesReleaseSpec, c fuzz.Continue) {
	v := &version.Version{}
	c.Fuzz(v)
	spec.Version = v.String()
	c.Fuzz(&spec.Kubernetes)
	c.Fuzz(&spec.OSImages)
	c.Fuzz(&spec.BootstrapPackages)
}

// FuzzTKRSpecKubernetes fuzzes the passed TKR KubernetesSpec.
func FuzzTKRSpecKubernetes(kubernetesSpec *KubernetesSpec, c fuzz.Continue) {
	v := &version.Version{}
	c.Fuzz(v)
	kubernetesSpec.Version = v.String()
	c.Fuzz(&kubernetesSpec.ImageRepository)
	c.Fuzz(kubernetesSpec.Etcd)
	c.Fuzz(kubernetesSpec.Pause)
	c.Fuzz(kubernetesSpec.CoreDNS)
	c.Fuzz(kubernetesSpec.KubeVIP)
}
