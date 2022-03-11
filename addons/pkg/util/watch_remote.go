// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/cluster-api/controllers/remote"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// RemoteObjectTracker is a helper struct to deal with watching remote clusters
type RemoteObjectTracker struct {
	log    logr.Logger
	client client.Client
	scheme *runtime.Scheme

	lock             sync.RWMutex
	ClusterAccessors map[client.ObjectKey]*clusterAccessorWithWatch
}

// NewRemoteObjectTracker creates a new RemoteObjectTracker
func NewRemoteObjectTracker(manager ctrl.Manager) *RemoteObjectTracker {
	return &RemoteObjectTracker{
		log:              ctrl.Log.WithName("RemoteObjectTracker"),
		client:           manager.GetClient(),
		scheme:           manager.GetScheme(),
		ClusterAccessors: make(map[client.ObjectKey]*clusterAccessorWithWatch),
	}
}

// clusterAccessorWithWatch represents the combination of a client and watches for a remote cluster
type clusterAccessorWithWatch struct {
	client  client.Client
	watches sets.String
}

// getClusterAccessor first tries to return an already-created clusterAccessor for cluster, falling back to creating a
// new clusterAccessor if needed. Note, this method requires t.lock to already be held before being called
func (t *RemoteObjectTracker) getClusterAccessor(ctx context.Context, cluster client.ObjectKey) (*clusterAccessorWithWatch, error) {
	a := t.ClusterAccessors[cluster]
	if a != nil {
		return a, nil
	}
	remoteClient, err := GetClusterClient(ctx, t.client, t.scheme, cluster)
	if err != nil {
		return nil, errors.Wrap(err, "error creating client for remote cluster")
	}

	a = &clusterAccessorWithWatch{
		client:  remoteClient,
		watches: sets.NewString(),
	}
	t.ClusterAccessors[cluster] = a

	return a, nil
}

// Watch watches a remote cluster for resource events. If the watch already exists based on input.Name, this is a no-op
func (t *RemoteObjectTracker) Watch(ctx context.Context, input *remote.WatchInput) error {
	if input == nil || input.Name == "" {
		return errors.New("input.Name is required")
	}

	t.lock.Lock()
	defer t.lock.Unlock()

	a, err := t.getClusterAccessor(ctx, input.Cluster)
	if err != nil {
		return err
	}

	if a.watches.Has(input.Name) {
		t.log.Info("Watch already exists", "namespace", input.Cluster.Namespace, "cluster", input.Cluster.Name, "name", input.Name)
		return nil
	}

	// Need to create the watch
	if err := input.Watcher.Watch(&source.Kind{Type: input.Kind}, input.EventHandler, input.Predicates...); err != nil {
		return errors.Wrap(err, "error creating watch")
	}
	a.watches.Insert(input.Name)

	return nil
}

// DeleteAccessor removes the clusterAccessor from the tracker
func (t *RemoteObjectTracker) DeleteAccessor(cluster client.ObjectKey) {
	t.lock.Lock()
	defer t.lock.Unlock()

	_, exists := t.ClusterAccessors[cluster]
	if !exists {
		return
	}

	t.log.Info("Deleting ClusterAccessors", "cluster", cluster.String())

	delete(t.ClusterAccessors, cluster)
}

// GetClient returns the remote client for the given cluster
func (t *RemoteObjectTracker) GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	accessor, err := t.getClusterAccessor(ctx, cluster)
	if err != nil {
		return nil, err
	}

	return accessor.client, nil
}

// ClusterAccessorExists returns true if a clusterAccessor exists for the provided cluster key
func (t *RemoteObjectTracker) ClusterAccessorExists(cluster client.ObjectKey) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	_, exists := t.ClusterAccessors[cluster]
	return exists
}
