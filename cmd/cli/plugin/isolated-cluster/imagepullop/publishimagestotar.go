// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package imagepullop define imgpkg pull command
package imagepullop

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/yaml"

	tkrv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/pkg/tkr/v1"
	imgpkginterface "github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/isolated-cluster/imgpkginterface"
)

const outputDir = "tmp"

type PublishImagesToTarOptions struct {
	TkgImageRepo                string
	TkgVersion                  string
	PkgClient                   imgpkginterface.ImgpkgClient
	ImageDetails                map[string]string
	CaCertificate               string
	Insecure                    bool
	TkgCustomCompatibilityImage string
}

var pullImage = &PublishImagesToTarOptions{}

var PublishImagestotarCmd = &cobra.Command{
	Use:   "download-bundle",
	Short: "Download images/bundle from source repo into local disk (at current directory) as TAR files",
	Example: `
        # download images/bundle for TKG version v2.1.0 into local disk from repo projects.registry.vmware.com and authenticate source repo with default system CA certificate
	tanzu isolated-cluster download-bundle --source-repo mirror-registry.test/tkg --tkg-version v2.1.0 

	# download images/bundle for TKG version v2.1.0 into local disk from an internal registry without validating the certificate
        tanzu isolated-cluster download-bundle --source-repo mirror-registry.test/tkg --tkg-version v2.1.0 --insecure

	# download images/bundle for TKG version v2.1.0 into local disk from an internal mirror using a self-signed certificate
        tanzu isolated-cluster download-bundle --source-repo mirror-registry.test/tkg --tkg-version v2.1.0 --ca-certificate registry.crt
`,

	RunE:         downloadImagesToTar,
	SilenceUsage: true,
}

func init() {
	PublishImagestotarCmd.Flags().StringVarP(&pullImage.TkgImageRepo, "source-repo", "", "projects.registry.vmware.com/tkg", "OCI repo where TKG bundles or images are hosted")
	PublishImagestotarCmd.Flags().StringVarP(&pullImage.TkgVersion, "tkg-version", "", "", "TKG version (required)")
	_ = PublishImagestotarCmd.MarkFlagRequired("tkg-version")
	PublishImagestotarCmd.Flags().BoolVarP(&pullImage.Insecure, "insecure", "", false, "Trusts the server certificate without validating it (optional)")
	PublishImagestotarCmd.Flags().StringVarP(&pullImage.CaCertificate, "ca-certificate", "", "", "The private repository’s CA certificate  (optional)")
	PublishImagestotarCmd.Flags().StringVarP(&pullImage.TkgCustomCompatibilityImage, "tkg-custom-compatibility-image-path", "", "", "TKG compatibility image path in special case(optional)")
	pullImage.ImageDetails = map[string]string{}
}

func (p *PublishImagesToTarOptions) DownloadTkgCompatibilityImage() error {
	if p.TkgVersion == "" {
		return errors.New("TKG Version is empty")
	}

	tkgCompatibilityRelativeImagePath := "tkg-compatibility"

	if p.TkgCustomCompatibilityImage != "" {
		tkgCompatibilityRelativeImagePath = p.TkgCustomCompatibilityImage
	} else if !imgpkginterface.IsTKGRTMVersion(p.TkgVersion) {
		tkgCompatibilityRelativeImagePath = path.Join(p.TkgVersion, tkgCompatibilityRelativeImagePath)
	}

	tkgCompatibilityImagePath := path.Join(p.TkgImageRepo, tkgCompatibilityRelativeImagePath)
	imageTags := p.PkgClient.GetImageTagList(tkgCompatibilityImagePath)
	if len(imageTags) == 0 {
		return errors.New("image doesn't have any tags")
	}
	sourceImageName := tkgCompatibilityImagePath + ":" + imageTags[len(imageTags)-1]
	tarFilename := "tkg-compatibility" + "-" + imageTags[len(imageTags)-1] + ".tar"
	err := p.PkgClient.CopyImageToTar(sourceImageName, tarFilename, p.CaCertificate, p.Insecure)
	if err != nil {
		return err
	}
	p.ImageDetails[tarFilename] = tkgCompatibilityRelativeImagePath
	return nil
}

