// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	clientconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/object-propagation/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/object-propagation/propagation"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))
}

var (
	metricsAddr string
	input       string
)

func init() {
	flag.StringVar(&metricsAddr, "metrics-bind-addr", ":8080", "The address the metric endpoint binds to")
	flag.StringVar(&input, "input", "/dev/stdin", "Input file (default: /dev/stdin)")
	flag.Parse()

	setupLog.Info("Version", "version", buildinfo.Version, "buildDate", buildinfo.Date, "sha", buildinfo.SHA)
}

func main() {
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	configEntries := readConfig(input)

	ctx := signals.SetupSignalHandler()
	mgr := createManager()

	propagationConfigs := propagation.Configs(configEntries)
	propagationReconcilers := propagationReconcilers(ctx, mgr, propagationConfigs)
	setupWithManager(mgr, propagationReconcilers)

	startManager(ctx, mgr)
}

func readConfig(input string) []*config.Entry {
	bytes, err := os.ReadFile(input)
	if err != nil {
		panic(errors.Wrap(err, "reading config"))
	}
	result, err := config.Parse(bytes)
	if err != nil {
		panic(errors.Wrap(err, "parsing config"))
	}
	return result
}

func createManager() manager.Manager {
	// Setup Manager
	setupLog.Info("setting up manager")
	mgr, err := manager.New(clientconfig.GetConfigOrDie(), manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
	})
	if err != nil {
		panic(errors.Wrap(err, "unable to set up controller manager"))
	}
	return mgr
}

func propagationReconcilers(ctx context.Context, mgr manager.Manager, propagationConfigs []*propagation.Config) []managedComponent {
	var result []managedComponent
	for _, propagationConfig := range propagationConfigs {
		result = append(result, propagationReconciler(ctx, mgr, propagationConfig))
	}
	return result
}

func propagationReconciler(ctx context.Context, mgr manager.Manager, propagationConfig *propagation.Config) *propagation.Reconciler {
	return &propagation.Reconciler{
		Ctx:    ctx,
		Log:    mgr.GetLogger().WithName("object-propagation").WithName(propagationConfig.ObjectType.GetObjectKind().GroupVersionKind().Kind),
		Client: mgr.GetClient(),
		Config: *propagationConfig,
	}
}

func setupWithManager(mgr manager.Manager, managedComponents []managedComponent) {
	for _, c := range managedComponents {
		setupLog.Info("setting up component", "type", fullTypeName(c))
		if err := c.SetupWithManager(mgr); err != nil {
			panic(errors.Wrapf(err, "unable to setup component type '%s'", fullTypeName(c)))
		}
	}
}

func fullTypeName(c managedComponent) string {
	cType := reflect.TypeOf(c)
	for cType.Kind() == reflect.Ptr {
		cType = cType.Elem()
	}
	return fmt.Sprintf("%s.%s", cType.PkgPath(), cType.Name())
}

type managedComponent interface {
	SetupWithManager(ctrl.Manager) error
}

func startManager(ctx context.Context, mgr manager.Manager) {
	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		panic(errors.Wrap(err, "unable to run manager"))
	}
}
