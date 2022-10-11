// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aunum/log"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkg/carvelhelpers"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/oracle"
)

var oracleCmd = &cobra.Command{
	Use:   "oracle",
	Short: "TKR OS image operations for Oracle Cloud Infrastructure",
}

var (
	// OCID of the destination OCI compartment to import the public image into
	compartmentID string
)

var oraclePopulate = &cobra.Command{
	Use:   "populate",
	Short: "populate public OS Image specified in TKR to private OCI compartment and region",
	Example: `
	# Import a public image to the specified OCI compartment and region
	tanzu cluster osimage oracle populate --image https://objectstorage.us-sanjose-1.oraclecloud.com/n/axxxxxxxxxx8/b/exported-node-images/o/ubuntu-2004 \
		--tkr-path projects-stg.registry.vmware.com/tkg/tkr-oci:v1.23.5 --compartment <compartment ocid>`,
	Run: ociPopulateCmdInitRun,
}

func init() {
	oraclePopulate.PersistentFlags().StringVar(&compartmentID, "compartment", "", "The destination OCI compartment to import the public image into")

	oracleCmd.AddCommand(oraclePopulate)
}

// writeToFile writes a Kubernetes runtime.Object to the file, given its paths
func writeToFile(object runtime.Object, path string) error {
	serializer := k8sjson.NewSerializerWithOptions(
		k8sjson.DefaultMetaFactory, nil, nil,
		k8sjson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := serializer.Encode(object, file); err != nil {
		return err
	}
	return nil
}

// getTKRFromManifest reads the TanzuKubernetesRelease from its yaml manifest given path
func getTKRFromManifest(path string) (*v1alpha3.TanzuKubernetesRelease, error) {
	var tkr v1alpha3.TanzuKubernetesRelease
	tkrFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer tkrFile.Close()
	if err := utilyaml.NewYAMLOrJSONDecoder(tkrFile, 100).Decode(&tkr); err != nil {
		return nil, err
	}
	return &tkr, nil
}

// getTKRPath returns the absolute path for the TKR manifest
func getTKRPath() string {
	// default TKR manifest file name in the bundle
	tkrFileName := "TanzuKubernetesRelease.yml"
	return filepath.Join(outputDirectory, "config", tkrFileName)
}

// getOSImage returns the absolute path for the OSImage manifest to create
func getOSImage() string {
	return filepath.Join(outputDirectory, "config", fmt.Sprintf("%s.yaml", name))
}

// getOSImageName calculates the default OSImage name, given provided flags
func getOSImageName(tkrK8sVersion string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", osName, osType, osVersion, osArch, tkrK8sVersion)
}

// patchTKrWithOSImage patches the TKR bundle with the imported OSImage information
func patchTKrWithOSImage(tkr *v1alpha3.TanzuKubernetesRelease, imageID, compartment, region string) (err error) {
	tkr.Spec.OSImages = append(tkr.Spec.OSImages, v1.LocalObjectReference{Name: name})
	osImage := &v1alpha3.OSImage{
		TypeMeta: metav1.TypeMeta{Kind: "OSImage", APIVersion: v1alpha3.GroupVersion.Version},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.TkgNamespace,
		},
		Spec: v1alpha3.OSImageSpec{
			KubernetesVersion: tkr.Spec.Kubernetes.Version,
			OS: v1alpha3.OSInfo{
				Type:    osType,
				Name:    osName,
				Version: osVersion,
				Arch:    osArch,
			},
			Image: v1alpha3.MachineImageInfo{
				Type: "oci",
				Ref: map[string]interface{}{
					"id":          &imageID,
					"compartment": &compartment,
					"region":      region,
				}},
		},
	}
	if err := writeToFile(tkr, getTKRPath()); err != nil {
		return err
	}
	if err := writeToFile(osImage, getOSImage()); err != nil {
		return err
	}
	return nil
}

func ociPopulateCmdInitRun(_ *cobra.Command, _ []string) {
	var err error

	// download TKR bundle and load the TKR resource
	if outputDirectory == "" {
		outputDirectory, err = os.MkdirTemp("", "oci_image")
		if err != nil {
			log.Fatalf("unable to create a temporary directory: %v", err)
		}
	}
	if err = carvelhelpers.DownloadImageBundleAndSaveFilesToDir(tkrRegistryPath, outputDirectory); err != nil {
		log.Fatalf("unable to fetch TKR package image from %s: %v", tkrRegistryPath, err.Error())
	}
	log.Infof("downloaded TKR from %s to %s", tkrRegistryPath, outputDirectory)

	tkr, err := getTKRFromManifest(getTKRPath())
	if err != nil {
		log.Fatalf("unable to read TKR package image from %s: %v", getTKRPath(), err.Error())
	}
	if name == "" {
		name = getOSImageName(tkr.Spec.Kubernetes.Version)
	}
	log.Infof("start importing image: %s", name)

	// import the BYOI image from the public endpoint to user's compartment
	oracleClient, err := oracle.New()
	if err != nil {
		log.Fatal("unable to get Oracle client")
	}
	region, err := oracleClient.Region()
	if err != nil {
		log.Fatal("unable to determine region from OCI config provider")
	}
	ctx := context.Background()
	_, err = oracleClient.EnsureCompartmentExists(ctx, compartmentID)
	if err != nil {
		log.Fatalf("unable to get compartment %s or not exists: %v", compartmentID, err.Error())
	}
	log.Infof("destination compartment %s exists", compartmentID)

	log.Infof("start creating customized image from public endpoint %s in compartment %s", imageEndpoint, compartmentID)
	image, err := oracleClient.ImportImageSync(ctx, name, compartmentID, imageEndpoint)
	if err != nil {
		log.Fatalf("unable to import image: %v", err.Error())
	}
	log.Infof("created customized image from public endpoint %s in compartment %s, ocid is %s", imageEndpoint, compartmentID, *image.Id)

	// patch the TKR bundle with the imported OSImage
	if err = patchTKrWithOSImage(tkr, *image.Id, compartmentID, region); err != nil {
		log.Fatalf("unable to patch TKR package in %s: %v", outputDirectory, err)
	}
	log.Infof("TKR has been patched! To consume, \n\t\t tanzu management-cluster create -f <cluster-config>.yaml --additional-manifests %s", outputDirectory)
}
