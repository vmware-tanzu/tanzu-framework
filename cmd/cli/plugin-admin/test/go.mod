module github.com/vmware-tanzu/tanzu-framework/plugin/test

go 1.17

replace (
	github.com/vmware-tanzu/tanzu-framework/apis/cli => ../../../../apis/cli
	github.com/vmware-tanzu/tanzu-framework/apis/run => ../../../../apis/run
	github.com/vmware-tanzu/tanzu-framework/capabilities/client => ../../../../capabilities/client
	github.com/vmware-tanzu/tanzu-framework/cli/core => ../../../../cli/core
	github.com/vmware-tanzu/tanzu-framework/cli/runtime => ../../../../cli/runtime
)

require (
	github.com/aunum/log v0.0.0-20200821225356-38d2e2c8b489
	github.com/spf13/cobra v1.5.0
	github.com/vmware-tanzu/tanzu-framework/cli/core v0.0.0-00010101000000-000000000000
	github.com/vmware-tanzu/tanzu-framework/cli/runtime v0.0.0-00010101000000-000000000000
)

require (
	cloud.google.com/go v0.99.0 // indirect
	cloud.google.com/go/storage v1.18.2 // indirect
	github.com/AlecAivazis/survey/v2 v2.3.5 // indirect
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/adrg/xdg v0.4.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/briandowns/spinner v1.19.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cheggaaa/pb v1.0.29 // indirect
	github.com/cncf/udpa/go v0.0.0-20210930031921-04548b0d99d4 // indirect
	github.com/cncf/xds/go v0.0.0-20211130200136-a8f946100490 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.10.0 // indirect
	github.com/cppforlife/cobrautil v0.0.0-20220411122935-c28a9f274a4e // indirect
	github.com/cppforlife/color v1.9.1-0.20200716202919-6706ac40b835 // indirect
	github.com/cppforlife/go-cli-ui v0.0.0-20200716203538-1e47f820817f // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/cli v20.10.11+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.16+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.4 // indirect
	github.com/envoyproxy/go-control-plane v0.10.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.6.2 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/go-containerregistry v0.7.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/go-version v1.4.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/juju/fslock v0.0.0-20160525022230-4d5c94c67b4b // indirect
	github.com/k14s/imgpkg v0.17.0 // indirect
	github.com/k14s/kbld v0.32.0 // indirect
	github.com/k14s/semver/v4 v4.0.1-0.20210701191048-266d47ac6115 // indirect
	github.com/k14s/starlark-go v0.0.0-20200720175618-3a5c849cc368 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/onsi/gomega v1.19.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vbatts/tar-split v0.11.2 // indirect
	github.com/vito/go-interact v0.0.0-20171111012221-fa338ed9e9ec // indirect
	github.com/vmware-tanzu/carvel-imgpkg v0.23.1 // indirect
	github.com/vmware-tanzu/carvel-vendir v0.26.0 // indirect
	github.com/vmware-tanzu/carvel-ytt v0.40.0 // indirect
	github.com/vmware-tanzu/tanzu-framework/apis/cli v0.0.0-00010101000000-000000000000 // indirect
	github.com/vmware-tanzu/tanzu-framework/apis/run v0.0.0-00010101000000-000000000000 // indirect
	github.com/vmware-tanzu/tanzu-framework/capabilities/client v0.0.0-00010101000000-000000000000 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/oauth2 v0.0.0-20220309155454-6242fa91716a // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/api v0.63.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220308174144-ae0e22291548 // indirect
	google.golang.org/grpc v1.44.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.23.5 // indirect
	k8s.io/apiextensions-apiserver v0.23.5 // indirect
	k8s.io/apimachinery v0.23.5 // indirect
	k8s.io/client-go v0.23.5 // indirect
	k8s.io/component-base v0.23.5 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220124234850-424119656bbf // indirect
	k8s.io/kubectl v0.23.4 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/cluster-api v1.1.5 // indirect
	sigs.k8s.io/controller-runtime v0.11.2 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