func (p *PublishImagesToTarOptions) DownloadTkgBomAndComponentImages() (string, error) {
	if p.TkgImageRepo == "" || p.TkgVersion == "" {
		return "", errors.New("input parameter TkgImageRepo or TkgVersion is empty")
	}
	tkgBomImagePath := path.Join(p.TkgImageRepo, "tkg-bom")

	sourceImageName := tkgBomImagePath + ":" + p.TkgVersion
	tarnames := "tkg-bom" + "-" + p.TkgVersion + ".tar"
	err := p.PkgClient.CopyImageToTar(sourceImageName, tarnames, p.CaCertificate, p.Insecure)
	if err != nil {
		return "", errors.New("error while downloading tkg-bom")
	}
	p.ImageDetails[tarnames] = tkgBomImagePath
	err = p.PkgClient.PullImage(sourceImageName, outputDir)
	if err != nil {
		return "", err
	}
	// read the tkg-bom file
	tkgBomFilePath := filepath.Join(outputDir, fmt.Sprintf("tkg-bom-%s.yaml", p.TkgVersion))
	b, err := os.ReadFile(tkgBomFilePath)

	// read the tkg-bom file
	if err != nil {
		return "", errors.Wrapf(err, "read tkg-bom file from %s faild", tkgBomFilePath)
	}
	tkgBom, _ := tkrv1.NewBom(b)
	tempImageDetails := map[string]string{}

	// imgpkg copy each component's artifacts
	components, _ := tkgBom.Components()
	group, _ := errgroup.WithContext(context.Background())
	group.SetLimit(30)
	for _, compInfos := range components {
		for _, compInfo := range compInfos {
			for _, imageInfo := range compInfo.Images {
				sourceImageName = path.Join(p.TkgImageRepo, imageInfo.ImagePath) + ":" + imageInfo.Tag
				tarname := imgpkginterface.ReplaceSlash(imageInfo.ImagePath) + "-" + imageInfo.Tag + ".tar"
				tempImageDetails[tarname] = imageInfo.ImagePath
				i := sourceImageName
				j := tarname
				group.Go(func() error {
					return p.PkgClient.CopyImageToTar(i, j, p.CaCertificate, p.Insecure)
				})
			}
		}
	}
	err = group.Wait()
	if err != nil {
		return "", errors.Wrap(err, "error while downloading images")
	}
	maps.Copy(p.ImageDetails, tempImageDetails)
	return tkgBom.GetCompatibility(), nil
}

