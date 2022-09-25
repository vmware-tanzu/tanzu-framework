// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/vmware-tanzu/tanzu-framework/capabilities/controller/pkg/constants"
)

func TestGetConfigForServiceAccount(t *testing.T) {
	objs, secrets := getTestObjects()
	cl := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objs...).Build()
	ctx, cancel := context.WithTimeout(context.Background(), constants.ContextTimeout)
	defer cancel()
	testCases := []struct {
		description        string
		serviceAccountName string
		namespaceName      string
		client             client.Client
		host               string
		returnErr          bool
	}{
		{
			description:        "should successfully return config",
			serviceAccountName: "foo",
			namespaceName:      "default",
			client:             cl,
			host:               "localhost:31145",
			returnErr:          false,
		},
		{
			description:        "pass invalid service account",
			serviceAccountName: "bar",
			namespaceName:      "default",
			client:             cl,
			host:               "localhost:31146",
			returnErr:          true,
		},
		{
			description:        "pass service account name that doesn't exist",
			serviceAccountName: "baz",
			namespaceName:      "default",
			client:             cl,
			host:               "localhost:31147",
			returnErr:          true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			config, err := GetConfigForServiceAccount(ctx, cl, tc.namespaceName, tc.serviceAccountName, tc.host)
			if err != nil {
				if !tc.returnErr {
					t.Errorf("error not expected, but got error: %v", err)
				}
			} else if tc.returnErr {
				if err == nil {
					t.Errorf("error expected, but got nothing")
				}
			} else {
				if config.BearerToken != string(secrets[tc.serviceAccountName].Data[corev1.ServiceAccountTokenKey]) {
					t.Errorf("config object is not constructed properly")
				}
				res := bytes.Compare(config.TLSClientConfig.CAData, secrets[tc.serviceAccountName].Data[corev1.ServiceAccountRootCAKey])
				if res != 0 {
					t.Errorf("config object is not constructed properly")
				}
			}
		})
	}
}

