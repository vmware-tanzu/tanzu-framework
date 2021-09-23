module github.com/vmware-tanzu/tanzu-framework

go 1.16

replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.1.0
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.3.1 // indirect
	k8s.io/api => k8s.io/api v0.17.11
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.11
	k8s.io/client-go => k8s.io/client-go v0.17.11
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.0.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.14
	sigs.k8s.io/kind => sigs.k8s.io/kind v0.11.1
)

require (
	cloud.google.com/go/storage v1.10.0
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/Azure/azure-sdk-for-go v48.2.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.11
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.3
	github.com/Jeffail/gabs v1.4.0
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/Masterminds/semver v1.5.0
	github.com/Netflix/go-expect v0.0.0-20190729225929-0e00d9168667 // indirect
	github.com/adrg/xdg v0.2.1
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/aunum/log v0.0.0-20200821225356-38d2e2c8b489
	github.com/avinetworks/sdk v0.0.0-20201123134013-c157ef55b6f7
	github.com/aws/aws-sdk-go v1.36.26
	github.com/briandowns/spinner v1.16.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7
	github.com/elazarl/go-bindata-assetfs v1.0.1
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/fatih/color v1.12.0
	github.com/getkin/kin-openapi v0.66.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ldap/ldap/v3 v3.3.0
	github.com/go-logr/logr v0.4.0
	github.com/go-openapi/errors v0.19.2
	github.com/go-openapi/loads v0.19.4
	github.com/go-openapi/runtime v0.19.4
	github.com/go-openapi/spec v0.19.5
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/swag v0.19.5
	github.com/go-openapi/validate v0.19.5
	github.com/gobuffalo/flect v0.2.3 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/google/go-containerregistry v0.4.1
	github.com/google/uuid v1.2.0 // indirect
	github.com/googleapis/gnostic v0.4.1
	github.com/gorilla/websocket v1.4.2
	github.com/gosuri/uitable v0.0.4
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c // indirect
	github.com/imdario/mergo v0.3.12
	github.com/jeremywohl/flatten v1.0.1
	github.com/jessevdk/go-flags v1.4.0
	github.com/jinzhu/copier v0.2.8
	github.com/juju/fslock v0.0.0-20160525022230-4d5c94c67b4b
	github.com/k14s/imgpkg v0.6.0
	github.com/k14s/ytt v0.32.1-0.20210511155130-214258be2519
	github.com/lithammer/dedent v1.1.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/novln/docker-parser v1.0.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/ginkgo v1.16.2
	github.com/onsi/gomega v1.13.0
	github.com/otiai10/copy v1.4.2
	github.com/pelletier/go-toml/v2 v2.0.0-beta.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.29.0 // indirect
	github.com/rs/xid v1.2.1
	github.com/satori/go.uuid v1.2.0
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.1-0.20210427113832-6241f9ab9942
	github.com/vmware-labs/yaml-jsonpath v0.3.2
	github.com/vmware-tanzu/carvel-kapp-controller v0.25.0
	github.com/vmware-tanzu/carvel-secretgen-controller v0.5.0
	github.com/vmware-tanzu/carvel-vendir v0.23.0
	github.com/vmware/govmomi v0.23.1
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0
	go.uber.org/multierr v1.5.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/mod v0.4.2
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0 // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/api v0.40.0
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/cluster-bootstrap v0.21.2 // indirect
	k8s.io/klog/v2 v2.8.0
	k8s.io/kubectl v0.17.14
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b
	sigs.k8s.io/cluster-api v0.3.23
	sigs.k8s.io/cluster-api-provider-aws v0.6.6
	sigs.k8s.io/cluster-api-provider-azure v0.4.15
	sigs.k8s.io/cluster-api-provider-vsphere v0.7.10
	sigs.k8s.io/cluster-api/test/infrastructure/docker v0.0.0-20210720023132-dfeb8d447bdc
	sigs.k8s.io/controller-runtime v0.7.0
	sigs.k8s.io/controller-tools v0.4.1
	sigs.k8s.io/kind v0.11.1
	sigs.k8s.io/yaml v1.2.0
)
