package webhook

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"k8s.io/client-go/util/keyutil"
	"math/big"
	"time"

	"k8s.io/api/admissionregistration/v1"
	certutil "k8s.io/client-go/util/cert"
)

const (
	// CertificateBlockType is a possible value for pem.Block.Type.
	CertificateBlockType = "CERTIFICATE"
	// RSAPrivateKeyBlockType is a possible value for pem.Block.Type.
	RSAPrivateKeyBlockType = "RSA PRIVATE KEY"
	// The names of the files that should contain the CA certificate and the TLS key pair.
	CACertFile  = "ca.crt"
	TLSCertFile = "tls.crt"
	TLSKeyFile  = "tls.key"
)

type TLSCredentials struct {
	CAName          string
	Organization    string
	Hostname        string
	DnsNames        []string
	ValidFrom       time.Time
	MaxAge          time.Duration
	ca_cert_der     *[]byte
	ca_key_der      *rsa.PrivateKey
	server_cert_der *[]byte
	server_key_der  *rsa.PrivateKey
}

func pemEncoded(derBytes *[]byte, pemType string) io.Writer {
	buffer := bytes.Buffer{}
	if err := pem.Encode(&buffer, &pem.Block{Type: pemType, Bytes: *derBytes}); err != nil {
		return nil
	}
	return &buffer
}

func (tls *TLSCredentials) CAPEM() io.Writer {
	return pemEncoded(tls.ca_cert_der, CertificateBlockType)
}

func (tls *TLSCredentials) ServerPEM() io.Writer {
	return pemEncoded(tls.server_cert_der, CertificateBlockType)
}

func (tls *TLSCredentials) CARSAKeyPEM() io.Writer {
	key := x509.MarshalPKCS1PrivateKey(tls.ca_key_der)
	return pemEncoded(&key, RSAPrivateKeyBlockType)
}

func (tls *TLSCredentials) ServerRSAKeyPEM() io.Writer {
	key := x509.MarshalPKCS1PrivateKey(tls.server_key_der)
	return pemEncoded(&key, RSAPrivateKeyBlockType)
}

func NewTLSCredentials(caname string, organization string, hostname string, dnsnames []string, validfrom time.Time, maxage time.Duration) (*TLSCredentials, error) {
	var err error
	tls := TLSCredentials{
		CAName:       caname,
		Organization: organization,
		Hostname:     hostname,
		DnsNames:     dnsnames,
		ValidFrom:    validfrom,
		MaxAge:       maxage,
	}

	//create CA rsa key
	tls.ca_key_der, err = rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	//create self signed CA certificate
	caTemplate := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName:   tls.CAName,
			Organization: []string{tls.Organization},
		},
		NotBefore:             tls.ValidFrom.Add(-time.Hour), //we want the ca to be older than anything else
		NotAfter:              tls.ValidFrom.Add(tls.MaxAge),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	cert, err := x509.CreateCertificate(cryptorand.Reader, &caTemplate, &caTemplate, &tls.ca_key_der.PublicKey, tls.ca_key_der)
	if err != nil {
		return nil, err
	}

	tls.ca_cert_der = &cert

	//create server rsa key
	tls.server_key_der, err = rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	//Create and sign server certificate
	serverTemlate := x509.Certificate{
		DNSNames:     tls.DnsNames,
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("%s@%d", tls.Hostname, time.Now().Unix()),
		},
		NotBefore: tls.ValidFrom,
		NotAfter:  tls.ValidFrom.Add(tls.MaxAge),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	cacert, err := x509.ParseCertificate(*tls.ca_cert_der)
	if err != nil {
		return nil, err
	}
	cert, err = x509.CreateCertificate(cryptorand.Reader, &serverTemlate, cacert, &tls.server_key_der.PublicKey, tls.ca_key_der)
	if err != nil {
		return nil, err
	}
	tls.server_cert_der = &cert
	return &tls, nil
}

func (tls *TLSCredentials) writeFiles() error {
	//both ca and server certificates
	certBuffer := bytes.Buffer{}
	if err := pem.Encode(&certBuffer, &pem.Block{Type: CertificateBlockType, Bytes: *tls.server_cert_der}); err != nil {
		return err
	}
	if err := pem.Encode(&certBuffer, &pem.Block{Type: CertificateBlockType, Bytes: *tls.ca_cert_der}); err != nil {
		return err
	}
	// Server key
	keyBuffer := bytes.Buffer{}
	if err := pem.Encode(&keyBuffer, &pem.Block{Type: keyutil.RSAPrivateKeyBlockType, Bytes: x509.MarshalPKCS1PrivateKey(tls.server_key_der)}); err != nil {
		return err
	}

	if err := certutil.WriteCert(certfile_path(), certBuffer.Bytes()); err != nil {
		return err
	}
	if err := keyutil.WriteKey(keyfile_path(), keyBuffer.Bytes()); err != nil {
		return err
	}

	return nil
}

// from controller-rutime documentation at https://github.com/kubernetes-sigs/controller-runtime/blob/3dafc3c844685e4f672d9e07b88b68ae3f14a842/pkg/config/v1alpha1/types.go#L141
// CertDir is the directory that contains the server key and certificate.
// if not set, webhook server would look up the server key and certificate in
// {TempDir}/k8s-webhook-server/serving-certs. The server key and certificate
// must be named tls.key and tls.crt, respectively.

//We need to come up with the functions below but for now we can use the assumed defaults
func certfile_path() string {
	return "/tmp/k8s-webhook-server/serving-certs/tsl.crt"
}

func keyfile_path() string {
	return "/tmp/k8s-webhook-server/serving-certs/tsl.key"
}

func updateValidationWebhook(vWebhook *v1.ValidatingWebhookConfiguration, caCert []byte) bool {
	updated := false
	for idx, webhook := range vWebhook.Webhooks {
		if bytes.Equal(webhook.ClientConfig.CABundle, caCert) {
			continue
		}
		updated = true
		webhook.ClientConfig.CABundle = caCert
		vWebhook.Webhooks[idx] = webhook
	}
	return updated
}

func updateMutatingWebhook(mWebhook *v1.MutatingWebhookConfiguration, caCert []byte) bool {
	updated := false
	for idx, webhook := range mWebhook.Webhooks {
		if bytes.Equal(webhook.ClientConfig.CABundle, caCert) {
			continue
		}
		updated = true
		webhook.ClientConfig.CABundle = caCert
		mWebhook.Webhooks[idx] = webhook
	}
	return updated
}
