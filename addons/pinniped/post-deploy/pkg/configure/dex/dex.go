package dex

import (
	"context"

	certmanagerclientset "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/constants"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/schemas"
	"github.com/vmware-tanzu-private/core/addons/pinniped/post-deploy/pkg/vars"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Configurator struct {
	CertmanagerClientset certmanagerclientset.Interface
	K8SClientset         kubernetes.Interface
}

type DexInfo struct {
	DexSvcEndpoint        string
	SupervisorSvcEndpoint string
	DexNamespace          string
	DexConfigmapName      string
	ClientSecret          string
}

func (c Configurator) CreateOrUpdateDexConfigMap(ctx context.Context, dexInfo DexInfo) error {
	var err error
	zap.S().Info("Creating the ConfigMap of Dex")

	// create configmap under kube-public namespace
	var dexConfigMap *corev1.ConfigMap
	dexConfigMap, err = c.K8SClientset.CoreV1().ConfigMaps(vars.DexNamespace).Get(ctx, vars.DexConfigMapName, metav1.GetOptions{})
	if err != nil {
		zap.S().Error(err)
		return err
	}

	configStr := dexConfigMap.Data["config.yaml"]

	dexConf := &schemas.DexConfig{}
	err = yaml.Unmarshal([]byte(configStr), dexConf)
	if err != nil {
		zap.S().Error(err)
		return err
	}

	// change dex config values
	dexConf.Issuer = dexInfo.DexSvcEndpoint
	dexConf.StaticClients[0] = &schemas.StaticClient{
		Id:           constants.DexClientID,
		Name:         constants.DexClientID,
		RedirectURIs: []string{dexInfo.SupervisorSvcEndpoint + "/callback"},
		Secret:       dexInfo.ClientSecret,
	}
	for _, connector := range dexConf.Connectors {
		if connector.Type == "oidc" {
			connector.Config.RedirectURI = dexInfo.DexSvcEndpoint + "/callback"
		}
	}

	out, err := yaml.Marshal(dexConf)
	if err != nil {
		zap.S().Error(err)
		return err
	}

	// update dex config map
	copiedConfigMap := dexConfigMap.DeepCopy()
	copiedConfigMap.Data = map[string]string{
		"config.yaml": string(out),
	}
	_, err = c.K8SClientset.CoreV1().ConfigMaps(vars.DexNamespace).Update(ctx, copiedConfigMap, metav1.UpdateOptions{})
	if err != nil {
		zap.S().Error(err)
		return err
	}

	zap.S().Infof("Updated the ConfigMap %s/%s for Dex", vars.DexNamespace, vars.DexConfigMapName)
	return nil
}
