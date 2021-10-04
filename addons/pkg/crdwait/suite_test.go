// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package crdwait

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteName := "CRDWait Test Suite"
	RunSpecsWithDefaultAndCustomReporters(t, suiteName, []Reporter{printer.NewlineReporter{}, printer.NewProwReporter(suiteName)})
}

var (
	ctx = ctrl.SetupSignalHandler()

	cancel context.CancelFunc
)

var _ = BeforeSuite(func(done Done) {
	ctx, cancel = context.WithCancel(ctx)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	cancel()
})
