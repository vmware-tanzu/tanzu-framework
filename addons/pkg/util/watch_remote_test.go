// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
)

const (
	clusterNameLabel = "cluster-name"
	testClusterFirst = "test-cluster-one"
	testClusterScnd  = "test-cluster-two"
	testClusterThird = "test-cluster-three"
	testSecretFirst  = "test-secret-one"
	testSecretScnd   = "test-secret-two"
	testSecretThird  = "test-secret-three"
	timeout          = time.Second * 10
)

var (
	ctx               = ctrl.SetupSignalHandler()
	cfg               *rest.Config
	k8sClient         client.Client
	mgr               manager.Manager
	mgrContext        context.Context
	mgrCancel         context.CancelFunc
	remoteClientFirst client.Client
	remoteClientScnd  client.Client
	remoteClientThird client.Client
	tracker           *RemoteObjectTracker
	testEnv           *envtest.Environment
	scheme            = runtime.NewScheme()
	ns                = &corev1.Namespace{TypeMeta: metav1.TypeMeta{Kind: "Namespace"}, ObjectMeta: metav1.ObjectMeta{Name: testNamespace}}
)

type testReconciler struct {
	Log        logr.Logger
	Client     client.Client
	controller controller.Controller
	Tracker    *RemoteObjectTracker
}

var _ = Describe("remote watch", func() {
	Context("remote watch", func() {
		r := setup()
		defer teardown()

		// Creates clusters and watches
		remoteClientFirst = r.createAndWatchCluster(testClusterFirst, ns)
		remoteClientScnd = r.createAndWatchCluster(testClusterScnd, ns)
		remoteClientThird = r.createAndWatchCluster(testClusterThird, ns)

		// Creating secrets labeled with cluster names
		createSecret(remoteClientFirst, testSecretFirst, testClusterFirst)
		createSecret(remoteClientScnd, testSecretScnd, testClusterScnd)
		createSecret(remoteClientThird, testSecretThird, testClusterThird)

		// Deleting clusters and ensuring cluster's clusterAccessor is removed upon cluster deletion
		for _, clusterName := range []string{testClusterFirst, testClusterScnd, testClusterThird} {
			obj := &clusterapiv1beta1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns.Name,
					Name:      clusterName,
				},
			}
			Expect(k8sClient.Delete(ctx, obj)).To(Succeed())
			Eventually(func() bool {
				return tracker.ClusterAccessorExists(clusterapiutil.ObjectKey(obj))
			}, timeout).Should(BeFalse())
		}
	})
})

// createAndWatchCluster creates a new cluster & watch and ensures that clusterAccessor has a key for the cluster containing the name of the corresponding watch
func (r *testReconciler) createAndWatchCluster(clusterName string, namespace *corev1.Namespace) client.Client {
	// Creating a cluster
	testCluster := &clusterapiv1beta1.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: clusterName, Namespace: namespace.GetName()},
	}
	Expect(r.Client.Create(ctx, testCluster)).Should(Succeed())

	// Check the cluster can be fetched from the API server
	testClusterKey := clusterapiutil.ObjectKey(testCluster)
	Eventually(func() error {
		return r.Client.Get(ctx, testClusterKey, &clusterapiv1beta1.Cluster{})
	}, timeout).Should(Succeed())

	// Creating a cluster kubeconfig secret
	Expect(testutil.CreateKubeconfigSecret(cfg, clusterName, namespace.GetName(), r.Client)).To(Succeed())

	// Check the secret can be fetched from the API server
	kubeCfgSecret := &corev1.Secret{}
	secretKey := client.ObjectKey{Namespace: namespace.GetName(), Name: fmt.Sprintf("%s-kubeconfig", testCluster.GetName())}
	Eventually(func() error {
		return r.Client.Get(ctx, secretKey, kubeCfgSecret)
	}, timeout).Should(Succeed())

	// Getting the remote client for the cluster
	remoteClient, err := r.Tracker.GetClient(ctx, testClusterKey)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(remoteClient).ShouldNot(BeNil())

	// Checking the clusterAccessor map has a key for the cluster
	Expect(r.Tracker.ClusterAccessors).ShouldNot(BeNil())
	accessor := r.Tracker.ClusterAccessors[testClusterKey]
	Expect(accessor).ShouldNot(BeNil())
	Expect(len(accessor.watches)).To(Equal(0))

	// Watching a remote cluster for secret resource events
	watchInput := remote.WatchInput{
		Name:         "watchTest",
		Cluster:      clusterapiutil.ObjectKey(testCluster),
		Watcher:      r.controller,
		Kind:         &corev1.Secret{},
		EventHandler: handler.EnqueueRequestsFromMapFunc(r.secretToCluster),
		Predicates:   []predicate.Predicate{},
	}
	Expect(r.Tracker.Watch(ctx, &watchInput)).Should(Succeed())

	// Make sure "watchTest" has been inserted to the tracker's "watches" set
	// so that the watch doesn't get recreated on the next call to "Tracker.Watch"
	Expect(accessor.watches.Has(watchInput.Name)).To(BeTrue())
	Expect(accessor.client).To(Equal(remoteClient))

	return remoteClient
}

// SetupWithManager sets up the controller with the Manager
func (r *testReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager, options controller.Options) error {
	ctrl, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterapiv1beta1.Cluster{}).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.secretToCluster),
		).
		WithOptions(options).
		Build(r)
	Expect(err).ShouldNot(HaveOccurred())

	r.controller = ctrl
	return nil
}

