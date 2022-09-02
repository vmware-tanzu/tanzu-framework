module github.com/vmware-tanzu/tanzu-framework/tkg

go 1.17

replace (
	github.com/vmware-tanzu/tanzu-framework/apis/cli => ../apis/cli
    github.com/vmware-tanzu/tanzu-framework/apis/config => ../apis/config
    github.com/vmware-tanzu/tanzu-framework/apis/run => ../apis/run
    github.com/vmware-tanzu/tanzu-framework/cli/runtime => ../cli/runtime
)

require (
	github.com/go-logr/logr v1.2.3
	github.com/pkg/errors v0.9.1
)
