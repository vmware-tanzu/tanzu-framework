// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

var (
	scheme = runtime.NewScheme()
	cl     client.Client
)

func init() {
	utilruntime.Must(corev1alpha2.AddToScheme(scheme))
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(rbacv1.AddToScheme(scheme))
}

func TestE2E(t *testing.T) {
	defer ginkgo.GinkgoRecover()

	var err error
	cl, err = client.New(config.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		t.Error(err)
	}

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "readiness e2e suite")
}
