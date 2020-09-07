package client

import clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"

// GetClient from config.
func GetClient(cfg clientv1alpha1.Config) error {
	_, err := GetCurrentServer()
	if err != nil {
		return err
	}

	return nil
}
