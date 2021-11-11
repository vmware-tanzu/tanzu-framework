module github.com/vmware-tanzu/tanzu-framework/addons

go 1.16

replace (
	github.com/vmware-tanzu/tanzu-framework => ../
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.0.0
)

require (
	github.com/go-logr/logr v0.4.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/vmware-tanzu/carvel-kapp-controller v0.25.0
	github.com/vmware-tanzu/carvel-vendir v0.23.0
	github.com/vmware-tanzu/tanzu-framework v0.9.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.2
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
	sigs.k8s.io/cluster-api v1.0.0
	sigs.k8s.io/controller-runtime v0.10.2
)
