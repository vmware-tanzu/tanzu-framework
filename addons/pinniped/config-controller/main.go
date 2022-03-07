// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"k8s.io/klog/v2/klogr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/controllers"
)

func main() {
	setupLog := ctrl.Log.WithName("Pinniped Config Controller Set Up")
	setupLog.Info("starting")

	// Add types to scheme.
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		setupLog.Error(err, fmt.Sprintf("cannot add %s to scheme", corev1.SchemeGroupVersion))
		os.Exit(1)
	}
	if err := clusterapiv1beta1.AddToScheme(scheme); err != nil {
		setupLog.Error(err, fmt.Sprintf("cannot add %s to scheme", clusterapiv1beta1.GroupVersion))
	}

	// Create manager to run our controller.
	ctrl.SetLogger(klogr.New())
	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	controller := controllers.NewController(manager.GetClient())
	if err := controller.SetupWithManager(manager); err != nil {
		setupLog.Error(err, "unable to create Pinniped Config Controller")
		os.Exit(1)
	}
	// Tell manager to start running our controller.
	setupLog.Info("starting manager")
	if err := manager.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
}
