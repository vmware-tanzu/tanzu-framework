module github.com/vmware-tanzu/tanzu-framework

go 1.16

replace (
	github.com/briandowns/spinner => github.com/alonyb/spinner v1.12.7
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/k14s/kbld => github.com/anujc25/carvel-kbld v0.31.0-update-vendir
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.0.1
	sigs.k8s.io/kind => sigs.k8s.io/kind v0.11.1
)

require (
	cloud.google.com/go/storage v1.10.0
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/Azure/azure-sdk-for-go v58.1.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.21
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.8
	github.com/Jeffail/gabs v1.4.0
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/Masterminds/semver v1.5.0
	github.com/Netflix/go-expect v0.0.0-20190729225929-0e00d9168667 // indirect
	github.com/adrg/xdg v0.2.1
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/aunum/log v0.0.0-20200821225356-38d2e2c8b489
	github.com/avinetworks/sdk v0.0.0-20201123134013-c157ef55b6f7
	github.com/aws/aws-sdk-go v1.40.56
	github.com/awslabs/goformation/v4 v4.19.5
	github.com/briandowns/spinner v1.16.0
	github.com/cppforlife/go-cli-ui v0.0.0-20200716203538-1e47f820817f
	github.com/docker/docker v20.10.9+incompatible
	github.com/elazarl/go-bindata-assetfs v1.0.1
	github.com/fatih/color v1.13.0
	github.com/getkin/kin-openapi v0.66.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ldap/ldap/v3 v3.3.0
	github.com/go-logr/logr v0.4.0
	github.com/go-openapi/errors v0.19.2
	github.com/go-openapi/loads v0.19.4
	github.com/go-openapi/runtime v0.19.4
	github.com/go-openapi/spec v0.19.5
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-openapi/swag v0.19.14
	github.com/go-openapi/validate v0.19.8
	github.com/gobwas/glob v0.2.3
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/google/go-containerregistry v0.6.0
	github.com/googleapis/gnostic v0.5.5
	github.com/gorilla/websocket v1.4.2
	github.com/gosuri/uitable v0.0.4
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c // indirect
	github.com/imdario/mergo v0.3.12
	github.com/jeremywohl/flatten v1.0.1
	github.com/jessevdk/go-flags v1.4.0
	github.com/jinzhu/copier v0.2.8
	github.com/juju/fslock v0.0.0-20160525022230-4d5c94c67b4b
	github.com/k14s/imgpkg v0.17.0
	github.com/k14s/kbld v0.31.0
	github.com/k14s/semver/v4 v4.0.1-0.20210701191048-266d47ac6115
	github.com/k14s/ytt v0.32.1-0.20210511155130-214258be2519
	github.com/lithammer/dedent v1.1.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.16.0
	github.com/otiai10/copy v1.4.2
	github.com/pelletier/go-toml/v2 v2.0.0-beta.3
	github.com/pkg/errors v0.9.1
	github.com/rs/xid v1.2.1
	github.com/sanathkr/go-yaml v0.0.0-20170819195128-ed9d249f429b
	github.com/satori/go.uuid v1.2.0
	github.com/sergi/go-diff v1.2.0
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	github.com/spf13/afero v1.6.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.1-0.20210427113832-6241f9ab9942
	github.com/tj/assert v0.0.0-20171129193455-018094318fb0
	github.com/vmware-tanzu/carvel-kapp-controller v0.25.0
	github.com/vmware-tanzu/carvel-secretgen-controller v0.5.0
	github.com/vmware-tanzu/carvel-vendir v0.23.0
	github.com/vmware/govmomi v0.27.1
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0
	go.uber.org/multierr v1.6.0
	golang.org/x/mod v0.5.1
	golang.org/x/net v0.0.0-20211005215030-d2e5035098b3
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	google.golang.org/api v0.56.0
	google.golang.org/grpc v1.41.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/klog/v2 v2.10.0
	k8s.io/kubectl v0.22.2
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
	sigs.k8s.io/cluster-api v1.0.1
	sigs.k8s.io/cluster-api-provider-aws v1.0.0
	sigs.k8s.io/cluster-api-provider-azure v1.0.0
	sigs.k8s.io/cluster-api-provider-vsphere v1.0.2
	sigs.k8s.io/cluster-api/test v1.0.1
	sigs.k8s.io/controller-runtime v0.10.3
	sigs.k8s.io/controller-tools v0.7.0
	sigs.k8s.io/kind v0.11.1
	sigs.k8s.io/yaml v1.3.0
)
