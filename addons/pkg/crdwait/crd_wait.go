// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package crdwait provides an API to wait for CRDs to be available in a kubernetes api-server
package crdwait

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// CRDWaiter waits for api-resources and emits logs and events for missing resources
type CRDWaiter struct {
	Ctx           context.Context
	ClientSet     kubernetes.Interface
	APIReader     client.Reader
	Logger        logr.Logger
	Scheme        *runtime.Scheme
	PollInterval  time.Duration
	PollTimeout   time.Duration
	eventRecorder record.EventRecorder
}

// WaitForCRDs checks if CRDs are available by polling for resources.
func (c *CRDWaiter) WaitForCRDs(crds map[schema.GroupVersion]*sets.String, object runtime.Object, controllerName string) error {
	eventRecorder, eventBroadcaster := c.getEventRecorder(c.ClientSet, controllerName)
	go func() {
		<-c.Ctx.Done()
		eventBroadcaster.Shutdown()
	}()

	poller := func() (done bool, err error) {
		// Create new clientset for every invocation to avoid caching of resources
		allFound := true
		for gv, resources := range crds {
			// All resources found, do nothing
			if resources.Len() == 0 {
				delete(crds, gv)
				continue
			}
			groupVersion := gv.String()
			for _, resource := range resources.List() {
				// Get the Resources for this GroupVersion

				res := gv.WithResource(resource)
				crd := unstructured.Unstructured{}
				crd.SetGroupVersionKind(apiextensions.SchemeGroupVersion.WithKind("CustomResourceDefinition"))
				name := res.GroupResource().String()
				err := c.APIReader.Get(c.Ctx,
					client.ObjectKey{Name: name},
					&crd)

				if apierrors.IsNotFound(err) {
					c.Logger.Info("CRD not available yet", "name", name)
					eventRecorder.Eventf(object, corev1.EventTypeWarning,
						"Polling for CRD", "The CRD '%s' is not available yet", name)
					return false, nil
				}
				if err != nil {
					c.Logger.Error(err, "error retrieving CRD", "name", name)
					eventRecorder.Eventf(object, corev1.EventTypeWarning,
						"Error retrieving CRD %s", "Error trying to read CRD '%s'. Exiting", name)
					return false, err
				}
				resources.Delete(resource)
			}

			// Still waiting on some resources in this group version
			if resources.Len() != 0 {
				allFound = false
				c.Logger.Info("resources are not available yet", "api-resources", resources.List(), "GroupVersion", groupVersion)
				eventRecorder.Eventf(object, corev1.EventTypeWarning,
					"Polling for api-resources", "The api-resources '%s' in GroupVersion '%s' are not available yet", resources.List(), groupVersion)
			}
		}
		return allFound, nil
	}
	return wait.PollImmediate(c.PollInterval, c.PollTimeout, poller)
}

func (c *CRDWaiter) getEventRecorder(clientSet kubernetes.Interface, component string) (record.EventRecorder, record.EventBroadcaster) {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{
			Interface: clientSet.CoreV1().Events(corev1.NamespaceAll)})

	eventBroadcaster.StartEventWatcher(
		func(e *corev1.Event) {
			c.Logger.V(1).Info(e.Type, "object", e.InvolvedObject, "reason", e.Reason, "message", e.Message)
		})
	var recorder record.EventRecorder
	if c.eventRecorder == nil {
		recorder = eventBroadcaster.NewRecorder(
			c.Scheme,
			corev1.EventSource{Component: component})
	} else {
		recorder = c.eventRecorder
	}

	return recorder, eventBroadcaster
}
