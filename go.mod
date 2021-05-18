module github.com/vmware-tanzu-private/core

go 1.16

replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1 // indirect
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.0.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
)

require (
	cloud.google.com/go/storage v1.10.0
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/Azure/azure-sdk-for-go v46.0.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.4
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.1
	github.com/Jeffail/gabs v1.4.0
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/Masterminds/semver v1.5.0
	github.com/Netflix/go-expect v0.0.0-20190729225929-0e00d9168667 // indirect
	github.com/adrg/xdg v0.2.1
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/aunum/log v0.0.0-20200821225356-38d2e2c8b489
	github.com/avinetworks/sdk v0.0.0-20201123134013-c157ef55b6f7
	github.com/aws/aws-sdk-go v1.36.26
	github.com/caarlos0/spin v1.1.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7
	github.com/fabriziopandini/capi-conditions v0.0.0-20201102133039-7eb142d1b6d6
	github.com/fatih/color v1.10.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.2.0
	github.com/go-openapi/errors v0.19.2
	github.com/go-openapi/loads v0.19.4
	github.com/go-openapi/runtime v0.19.4
	github.com/go-openapi/spec v0.19.3
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/swag v0.19.5
	github.com/go-openapi/validate v0.19.5
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.4
	github.com/google/go-containerregistry v0.4.1
	github.com/gosuri/uitable v0.0.4
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/helloeave/json v1.15.3
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c // indirect
	github.com/jedib0t/go-pretty v4.3.0+incompatible
	github.com/jeremywohl/flatten v1.0.1
	github.com/jessevdk/go-flags v1.4.0
	github.com/juju/fslock v0.0.0-20160525022230-4d5c94c67b4b
	github.com/k14s/imgpkg v0.6.0
	github.com/lithammer/dedent v1.1.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.4
	github.com/otiai10/copy v1.4.2
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/vmware-tanzu-private/tkg-cli v1.3.1-rc.1.0.20210511054704-abb2b0362ada
	github.com/vmware-tanzu-private/tkg-providers v1.3.1-rc.1.0.20210507174302-4cc66882ac81 // indirect
	go.uber.org/multierr v1.5.0
	golang.org/x/mod v0.4.0
	golang.org/x/net v0.0.0-20201209123823-ac852fbbde11
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5
	golang.org/x/tools v0.1.0 // indirect
	google.golang.org/api v0.40.0
	google.golang.org/grpc v1.34.0
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.17.11
	k8s.io/apimachinery v0.17.11
	k8s.io/client-go v0.17.11
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	sigs.k8s.io/cluster-api v0.3.14
	sigs.k8s.io/cluster-api-provider-aws v0.6.5-0.20210309190705-ff3ed1b9f6f1
	sigs.k8s.io/controller-runtime v0.5.14
	sigs.k8s.io/yaml v1.2.0
)
