package util

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached/memory"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
)

type GVRHelper interface {
	GetGVR(gk schema.GroupKind) (*schema.GroupVersionResource, error)
	GetDiscoveryClient() discovery.DiscoveryInterface
}

type gvrHelper struct {
	cachedDiscoveryClient discovery.CachedDiscoveryInterface
	cachedLookups         map[schema.GroupKind]*schema.GroupVersionResource
	context               context.Context
}

func NewGVRHelper(ctx context.Context, discoveryClient discovery.DiscoveryInterface) GVRHelper {
	cachedDiscoveryClient := cacheddiscovery.NewMemCacheClient(discoveryClient)
	helper := &gvrHelper{
		cachedDiscoveryClient: cachedDiscoveryClient,
		cachedLookups:         make(map[schema.GroupKind]*schema.GroupVersionResource),
		context:               ctx,
	}
	go helper.periodicGVRCachesClean()
	return helper
}

// GetGVR returns a GroupVersionResource for a GroupKind
func (g *gvrHelper) GetGVR(gk schema.GroupKind) (*schema.GroupVersionResource, error) {
	if gvr, ok := g.cachedLookups[gk]; ok {
		return gvr, nil
	}
	gvr, err := g.gvrForGroupKind(gk)
	if err != nil {
		return nil, err
	}
	g.cachedLookups[gk] = gvr
	return gvr, nil
}

func (g *gvrHelper) GetDiscoveryClient() discovery.DiscoveryInterface {
	return g.cachedDiscoveryClient
}

// periodicGVRCachesClean invalidates caches used for GVR lookup
func (g *gvrHelper) periodicGVRCachesClean() {
	ticker := time.NewTicker(constants.DiscoveryCacheInvalidateInterval)
	for {
		select {
		case <-g.context.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			g.cachedDiscoveryClient.Invalidate()
			g.cachedLookups = make(map[schema.GroupKind]*schema.GroupVersionResource)
		}
	}
}

func (g *gvrHelper) gvrForGroupKind(gk schema.GroupKind) (*schema.GroupVersionResource, error) {
	apiResourceList, err := g.cachedDiscoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	for _, apiResource := range apiResourceList {
		gv, err := schema.ParseGroupVersion(apiResource.GroupVersion)
		if err != nil {
			return nil, err
		}
		if gv.Group == gk.Group {
			for i := 0; i < len(apiResource.APIResources); i++ {
				if apiResource.APIResources[i].Kind == gk.Kind {
					return &schema.GroupVersionResource{Group: gv.Group, Resource: apiResource.APIResources[i].Name, Version: gv.Version}, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("unable to find server preferred resource %s/%s", gk.Group, gk.Kind)
}
