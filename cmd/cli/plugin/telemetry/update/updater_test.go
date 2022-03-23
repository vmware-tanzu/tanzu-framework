package update_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/telemetry/update"

	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/telemetry/kubernetes"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestUpdateCeip_NoPreviousConfiguration(t *testing.T) {
	var namespaceCreated, ceipUpdated bool
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		switch r.URL.Path {
		case "/api/v1/namespaces/vmware-system-telemetry":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "namespace not found",
					"reason": "NotFound",
					"details": {
						"name": "vmware-system-telemetry",
						"group": "",
						"kind": "namespace"
					},
					"code": 404}`))
		case "/api/v1/namespaces":
			assert.Equal(t, r.Method, http.MethodPost)
			namespaceCreated = true
			buf, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(buf), `"name":"vmware-system-telemetry"`)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				  "kind": "Namespace",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-system-telemetry"
				  }
				}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-cluster-ceip":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "configmap not found",
					"reason": "NotFound",
					"details": {
						"name": "vmware-telemetry-cluster-ceip",
						"group": "",
						"kind": "configmap"
					},
					"code": 404}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps":
			assert.Equal(t, r.Method, http.MethodPost)
			buf, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(buf), `"level":"standard"`)
			ceipUpdated = true
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				  "kind": "ConfigMap",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-telemetry-cluster-ceip"
				  }
				}`))

		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	err = subject.UpdateCeip(context.Background(), true)
	assert.NoError(t, err)
	assert.True(t, namespaceCreated)
	assert.True(t, ceipUpdated)
}

func TestUpdateCeip_CreateConfigMapError(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		switch r.URL.Path {
		case "/api/v1/namespaces/vmware-system-telemetry":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				  "kind": "Namespace",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-system-telemetry"
				  }
				}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-cluster-ceip":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "configmap not found",
					"reason": "NotFound",
					"details": {
						"name": "vmware-telemetry-cluster-ceip",
						"group": "",
						"kind": "configmap"
					},
					"code": 404}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps":
			assert.Equal(t, r.Method, http.MethodPost)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{
			  "kind": "Status",
			  "apiVersion": "v1",
			  "metadata": {},
			  "status": "Failure",
			  "message": "access to configmap forbidden",
			  "reason": "Forbidden",
			  "details": {
				"name": "vmware-system-telemetry-ceip",
				"kind": "configmap"
			  },
			  "code": 403
			}`))

		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	err = subject.UpdateCeip(context.Background(), true)
	assert.Error(t, err)
	if err == nil {
		t.Logf("err is not nil")
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "forbidden")
}

func TestUpdateCeip_GetConfigMapError(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)

		switch r.URL.Path {
		case "/api/v1/namespaces/vmware-system-telemetry":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				  "kind": "Namespace",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-system-telemetry"
				  }
				}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-cluster-ceip":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "access to configmap forbidden",
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

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	err = subject.UpdateCeip(context.Background(), true)
	assert.Error(t, err)
	if err == nil {
		t.Logf("err is not nil")
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "forbidden")
}

func TestUpdateCeip_UpdateConfigMapError(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		switch r.URL.Path {
		case "/api/v1/namespaces/vmware-system-telemetry":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				  "kind": "Namespace",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-system-telemetry"
				  }
				}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-cluster-ceip":
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"apiVersion": "v1",
					"data": {
						"level": "disabled"
					},
					"kind": "ConfigMap",
					"metadata": {
						"name": "vmware-telemetry-cluster-ceip",
						"namespace": "vmware-system-telemetry",
						"uid": "464177f9-4e0a-4c83-b1cc-a8197788de24"}
					}`))

			case http.MethodPut:
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "access to configmap forbidden",
					"reason": "Forbidden",
					"details": {
						"name": "vmware-system-telemetry-ceip",
						"kind": "configmap"
					},
					"code": 403
				}`))
			}
		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	err = subject.UpdateCeip(context.Background(), true)
	assert.Error(t, err)
	if err == nil {
		t.Logf("err is not nil")
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "forbidden")
}

func TestUpdateCeip_PreviousConfigurationExists(t *testing.T) {
	var ceipUpdated bool
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)

		switch r.URL.Path {
		case "/api/v1/namespaces/vmware-system-telemetry":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				  "kind": "Namespace",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-system-telemetry"
				  }
				}`))
		case "/api/v1/namespaces":
			assert.Equal(t, r.Method, http.MethodPost)
			buf, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(buf), `"name":"vmware-system-telemetry"`)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				  "kind": "Namespace",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-system-telemetry"
				  }
				}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-cluster-ceip":
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"apiVersion": "v1",
					"data": {
						"level": "disabled"
					},
					"kind": "ConfigMap",
					"metadata": {
						"name": "vmware-telemetry-cluster-ceip",
						"namespace": "vmware-system-telemetry",
						"uid": "464177f9-4e0a-4c83-b1cc-a8197788de24"}
					}`))

			case http.MethodPut:
				w.WriteHeader(http.StatusOK)
				buf, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.Contains(t, string(buf), `"level":"standard"`)
				ceipUpdated = true
				w.Write([]byte(`{
					"apiVersion": "v1",
					"data": {
						"level": "standard"
					},
					"kind": "ConfigMap",
					"metadata": {
						"name": "vmware-telemetry-cluster-ceip",
						"namespace": "vmware-system-telemetry",
						"uid": "464177f9-4e0a-4c83-b1cc-a8197788de24"}
					}`))
			}

		case "/api/v1/namespaces/vmware-system-telemetry/configmaps":
			assert.Equal(t, r.Method, http.MethodPost)
			buf, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(buf), `"level":"standard"`)
			ceipUpdated = true
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				  "kind": "ConfigMap",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-telemetry-cluster-ceip"
				  }
				}`))

		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	err = subject.UpdateCeip(context.Background(), true)
	assert.NoError(t, err)
	assert.True(t, ceipUpdated)
}