// Reconcile reconciles Clusters and removes cluster accessor for any Cluster that cannot be retrieved
func (r *testReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// get cluster object
	cluster := &clusterapiv1beta1.Cluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("Cluster not found")
			r.Tracker.DeleteAccessor(req.NamespacedName)
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "unable to fetch cluster")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// secretToCluster returns a list of Requests with Cluster ObjectKey
func (r *testReconciler) secretToCluster(o client.Object) []ctrl.Request {
	var secret *corev1.Secret

	switch obj := o.(type) {
	case *corev1.Secret:
		secret = obj
	default:
		return nil
	}
	labels := secret.GetLabels()
	if labels == nil {
		return nil
	}
	clusterName, ok := labels[clusterNameLabel]
	if !ok || clusterName == "" {
		return nil
	}
	cluster, err := clusterapiutil.GetClusterByName(ctx, r.Client, secret.Namespace, clusterName)
	if err != nil || cluster == nil {
		return nil
	}
	if !cluster.GetDeletionTimestamp().IsZero() {
		return nil
	}
	return []ctrl.Request{{
		NamespacedName: clusterapiutil.ObjectKey(cluster),
	}}
}

// setup sets up the test environment and registers the testReconciler controller with the manager
func setup() *testReconciler {
	var err error

	testEnv = &envtest.Environment{CRDInstallOptions: envtest.CRDInstallOptions{
		CleanUpAfterUse: true},
		ErrorIfCRDPathMissing: true,
	}

	externalDeps := map[string][]string{
		"sigs.k8s.io/cluster-api": {"config/crd/bases",
			"controlplane/kubeadm/config/crd/bases"},
	}
	externalCRDPaths, err := testutil.GetExternalCRDPaths(externalDeps)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(externalCRDPaths).ShouldNot(BeEmpty())
	testEnv.CRDDirectoryPaths = externalCRDPaths
	testEnv.CRDDirectoryPaths = append(testEnv.CRDDirectoryPaths,
		filepath.Join("..", "..", "..", "config", "crd", "bases"))
	testEnv.ErrorIfCRDPathMissing = true

	cfg, err = testEnv.Start()
	Expect(err).ShouldNot(HaveOccurred())
	Expect(cfg).ShouldNot(BeNil())

	Expect(corev1.AddToScheme(scheme)).Should(Succeed())

	Expect(clusterapiv1beta1.AddToScheme(scheme)).Should(Succeed())

	// Setting up a new manager
	mgr, err = manager.New(testEnv.Config, manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
	})
	Expect(err).ShouldNot(HaveOccurred())

	// Getting manager's local client
	k8sClient = mgr.GetClient()

	// Setting up a RemoteObjectTracker
	tracker = NewRemoteObjectTracker(mgr)
	Expect(tracker).ShouldNot(BeNil())

	// Creating the testReconciler
	r := &testReconciler{
		Log:     logr.New(log.NullLogSink{}),
		Client:  k8sClient,
		Tracker: tracker,
	}
	Expect(r.SetupWithManager(ctx, mgr, controller.Options{})).To(Succeed())

	// Starting the manager
	mgrContext, mgrCancel = context.WithCancel(ctx)
	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(mgrContext)).To(Succeed())
	}()

	// Creating a namespace for the test
	Expect(k8sClient.Create(ctx, ns)).Should(Succeed())

	return r
}

// teardown removes all created resources
func teardown() {
	// Deleting any Secrets
	Expect(cleanupTestSecrets(ctx, k8sClient)).To(Succeed())
	Expect(cleanupTestSecrets(ctx, remoteClientFirst)).To(Succeed())
	Expect(cleanupTestSecrets(ctx, remoteClientScnd)).To(Succeed())
	Expect(cleanupTestSecrets(ctx, remoteClientThird)).To(Succeed())
	// Deleting any Clusters
	Expect(cleanupTestClusters(ctx, k8sClient)).To(Succeed())
	// Deleting Namespace
	Expect(k8sClient.Delete(ctx, ns)).Should(Succeed())
	// Stopping the manager
	mgrCancel()
}

// createSecret creates a new secret labeled with the provided cluster name using the provided client
func createSecret(c client.Client, secretName, clusterName string) {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: testNamespace,
			Labels:    map[string]string{clusterNameLabel: clusterName},
		},
	}
	_, err := controllerutil.CreateOrPatch(ctx, c, secret, nil)
	Expect(err).ShouldNot(HaveOccurred())
}

// cleanupTestSecrets deletes all secret resources using the provided client
func cleanupTestSecrets(ctx context.Context, c client.Client) error {
	secretList := &corev1.SecretList{}
	if err := c.List(ctx, secretList); err != nil {
		return err
	}
	for _, secret := range secretList.Items {
		s := secret
		if err := c.Delete(ctx, &s); err != nil {
			return err
		}
	}
	return nil
}

// cleanupTestClusters deletes all cluster resources using the provided client
func cleanupTestClusters(ctx context.Context, c client.Client) error {
	clusterList := &clusterapiv1beta1.ClusterList{}
	if err := c.List(ctx, clusterList); err != nil {
		return err
	}
	for _, cluster := range clusterList.Items {
		o := cluster
		if err := c.Delete(ctx, &o); err != nil {
			return err
		}
	}
	return nil
}
