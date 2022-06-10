import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

const (
	clusterNameMng      = "ggao-aws-june"
	clusterNameWlc      = "ggao-aws-wc"
	systemNS            = "tkg-system"
)

// getClusterBootstrap gets ClusterBootstrap resource with the provided object key
func getClusterBootstrap(ctx context.Context, k8sClient client.Client, namespace, clusterName string) *runtanzuv1alpha3.ClusterBootstrap {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	objKey := client.ObjectKey{Namespace: namespace, Name: clusterName}

	Eventually(func() bool {
		err := k8sClient.Get(ctx, objKey, clusterBootstrap)
		return err == nil
	}, waitTimeout, pollingInterval).Should(BeTrue())

	Expect(clusterBootstrap).ShouldNot(BeNil())
	return clusterBootstrap
}

// getPackageInstall get PackageInstall resource with the provided object key
func getPackageInstall(ctx context.Context, k8sClient client.Client, namespace, pkgiName string) *kapppkgiv1alpha1.PackageInstall {
	pkgInstall := &kapppkgiv1alpha1.PackageInstall{}
	objKey := client.ObjectKey{Namespace: namespace, Name: pkgiName}

	Eventually(func() bool {
		err := k8sClient.Get(ctx, objKey, pkgInstall)
		return err == nil
	}, waitTimeout, pollingInterval).Should(BeTrue())

	Expect(pkgInstall).ShouldNot(BeNil())

	return pkgInstall
}

func generatePkgiNameWithVersion(packageName, PackageVersion string) string {
	return fmt.Sprintf("%s.%s", packageName, PackageVersion)
}

func checkUtkgAddons(ctx context.Context, cl client.Client, t *testing.T) {
	assert := assert.New(t)

	// Get ClusterBootstrap and return error if not found
	//mngCluster := &clusterapiv1beta1.Cluster{}
	//clusterBootstrap := getClusterBootstrap(client.ObjectKeyFromObject(mngCluster))
	clusterBootstrap := getClusterBootstrap(ctx, cl, systemNS, clusterNameMng)

	//wlcCluster := &clusterapiv1beta1.Cluster{}

	// packageInstall name for for both management and workload clusters should follow the <cluster name>-<addon short name>
	// packageInstall name and version should match info in clusterBootstrap for all packages, format is <package name>.<package version>
	antreaPkgiName := util.GeneratePackageInstallName(clusterNameWlc, "antrea")
	antreaPackageInstall := getPackageInstall(ctx, cl, systemNS, antreaPkgiName)
	assert.Equal(clusterBootstrap.Spec.CNI.RefName, generatePkgiNameWithVersion(antreaPackageInstall.Spec.PackageRef.RefName, antreaPackageInstall.Spec.PackageRef.VersionSelection.Constraints))

	csiPkgiName := util.GeneratePackageInstallName(clusterNameWlc, "csi")
	csiPackageInstall := getPackageInstall(ctx, cl, systemNS, csiPkgiName)
	assert.Equal(clusterBootstrap.Spec.CNI.RefName, generatePkgiNameWithVersion(csiPackageInstall.Spec.PackageRef.RefName, csiPackageInstall.Spec.PackageRef.VersionSelection.Constraints))

	cpiPkgiName := util.GeneratePackageInstallName(clusterNameWlc, "cpi")
	cpiPackageInstall := getPackageInstall(ctx, cl, systemNS, cpiPkgiName)
	assert.Equal(clusterBootstrap.Spec.CNI.RefName, generatePkgiNameWithVersion(cpiPackageInstall.Spec.PackageRef.RefName, cpiPackageInstall.Spec.PackageRef.VersionSelection.Constraints))

	kappPkgiName := util.GeneratePackageInstallName(clusterNameWlc, "kapp")
	kappPackageInstall := getPackageInstall(ctx, cl, systemNS, kappPkgiName)
	assert.Equal(clusterBootstrap.Spec.CNI.RefName, generatePkgiNameWithVersion(kappPackageInstall.Spec.PackageRef.RefName, kappPackageInstall.Spec.PackageRef.VersionSelection.Constraints))

	msPkgiName := util.GeneratePackageInstallName(clusterNameWlc, "metrics-server")
	msPackageInstall := getPackageInstall(ctx, cl, systemNS, msPkgiName)
	assert.Equal(clusterBootstrap.Spec.AdditionalPackages[0].RefName, generatePkgiNameWithVersion(msPackageInstall.Spec.PackageRef.RefName, msPackageInstall.Spec.PackageRef.VersionSelection.Constraints))

	scPkgiName := util.GeneratePackageInstallName(clusterNameWlc, "secretgen-controller")
	scPackageInstall := getPackageInstall(ctx, cl, systemNS, scPkgiName)
	assert.Equal(clusterBootstrap.Spec.AdditionalPackages[1].RefName, generatePkgiNameWithVersion(scPackageInstall.Spec.PackageRef.RefName, scPackageInstall.Spec.PackageRef.VersionSelection.Constraints))

	ppdPkgiName := util.GeneratePackageInstallName(clusterNameWlc, "pinniped")
	ppdPackageInstall := getPackageInstall(ctx, cl, systemNS, ppdPkgiName)
	assert.Equal(clusterBootstrap.Spec.AdditionalPackages[2].RefName, generatePkgiNameWithVersion(msPackageInstall.Spec.PackageRef.RefName, msPackageInstall.Spec.PackageRef.VersionSelection.Constraints))
}

