package webhooks_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/vmware-tanzu/tanzu-framework/addons/webhooks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	pkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

const (
	ServiceName     = "webhook-service"
	NameSpace       = "tkg-system"
	CertDirPath     = "./testdata/webhookserver"
	CertFileName    = "cert.pem"
	KeyFileName     = "key.pem"
	waitTimeout     = time.Second * 20
	pollingInterval = time.Second * 2
)

var (
	cfg       *rest.Config
	testEnv   *envtest.Environment
	scheme    = runtime.NewScheme()
	k8sClient client.Client
	ctx       = ctrl.SetupSignalHandler()
	fakeMgr   *manager.Manager
)

func TestWebhooks(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Addon Webhooks Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {

	By("bootstrapping test environment")

	testEnv = &envtest.Environment{CRDInstallOptions: envtest.CRDInstallOptions{CleanUpAfterUse: true},
		ErrorIfCRDPathMissing: true,
	}

	externalDeps := map[string][]string{
		"sigs.k8s.io/cluster-api":                        {"config/crd/bases", "controlplane/kubeadm/config/crd/bases"},
		"github.com/vmware-tanzu/carvel-kapp-controller": {"config/crds.yml"},
	}

	externalCRDPaths, err := testutil.GetExternalCRDPaths(externalDeps)
	Expect(err).NotTo(HaveOccurred())
	Expect(externalCRDPaths).ToNot(BeEmpty())

	testEnv.CRDDirectoryPaths = externalCRDPaths

	testEnv.CRDDirectoryPaths = append(testEnv.CRDDirectoryPaths,
		filepath.Join("..", "..", "config", "crd", "bases"),
	)

	testEnv.ErrorIfCRDPathMissing = true

	cfg, err = testEnv.Start()

	testEnv.ControlPlane.APIServer.Configure().Append("admission-control", "MutatingAdmissionWebhook")

	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = runtanzuv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = runtanzuv1alpha3.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = kappctrl.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clusterapiv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = controlplanev1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = pkgiv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = cniv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	options := manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		CertDir:            CertDirPath,
	}

	mgr, err := ctrl.NewManager(testEnv.Config, options)
	Expect(err).ToNot(HaveOccurred())
	mgr.GetWebhookServer().CertName = CertFileName
	mgr.GetWebhookServer().KeyName = KeyFileName
	mgr.GetWebhookServer().Host = "127.0.0.1"
	mgr.GetWebhookServer().Port = 9443
	fakeMgr = &mgr
	Expect(err).ToNot(HaveOccurred())

	// setup ClusterBootstrap webhook
	clusterbootstrapWebhook := webhooks.ClusterBootstrap{
		Client: k8sClient,
	}
	err = clusterbootstrapWebhook.SetupWebhookWithManager(mgr)
	Expect(err).ShouldNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).To(Succeed())
	}()

	close(done)
}, 120)

var _ = AfterSuite(func() {
	By("running after suite")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