func TestUpdateCeip_FailToFindNamespace(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		if r.URL.Path == "/api/v1/namespaces/vmware-system-telemetry" {
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{
			  "kind": "Status",
			  "apiVersion": "v1",
			  "metadata": {
			
			  },
			  "status": "Failure",
			  "message": "namespaces \"vmware-system-telemetry\" is forbidden: User \"system:anonymous\" cannot get resource \"namespaces\" in API group \"\" in the namespace \"vmware-system-telemetry\"",
			  "reason": "Forbidden",
			  "details": {
				"name": "vmware-system-telemetry",
				"kind": "namespaces"
			  },
			  "code": 403
			}`))
		}
	})

	assert.NoError(t, err)
	defer srv.Close()

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	err = subject.UpdateCeip(context.Background(), true)
	assert.Error(t, err)
	if err == nil {
		t.Logf("err is not nil")
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "forbidden")
	assert.Contains(t, err.Error(), "vmware-system-telemetry")
}

func TestUpdateIdentifiers_NoPreviousConfiguration(t *testing.T) {
	var namespaceCreated, identifiersCreated, identifiersUpdated bool
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		switch r.URL.Path {
		case "/api/v1/namespaces/vmware-system-telemetry":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "namespace not found",
					"reason": "NotFound",
					"details": {
						"name": "vmware-system-telemetry",
						"group": "",
						"kind": "namespace"
					},
					"code": 404}`))
		case "/api/v1/namespaces":
			assert.Equal(t, r.Method, http.MethodPost)
			namespaceCreated = true
			buf, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(buf), `"name":"vmware-system-telemetry"`)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				  "kind": "Namespace",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-system-telemetry"
				  }
				}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-identifiers":
			if r.Method == http.MethodPut {
				buf, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.Contains(t, string(buf), `"data":{"key-1":"val-1","key-2":"val-2"`)
				identifiersUpdated = true
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
				  "kind": "ConfigMap",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-telemetry-identifiers"
				  }
				}`))
			} else if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "configmap not found",
					"reason": "NotFound",
					"details": {
						"name": "vmware-telemetry-identifiers",
						"group": "",
						"kind": "configmap"
					},
					"code": 404}`))
			}
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps":
			assert.Equal(t, r.Method, http.MethodPost)
			buf, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(buf), `"name":"vmware-telemetry-identifiers"`)
			identifiersCreated = true
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				  "kind": "ConfigMap",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-telemetry-identifiers"
				  }
				}`))

		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	vals := []update.UpdateVal{
		{Changed: true, Key: "key-1", Value: "val-1"},
		{Changed: true, Key: "key-2", Value: "val-2"},
	}
	err = subject.UpdateIdentifiers(context.Background(), vals)
	assert.NoError(t, err)
	assert.True(t, namespaceCreated)
	assert.True(t, identifiersCreated)
	assert.True(t, identifiersUpdated)
}

func TestUpdateIdentifiers_UpdateOnlyChangedIdentifiers(t *testing.T) {
	var identifiersUpdated bool
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		switch r.URL.Path {
		case "/api/v1/namespaces/vmware-system-telemetry":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				  "kind": "Namespace",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-system-telemetry"
				  }
				}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-identifiers":
			if r.Method == http.MethodPut {
				buf, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.Contains(t, string(buf), `"data":{"key-1":"val-1","key-2":"val-2","key-3":"val-3"`)
				identifiersUpdated = true
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
				  "kind": "ConfigMap",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-telemetry-identifiers"
				  }
				}`))
			} else if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"apiVersion": "v1",
					"data": {
						"key-1": "val-1",
						"key-2": "no val"
					},
					"kind": "ConfigMap",
					"metadata": {
						"name": "vmware-telemetry-identifiers",
						"namespace": "vmware-system-telemetry",
						"uid": "464177f9-4e0a-4c83-b1cc-a8197788de24"
					}
				}`))
			}
		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	vals := []update.UpdateVal{
		{Changed: false, Key: "key-1", Value: "default"},
		{Changed: true, Key: "key-2", Value: "val-2"},
		{Changed: true, Key: "key-3", Value: "val-3"},
	}
	err = subject.UpdateIdentifiers(context.Background(), vals)
	assert.NoError(t, err)
	assert.True(t, identifiersUpdated)
}

func TestUpdateIdentifiers_CreateConfigMapError(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		switch r.URL.Path {
		case "/api/v1/namespaces/vmware-system-telemetry":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				  "kind": "Namespace",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-system-telemetry"
				  }
				}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-identifiers":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "configmap not found",
					"reason": "NotFound",
					"details": {
						"name": "vmware-telemetry-cluster-identifiers",
						"group": "",
						"kind": "configmap"
					},
					"code": 404}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps":
			assert.Equal(t, r.Method, http.MethodPost)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{
			  "kind": "Status",
			  "apiVersion": "v1",
			  "metadata": {},
			  "status": "Failure",
			  "message": "access to configmap forbidden",
			  "reason": "Forbidden",
			  "details": {
				"name": "vmware-system-telemetry-identifiers",
				"kind": "configmap"
			  },
			  "code": 403
			}`))
		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	err = subject.UpdateIdentifiers(context.Background(), nil)
	assert.Error(t, err)
	if err == nil {
		t.Logf("err is not nil")
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "forbidden")
}

