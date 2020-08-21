package client

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client.tanzu.cloud.vmware.com/v1alpha1"
)

const (
	// EnvContextKey is the environment variable that points to a tanzu context.
	EnvContextKey = "TANZU_CONTEXT"

	// LocalDirName is the name of the local directory in which tanzu state is stored.
	LocalDirName = ".tanzu"

	// ContextName is the name of the context
	ContextName = "context.yaml"
)

// LocalDir returns the local directory in which tanzu state is stored.
func LocalDir() (path string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return path, err
	}
	path = filepath.Join(home, LocalDirName)
	return
}

// ContextPath returns the tanzu context path, checking for enviornment overrides.
func ContextPath() (path string, err error) {
	localDir, err := LocalDir()
	if err != nil {
		return path, err
	}
	var ok bool
	path, ok = os.LookupEnv(EnvContextKey)
	if !ok {
		path = filepath.Join(localDir, ContextName)
		return
	}
	return
}

// GetContext retrieves the context from the local directory.
func GetContext() (ctx *clientv1alpha1.Context, err error) {
	ctxPath, err := ContextPath()
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(ctxPath)
	if err != nil {
		return nil, err
	}
	// TODO (pbarker): this needs the k8s serializer.
	err = yaml.Unmarshal(b, ctx)
	return
}

// StoreContext stores the context in the local directory.
func StoreContext(ctx *clientv1alpha1.Context) error {
	ctxPath, err := ContextPath()
	if err != nil {
		return err
	}
	_, err = os.Stat(ctxPath)
	if os.IsNotExist(err) {
		localDir, err := LocalDir()
		if err != nil {
			return err
		}
		err = os.MkdirAll(localDir, 0755)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	// TODO (pbarker): needs k8s serializer.
	b, err := yaml.Marshal(ctx)
	if err != nil {
		return err
	}

	// TODO (pbarker): need to consider races.
	return ioutil.WriteFile(ctxPath, b, 0644)
}

// DeleteContext deletes the context from the local directory.
func DeleteContext() error {
	ctxPath, err := ContextPath()
	if err != nil {
		return err
	}
	return os.Remove(ctxPath)
}