func getTestObjects() ([]runtime.Object, map[string]*corev1.Secret) {
	fooSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"token":     []byte("eyJhbGciOiJSUzI1NiIsImtpZCI6InlSV3V1T3RFWDJvUDN0MGtGQ3BiUVRNUkd0SFotX0hvUHJaMEFuNGF4ZTAifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Im15LXNhLXRva2VuLWxuY3FwIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6Im15LXNhIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiOGIxMWUwZWMtYTE5Ny00YWMyLWFjNDQtODczZGJjNTMwNGJlIiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OmRlZmF1bHQ6bXktc2EifQ.je34lgxiMKgwD1Paac_T1FMXwWXCBfhcqXP0A6UEvOAzzOqXhiQGF7jhctRxXfQQITK4CkdVftanRR3OEDNMLU1PW5ulWxmU6SbC3vfJOz3-RO_pNVCfUo-eDinSuZnwn3s23ceOJ3r3bM8rpk0vYdX2EYPDb-2wxr23gTqR5qeNT-muq-jbKWTO0btXV_pTscNqWRFHW2AU5GaPincVUx5mq1-stHWN8kcLoz8_zKdgRrFa_vrQcosVg6BEnLHKv5m_THZGp2mO0bkHWruClDP7KsKL8UZeloL3xcwPkM4VPAoetl9y39oR-JmhwEIHq-a_psiXyXODAN8I72lFiQ%"),
			"namespace": []byte("default"),
			"ca.crt":    []byte("-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl\ncm5ldGVzMB4XDTIxMDgxNzE4NTg1MVoXDTMxMDgxNTE4NTg1MVowFTETMBEGA1UE\nAxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKbD\ndordCz868nGBmvu+IxsyFE4N8xFMm3k8K85omiqsKgr+z2GhiWaO1WHv+EjQFv9k\n1HSJ/jcoh1G7wtc0d1pZbsZWWscR2I50REKcTI55kES12NJg/oGK2fjSN4eQm3Ag\nFQTj0x+9FvB7ydJCZ9xGyMHh8yhGBc2JgU/aR8aeY2yrxnyCTLS1QxU0E1LxBpLg\nHV/JjSWFxULeDQtufP4lNl/Fmi9UFoEKgE5xwgLzTF3UpkCodKwdlVVcsc2VStof\n9peuA5nFXoA6dyHuG+PEDCKH72oYtN3I3NzFFn4ecCRVt2tWQOhsGI1lrEeg1KSS\n0eexH6a2uTv9jxfXG2sCAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB\n/wQFMAMBAf8wHQYDVR0OBBYEFAK8hmyS1mlR3ipnSRc/oz07QluJMA0GCSqGSIb3\nDQEBCwUAA4IBAQAfMmX+sPye95+wQAVglnwG8BEFboGb7Jbqd9Lm9i+AWjhVnrqX\nZdnnSh1o8xbuAqVv4etbZruh5wXJw57yMCoGg/K8jJwoGiFDFFHuzzotLt9gkrQv\n2tSazl1yUEDCZcwaymXiJIoqMIjr1bHaAp92ycgx++hSNszwyT3uKbuScYTODTFg\ntbFQy5UTbfvLOBF7DM5sRqpWGSu6fZw6yEgYPAYTmwLN9Cb08onohTouNF71Mobr\n7jdXWv2CLtq7VC3rqpWAO81idsrbx2kxL7UEAONaGNqPfmoZgi1S2jAxBygQrcMV\nsZAZ38xYxkCcxsBIKW+/kL6jramz7IJJYvg4\n-----END CERTIFICATE-----"),
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}

	fooServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Secrets: []corev1.ObjectReference{
			corev1.ObjectReference{
				Name: "foo",
			},
		},
	}

	barSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"namespace": []byte("default"),
			"ca.crt":    []byte("-----BEGIN CERTIFICATE-----\nMIIC5zCCAc+gAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl\ncm5ldGVzMB4XDTIxMDgxNzE4NTg1MVoXDTMxMDgxNTE4NTg1MVowFTETMBEGA1UE\nAxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKbD\ndordCz868nGBmvu+IxsyFE4N8xFMm3k8K85omiqsKgr+z2GhiWaO1WHv+EjQFv9k\n1HSJ/jcoh1G7wtc0d1pZbsZWWscR2I50REKcTI55kES12NJg/oGK2fjSN4eQm3Ag\nFQTj0x+9FvB7ydJCZ9xGyMHh8yhGBc2JgU/aR8aeY2yrxnyCTLS1QxU0E1LxBpLg\nHV/JjSWFxULeDQtufP4lNl/Fmi9UFoEKgE5xwgLzTF3UpkCodKwdlVVcsc2VStof\n9peuA5nFXoA6dyHuG+PEDCKH72oYtN3I3NzFFn4ecCRVt2tWQOhsGI1lrEeg1KSS\n0eexH6a2uTv9jxfXG2sCAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB\n/wQFMAMBAf8wHQYDVR0OBBYEFAK8hmyS1mlR3ipnSRc/oz07QluJMA0GCSqGSIb3\nDQEBCwUAA4IBAQAfMmX+sPye95+wQAVglnwG8BEFboGb7Jbqd9Lm9i+AWjhVnrqX\nZdnnSh1o8xbuAqVv4etbZruh5wXJw57yMCoGg/K8jJwoGiFDFFHuzzotLt9gkrQv\n2tSazl1yUEDCZcwaymXiJIoqMIjr1bHaAp92ycgx++hSNszwyT3uKbuScYTODTFg\ntbFQy5UTbfvLOBF7DM5sRqpWGSu6fZw6yEgYPAYTmwLN9Cb08onohTouNF71Mobr\n7jdXWv2CLtq7VC3rqpWAO81idsrbx2kxL7UEAONaGNqPfmoZgi1S2jAxBygQrcMV\nsZAZ38xYxkCcxsBIKW+/kL6jramz7IJJYvg4\n-----END CERTIFICATE-----"),
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}

	barServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "default",
		},
		Secrets: []corev1.ObjectReference{
			corev1.ObjectReference{
				Name: "bar",
			},
		},
	}

	secrets := map[string]*corev1.Secret{
		"foo": fooSecret,
		"bar": barSecret,
	}

	// Objects to track in the fake client.
	return []runtime.Object{fooServiceAccount, fooSecret, barServiceAccount, barSecret}, secrets
}