func TestUpdateIdentifiers_GetConfigMapError(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)

		switch r.URL.Path {
		case "/api/v1/namespaces/vmware-system-telemetry":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				  "kind": "Namespace",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-system-telemetry"
				  }
				}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-identifiers":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "access to configmap forbidden",
					"reason": "Forbidden",
					"details": {
						"name": "vmware-telemetry-cluster-identifiers",
						"group": "",
						"kind": "configmap"
					},
					"code": 403}`))
		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	err = subject.UpdateIdentifiers(context.Background(), nil)
	assert.Error(t, err)
	if err == nil {
		t.Logf("err is not nil")
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "forbidden")
}

func TestUpdateIdentifiers_UpdateConfigMapError(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		switch r.URL.Path {
		case "/api/v1/namespaces/vmware-system-telemetry":
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				  "kind": "Namespace",
				  "apiVersion": "v1",
				  "metadata": {
					"name": "vmware-system-telemetry"
				  }
				}`))
		case "/api/v1/namespaces/vmware-system-telemetry/configmaps/vmware-telemetry-identifiers":
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"apiVersion": "v1",
					"data": {
						"key-1": "val-1"
					},
					"kind": "ConfigMap",
					"metadata": {
						"name": "vmware-telemetry-identifiers",
						"namespace": "vmware-system-telemetry",
						"uid": "464177f9-4e0a-4c83-b1cc-a8197788de24"}
					}`))

			case http.MethodPut:
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{
					"kind": "Status",
					"apiVersion": "v1",
					"metadata": {},
					"status": "Failure",
					"message": "access to configmap forbidden",
					"reason": "Forbidden",
					"details": {
						"name": "vmware-telemetry-identifiers",
						"kind": "configmap"
					},
					"code": 403
				}`))
			}
		}
	})
	assert.NoError(t, err)
	defer srv.Close()

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	err = subject.UpdateIdentifiers(context.Background(), nil)
	assert.Error(t, err)
	if err == nil {
		t.Logf("err is not nil")
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "forbidden")
}

func TestUpdateIdentifiers_FailToFindNamespace(t *testing.T) {
	clf, srv, err := kubernetes.GetKubernetesClientServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", runtime.ContentTypeJSON)
		if r.URL.Path == "/api/v1/namespaces/vmware-system-telemetry" {
			assert.Equal(t, r.Method, http.MethodGet)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{
			  "kind": "Status",
			  "apiVersion": "v1",
			  "metadata": {
			
			  },
			  "status": "Failure",
			  "message": "namespaces \"vmware-system-telemetry\" is forbidden: User \"system:anonymous\" cannot get resource \"namespaces\" in API group \"\" in the namespace \"vmware-system-telemetry\"",
			  "reason": "Forbidden",
			  "details": {
				"name": "vmware-system-telemetry",
				"kind": "namespaces"
			  },
			  "code": 403
			}`))
		}
	})

	assert.NoError(t, err)
	defer srv.Close()

	client, _ := clf()
	subject := &update.Updater{
		Client: client,
	}

	err = subject.UpdateIdentifiers(context.Background(), nil)
	assert.Error(t, err)
	if err == nil {
		t.Logf("err is not nil")
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "forbidden")
	assert.Contains(t, err.Error(), "vmware-system-telemetry")
}
