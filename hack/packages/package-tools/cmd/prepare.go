// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	_ "embed" // required to embed file
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/vmware-tanzu/tanzu-framework/hack/packages/package-tools/utils"
)

// prepareCmd is for preparing package tooling for building packages
var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare for building packages",
	Long:  "Prepare is for preparing package tooling like downloading Carvel binaries etc. for building packages",
	RunE:  runPrepare,
}

//go:embed config/carvel-tools-config.yaml
var carvelTools []byte

var clean bool

func init() {
	rootCmd.AddCommand(prepareCmd)
	prepareCmd.Flags().BoolVar(&clean, "clean", false, "Deletes the existing Carvel tool binaries")
}

func runPrepare(cmd *cobra.Command, args []string) error {
	if err := downloadCarvelBinaries(); err != nil {
		return fmt.Errorf("couldn't download carvel binaries: %w", err)
	}
	return nil
}

// downloadCarvelBinaries downloads the carvel binaries and places them in hack/tools/bin in project root dir
func downloadCarvelBinaries() error {
	c := new(CarvelTools)
	err := yaml.Unmarshal(carvelTools, c)
	if err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	projectRootDir, err := utils.GetProjectRootDir()
	if err != nil {
		return err
	}

	toolsBinDir := filepath.Join(projectRootDir, "hack", "tools", "bin")

	// clean hack/tools/bin directory
	if clean {
		if _, err := os.Stat(toolsBinDir); !os.IsNotExist(err) {
			if err := os.RemoveAll(toolsBinDir); err != nil {
				return fmt.Errorf("couldn't delete carvel binaries: %w", err)
			}
		}
	}

	if err := utils.CreateDir(toolsBinDir); err != nil {
		return err
	}

	for _, tool := range c.Tools {
		fmt.Printf("Downloading %q binary, version: %q \n", tool.Name, tool.Version)

		// resolve the url template and get the url
		t, err := template.New("url").Parse(tool.Url)
		if err != nil {
			return err
		}
		var url bytes.Buffer
		err = t.Execute(&url, struct {
			VERSION string
			OS      string
			ARCH    string
		}{
			tool.Version,
			runtime.GOOS,
			runtime.GOARCH,
		})
		if err != nil {
			return err
		}

		resp, err := http.Get(url.String())
		if err != nil {
			return fmt.Errorf("couldn't download %s binary: %w", tool.Name, err)
		}
		defer resp.Body.Close()

		// Create the binary file
		out, err := os.OpenFile(filepath.Join(toolsBinDir, tool.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}
	}
	return nil
}
