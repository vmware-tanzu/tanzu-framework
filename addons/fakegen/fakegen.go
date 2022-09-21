package fakegen

import (
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate counterfeiter -o ../pkg/fakeclusterclient/crtclusterclient.go --fake-name CRTClusterClient . CrtClient

// CrtClient clientset interface
type CrtClient interface {
	crtclient.Client
}
