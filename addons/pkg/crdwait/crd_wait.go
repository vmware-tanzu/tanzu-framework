// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package crdwait provides an API to wait for CRDs to be available in a kubernetes api-server
package crdwait

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
)

type clientSetProvider func() (kubernetes.Interface, error)

// CRDWaiter waits for api-resources and emits logs and events for missing resources
type CRDWaiter struct {
	Ctx           context.Context
	ClientSetFn   clientSetProvider
	Logger        logr.Logger
	Scheme        *runtime.Scheme
	PollInterval  time.Duration
	PollTimeout   time.Duration
	eventRecorder record.EventRecorder
}

// WaitForCRDs checks if CRDs are available by polling for resources.
func (c *CRDWaiter) WaitForCRDs(crds map[schema.GroupVersion]*sets.String, object runtime.Object, controllerName string) error {
	cs, err := c.ClientSetFn()
	if err != nil {
		return err
	}

	eventRecorder, eventBroadcaster := c.getEventRecorder(cs, controllerName)
	go func() {
		<-c.Ctx.Done()
		eventBroadcaster.Shutdown()
	}()

	poller := func() (done bool, err error) {
		// Create new clientset for every invocation to avoid caching of resources
		cs, err := c.ClientSetFn()
		if err != nil {
			return false, err
		}

		allFound := true
		for gv, resources := range crds {
			// All resources found, do nothing
			if resources.Len() == 0 {
				delete(crds, gv)
				continue
			}

			// Get the Resources for this GroupVersion
			groupVersion := gv.String()

			resourceList, err := cs.Discovery().ServerResourcesForGroupVersion(groupVersion)
			if err != nil {
				c.Logger.Info("error retrieving GroupVersion", "GroupVersion", groupVersion)
				eventRecorder.Eventf(object, corev1.EventTypeWarning,
					"Polling for GroupVersion", "The GroupVersion '%s' is not available yet", groupVersion)
				return false, nil
			}

			// Remove each found resource from the resources set that we are waiting for
			for i := 0; i < len(resourceList.APIResources); i++ {
				resources.Delete(resourceList.APIResources[i].Name)
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

// GetClientSet returns ClientSet. Tests can return fake clientSet
func GetClientSet() (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(ctrl.GetConfigOrDie())
}
