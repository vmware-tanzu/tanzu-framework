module github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy

go 1.16

require (
	github.com/jetstack/cert-manager v1.1.0
	github.com/stretchr/testify v1.6.1
	go.pinniped.dev/generated/1.19/apis v0.0.0-00010101000000-000000000000
	go.pinniped.dev/generated/1.19/client v0.0.0-20210916124603-454b792afbd9 // Commit SHA 454b792afbd9 is tag v0.12.0.
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.19.5
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.19.5
	sigs.k8s.io/controller-runtime v0.7.0
)

// Import an nested go modules have some known issues. The following replace temporarily fixes it
// https://github.com/golang/go/issues/34055
//
// Commit SHA 454b792afbd9 is tag v0.12.0.
replace go.pinniped.dev/generated/1.19/apis => go.pinniped.dev/generated/1.19/apis v0.0.0-20210916124603-454b792afbd9
