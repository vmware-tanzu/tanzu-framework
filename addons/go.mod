module github.com/vmware-tanzu/tanzu-framework/addons

go 1.18

replace (
	github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs => ../apis/addonconfigs
	github.com/vmware-tanzu/tanzu-framework/apis/run => ../apis/run
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.2.6
)

require (
	github.com/go-logr/logr v1.2.3
	github.com/google/gnostic v0.6.9
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.20.2
	github.com/oracle/cluster-api-provider-oci v0.6.0
	github.com/pkg/errors v0.9.1
	github.com/vmware-tanzu/carvel-kapp-controller v0.35.0
	github.com/vmware-tanzu/carvel-secretgen-controller v0.5.0
	github.com/vmware-tanzu/carvel-vendir v0.26.0
	github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs v0.0.0-20220907220230-c1137d344dd3
	github.com/vmware-tanzu/tanzu-framework/apis/run v0.0.0-20220907220230-c1137d344dd3
	github.com/vmware-tanzu/vm-operator-api v0.1.4-0.20211202185235-43eb44c09ecd
	github.com/vmware-tanzu/vm-operator/external/tanzu-topology v0.0.0-20211209213435-0f4ab286f64f
	golang.org/x/text v0.4.0
	golang.org/x/tools v0.1.12
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.24.6
	k8s.io/apiextensions-apiserver v0.24.6
	k8s.io/apimachinery v0.24.6
	k8s.io/client-go v0.24.6
	k8s.io/klog/v2 v2.70.1
	k8s.io/utils v0.0.0-20220812165043-ad590609e2e5
	knative.dev/pkg v0.0.0-20220302134643-d2cdc682d974
	sigs.k8s.io/cluster-api v1.2.6
	sigs.k8s.io/cluster-api-provider-vsphere v1.5.0
	sigs.k8s.io/controller-runtime v0.12.3
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go/compute v1.7.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.23 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.18 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/coredns/caddy v1.1.1 // indirect
	github.com/coredns/corefile-migration v1.0.17 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gobuffalo/flect v0.2.5 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/go-version v1.3.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k14s/semver/v4 v4.0.1-0.20210701191048-266d47ac6115 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/oracle/oci-go-sdk/v65 v65.18.0 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/sony/gobreaker v0.5.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vmware-tanzu/net-operator-api v0.0.0-20210401185409-b0dc6c297707 // indirect
	github.com/vmware-tanzu/vm-operator/external/ncp v0.0.0-20211209213435-0f4ab286f64f // indirect
	github.com/vmware/govmomi v0.27.1 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.22.0 // indirect
	golang.org/x/crypto v0.0.0-20220817201139-bc19a97f63c8 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/net v0.0.0-20220906165146-f3363e06e74c // indirect
	golang.org/x/oauth2 v0.0.0-20220822191816-0ebed06d0094 // indirect
	golang.org/x/sys v0.0.0-20220818161305-2296e01440c6 // indirect
	golang.org/x/term v0.0.0-20220722155259-a9ba230a4035 // indirect
	golang.org/x/time v0.0.0-20220722155302-e5dcc9cfc0b9 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/gcfg.v1 v1.2.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	k8s.io/cluster-bootstrap v0.24.6 // indirect
	k8s.io/component-base v0.24.6 // indirect
	k8s.io/kube-openapi v0.0.0-20220803164354-a70c9af30aea // indirect
	k8s.io/kubectl v0.24.6 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)
