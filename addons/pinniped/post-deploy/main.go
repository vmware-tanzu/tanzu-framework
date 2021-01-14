package main

import (
	"flag"
	"os"

	certmanagerclientset "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/configure"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/vars"
	conciergeclientset "go.pinniped.dev/generated/1.19/client/concierge/clientset/versioned"
	supervisorclientset "go.pinniped.dev/generated/1.19/client/supervisor/clientset/versioned"

	"k8s.io/client-go/kubernetes"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	k8sClientset         kubernetes.Interface
	supervisorClientset  supervisorclientset.Interface
	conciergeClientset   conciergeclientset.Interface
	certmanagerClientset certmanagerclientset.Interface
)

func init() {
	cfg, err := k8sconfig.GetConfig()
	if err != nil {
		panic(err)
	}
	k8sClientset = kubernetes.NewForConfigOrDie(cfg)
	supervisorClientset = supervisorclientset.NewForConfigOrDie(cfg)
	conciergeClientset = conciergeclientset.NewForConfigOrDie(cfg)
	certmanagerClientset = certmanagerclientset.NewForConfigOrDie(cfg)
}

func main() {
	// optional
	flag.StringVar(&vars.SupervisorNamespace, "supervisor-namespace", vars.SupervisorNamespace, "The namespace of Pinniped supervisor")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.SupervisorSvcName, "supervisor-svc-name", "", "The name of Pinniped supervisor service")

	flag.StringVar(&vars.ConciergeNamespace, "concierge-namespace", vars.ConciergeNamespace, "The namespace of Pinniped concierge")

	// required for management cluster: no
	// required for workload cluster: yes
	flag.StringVar(&vars.SupervisorSvcEndpoint, "supervisor-svc-endpoint", "", "The endpoint of Pinniped supervisor service")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.FederationDomainName, "federationdomain-name", "", "The name of Pinniped FederationDomain")

	// required for management cluster: yes
	// required for workload cluster: yes
	flag.StringVar(&vars.JWTAuthenticatorName, "jwtauthenticator-name", "", "The name of Pinniped JWTAuthenticator")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.SupervisorCertName, "supervisor-cert-name", "", "The name of Certificate for Pinniped supervisor service")

	// required for management cluster: no
	// required for workload cluster: yes
	flag.StringVar(&vars.SupervisorCABundleData, "supervisor-ca-bundle-data", "", "The CA data of Pinniped supervisor service")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.DexNamespace, "dex-namespace", vars.DexNamespace, "The namespace of Dex")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.DexSvcName, "dex-svc-name", vars.DexSvcName, "The name of Dex service")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.DexCertName, "dex-cert-name", vars.DexCertName, "The name of Dex TLS certificate")

	// required for management cluster: yes
	// required for workload cluster: no
	flag.StringVar(&vars.DexConfigMapName, "dex-configmap-name", vars.DexConfigMapName, "The name of Dex config map")

	flag.Parse()

	loggerMgr := initZapLog()
	zap.ReplaceGlobals(loggerMgr)
	defer loggerMgr.Sync() // nolint
	logger := loggerMgr.Sugar()

	err := configure.TKGAuthentication(
		configure.Clients{
			K8SClientset:         k8sClientset,
			SupervisorClientset:  supervisorClientset,
			ConciergeClientset:   conciergeClientset,
			CertmanagerClientset: certmanagerClientset},
	)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}

func initZapLog() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.CallerKey = "caller"
	logger, _ := config.Build()
	return logger
}
