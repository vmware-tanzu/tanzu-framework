package client

import (
	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
)

// CoreRepositoryName is the core repository name.
const CoreRepositoryName = "core"

// CoreGCPBucketRepository is the default GCP bucket repository.
var CoreGCPBucketRepository = clientv1alpha1.GCPPluginRepository{
	BucketName: "tanzu-cli",
	Name:       CoreRepositoryName,
}

// TMCRepositoryName is the TMC repository name.
const TMCRepositoryName = "tmc"

// TMCGCPBucketRepository is the GCP bucket repository for TMC plugins.
var TMCGCPBucketRepository = clientv1alpha1.GCPPluginRepository{
	BucketName: "tmc-cli-plugins",
	Name:       TMCRepositoryName,
}

// TKGRepositoryName is the TKG repository name.
const TKGRepositoryName = "tkg"

// TKGGCPBucketRepository is the GCP bucket repository for TKG plugins.
var TKGGCPBucketRepository = clientv1alpha1.GCPPluginRepository{
	BucketName: "tanzu-cli-tkg-plugins",
	Name:       TKGRepositoryName,
}

// DefaultRepositories are the default repositories for the CLI.
var DefaultRepositories []clientv1alpha1.PluginRepository = []clientv1alpha1.PluginRepository{
	clientv1alpha1.PluginRepository{
		GCPPluginRepository: &CoreGCPBucketRepository,
	},
	clientv1alpha1.PluginRepository{
		GCPPluginRepository: &TMCGCPBucketRepository,
	},
	clientv1alpha1.PluginRepository{
		GCPPluginRepository: &TKGGCPBucketRepository,
	},
}
