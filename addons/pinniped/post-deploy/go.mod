module github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy

go 1.17

// Right now, we depend on Kube 1.23, but Pinniped 1.20.
//
// This is because we need to depend on Pinniped v0.12.1 (because that is the
// latest Pinniped version running on TKG clusters), but Pinniped v0.12.1 does
// not contain generated Pinniped APIs/clients to go with Kube 1.23 (i.e.,
// https://github.com/vmware-tanzu/pinniped/tree/v0.12.1/generated does not
// contain a 1.23 directory).
//
// For now, we will depend on Pinniped 1.20 generated code because go mod logic
// will automatically update our dependency graph to use the later Kube version
// in the graph (i.e., our post-deploy job will use Kube v0.23.0, derived from
// the dependencies k8s.io/{api,apimachinery,client-go} v0.23.0). This seems
// like the best of the worst ways to go.
//
// In the future, we should always try to depend on the same version of Kube and
// Pinniped generated code (i.e., when we update to depend on Kube 1.24, we
// should also depend on go.pinniped.dev/generated/1.24).

require (
	github.com/jetstack/cert-manager v1.1.0
	github.com/stretchr/testify v1.7.0
	go.pinniped.dev/generated/1.20/apis v0.0.0-00010101000000-000000000000
	go.pinniped.dev/generated/1.20/client v0.0.0-20220209183828-4d6a2af89419 // Commit SHA 4d6a2af89419 is tag v0.12.1.
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.pinniped.dev/generated/1.19/apis v0.0.0-20220310140840-61c8d5452705 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 // indirect
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f // indirect
	golang.org/x/sys v0.0.0-20210831042530-f4d43177bf5e // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apiextensions-apiserver v0.19.2 // indirect
	k8s.io/klog/v2 v2.30.0 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

// Import an nested go modules have some known issues. The following replace temporarily fixes it
// https://github.com/golang/go/issues/34055
//
// Commit SHA 4d6a2af89419 is tag v0.12.1.
replace go.pinniped.dev/generated/1.20/apis => go.pinniped.dev/generated/1.19/apis v0.0.0-20220209183828-4d6a2af89419
