// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package imageop

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/yaml"
)

type PublishImagesFromTarOptions struct {
	TkgTarFilePath             string
	CustomImageRepoCertificate string
	PkgClient                  ImgpkgClient
	DestinationRepository      string
	Insecure                   bool
}

var pushImage = &PublishImagesFromTarOptions{}

var PublishImagesfromtarCmd = &cobra.Command{
	Use:          "upload-bundle",
	Short:        "Upload images to private repository.",
	RunE:         publishImagesFromTar,
	SilenceUsage: false,
}

func init() {
	PublishImagesfromtarCmd.Flags().StringVarP(&pushImage.TkgTarFilePath, "source-directory", "", "", "Path to the directory that contains the TAR file  (required)")
	_ = PublishImagesfromtarCmd.MarkFlagRequired("source-directory")
	PublishImagesfromtarCmd.Flags().StringVarP(&pushImage.DestinationRepository, "destination-repo", "", "", "Private OCI repository where the images should be hosted in air-gapped (required)")
	_ = PublishImagesfromtarCmd.MarkFlagRequired("destination-repo")
	PublishImagesfromtarCmd.Flags().StringVarP(&pushImage.CustomImageRepoCertificate, "destination-ca-certificate", "", "", "The private repository’s CA certificate  (optional)")
	PublishImagesfromtarCmd.Flags().BoolVarP(&pushImage.Insecure, "destination-insecure", "", false, "Trusts the private repository’s certificate without validating it (optional)")
}

func (p *PublishImagesFromTarOptions) PushImageToRepo() error {
	yamlFile := path.Join(p.TkgTarFilePath, "publish-images-fromtar.yaml")
	yfile, err := os.ReadFile(yamlFile)
	if err != nil {
		return errors.Wrapf(err, "Error while reading %s file", yamlFile)
	}

	data := make(map[string]string)
	err = yaml.Unmarshal(yfile, &data)

	if err != nil {
		return errors.Wrapf(err, "Error while parsing publish-images-fromtar.yaml file")
	}
	group, _ := errgroup.WithContext(context.Background())
	for tarfile, path := range data {
		fileName := filepath.Join(p.TkgTarFilePath, tarfile)
		destPath := filepath.Join(p.DestinationRepository, path)
		group.Go(
			func() error {
				err = p.PkgClient.CopyImageFromTar(fileName, destPath, p.CustomImageRepoCertificate, p.Insecure)
				if err != nil {
					return err
				}
				return nil
			})
	}
	err = group.Wait()
	if err != nil {
		return errors.Wrap(err, "error while uploading the images")
	}

	return nil
}

func publishImagesFromTar(cmd *cobra.Command, args []string) error {
	pushImage.PkgClient = &imgpkgClient{}
	if !pushImage.Insecure && pushImage.CustomImageRepoCertificate == "" {
		return fmt.Errorf("CA certificate is empty and Insecure option is disabled")
	}

	err := pushImage.PushImageToRepo()
	if err != nil {
		return err
	}
	return nil
}
