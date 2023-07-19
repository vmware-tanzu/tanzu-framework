# Tanzu Cluster Essentials

Cluster essentials contains the basic set of components that are required to set up the Tanzu Runtime Core. The Tanzu Cluster Essentials include the following:

1. Kapp-controller
2. secret-gen controller.

## APIs

Install(ctx context.Context, config *rest.Config, imgpkgBundlePath string, timeout time.Duration, bundleDestDir string) error

Install install the cluster essentials packages
It takes kubernetes rest config of cluster (config), cluster essentials bundler path (imgpkgBundlePath)
timeout for blocking call (if zero is given, it will use default timeout of 15 min),
bundleDestDir to use as temporary dir to download bundles(if empty, it will use \tmp as default path)
It return any error encountered

## Usage Example

```go
import (
        "github.com/vmware-tanzu/tanzu-framework/cluster-essentials/pkg/client"
        "k8s.io/client-go/tools/clientcmd"
        "time"
        "context"
)
...
        clusterEssentialRepo := "public.ecr.aws/f1l6q4s3/cluster-essentials"
        clusterEssentialVersion := "v0.0.1"
        clusterKubeconfigPath := "" //kube config file with absolute path
        config, err := clientcmd.LoadFromFile(clusterKubeconfigPath)
        rawConfig, err := clientcmd.Write(*config)
        restConfig, err := clientcmd.RESTConfigFromKubeConfig(rawConfig)
        timeout := time.Duration(0)
        ctx := context.Background()
        err = client.Install(ctx, restConfig, clusterEssentialRepo, clusterEssentialVersion, timeout)
```