func (p *PublishImagesToTarOptions) DownloadTkrCompatibilityImage(tkrCompatibilityRelativeImagePath string) (tkgVersion []string, err error) {
	if p.TkgImageRepo == "" || p.TkgVersion == "" {
		return nil, errors.New("input parameter source image repo or TKG Version is empty")
	}

	// get the latest tag of tkr-compatibility image
	tkrCompatibilityImagePath := path.Join(p.TkgImageRepo, tkrCompatibilityRelativeImagePath)
	imageTags := p.PkgClient.GetImageTagList(tkrCompatibilityImagePath)
	if len(imageTags) == 0 {
		return nil, errors.New("image doesn't have any tags")
	}
	// inspect the tkr-compatibility image to get the list of compatible tkrs
	tkrCompatibilityImageURL := tkrCompatibilityImagePath + ":" + imageTags[len(imageTags)-1]

	sourceImageName := tkrCompatibilityImageURL
	err1 := p.PkgClient.PullImage(sourceImageName, outputDir)
	if err1 != nil {
		return nil, err1
	}
	files, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, errors.Wrapf(err, "read directory tmp failed")
	}
	if len(files) != 1 || files[0].IsDir() {
		return nil, fmt.Errorf("tkr-compatibility image should only have exactly one file inside")
	}
	tkrCompatibilityFilePath := filepath.Join(outputDir, files[0].Name())
	b, err := os.ReadFile(tkrCompatibilityFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "read tkr-compatibility file from %s failed", tkrCompatibilityFilePath)
	}
	tkrCompatibility := &tkrv1.CompatibilityMetadata{}
	if err := yaml.Unmarshal(b, tkrCompatibility); err != nil {
		return nil, errors.Wrapf(err, "Unmarshal tkr-compatibility file %s failed", tkrCompatibilityFilePath)
	}
	// find the corresponding tkg-bom entry
	var tkrVersions []string
	var found = false
	for _, compatibilityInfo := range tkrCompatibility.ManagementClusterVersions {
		if compatibilityInfo.TKGVersion == p.TkgVersion {
			found = true
			tkrVersions = compatibilityInfo.SupportedKubernetesVersions
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("couldn't find the corresponding tkg-bom version in the tkr-compatibility file")
	}
	// imgpkg copy the tkr-compatibility image
	sourceImageName = tkrCompatibilityImageURL
	tarFilename := "tkr-compatibility" + "-" + imageTags[len(imageTags)-1] + ".tar"
	err = p.PkgClient.CopyImageToTar(sourceImageName, tarFilename, p.CaCertificate, p.Insecure)
	if err != nil {
		return nil, err
	}
	p.ImageDetails[tarFilename] = tkrCompatibilityRelativeImagePath
	return tkrVersions, nil
}

func (p *PublishImagesToTarOptions) DownloadTkrBomAndComponentImages(tkrVersion string) error {
	if p.TkgImageRepo == "" {
		return errors.New("Source Repo is empty")
	}
	tkrTag := imgpkginterface.UnderscoredPlus(tkrVersion)
	tkrBomImagePath := path.Join(p.TkgImageRepo, "tkr-bom")
	sourceImageName := tkrBomImagePath + ":" + tkrTag
	tarFilename := "tkr-bom" + "-" + tkrTag + ".tar"
	err := p.PkgClient.CopyImageToTar(sourceImageName, tarFilename, p.CaCertificate, p.Insecure)
	if err != nil {
		return err
	}
	p.ImageDetails[tarFilename] = "tkr-bom"
	sourceImageName = tkrBomImagePath + ":" + tkrTag
	err = p.PkgClient.PullImage(sourceImageName, outputDir)
	if err != nil {
		return err
	}
	// read the tkr-bom file
	tkrBomFilePath := path.Join(outputDir, fmt.Sprintf("tkr-bom-%s.yaml", tkrVersion))
	b, err := os.ReadFile(tkrBomFilePath)
	if err != nil {
		return errors.Wrapf(err, "read tkr-bom file from %s failed", tkrBomFilePath)
	}
	tkgBom, _ := tkrv1.NewBom(b)
	// imgpkg copy each component's artifacts
	components, _ := tkgBom.Components()
	tempImageDetails := map[string]string{}
	group, _ := errgroup.WithContext(context.Background())
	group.SetLimit(30)
	for _, compInfos := range components {
		for _, compInfo := range compInfos {
			for _, imageInfo := range compInfo.Images {
				sourceImageName = filepath.Join(p.TkgImageRepo, imageInfo.ImagePath) + ":" + imageInfo.Tag
				tarname := imgpkginterface.ReplaceSlash(imageInfo.ImagePath) + "-" + imageInfo.Tag + ".tar"
				tempImageDetails[tarname] = imageInfo.ImagePath
				i := sourceImageName
				j := tarname
				group.Go(func() error {
					return p.PkgClient.CopyImageToTar(i, j, p.CaCertificate, p.Insecure)
				})
			}
		}
	}
	err = group.Wait()
	if err != nil {
		return errors.Wrap(err, "error while downloading images")
	}
	maps.Copy(p.ImageDetails, tempImageDetails)
	return nil
}

