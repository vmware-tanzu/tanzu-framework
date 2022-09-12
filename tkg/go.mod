module github.com/vmware-tanzu/tanzu-framework/tkg

go 1.17

replace (
	github.com/vmware-tanzu/tanzu-framework/apis/cli => ../apis/cli
	github.com/vmware-tanzu/tanzu-framework/apis/config => ../apis/config
	github.com/vmware-tanzu/tanzu-framework/apis/run => ../apis/run
	github.com/vmware-tanzu/tanzu-framework/cli/runtime => ../cli/runtime
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.1.5
)
