package clusterbootstrapclone

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClusterbootstrapclone(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterbootstrapClone Suite")
}
