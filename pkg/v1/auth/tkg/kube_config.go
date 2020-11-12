package tkgauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func PrepareKubeconfigWithPinnipedPlugin(clusterInfoPath string, endpoint string) (string, string, error) {
	config, err := clientcmd.LoadFromFile(clusterInfoPath)
	if err != nil {
		return "", "", errors.Wrapf(err, "Error loading from the clusterInfo file")
	}

	clustername := ""
	//clusters := config.Clusters
	for clustername = range config.Clusters {
		break
	}
	username := "tanzu-cli-" + clustername
	AuthInfos, err := getUserInfoWithPinnipedPlugin(endpoint, username)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to generate kubeconfig authInfo data with pinniped plugin ")
	}

	contextName := fmt.Sprintf("%s@%s", username, clustername)
	contexts := map[string]*clientcmdapi.Context{
		contextName: &clientcmdapi.Context{
			Cluster:  clustername,
			AuthInfo: username,
		},
	}

	config.AuthInfos = AuthInfos
	config.Contexts = contexts
	config.CurrentContext = contextName

	filename, err := CreateTempFile("", "tmp_kubeconfig")
	if err != nil {
		return "", "", errors.Wrap(err, "unable to save kubeconfig to temporary file")
	}
	err = clientcmd.WriteToFile(*config, filename)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to write kubeconfig to temporary file")
	}

	return filename, contextName, nil
}

func getUserInfoWithPinnipedPlugin(endpoint, username string) (map[string]*clientcmdapi.AuthInfo, error) {

	var pinnipedUserInfo string = `
	{
		"exec": {
	  	 "apiVersion": "client.authentication.k8s.io/v1beta1",
	  	 "args": [
		 	 "login",
		 	 "oidc",
		 	 "--issuer",
		 	 "%s",
		  	"--client-id",
		  	"pinniped-cli",
		  	"--listen-port",
		  	"48095"
	   	 ],
	   	 "command": "/Users/pkalle/project/pinniped/pinniped",
	     "installHint": "The Pinniped CLI is required to authenticate to the current cluster.\nFor more information, please visit https://pinniped.dev"
		}
	 }`

	user := fmt.Sprintf(pinnipedUserInfo, endpoint)
	authInfo := &clientcmdapi.AuthInfo{}

	if err := json.Unmarshal([]byte(user), authInfo); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal Users data to AuthInfos map")
	}
	authInfos := map[string]*clientcmdapi.AuthInfo{}
	authInfos[username] = authInfo
	return authInfos, nil

}

// GetServerKubernetesVersion uses the kubeconfig to get the server k8s version.
func GetServerKubernetesVersion(kubeconfigPath, context string) (string, error) {
	var discoveryClient discovery.DiscoveryInterface
	kubeConfigBytes, err := loadKubeconfigAndEnsureContext(kubeconfigPath, context)
	if err != nil {
		return "", errors.Errorf("unable to read kubeconfig")
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigBytes)
	if err != nil {
		return "", errors.Errorf("Unable to set up rest config due to : %v", err)
	}

	discoveryClient, err = discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return "", errors.Errorf("Error getting discovery client due to : %v", err)
	}

	if _, err := discoveryClient.ServerVersion(); err != nil {
		return "", errors.Errorf("Failed to invoke API on cluster : %v", err)
	}

	return "", nil
}
func getTanzuLocalKubeDir() (string, error) {
	tanzuConfigDir, err := client.LocalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(tanzuConfigDir, ".kube"), nil
}

// CreateTempFile creates temporary file
func CreateTempFile(dir, prefix string) (string, error) {
	f, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func loadKubeconfigAndEnsureContext(kubeConfigPath, context string) ([]byte, error) {
	config, err := clientcmd.LoadFromFile(kubeConfigPath)

	if err != nil {
		return []byte{}, err
	}
	if context != "" {
		config.CurrentContext = context
	}

	return clientcmd.Write(*config)
}
