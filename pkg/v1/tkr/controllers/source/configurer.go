package source

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	configMapName   = "tkr-controller-config"
	caCertsKey      = "caCerts"
	systemCertsFile = "/etc/pki/tls/certs/ca-bundle.crt"
)

func (r *reconciler) Configure() error {
	configMap := &corev1.ConfigMap{}
	err := r.client.Get(context.Background(), types.NamespacedName{Namespace: constants.TKRNamespace, Name: configMapName}, configMap)
	// Not configure anything if the ConfigMap is not found
	if k8serr.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return errors.Wrapf(err, "unable to find the ConfigMap %s", configMapName)
	}
	err = addTrustedCerts(configMap.Data[caCertsKey])
	if err != nil {
		return errors.Wrap(err, "failed to add certs")
	}

	return nil
}

func addTrustedCerts(certChain string) (err error) {
	if certChain == "" {
		return nil
	}

	var file *os.File
	file, err = os.OpenFile(systemCertsFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return errors.Wrap(err, "failed to open certs file")
	}

	_, err = file.Write([]byte("\n" + certChain))
	if err != nil {
		_ = file.Close()
		return err
	}

	return file.Close()
}
