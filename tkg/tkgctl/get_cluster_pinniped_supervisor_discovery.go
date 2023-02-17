package tkgctl

import (
	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
)

// TODO(Ben): rename and drop "cluster"
type GetClusterPinnipedSupervisorDiscoveryOptions struct {
	// the .well-known/openid-configuration discovery endpoint for a pinniped supervisor
	Endpoint string
	// a certificate bundle to trust in order to communicate with the pinniped supervisor
	CABundle string
}

// NOTE: like tkgctl.GetClusterPinnipedInfo() this should really just be a wrapper to
// make a call into tkgClient.GetClusterPinnipedInfo()
func (t *tkgctl) GetPinnipedSupervisorDiscovery(options GetClusterPinnipedSupervisorDiscoveryOptions) (*client.PinnipedSupervisorDiscoveryInfo, error) {

	getClusterPinnipedSupervisorDiscoveryOptions := client.GetPinnipedSupervisorDiscoveryOptions{
		Endpoint: options.Endpoint,
		CABundle: options.CABundle,
	}

	discoveryInfo, err := t.tkgClient.GetPinnipedSupervisorDiscovery(getClusterPinnipedSupervisorDiscoveryOptions)
	if err != nil {
		return nil, err
	}
	return discoveryInfo, nil
}
