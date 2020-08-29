module github.com/vmware-tanzu-private/core

go 1.13

require (
	cloud.google.com/go v0.46.3
	cloud.google.com/go/storage v1.0.0
	github.com/99designs/keyring v1.1.2
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/adrg/xdg v0.2.1
	github.com/aunum/log v0.0.0-20200821225356-38d2e2c8b489
	github.com/caarlos0/spin v1.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.3.2
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	gitlab.eng.vmware.com/olympus/api v0.0.0-20200826184856-2a2587ac1169
	gitlab.eng.vmware.com/olympus/api-machinery v0.0.0-20200811223306-665e65790d28
	gitlab.eng.vmware.com/olympus/oauth2cli v0.1.1
	go.uber.org/multierr v1.1.0
	golang.org/x/mod v0.3.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gonum.org/v1/netlib v0.0.0-20190331212654-76723241ea4e // indirect
	google.golang.org/api v0.13.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/apimachinery v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06 // indirect
)