func (p *PublishImagesToTarOptions) DownloadTkgPackagesImages(tkrVersions []string) error {
	tkgBomImagePath := path.Join(p.TkgImageRepo, "tkg-bom")
	sourceImageName := tkgBomImagePath + ":" + p.TkgVersion
	err := p.PkgClient.PullImage(sourceImageName, outputDir)
	if err != nil {
		return err
	}
	tkgBomFilePath := filepath.Join(outputDir, fmt.Sprintf("tkg-bom-%s.yaml", p.TkgVersion))
	b, err := os.ReadFile(tkgBomFilePath)

	// read the tkg-bom file
	if err != nil {
		return errors.Wrapf(err, "read tkg-bom file from %s faild", tkgBomFilePath)
	}

	tkgBom, _ := tkrv1.NewTkgBom(b)

	tkgPackageRepo, err := tkgBom.GetTKRPackageRepo()
	if err != nil {
		return err
	}
	tkgPackageRepoStruct := reflect.ValueOf(tkgPackageRepo)

	tkgPackage, err := tkgBom.GetTKRPackage()
	if err != nil {
		return err
	}
	tkgPackageStruct := reflect.ValueOf(tkgPackage)
	tempImageDetails := map[string]string{}
	group, _ := errgroup.WithContext(context.Background())
	group.SetLimit(30)
	for _, tkrVersion := range tkrVersions {
		tkrVersion = imgpkginterface.UnderscoredPlus(tkrVersion)
		for i := 0; i < tkgPackageRepoStruct.NumField(); i++ {
			imageName := tkgPackageRepoStruct.Field(i).Interface().(string)
			sourceImageName := filepath.Join(p.TkgImageRepo, imageName) + ":" + tkrVersion
			tarname := imageName + "-" + tkrVersion + ".tar"
			tempImageDetails[tarname] = imageName
			i := sourceImageName
			j := tarname
			group.Go(func() error {
				return p.PkgClient.CopyImageToTar(i, j, p.CaCertificate, p.Insecure)
			})
		}
		for i := 0; i < tkgPackageStruct.NumField(); i++ {
			imageName := tkgPackageStruct.Field(i).Interface().(string)
			sourceImageName := filepath.Join(p.TkgImageRepo, imageName) + ":" + tkrVersion
			tarname := imageName + "-" + tkrVersion + ".tar"
			tempImageDetails[tarname] = imageName
			i := sourceImageName
			j := tarname
			group.Go(func() error {
				return p.PkgClient.CopyImageToTar(i, j, p.CaCertificate, p.Insecure)
			})
		}
	}
	err = group.Wait()
	if err != nil {
		return errors.Wrap(err, "error while downloading images")
	}
	maps.Copy(p.ImageDetails, tempImageDetails)
	return nil
}

func downloadImagesToTar(cmd *cobra.Command, args []string) error {
	pullImage.PkgClient = &imgpkginterface.Imgpkg{}
	if !strings.HasPrefix(pullImage.TkgVersion, "v") {
		return fmt.Errorf("invalid TKG Tag %s", pullImage.TkgVersion)
	}
	err := pullImage.DownloadTkgCompatibilityImage()
	if err != nil {
		return err
	}
	tkrCompatibilityRelativeImagePath, err := pullImage.DownloadTkgBomAndComponentImages()

	if err != nil {
		return err
	}
	tkrVersions, err := pullImage.DownloadTkrCompatibilityImage(tkrCompatibilityRelativeImagePath)
	if err != nil {
		return errors.Wrapf(err, "Error while retrieving tkrVersions")
	}

	err = pullImage.DownloadTkgPackagesImages(tkrVersions)
	if err != nil {
		fmt.Printf("TkgPackagesImages Not found %s", err)
	}

	for _, tkrVersion := range tkrVersions {
		err = pullImage.DownloadTkrBomAndComponentImages(tkrVersion)
		if err != nil {
			return err
		}
	}
	data, _ := yaml.Marshal(&pullImage.ImageDetails)
	err2 := os.WriteFile("publish-images-fromtar.yaml", data, 0666)
	if err2 != nil {
		return errors.Wrapf(err2, "Error while writing publish-images-fromtar.yaml file")
	}
	return nil
}
