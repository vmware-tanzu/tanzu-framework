# Testdata

## The crds directory

All CRDs are available under `config/crd/bases` which are aligned with the release. However, some newly introduced features
need a newer version of CRD for testing. We copy the CRDs from `config/crd/bases` and make some modifications(e.g., storage version).

The downside of this approach is the potential inconsistency. Before we have a better solution, we might have to periodically
sync the testdata manually.
