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
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkg/carvelhelpers"
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

// getOSImageFromManifest reads the OSImage from its yaml manifest given path
func getOSImageFromManifest(path string) (*v1alpha3.OSImage, error) {
	var osImage v1alpha3.OSImage
	osImageFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer osImageFile.Close()
	if err := utilyaml.NewYAMLOrJSONDecoder(osImageFile, 100).Decode(&osImage); err != nil {
		return nil, err
	}
	return &osImage, nil
}

// TKRPath returns the absolute path for the TKR manifest
func TKRPath() string {
	// default TKR manifest file name in the bundle
	tkrFileName := "TanzuKubernetesRelease.yml"
	return filepath.Join(outputDirectory, "config", tkrFileName)
}

// OSImagePath returns the absolute path for the OSImage manifest to create, provided its name
func OSImagePath(name string) string {
	return filepath.Join(outputDirectory, "config", fmt.Sprintf("OSImage-%s.yaml", name))
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

	tkr, err := getTKRFromManifest(TKRPath())
	if err != nil {
		log.Fatalf("unable to read TKR package image from %s: %v", TKRPath(), err.Error())
	}

	// import the BYOI image from the public endpoint to user's compartment
	oracleClient, err := oracle.New()
	if err != nil {
		log.Fatalf("unable to get Oracle client: %v", err)
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

	var waitGroup errgroup.Group
	for _, osImageRef := range tkr.Spec.OSImages {
		name := osImageRef.Name
		waitGroup.Go(func() error {
			osImage, err := getOSImageFromManifest(OSImagePath(name))
			if err != nil {
				return err
			}
			if osImage.Spec.Image.Type != "oci" {
				log.Infof("skip patching image %s, image type is %s, not 'oci'", osImage.Spec.Image.Type)
				return nil
			}
			mapRef := osImage.Spec.Image.Ref
			imageURL, exists := mapRef["imageURL"]
			if !exists {
				log.Infof("skip patching image %s, image ref does not contains key 'imageURL'")
				return nil
			}
			imageURLStr, ok := imageURL.(string)
			if !ok {
				log.Infof("skip patching image %s, image ref value of key 'imageURL' is not a string")
				return nil
			}

			image, err := oracleClient.ImportImageSync(ctx, name, compartmentID, imageURLStr)
			if err != nil {
				return err
			}
			log.Infof("finish importing image %s, ocid is %s", name, *image.Id)

			newOSImage := osImage.DeepCopy()
			newOSImage.Spec.Image.Ref["compartment"] = compartmentID
			newOSImage.Spec.Image.Ref["id"] = *image.Id
			newOSImage.Spec.Image.Ref["region"] = region
			if err := writeToFile(newOSImage, OSImagePath(name)); err != nil {
				log.Infof("failed to patch OSImage %s: %v", name, err)
				return err
			}
			return nil
		})
	}
	if err := waitGroup.Wait(); err != nil {
		log.Fatal(err)
	}
	log.Infof("TKR has been patched! To consume, \n\t\t tanzu management-cluster create -f <cluster-config>.yaml --additional-manifests %s", outputDirectory)
}
