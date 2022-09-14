// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package helper implements helper functions used for unit tests
package helper

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
)

// ###################### Fake CAPI objects creation helper ######################

// GetFakeClusterInfo returns the cluster-info configmap
func GetFakeClusterInfo(server string, cert *x509.Certificate) string {
	clusterInfoJSON := `
	{
		"kind": "ConfigMap",
		"apiVersion": "v1",
    	"data": {
        "kubeconfig": "apiVersion: v1\nclusters:\n- cluster:\n    certificate-authority-data: %s\n    server: %s\n  name: \"\"\ncontexts: null\ncurrent-context: \"\"\nkind: Config\npreferences: {}\nusers: null\n"
    	},
		"metadata": {
		  "name": "cluster-info",
		  "namespace": "kube-public"
		}
	}`
	certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	clusterInfoJSON = fmt.Sprintf(clusterInfoJSON, base64.StdEncoding.EncodeToString(certBytes), server)

	return clusterInfoJSON
}

// PinnipedInfo contains settings for the supervisor.
type PinnipedInfo struct {
	ClusterName              string `json:"cluster_name"`
	ConciergeEndpoint        string `json:"concierge_endpoint"`
	Issuer                   string `json:"issuer"`
	IssuerCABundleData       string `json:"issuer_ca_bundle_data"`
	ConciergeIsClusterScoped bool   `json:"concierge_is_cluster_scoped,string"`
}

// GetFakePinnipedInfo returns the pinniped-info configmap
func GetFakePinnipedInfo(pinnipedInfo PinnipedInfo) string {
	data, err := json.Marshal(pinnipedInfo)
	if err != nil {
		err = fmt.Errorf("could not marshal Pinniped info into JSON: %w", err)
	}

	pinnipedInfoJSON := `
	{
		"kind": "ConfigMap",
		"apiVersion": "v1",
		"metadata": {
	  	  "name": "pinniped-info",
	  	  "namespace": "kube-public"
		},
		"data": %s
	}`
	pinnipedInfoJSON = fmt.Sprintf(pinnipedInfoJSON, string(data))
	return pinnipedInfoJSON
}

// NewCLIPlugin returns new NewCLIPlugin object
func NewCLIPlugin(options TestCLIPluginOption) v1alpha1.CLIPlugin {
	artifacts := []v1alpha1.Artifact{
		{
			Image: "fake.image.repo.com/tkg/plugin/test-darwin-plugin:v1.4.0",
			OS:    "darwin",
			Arch:  "amd64",
		},
		{
			Image: "fake.image.repo.com/tkg/plugin/test-linux-plugin:v1.4.0",
			OS:    "linux",
			Arch:  "amd64",
		},
		{
			Image: "fake.image.repo.com/tkg/plugin/test-windows-plugin:v1.4.0",
			OS:    "windows",
			Arch:  "amd64",
		},
	}
	cliplugin := v1alpha1.CLIPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.Name,
		},
		Spec: v1alpha1.CLIPluginSpec{
			Description:        options.Description,
			RecommendedVersion: options.RecommendedVersion,
			Artifacts: map[string]v1alpha1.ArtifactList{
				"v1.0.0": artifacts,
			},
		},
	}
	return cliplugin
}
