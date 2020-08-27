package client

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"

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
		return path, errors.Wrap(err, "could not locate local tanzu dir")
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
		return nil, errors.Wrap(err, "failed to read context file")
	}
	scheme, err := clientv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scheme")
	}
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	var c clientv1alpha1.Context
	_, _, err = s.Decode(b, nil, &c)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode context file")
	}
	return &c, nil
}

// StoreContext stores the context in the local directory.
func StoreContext(ctx *clientv1alpha1.Context) error {
	ctxPath, err := ContextPath()
	if err != nil {
		return errors.Wrap(err, "could not find context path")
	}

	_, err = os.Stat(ctxPath)
	if os.IsNotExist(err) {
		localDir, err := LocalDir()
		if err != nil {
			return errors.Wrap(err, "could not find local tanzu dir for OS")
		}
		err = os.MkdirAll(localDir, 0755)
		if err != nil {
			return errors.Wrap(err, "could not make local tanzu directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "could not create context path")
	}

	scheme, err := clientv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to create scheme")
	}

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	buf := new(bytes.Buffer)
	if err := s.Encode(ctx, buf); err != nil {
		return errors.Wrap(err, "failed to encode context file")
	}
	// TODO (pbarker): need to consider races.
	if err = ioutil.WriteFile(ctxPath, buf.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "failed to write context file")
	}
	return nil
}

// DeleteContext deletes the context from the local directory.
func DeleteContext() error {
	ctxPath, err := ContextPath()
	if err != nil {
		return err
	}
	err = os.Remove(ctxPath)
	if err != nil {
		return errors.Wrap(err, "could not remove context")
	}
	return nil
}
