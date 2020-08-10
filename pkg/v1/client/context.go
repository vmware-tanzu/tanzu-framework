package client

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client.tanzu.cloud.vmware.com/v1alpha1"
)

const (
	// LocalDir is the default local directory in which contexts are stored.
	LocalDir = "~/.tanzu"

	// ContextName is the name of the context
	ContextName = "context.yaml"
)

// GetContext retrieves the context from the local directory.
func GetContext() (ctx *clientv1alpha1.Context, err error) {
	b, err := ioutil.ReadFile(contextPath)
	if err != nil {
		return nil, err
	}
	return yaml.Unmarshal(b, ctx)
}

// StoreContext stores the context in the local directory.
func StoreContext(ctx *clientv1alpha1.Context) error {
	f, err := os.Stat(contextPath())
	if os.IsNotExist(err) {
		err := os.MkdirAll(LocalDir, 0755)
	} else {
		return err
	}
	b, err := yaml.Marshal(ctx)
	if err != nil {
		return nil, err
	}
	return ioutil.WriteFile(contextPath, b)
}

// DeleteContext deletes the context from the local directory.
func DeleteContext() error {
	return os.Remove(contextPath())
}

func contextPath() string {
	return filepath.Join(LocalDir, ContextName)
}
