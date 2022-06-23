// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/tanzu-auth-controller-manager/controllers"

	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var metricsAddr string

func main() {
	flag.StringVar(&metricsAddr, "metrics-bind-addr", ":8080", "The address the metric endpoint binds to.")
	disableCascadeV1alpha1 := flag.Bool("disable-cascade-v1alpha1", false, "whether to disable the v1alpha1 control loop")
	klog.InitFlags(nil)
	flag.Parse()
	ctrl.SetLogger(klogr.New())
	setupLog := ctrl.Log.WithName("tanzu auth controller manager").WithName("set up")
	setupLog.Info("starting set up")
	if err := reallyMain(setupLog, *disableCascadeV1alpha1); err != nil {
		setupLog.Error(err, "error running controller")
		os.Exit(1)
	}
}

func reallyMain(setupLog logr.Logger, disableCascadeV1alpha1 bool) error {
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
		MetricsBindAddress: metricsAddr,
		Scheme:             scheme,
	})
	if err != nil {
		return fmt.Errorf("unable to start manager: %w", err)
	}

	// Register our controllers with the manager.
	if !disableCascadeV1alpha1 {
		if err := controllers.NewV1Controller(manager.GetClient()).SetupWithManager(manager); err != nil {
			return fmt.Errorf("unable to create %s: %w", controllers.CascadeControllerV1alpha1Name, err)
		}
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
