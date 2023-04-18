// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package certs

import "time"

const (
	// CACertName is the name of the CA certificate.
	CACertName = "ca.crt"
	// ServerCertName is the name of the serving certificate.
	ServerCertName = "tls.crt"
	// ServerKeyName is the name of the server private key.
	ServerKeyName = "tls.key"

	// certExpirtationBuffer specifies the amount of time in addition to the
	// rotation interval that generated certificates expire.
	certExpirtationBuffer = time.Minute * 30

	// defaultRotationInterval is the default interval at which certificates
	// are rotated. This value is used if the webhook server secret is missing
	// the annotation that specifies the rotation interval.
	defaultRotationInterval = time.Hour * 24
)
