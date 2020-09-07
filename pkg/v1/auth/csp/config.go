package csp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	authv1alpha1 "github.com/vmware-tanzu-private/core/apis/auth/v1alpha1"
	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
)

const (
	// LocalDirName is the name of the local directory in which tanzu state is stored.
	LocalDirName = "csp"

	// EnvConfigKey is the environment variable that points to a csp config.
	EnvConfigKey = "CSP_CONFIG"
)

// LocalDir returns the local directory in which tanzu state is stored.
func LocalDir() (path string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return path, errors.Wrap(err, "could not locate local tanzu dir")
	}
	path = filepath.Join(home, client.LocalDirName, LocalDirName)
	return
}

// ConfigPath returns the tanzu config path, checking for enviornment overrides.
func ConfigPath(name string) (path string, err error) {
	localDir, err := LocalDir()
	if err != nil {
		return path, err
	}
	var ok bool
	path, ok = os.LookupEnv(EnvConfigKey)
	if !ok {
		path = filepath.Join(localDir, name)
		return
	}
	return
}

// GetConfig retrieves the config from the local directory.
func GetConfig(name string) (cfg *authv1alpha1.CSPConfig, err error) {
	cfgPath, err := ConfigPath(name)
	if err != nil {
		return nil, err
	}
	return GetConfigFromPath(cfgPath)
}

// GetConfigFromPath gets the config from path.
func GetConfigFromPath(path string) (cfg *authv1alpha1.CSPConfig, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read Config file")
	}
	scheme, err := authv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scheme")
	}
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	var c authv1alpha1.CSPConfig
	_, _, err = s.Decode(b, nil, &c)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode Config file")
	}
	return &c, nil
}

// StoreConfig stores the config in the local directory.
func StoreConfig(cfg *authv1alpha1.CSPConfig) error {
	cfgPath, err := ConfigPath(cfg.GetName())
	if err != nil {
		return errors.Wrap(err, "could not find Config path")
	}

	_, err = os.Stat(cfgPath)
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
		return errors.Wrap(err, "could not create Config path")
	}

	scheme, err := authv1alpha1.SchemeBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to create scheme")
	}

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme, scheme)
	buf := new(bytes.Buffer)
	if err := s.Encode(cfg, buf); err != nil {
		return errors.Wrap(err, "failed to encode Config file")
	}
	// TODO (pbarker): need to consider races.
	if err = ioutil.WriteFile(cfgPath, buf.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "failed to write Config file")
	}
	return nil
}

// DeleteConfig deletes the config from the local directory.
func DeleteConfig(name string) error {
	cfgPath, err := ConfigPath(name)
	if err != nil {
		return err
	}
	err = os.Remove(cfgPath)
	if err != nil {
		return errors.Wrap(err, "could not remove Config")
	}
	return nil
}

// EndpointFromServer returns the endpoint from server.
func EndpointFromServer(s clientv1alpha1.Server) (endpoint string, err error) {
	switch s.Type {
	case clientv1alpha1.KubernetesServer:
		// TODO (pbarker): implement kubernetes server
		return endpoint, fmt.Errorf("type %q not yet implemented", s.Type)
	case clientv1alpha1.TanzuServer:
		cfg, err := GetConfigFromPath(s.Path)
		if err != nil {
			return endpoint, nil
		}
		return cfg.Spec.Endpoint, nil
	default:
		return endpoint, fmt.Errorf("unknown server type %q", s.Type)
	}
}
