module github.com/vmware-tanzu/tanzu-framework

go 1.18

replace (
	github.com/vmware-tanzu/tanzu-framework/apis/config => ./apis/config
	github.com/vmware-tanzu/tanzu-framework/capabilities/client => ./capabilities/client
	github.com/vmware-tanzu/tanzu-framework/util => ./util
)

// Legacy tags before v0.1.0 was created
retract [v1.4.0-pre-alpha-1, v1.4.0-pre-alpha-2]
