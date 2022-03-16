# Kubernetes Query Builder SDK

The discovery SDK provides means to prepare and query for the state of a clusters objects, schema and API resources to determine the details about a cluster and how best to interact with it.

The longer term goal is to also become the most query and resource efficient means of doing so, perhaps with some limited caching ability.

The Query Builder SDK can query for:

- Objects
  - Annotation
- Resources
  - WithFields
- OpenAPI Schema

Once created, these prepared queries can be exported and used whenever necessary.

## Example

```go
import (
    "sigs.k8s.io/controller-runtime/pkg/client/config"
    "github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/capabilities/discovery"
)

cfg := config.GetConfig()

clusterQueryClient, err := discovery.NewClusterQueryClientForConfig(cfg)
if err != nil {
    log.Error(err)
}

// Define objects to query.
var pod = corev1.ObjectReference{
    Kind:       "Pod",
    Name:       "testpod",
    Namespace:  "testns",
    APIVersion: "v1",
}

var testAnnotations = map[string]string{
    "cluster.x-k8s.io/provider": "infrastructure-fake",
}

// Define queries.
var testObject = Object("podObj", &pod).WithAnnotations(testAnnotations)

var testGVR = Group("podResource", testapigroup.SchemeGroupVersion.Group).
    WithVersions("v1").
    WithResource("pods").
    WithFields("spec.containers", "spec.serviceAccountName") // Only works on resources with structural schema.

// Build query client.
c := clusterQueryClient.Query(testObject, testGVR)

// Execute returns combined result of all queries.
found, err := c.Execute()
if err != nil {
    log.Error(err)
}

if found {
    log.Info("Queries successful")
}

// Inspect granular results of each query using the Results method (should be called after Execute).
if result := c.Results().ForQuery("podResource"); result != nil {
    if result.Found {
        log.Info("Pod resource found")
    } else {
        log.Infof("Pod resource not found. Reason: %s", result.NotFoundReason)
    }
}
```
