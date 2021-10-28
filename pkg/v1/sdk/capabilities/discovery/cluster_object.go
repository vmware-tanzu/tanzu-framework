// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

// Object represents any runtime.Object that could exist on a cluster, with ability to specify:
// WithAnnotations()
// WithLabels()
// WithConditions()
func Object(queryName string, obj *corev1.ObjectReference) *QueryObject {
	return &QueryObject{
		name:     queryName,
		object:   obj,
		presence: true,
	}
}

// QueryObject allows for resource querying
type QueryObject struct {
	name        string
	object      *corev1.ObjectReference
	annotations []resourceAnnotation
	presence    bool
	//	conditions []resourceCondition
}

// Name is the name of the query.
func (q *QueryObject) Name() string {
	return q.name
}

// WithoutAnnotations ensures lack of presence annotations on a resource
func (q *QueryObject) WithoutAnnotations(a map[string]string) *QueryObject {
	for k, v := range a {
		q.annotations = append(q.annotations, resourceAnnotation{
			key:      k,
			value:    v,
			presence: false,
		})
	}

	return q
}

// WithAnnotations matches annotations on a resource
func (q *QueryObject) WithAnnotations(a map[string]string) *QueryObject {
	for k, v := range a {
		q.annotations = append(q.annotations, resourceAnnotation{
			key:      k,
			value:    v,
			presence: true,
		})
	}

	return q
}

// Run the object discovery
func (q *QueryObject) Run(config *clusterQueryClientConfig) (bool, error) {
	groupResources, err := restmapper.GetAPIGroupResources(config.discoveryClientset)
	if err != nil {
		return false, err
	}

	// Ensure object presence or lack
	objectExists, err := q.QueryObjectExists(groupResources, config)
	if err != nil {
		return false, err
	}
	// Ensure the state of the resource matches intent
	if q.presence != objectExists {
		return false, nil
	}

	return true, nil
}

// QueryObjectExists uses dynamic and unstructured APIs to reason about object state
func (q *QueryObject) QueryObjectExists(resources []*restmapper.APIGroupResources, config *clusterQueryClientConfig) (bool, error) {
	u, err := q.objectExists(resources, config)
	if err != nil {
		return false, err
	}
	if u == nil {
		return false, nil
	}

	if !q.checkAnnotations(u) {
		return false, nil
	}

	return true, nil
}

func (q *QueryObject) objectExists(resources []*restmapper.APIGroupResources, config *clusterQueryClientConfig) (obj *unstructured.Unstructured, err error) {
	gvk := q.object.GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}

	rm := restmapper.NewDiscoveryRESTMapper(resources)
	mapping, err := rm.RESTMapping(gk, gvk.Version)
	if err != nil {
		return nil, err
	}

	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		dr = config.dynamicClient.Resource(mapping.Resource).Namespace(q.object.Namespace)
	} else {
		dr = config.dynamicClient.Resource(mapping.Resource)
	}

	o, err := dr.Get(context.Background(), q.object.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return o, nil
}

func (q *QueryObject) checkAnnotations(u *unstructured.Unstructured) bool {
	for _, v := range q.annotations {
		val, ok := u.GetAnnotations()[v.key]
		if ok {
			if !v.presence {
				return false
			}
			if v.value != "" && v.value != val {
				return false
			}
		} else if v.presence {
			return false
		}
	}
	return true
}

// Reason for failures, in a standard structure
func (q *QueryObject) Reason() string {
	return fmt.Sprintf("kind=%s status=unmatched presence=%t", q.object.Kind, q.presence)
}

func (q *QueryObject) annotationsMap(presence bool) map[string]string {
	annotations := make(map[string]string)
	for _, a := range q.annotations {
		if a.presence == presence {
			annotations[a.key] = a.value
		}
	}
	return annotations
}

type resourceAnnotation struct {
	key      string
	value    string
	presence bool
}
