/*
 * Copyright (c) 2021 VMware, Inc. All rights reserved.
 *
 * Proprietary and confidential.
 *
 * Unauthorized copying or use of this file, in any medium or form,
 * is strictly prohibited.
 */

package kubernetes

import (
	"net/http"
	"net/http/httptest"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// GetKubernetesClientServer creates an httptest.Server that allows k8s based code to be tested locally.
func GetKubernetesClientServer(h func(http.ResponseWriter, *http.Request)) (func() (dynamic.Interface, error), *httptest.Server, error) {
	cfgget, srv, err := GetConfigAndServer(h)
	if err != nil {
		return nil, nil, err
	}

	cfg, err := cfgget()
	if err != nil {
		srv.Close()
		return nil, nil, err
	}

	cl, err := dynamic.NewForConfig(cfg)
	if err != nil {
		srv.Close()
		return nil, nil, err
	}

	return func() (dynamic.Interface, error) { return cl, nil }, srv, nil
}

// GetConfigAndServer creates an httptest.Server that allows k8s based code to be tested locally
// Returns a function to acquire a rest.Config, leaving k8s client creation up to the caller
func GetConfigAndServer(h func(http.ResponseWriter, *http.Request)) (func() (*rest.Config, error), *httptest.Server, error) {
	srv := httptest.NewServer(http.HandlerFunc(h))
	c := &rest.Config{
		Host: srv.URL,
	}

	return func() (*rest.Config, error) { return c, nil }, srv, nil
}
