module github.com/vmware-tanzu/tanzu-framework/pkg/v1/providers/tests

replace github.com/vmware-tanzu/tanzu-framework/ => ../../../..

go 1.16

require (
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/vmware-labs/yaml-jsonpath v0.3.2
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
)
