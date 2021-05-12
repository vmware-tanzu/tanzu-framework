/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package helper contains helper function for unit tests
package helper

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

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

// GetFakePinnipedInfo returns the pinniped-info configmap
func GetFakePinnipedInfo(clustername, issuer, issuerCA string) string {
	pinnipedInfoJSON := `
	{
		"kind": "ConfigMap",
		"apiVersion": "v1",
		"metadata": {
	  	  "name": "pinniped-info",
	  	  "namespace": "kube-public"
		},
		"data": {
		  "cluster_name": "%s",
		  "issuer": "%s",
		  "issuer_ca_bundle_data": "%s"
		}
	}`
	pinnipedInfoJSON = fmt.Sprintf(pinnipedInfoJSON, clustername, issuer, issuerCA)
	return pinnipedInfoJSON
}
