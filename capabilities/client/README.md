# Kubernetes Query Builder SDK

The discovery SDK provides means to prepare and query for the state of a clusters objects, schema and API resources to determine the details about a cluster and how best to interact with it.

The longer term goal is to also become the most query and resource efficient means of doing so, perhaps with some limited caching ability.

The Query Builder SDK can query for:

- Objects
  - Annotation
  - Labels
  - Conditions
- Resources
  - WithFields
- OpenAPI Schema

Once created, these prepared queries can be exported and used whenever necessary .

## Example

```go
// Define some standard objects, GVRs or schema
var ns = corev1.ObjectReference{
    Kind:      "namespace",
    Name:      "ian",
    Namespace: "ian",
}

// Its important to our use case for this annotation to also match
var testAnnotations = map[string]string{
    "cluster.x-k8s.io/provider": "infrastructure-fake",
}

// We want to ensure this version of pipelines exists, perhaps with specific fields
var testGVR = schema.GroupVersionResource{
    Group:    "tekton.dev",
    Version:  "v1beta1",
    Resource: "pipelines",
}

// We will now generate a partial query to match our specific namespace
var testResource1 = Object(ns).WithAnnotations(testAnnotations).WithConditions([C])

// Our usecase requires us to match a specific GVR
var testGVR1 = GVR(testGVR).WithFields([field])

// This partial schema is importan to be available - it doesnt matter what GVR provides it, but the schema needs to be available.
var testSchema1 = PartialSchema(schema).NotExist()

c, err := NewClusterQueryClient()
if err != nil {
    return err
}

// Build your query and use it where its necessary
OurCoolQuery := c.Query(testResource1, testGVR1, testSchema1).Prepare()

ok, err := OurCoolQuery()
if err != nil {
    return err
}
if ok {
    log.Log("W00T")
}
```
