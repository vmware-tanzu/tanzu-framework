package status_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/telemetry/cmd/fakes"
	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/telemetry/kubernetes"
	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/telemetry/status"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestStatus(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)

		configMap := strings.Split(r.URL.Path, "/api/v1/namespaces/vmware-system-telemetry/configmaps/")[1]
		switch configMap {
		case "vmware-telemetry-identifiers":
			w.Write([]byte(`{
				"kind": "ConfigMap",
				"apiVersion": "v1",
				"metadata": {
					"name": "vmware-telemetry-identifiers",
					"namespace": "vmware-system-telemetry",
					"uid": "11gg33aa-ffff-zzzz-1234-123412341234",
					"resourceVersion": "99988866"
				},
				"data": {
					"customer_csp_org_id": "XXXX",
					"customer_entitlement_account_number": "XXXX",
					"env_is_prod": "true"
				}
			}`))
		case "vmware-telemetry-cluster-ceip":
			w.Write([]byte(`{
				"kind": "ConfigMap",
				"apiVersion": "v1",
				"metadata": {
					"name": "vmware-telemetry-cluster-ceip",
					"namespace": "vmware-system-telemetry",
					"uid": "11gg33aa-ffff-zzzz-1234-123412341234",
					"resourceVersion": "99988866"
				},
				"data": {
					"level": "standard"
				}
			}`))
		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	out := &fakes.FakeOutputWriter{}
	client, _ := clf()

	subject := &status.Printer{Client: client, Out: out}
	err = subject.PrintStatus()

	assert.NoError(t, err)
	assert.Equal(t, 1, out.AddRowCallCount())
	assert.Equal(t, 1, out.RenderCallCount())

	ceipMap, ok := out.AddRowArgsForCall(0)[0].(string)
	assert.True(t, ok)
	idsMap, ok := out.AddRowArgsForCall(0)[1].(string)
	assert.True(t, ok)

	assert.Contains(t, ceipMap, "level: standard")

	assert.Contains(t, idsMap, `customer_csp_org_id: XXXX`)
	assert.Contains(t, idsMap, `customer_entitlement_account_number: XXXX`)
	assert.Contains(t, idsMap, `env_is_prod: "true"`)
}

func TestStatus_IgnoreNotFoundErrors(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)

		configMap := strings.Split(r.URL.Path, "/api/v1/namespaces/vmware-system-telemetry/configmaps/")[1]
		switch configMap {
		case "vmware-telemetry-cluster-ceip":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "ceip configmap not found",
					"reason": "NotFound",
					"details": {
						"name": "vmware-telemetry-cluster-ceip",
						"group": "",
						"kind": "configmap"
					},
					"code": 404}`))
		case "vmware-telemetry-identifiers":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "ids configmap not found",
					"reason": "NotFound",
					"details": {
						"name": "vmware-telemetry-identifiers",
						"group": "",
						"kind": "configmap"
					},
					"code": 404}`))
		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	out := &fakes.FakeOutputWriter{}
	client, _ := clf()

	subject := &status.Printer{Client: client, Out: out}
	err = subject.PrintStatus()

	assert.NoError(t, err)
	assert.Equal(t, 1, out.AddRowCallCount())
	assert.Equal(t, 1, out.RenderCallCount())

	ceipErrMsg, ok := out.AddRowArgsForCall(0)[0].(string)
	assert.True(t, ok)
	idsErrMsg, ok := out.AddRowArgsForCall(0)[1].(string)
	assert.True(t, ok)

	assert.Contains(t, ceipErrMsg, "vmware-system-telemetry/vmware-telemetry-cluster-ceip configmap resource not found")
	assert.Contains(t, idsErrMsg, "vmware-system-telemetry/vmware-telemetry-identifiers configmap resource not found")
}

func TestStatus_CeipConfigMapError(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)

		configMap := strings.Split(r.URL.Path, "/api/v1/namespaces/vmware-system-telemetry/configmaps/")[1]
		switch configMap {
		case "vmware-telemetry-cluster-ceip":
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "ceip configmap forbidden",
					"reason": "Forbidden",
					"details": {
						"name": "vmware-telemetry-cluster-ceip",
						"group": "",
						"kind": "configmap"
					},
					"code": 403}`))
		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	out := &fakes.FakeOutputWriter{}
	client, _ := clf()

	subject := &status.Printer{Client: client, Out: out}
	err = subject.PrintStatus()
	if err == nil {
		t.Logf("err is not nil")
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "forbidden")
	assert.Equal(t, 0, out.AddRowCallCount())
	assert.Equal(t, 0, out.RenderCallCount())
}

func TestStatus_IdsConfigMapError(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)

		configMap := strings.Split(r.URL.Path, "/api/v1/namespaces/vmware-system-telemetry/configmaps/")[1]
		switch configMap {
		case "vmware-telemetry-cluster-ceip":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "ceip configmap not found",
					"reason": "NotFound",
					"details": {
						"name": "vmware-telemetry-cluster-ceip",
						"group": "",
						"kind": "configmap"
					},
					"code": 404}`))

		case "vmware-telemetry-identifiers":
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "ids configmap forbidden",
					"reason": "Forbidden",
					"details": {
						"name": "vmware-telemetry-identifiers",
						"group": "",
						"kind": "configmap"
					},
					"code": 403}`))
		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	out := &fakes.FakeOutputWriter{}
	client, _ := clf()

	subject := &status.Printer{Client: client, Out: out}
	err = subject.PrintStatus()

	assert.Error(t, err)
	if err == nil {
		t.Logf("err is not nil")
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "forbidden")
	assert.Equal(t, 0, out.AddRowCallCount())
	assert.Equal(t, 0, out.RenderCallCount())
}
