package update

import (
	"context"
	"fmt"

	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/telemetry/kubernetes"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtimeschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Updater struct {
	Client dynamic.Interface
}

type UpdateVal struct {
	Changed bool
	Value   string
	Key     string
}

func (u *Updater) UpdateCeip(ctx context.Context, userOptIn bool) error {
	err := u.findOrCreateNamespace(ctx)
	if err != nil {
		return err
	}

	var ceipVal string
	if userOptIn {
		ceipVal = "standard"
	} else {
		ceipVal = "disabled"
	}

	ceipConfigMap := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": kubernetes.CeipConfigMapName,
			},
			"data": map[string]interface{}{
				"level": ceipVal,
			},
		},
	}

	_, err = u.Client.
		Resource(runtimeschema.GroupVersionResource{Version: "v1", Resource: "configmaps"}).
		Namespace(kubernetes.TelemetryNamespace).
		Get(ctx, kubernetes.CeipConfigMapName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = u.Client.
			Resource(runtimeschema.GroupVersionResource{Version: "v1", Resource: "configmaps"}).
			Namespace(kubernetes.TelemetryNamespace).
			Create(ctx, ceipConfigMap, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		_, err = u.Client.
			Resource(runtimeschema.GroupVersionResource{Version: "v1", Resource: "configmaps"}).
			Namespace(kubernetes.TelemetryNamespace).
			Update(ctx, ceipConfigMap, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *Updater) UpdateIdentifiers(ctx context.Context, identifierVals []UpdateVal) error {
	err := u.findOrCreateNamespace(ctx)
	if err != nil {
		return err
	}

	sharedIdsData := make(map[string]interface{})

	existingConfigMap, err := u.Client.
		Resource(runtimeschema.GroupVersionResource{Version: "v1", Resource: "configmaps"}).
		Namespace(kubernetes.TelemetryNamespace).
		Get(ctx, kubernetes.SharedIdsConfigMapName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		fmt.Println("did not find existing identifiers config map")
		sharedIdsConfigMap := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": kubernetes.SharedIdsConfigMapName,
				},
				"data": map[string]interface{}{},
			},
		}

		_, err = u.Client.
			Resource(runtimeschema.GroupVersionResource{Version: "v1", Resource: "configmaps"}).
			Namespace(kubernetes.TelemetryNamespace).
			Create(ctx, sharedIdsConfigMap, metav1.CreateOptions{})
		if err != nil {
			return err
		}

	} else if err != nil {
		return err
	} else {
		fmt.Println("found existing identifiers config map")
		if existingConfigMap.Object["data"] != nil {
			sharedIdsData = existingConfigMap.Object["data"].(map[string]interface{})
		}
	}

	for _, val := range identifierVals {
		if val.Changed {
			sharedIdsData[val.Key] = val.Value
		}
	}

	newIdsMap := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": kubernetes.SharedIdsConfigMapName,
			},
			"data": sharedIdsData,
		},
	}

	fmt.Println("Updating config map ....")
	_, err = u.Client.
		Resource(runtimeschema.GroupVersionResource{Version: "v1", Resource: "configmaps"}).
		Namespace(kubernetes.TelemetryNamespace).
		Update(ctx, newIdsMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (u *Updater) findOrCreateNamespace(ctx context.Context) error {
	_, err := u.Client.
		Resource(runtimeschema.GroupVersionResource{Version: "v1", Resource: "namespaces"}).
		Get(ctx, kubernetes.TelemetryNamespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		fmt.Println("did not find namespace, creating namespace")
		nsObj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": kubernetes.TelemetryNamespace,
				},
			},
		}

		_, err = u.Client.
			Resource(runtimeschema.GroupVersionResource{Version: "v1", Resource: "namespaces"}).
			Create(ctx, nsObj, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}
