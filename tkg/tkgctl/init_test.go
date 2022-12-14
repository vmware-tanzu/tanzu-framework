package tkgctl

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
)

var _ = Describe("Init", func() {
	var (
		tkgClient              *fakes.Client
		tkgconfigreaderwriter  *fakes.TKGConfigReaderWriter
		initRegionOptions      InitRegionOptions
		err                    error
		tkgctlClient           *tkgctl
		tkgConfigUpdaterClient *fakes.TKGConfigUpdaterClient
		tkgBomClient           *fakes.TKGConfigBomClient
	)

	BeforeEach(func() {
		tkgClient = &fakes.Client{}
		tkgconfigreaderwriter = &fakes.TKGConfigReaderWriter{}
		tkgConfigUpdaterClient = &fakes.TKGConfigUpdaterClient{}
		tkgBomClient = &fakes.TKGConfigBomClient{}

		tkgctlClient = &tkgctl{
			tkgClient:              tkgClient,
			tkgConfigReaderWriter:  tkgconfigreaderwriter,
			tkgConfigUpdaterClient: tkgConfigUpdaterClient,
			tkgBomClient:           tkgBomClient,
		}
		initRegionOptions = InitRegionOptions{
			Plan:                   "dev",
			ClusterName:            "foobar",
			InfrastructureProvider: "FOOBAR",
			CniType:                "calico",
			UseExistingCluster:     true,
			UI:                     false,
			ClusterConfigFile:      "../fakes/config/config.yaml",
		}

	})

	Context("When _ALLOW_CALICO_ON_MANAGEMENT_CLUSTER is not set", func() {
		It("should return an error", func() {
			err = tkgctlClient.Init(initRegionOptions)
			Expect(err.Error()).To(Equal("Calico management-cluster creation is forbidden..."))
		})
	})

	Context("When _ALLOW_CALICO_ON_MANAGEMENT_CLUSTER is set", func() {
		BeforeEach(func() {
			os.Setenv("_ALLOW_CALICO_ON_MANAGEMENT_CLUSTER", "true")
		})

		It("should succeed", func() {
			err = tkgctlClient.Init(initRegionOptions)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			os.Unsetenv("_ALLOW_CALICO_ON_MANAGEMENT_CLUSTER")
		})
	})
})
