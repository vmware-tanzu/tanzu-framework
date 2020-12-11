module github.com/vmware-tanzu-private/core

go 1.13

require (
	cloud.google.com/go/storage v1.0.0
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/Jeffail/gabs v1.4.0
	github.com/Netflix/go-expect v0.0.0-20190729225929-0e00d9168667 // indirect
	github.com/adrg/xdg v0.2.1
	github.com/amenzhinsky/go-memexec v0.3.0
	github.com/aunum/log v0.0.0-20200821225356-38d2e2c8b489
	github.com/caarlos0/spin v1.1.0

	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/gobwas/glob v0.2.3
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c // indirect
	github.com/jeremywohl/flatten v1.0.1

	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1

	github.com/vmware-tanzu-private/tkg-cli v1.3.0-pre-alpha.0.20201125210554-f7d01631a634
	github.com/vmware-tanzu-private/tkg-providers v1.3.0-pre-alpha.0.20201119173244-e73afc2fd3f0 // indirect

	go.opencensus.io v0.22.2 // indirect
	go.uber.org/multierr v1.1.0
	golang.org/x/mod v0.3.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.13.0
	google.golang.org/grpc v1.26.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/apimachinery v0.17.11
	k8s.io/client-go v0.17.11
	sigs.k8s.io/controller-runtime v0.5.11
)
