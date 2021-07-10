module github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy

go 1.16

require (
	github.com/google/go-cmp v0.5.3 // indirect
	github.com/jetstack/cert-manager v1.1.0
	github.com/nxadm/tail v1.4.6 // indirect
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.3 // indirect
	go.pinniped.dev/generated/1.19/apis v0.0.0-00010101000000-000000000000
	go.pinniped.dev/generated/1.19/client v0.0.0-20201219022151-546b8b5d25c6
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9 // indirect
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43 // indirect
	golang.org/x/sys v0.0.0-20210110051926-789bb1bd4061 // indirect
	golang.org/x/tools v0.0.0-20200904185747-39188db58858 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.5
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.19.5
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	sigs.k8s.io/controller-runtime v0.7.0
)

// Import an nested go modules have some known issues. The following replace temporarily fixes it
// https://vmware.slack.com/archives/C0158N04WDD/p1602106922371100
// https://github.com/golang/go/issues/34055
replace go.pinniped.dev/generated/1.19/apis => go.pinniped.dev/generated/1.19/apis v0.0.0-20201219022151-546b8b5d25c6
