// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/controllers"

	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	ctrl.SetLogger(klogr.New())
	setupLog := ctrl.Log.WithName("pinniped config controller").WithName("set up")
	setupLog.Info("starting set up")
	if err := reallyMain(setupLog); err != nil {
		setupLog.Error(err, "error running controller")
		os.Exit(1)
	}
}

func reallyMain(setupLog logr.Logger) error {
	// Add types our controller uses to scheme.
	scheme := runtime.NewScheme()
	addToSchemes := []func(*runtime.Scheme) error{
		corev1.AddToScheme,
		clusterapiv1beta1.AddToScheme,
	}
	for _, addToScheme := range addToSchemes {
		if err := addToScheme(scheme); err != nil {
			return fmt.Errorf("cannot add to scheme: %w", err)
		}
	}

	// Create manager to run our controller.
	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("unable to start manager: %w", err)
	}

	// Register our controllers with the manager.
	if err := controllers.NewV1Controller(manager.GetClient()).SetupWithManager(manager); err != nil {
		return fmt.Errorf("unable to create %s: %w", controllers.CascadeControllerV1alpha1Name, err)
	}
	if err := controllers.NewV3Controller(manager.GetClient()).SetupWithManager(manager); err != nil {
		return fmt.Errorf("unable to create %s: %w", controllers.CascadeControllerV1alpha3Name, err)
	}

	// Tell manager to start running our controller.
	setupLog.V(1).Info("starting manager")
	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("unable to start manager: %w", err)
	}

	return nil
}
