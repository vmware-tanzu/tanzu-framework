// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package testutil implements helper utilities used in tests.
package testutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/tools/go/packages"

	"github.com/onsi/ginkgo"
	adminregv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"k8s.io/apimachinery/pkg/util/wait"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"knative.dev/pkg/webhook/certificates/resources"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/webhooks"
)

// CreateResources using unstructured objects from a yaml/json file provided by decoder
func CreateResources(f *os.File, cfg *rest.Config, dynamicClient dynamic.Interface) error {
	var err error
	data, err := os.ReadFile(f.Name())
	if err != nil {
		return err
	}
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(data), 100)
	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		return err
	}
	for {
		resource, unstructuredObj, err := getResource(decoder, mapper, dynamicClient)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		_, err = resource.Create(context.Background(), unstructuredObj, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// parseObjects gets a decoder and mapper for reading unstructured objects from a yaml/json file.
func parseObjects(f *os.File, cfg *rest.Config) (*yamlutil.YAMLOrJSONDecoder, meta.RESTMapper, error) {
	data, err := os.ReadFile(f.Name())
	if err != nil {
		return nil, nil, err
	}

	dataReader := bytes.NewReader(data)
	decoder := yamlutil.NewYAMLOrJSONDecoder(dataReader, 100)
	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)

	return decoder, mapper, err
}

// DeleteResources using unstructured objects from a yaml/json file provided by decoder.
func DeleteResources(f *os.File, cfg *rest.Config, dynamicClient dynamic.Interface, waitForDeletion bool) error {
	deletionPropagation := metav1.DeletePropagationForeground
	gracePeriodSeconds := int64(0)

	decoder, mapper, err := parseObjects(f, cfg)
	if err != nil {
		return err
	}

	for {
		resource, unstructuredObj, err := getResource(decoder, mapper, dynamicClient)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		if err := resource.Delete(context.Background(), unstructuredObj.GetName(),
			metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds,
				PropagationPolicy: &deletionPropagation}); err != nil {
			return err
		}
	}

	if waitForDeletion {
		// verify deleted
		decoder, mapper, err := parseObjects(f, cfg)
		if err != nil {
			return err
		}

		for {
			resource, unstructuredObj, err := getResource(decoder, mapper, dynamicClient)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return err
				}
			}

			fmt.Fprintln(ginkgo.GinkgoWriter, "wait for deletion", unstructuredObj.GetName())
			if err := wait.Poll(time.Second*5, time.Second*10, func() (done bool, err error) {
				obj, err := resource.Get(context.Background(), unstructuredObj.GetName(), metav1.GetOptions{})

				if err == nil {
					fmt.Fprintln(ginkgo.GinkgoWriter, "remove finalizers", obj.GetFinalizers(), unstructuredObj.GetName())
					obj.SetFinalizers(nil)
					_, err = resource.Update(context.Background(), obj, metav1.UpdateOptions{})
					if err != nil {
						return false, err
					}
					return false, nil
				}
				if apierrors.IsNotFound(err) {
					return true, nil
				}
				return false, err
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// EnsureResources verifies that resources exist, creating it if necessary
func EnsureResources(f *os.File, cfg *rest.Config, dynamicClient dynamic.Interface) error {
	decoder, mapper, err := parseObjects(f, cfg)
	if err != nil {
		return err
	}

	for {
		resource, unstructuredObj, err := getResource(decoder, mapper, dynamicClient)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		_, err = resource.Get(context.Background(), unstructuredObj.GetName(), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// create it
			if _, err := resource.Create(context.Background(), unstructuredObj, metav1.CreateOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateKubeconfigSecret create a secret with kubeconfig token for the cluster provided by client
func CreateKubeconfigSecret(cfg *rest.Config, clusterName, namespace string, crClient client.Client) error {
	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters[clusterName] = &clientcmdapi.Cluster{
		Server:                   cfg.Host,
		CertificateAuthorityData: cfg.CAData,
	}
	contexts := make(map[string]*clientcmdapi.Context)
	contextName := fmt.Sprintf("%s@%s", cfg.Username, clusterName)
	contexts[contextName] = &clientcmdapi.Context{
		Cluster:   clusterName,
		Namespace: namespace,
		AuthInfo:  cfg.Username,
	}
	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos[cfg.Username] = &clientcmdapi.AuthInfo{
		ClientKeyData:         cfg.KeyData,
		ClientCertificateData: cfg.CertData,
	}
	clientConfig := clientcmdapi.Config{
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: contextName,
		AuthInfos:      authinfos,
	}
	kubeconfig, err := clientcmd.Write(clientConfig)
	if err != nil {
		return err
	}
	kc := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name(clusterName, secret.Kubeconfig),
			Namespace: namespace,
			Labels: map[string]string{
				clusterapiv1beta1.ClusterLabelName: clusterName,
			},
		},
		Data: map[string][]byte{
			secret.KubeconfigDataName: kubeconfig,
		},
		Type: clusterapiv1beta1.ClusterSecretType,
	}

	return crClient.Create(context.Background(), kc)
}

func getResource(decoder *yamlutil.YAMLOrJSONDecoder, mapper meta.RESTMapper, dynamicClient dynamic.Interface) (
	dynamic.ResourceInterface, *unstructured.Unstructured, error) { // nolint:whitespace
	var rawObj runtime.RawExtension
	if err := decoder.Decode(&rawObj); err != nil {
		return nil, nil, err
	}

	obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, nil, err
	}

	unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, nil, err
	}

	var resource dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if unstructuredObj.GetNamespace() == "" {
			unstructuredObj.SetNamespace("default")
		}
		resource = dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
	} else {
		resource = dynamicClient.Resource(mapping.Resource)
	}
	return resource, unstructuredObj, nil
}

// GetExternalCRDPaths gets paths for external CRDs by introspecting versions of the go dependencies
func GetExternalCRDPaths(externalDeps map[string][]string) ([]string, error) {
	packageConfig := &packages.Config{
		Mode: packages.NeedModule,
	}
	var crdPaths []string
	for dep, crdDirs := range externalDeps {
		for _, crdDir := range crdDirs {
			pkgs, err := packages.Load(packageConfig, dep)
			if err != nil {
				return nil, err
			}

			pkg := pkgs[0]
			if pkg.Errors != nil {
				errs := []error{}
				for _, err := range pkg.Errors {
					errs = append(errs, err)
				}
				return nil, utilerrors.NewAggregate(errs)
			}
			crdPaths = append(crdPaths, filepath.Join(pkg.Module.Dir, crdDir))
		}
	}

	logf.Log.Info("external CRD paths", "crdPaths", crdPaths)
	return crdPaths, nil
}

type WebhookCertificatesDetails struct {
	CertPath           string
	KeyPath            string
	WebhookScrtName    string
	AddonNamespace     string
	WebhookServiceName string
	LabelSelector      labels.Selector
}

func SetupWebhookCertificates(ctx context.Context, k8sClient client.Client, k8sConfig *rest.Config, certDetails *WebhookCertificatesDetails) error {
	scrt, err := webhooks.InstallNewCertificates(ctx, k8sConfig, certDetails.CertPath, certDetails.KeyPath,
		certDetails.WebhookScrtName, certDetails.AddonNamespace, certDetails.WebhookServiceName, certDetails.LabelSelector.String())
	if err != nil {
		return err
	}
	vwcfgs := &adminregv1.ValidatingWebhookConfigurationList{}
	err = k8sClient.List(ctx, vwcfgs, &client.ListOptions{LabelSelector: certDetails.LabelSelector})
	if err != nil {
		return err
	}
	for i := range vwcfgs.Items {
		wcfg := vwcfgs.Items[i]
		for j := range wcfg.Webhooks {
			whook := wcfg.Webhooks[j]
			if !bytes.Equal(whook.ClientConfig.CABundle, scrt.Data[resources.CACert]) {
				return fmt.Errorf("validating Webhook CA Bundlle is not updated correctly")
			}
		}
	}

	mwcfgs := &adminregv1.MutatingWebhookConfigurationList{}
	err = k8sClient.List(ctx, mwcfgs, &client.ListOptions{LabelSelector: certDetails.LabelSelector})
	if err != nil {
		return err
	}
	for i := range mwcfgs.Items {
		wcfg := mwcfgs.Items[i]
		for j := range wcfg.Webhooks {
			whook := wcfg.Webhooks[j]
			if !bytes.Equal(whook.ClientConfig.CABundle, scrt.Data[resources.CACert]) {
				return fmt.Errorf("mutating Webhook CA Bundlle is not updated correctly")
			}
		}
	}
	return nil
}
