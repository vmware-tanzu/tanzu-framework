package kubernetes

import (
	"errors"
	"os"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// GetDynamicClient gets a dynamic client for the targeted cluster
// We expect a KUBECONFIG env var to be set for cluster targeting
func GetDynamicClient() (dynamic.Interface, error) {
	path := os.Getenv("KUBECONFIG")
	if len(path) == 0 {
		return nil, errors.New("KUBECONFIG must be set for the telemetry command")
	}

	clientConfig, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return dynamicClient, nil
}
