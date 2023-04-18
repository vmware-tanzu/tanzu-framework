# Certificate Manager Lite

This package provides lightweight certificate management functionality for a controller manager's webhook server. It is
intended for developers who want to avoid depending on [cert-manager](https://cert-manager.io) just to manage
self-signed certificates for webhooks.

Certificate Manager is designed to be invoked in your controller manager. It runs in the background as a goroutine,
rotates certificates required by the webhook server in a secret and keeps the
`{Mutating,Validating}WebhookConfiguration`'s `caBundle` up-to-date.

## Usage

Certificate Manager requires the following setup to be used by a controller manager.

#### Create an empty secret to hold webhook certs

Certificate Manager updates the certificate data to this secret whenever a rotation occurs.

```yaml
apiVersion: v1
kind: Secret
metadata:
  annotations:
    tanzu.vmware.com/foo-webhook-rotation-interval: 6h
  name: tanzu-foo-webhook-server-cert
  namespace: default
type: Opaque
```

By default, Certificate Manager rotates certificates every 24 hours (plus a 30m grace period). To customize the rotation
interval, use an annotation like `tanzu.vmware.com/foo-webhook-rotation-interval: 6h` above.

The [Start Certificate Manager in the Controller Manager](#start-certificate-manager-in-the-controller-manager) section
shows how to configure Certificate Manager to look for this annotation.

#### Add a label to the WebhookConfiguration objects

Add a label of your choosing to the `{Mutating,Validating}WebhookConfiguration` objects to denote that their
certificates are being managed by the Certificate Manager (the example below
uses `tanzu.vmware.com/foo-webhook-managed-certs: "true"` label). This label will be used by the Certificate Manager to
select these `{Mutating,Validating}WebhookConfiguration` objects and write `caBundle` to them.

The [Start Certificate Manager in the Controller Manager](#start-certificate-manager-in-the-controller-manager) section
shows how to configure Certificate Manager to look for this label.

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: tanzu-foo-validating-webhook-core
  labels:
    tanzu.vmware.com/foo-webhook-managed-certs: "true"
webhooks:
...
```

#### Configure your Controller Manager deployment

1. Mount the secret created above to the controller manager pod.
2. Pass the configuration necessary for the Certificate Manager as arguments to the controller manager binary. You will
   need to declare appropriate flags in the controller manager for passing arguments. Certificate Manager requires the
   following configuration options:
    * `{Mutating,Validating}WebhookConfiguration` label (added in the previous step).
    * Namespace of the webhook `Service`.
    * Name of the webhook `Service`.
    * Namespace of the webhook `Secret` (created above).
    * Name of the webhook `Secret`.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: tanzu-foo-manager
  name: tanzu-foo-controller-manager
  namespace: default
spec:
  ...
  spec:
    ...
    containers:
        - image: foo-controller-manager:latest
          name: manager
          args:
            - "--webhook-config-label=tanzu.vmware.com/foo-webhook-managed-certs=true"
            - "--webhook-service-namespace=default"
            - "--webhook-service-name=tanzu-foo-webhook-service"
            - "--webhook-secret-namespace=default"
            - "--webhook-secret-name=tanzu-foo-webhook-server-cert"
          volumeMounts:
            - mountPath: /tmp/k8s-webhook-server/serving-certs
              name: cert
              readOnly: true
          ...
    volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: tanzu-foo-webhook-server-cert
```

3. Add RBAC rules to your controller manager deployment to read and write `Secret`, `MutatingWebhookConfiguration` and
   `ValidatingWebhookConfiguration` objects.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
  name: tanzu-foo-manager-clusterrole
rules:
  - apiGroups:
      - ""
    resources:
      - secrets
    verbs:
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - "admissionregistration.k8s.io"
    resources:
      - mutatingwebhookconfigurations
      - validatingwebhookconfigurations
    verbs:
      - get
      - list
      - patch
      - update
      - watch
```

#### Start Certificate Manager in the Controller Manager

In the controller manager's `main.go`, initialize a `CertificateManager` object and invoke the `Start` method to start
certificate management.

```go
package main

import "github.com/vmware-tanzu/tanzu-framework/util/webhook/certs"

func main() {
	// Declare flags.
	flag.StringVar(&webhookConfigLabel, "webhook-config-label", defaultWebhookConfigLabel, "The label used to select webhook configurations to update the certs for.")
	flag.StringVar(&webhookServiceNamespace, "webhook-service-namespace", defaultWebhookServiceNamespace, "The namespace in which webhook service is installed.")
	flag.StringVar(&webhookServiceName, "webhook-service-name", defaultWebhookServiceName, "The name of the webhook service.")
	flag.StringVar(&webhookSecretNamespace, "webhook-secret-namespace", defaultWebhookSecretNamespace, "The namespace in which webhook secret is installed.")
	flag.StringVar(&webhookSecretName, "webhook-secret-name", defaultWebhookSecretName, "The name of the webhook secret.")

	// Initialize CertificateManager.
	certManagerOpts := certs.Options{
		Logger:                        ctrl.Log.WithName("foo-webhook-cert-manager"),
		CertDir:                       webhookSecretVolumeMountPath,
		WebhookConfigLabel:            webhookConfigLabel,
		RotationIntervalAnnotationKey: "tanzu.vmware.com/foo-webhook-rotation-interval",
		NextRotationAnnotationKey:     "tanzu.vmware.com/foo-webhook-next-rotation",
		RotationCountAnnotationKey:    "tanzu.vmware.com/featuregates-webhook-rotation-count",
		SecretName:                    webhookSecretName,
		SecretNamespace:               webhookSecretNamespace,
		ServiceName:                   webhookServiceName,
		ServiceNamespace:              webhookServiceNamespace,
	}

	// Other setup code...

	// Initialize certificate manager.
	signalHandler := ctrl.SetupSignalHandler()

	certManager, err := certs.New(certManagerOpts)
	if err != nil {
		log.Error(err, "failed to create certificate manager")
		os.Exit(1)
	}

	// Start cert manager.
	if err := certManager.Start(signalHandler); err != nil {
		log.Error(err, "failed to start certificate manager")
		os.Exit(1)
	}

	// Wait for cert dir to be ready.
	if err := certManager.WaitForCertDirReady(); err != nil {
		log.Error(err, "certificates not ready")
		os.Exit(1)
	}

	// Start controller manager.
	if err := mgr.Start(signalHandler); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
```
