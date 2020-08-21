package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/mod/semver"
)

// CoreName is the name of the core binary.
const CoreName = "core"

// HasUpdate tells whether the core plugin has an update.
func HasUpdate(repo Repository) (update bool, version string, err error) {
	manifest, err := repo.Manifest()
	if err != nil {
		return update, version, err
	}
	valid := semver.IsValid(manifest.Version)
	if !valid {
		err = fmt.Errorf("core manifest version %q is not a valid semantic version", manifest.Version)
	}
	valid = semver.IsValid(BuildVersion)
	if !valid {
		err = fmt.Errorf("core build version %q is not a valid semantic version", BuildVersion)
	}
	compared := semver.Compare(manifest.Version, BuildVersion)
	if compared == 1 {
		return true, manifest.Version, nil
	}
	return false, version, nil
}

// Update the core CLI.
func Update(repo Repository) error {
	var executable string

	update, version, err := HasUpdate(repo)
	if err != nil {
		return err
	}
	if !update {
		return nil
	}
	b, err := repo.Fetch(CoreName, version, BuildArch())
	if err != nil {
		return err
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	newCliFile := filepath.Join(dir, "tanzu_new")
	outFile, err := os.OpenFile(newCliFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, bytes.NewReader(b))
	outFile.Close()

	if BuildArch().IsWindows() {
		executable = outFile.Name() + ".exe"
		err = os.Rename(outFile.Name(), executable)
		if err != nil {
			return err
		}
	} else {
		executable, err = os.Executable()
		if err != nil {
			return err
		}
		err = os.Rename(outFile.Name(), executable)
		if err != nil {
			return err
		}
	}
	return nil
}
