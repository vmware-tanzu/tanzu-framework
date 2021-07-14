module github.com/vmware-tanzu/tanzu-framework/addons

go 1.16

replace (
	github.com/vmware-tanzu/tanzu-framework => ../
	k8s.io/api => k8s.io/api v0.19.2
	k8s.io/client-go => k8s.io/client-go v0.19.2
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v0.3.11-0.20201204161359-8437691189ad
)

require (
	github.com/go-logr/logr v0.4.0
	github.com/onsi/ginkgo v1.16.2
	github.com/onsi/gomega v1.13.0
	github.com/pkg/errors v0.9.1
	github.com/vmware-tanzu/tanzu-framework v0.0.0-20201211200158-5838874f2c38
	github.com/vmware-tanzu/carvel-kapp-controller v0.20.0-rc.1
	github.com/vmware-tanzu/carvel-vendir v0.19.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b
	sigs.k8s.io/cluster-api v0.3.20-0.20210628204229-9fcfbce8e5c6
	sigs.k8s.io/controller-runtime v0.7.0
	sigs.k8s.io/yaml v1.2.0
)
