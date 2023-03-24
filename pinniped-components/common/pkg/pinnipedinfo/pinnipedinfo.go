// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pinnipedinfo

// PinnipedInfo contains settings for the supervisor.
type PinnipedInfo struct {
	ClusterName        string `json:"cluster_name"`
	Issuer             string `json:"issuer"`
	IssuerCABundleData string `json:"issuer_ca_bundle_data"`

	// ConciergeEndpoint does not appear to be set anywhere in tanzu-framework.
	// It appears that `pinniped kubeconfig get` will autodetect this endpoint from the current Kubeconfig context,
	// when someone invokes `tanzu pinniped-auth login` via a Kubeconfig.
	// See https://github.com/vmware-tanzu/pinniped/blob/77041760ccf3747972faa9b029fb85f0cb2b592c/cmd/pinniped/cmd/kubeconfig.go#L428-L436
	ConciergeEndpoint string `json:"concierge_endpoint,omitempty"`
}
