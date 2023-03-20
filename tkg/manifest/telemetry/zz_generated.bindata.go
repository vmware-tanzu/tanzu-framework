// Code generated by go-bindata. DO NOT EDIT.
// sources:
// tkg/manifest/telemetry/config-aws.yaml
// tkg/manifest/telemetry/config-azure.yaml
// tkg/manifest/telemetry/config-docker.yaml
// tkg/manifest/telemetry/config-vsphere.yaml
// tkg/manifest/telemetry/zz_generated.bindata.go

package telemetry

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  fileInfoEx
}

type fileInfoEx interface {
	os.FileInfo
	MD5Checksum() string
}

type bindataFileInfo struct {
	name        string
	size        int64
	mode        os.FileMode
	modTime     time.Time
	md5checksum string
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) MD5Checksum() string {
	return fi.md5checksum
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _bindataTkgManifestTelemetryConfigawsYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x54\xdd\x6a\x1b\x4d\x0c\xbd\xdf\xa7\x10\x86\x10\x08\xdf\xda\x0e\x1f\x84\xb2\x90\x8b\x36\x85\x86\x50\x92\xe0\x84\xd2\x52\x4a\xd1\xce\xca\xf6\xc4\xf3\xb3\x8c\x34\x9b\xba\x4f\x5f\x66\xbd\x0e\xfe\x8d\x9d\x42\xc7\x57\x3b\x92\xce\x39\xd2\x1c\x39\xcf\xf3\x0c\x6b\xfd\x85\x02\x6b\xef\x0a\x68\xce\xb3\x99\x76\x55\x01\xb7\x68\x89\x6b\x54\x94\x59\x12\xac\x50\xb0\xc8\x00\x1c\x5a\x2a\x40\x66\x93\x9c\xe7\x2c\x64\x73\x21\x43\x96\x24\xcc\xb3\x6c\x2f\xd4\x03\x85\x46\x2b\x7a\xaf\x94\x8f\x4e\xf6\xe0\xbd\x00\xe5\x8c\x5d\xa0\xa5\x7f\x8d\x6d\x01\x7f\x65\x22\x0b\x85\x91\x37\xb4\xc6\x1f\x4a\x54\x7d\x8c\x32\xf5\x41\xff\x46\xd1\xde\xf5\x67\xef\xb8\xaf\xfd\xa0\x39\x3f\x28\x42\x2d\x40\xf3\x90\x50\x43\x34\xc4\x29\x33\x07\xac\xf5\xa7\xe0\x63\xcd\x05\x7c\xef\xf5\x7e\x64\x00\x00\x81\xd8\xc7\xa0\xa8\xbd\x63\x52\x81\x84\x7b\xff\x41\xef\xa5\x87\xf6\x4b\x79\x37\xd6\x13\x8b\x35\x77\x65\x0d\x85\xb2\x2d\x99\x90\xa4\x04\xa3\x59\xda\xd0\x06\x4b\x27\xa5\xff\x2b\x5f\xc8\xdf\xc5\xda\xe5\xb4\x44\x16\xd5\x54\x3b\xaa\xa8\x36\x7e\x6e\xc9\xc9\x5b\x09\xb5\x1b\x07\x64\x09\x51\x49\x0c\xd4\x3f\x86\x1f\x9f\x79\x55\x02\x3e\x73\xa7\x42\xc8\xd6\x06\x85\xde\xdc\xb4\x77\x12\xbc\xa9\x0d\xba\xe3\x14\xcc\x62\x49\x58\xd9\xd5\xba\x03\x9c\xbb\x2d\xf4\x41\xbb\x4a\xbb\xc9\xbf\x71\x52\x5e\x76\xe8\x1c\xcb\x27\x52\xd2\x99\x6a\xe7\x9a\x24\xe5\x7b\xd7\xe3\xf0\x82\x24\xba\x11\x8d\x13\xc1\x72\xb0\xaf\x34\x92\x01\x6c\x2f\xd3\x11\x9b\xb1\xb5\xf4\x25\x8a\x9a\x0e\x5e\x56\xff\x2a\x78\x77\xe3\xcb\x43\x43\x3a\x62\xe1\x01\x0c\x96\x64\xda\x89\x01\x6c\x3a\x62\xb0\x54\xb5\x00\x3f\x3d\xe1\xd3\x8c\x6b\x52\x29\x9b\xd5\x94\xaa\x68\xa8\x80\xde\x10\xce\x06\x17\x70\x96\x7e\xbd\x0c\x40\x79\xa7\x62\x08\xe4\xd4\xfc\xde\x1b\xad\xe6\x05\x8c\xa8\x36\xe9\x4f\x0f\x60\x8c\xda\x50\x75\xe3\x4b\xbe\xd6\x2c\x3e\xcc\x3f\x6b\xab\xa5\x80\xf3\x61\x06\xf0\xe4\xcb\xc7\xce\xd9\x0b\x41\x4b\xb2\x74\x64\x2d\xb2\x19\x6d\xbf\xd7\x9e\x7a\xcf\x0b\x2f\x0e\x2a\xd1\x0d\x7d\x24\xac\x8c\x76\xf4\x40\xca\xbb\x8a\x0b\xf8\xff\x62\x38\x5c\xc9\x4a\xb6\x47\xed\x28\xf0\x2a\x4d\xf2\xd6\xee\x69\xaf\x1e\x6d\x71\xb2\x9c\xd9\x7a\x44\x79\x6b\xd1\x55\xc5\xc6\x75\x82\x1d\xbc\x86\x98\x12\xf2\x9c\xbd\xf3\x65\xf4\xf3\x64\xfa\xbc\x46\x99\x5e\x0e\x96\x57\x3b\xf3\x93\x3f\xd2\xa0\x29\x8f\xc1\x5c\x9e\xf0\xce\xa4\x85\x09\xb6\xa3\xe4\x9a\x4d\x99\xcb\xde\xaf\x1f\x1f\xef\x7f\xde\x8f\xee\xbe\x7e\xdb\x42\x6c\xd0\xc4\xdd\xad\xaf\x16\x3f\xfc\x75\xf5\xed\xdd\x9b\x4a\x03\xb1\x60\x90\xa5\x17\x6f\xa9\xa1\x90\xfd\x09\x00\x00\xff\xff\x98\x14\x16\x27\xa2\x07\x00\x00")

func bindataTkgManifestTelemetryConfigawsYamlBytes() ([]byte, error) {
	return bindataRead(
		_bindataTkgManifestTelemetryConfigawsYaml,
		"tkg/manifest/telemetry/config-aws.yaml",
	)
}

func bindataTkgManifestTelemetryConfigawsYaml() (*asset, error) {
	bytes, err := bindataTkgManifestTelemetryConfigawsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name:        "tkg/manifest/telemetry/config-aws.yaml",
		size:        1954,
		md5checksum: "",
		mode:        os.FileMode(420),
		modTime:     time.Unix(1, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTkgManifestTelemetryConfigazureYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x54\x5d\x6b\x1b\x3b\x10\x7d\xdf\x5f\x31\x18\x42\x20\x44\xb6\xc3\x85\x70\x59\xc8\xc3\xbd\x29\x34\x84\x92\x04\x27\x94\x96\x52\xca\xac\x76\x6c\x2b\xd6\xc7\x22\x8d\x96\x3a\xbf\xbe\x68\xbd\x0e\x6b\x7b\x1d\x3b\x85\x68\x9f\x56\x33\x73\xce\xd1\xe8\x8c\x84\x10\x19\x56\xea\x2b\xf9\xa0\x9c\xcd\xa1\xbe\xc8\x16\xca\x96\x39\xdc\xa1\xa1\x50\xa1\xa4\xcc\x10\x63\x89\x8c\x79\x06\x60\xd1\x50\x0e\xbc\x98\x89\xb0\x0c\x4c\x46\x30\x69\x32\xc4\x7e\x99\x65\x7b\xa1\x1e\xc9\xd7\x4a\xd2\x7f\x52\xba\x68\x79\x0f\xde\x2b\x90\x08\xd8\x06\x1a\xfa\xb7\xd8\x56\xf0\xd7\x3a\x06\x26\x3f\x71\x9a\x36\xf8\x7d\x81\x72\x88\x91\xe7\xce\xab\x17\x64\xe5\xec\x70\xf1\x6f\x18\x2a\x37\xaa\x2f\x0e\x8a\x90\x2b\x50\xe1\x13\xaa\x8f\x9a\x42\xca\x14\x80\x95\xfa\xec\x5d\xac\x42\x0e\x3f\x06\x83\x9f\x19\x00\x80\xa7\xe0\xa2\x97\xd4\xec\x05\x92\x9e\x38\x0c\xce\x61\xf0\x7a\x86\xe6\x4f\x3a\x3b\x55\x33\x83\x55\x68\xcb\x6a\xf2\x45\x53\x32\x23\x4e\x09\x5a\x05\x6e\x42\x5b\x2c\xad\x94\xe1\x6f\xb1\x92\xdf\xc7\xda\xe6\x7c\x28\xf4\x39\x0c\x0c\xca\xb9\xb2\x54\x52\xa5\xdd\xd2\x90\xe5\xf7\x12\x2a\x3b\xf5\x18\xd8\x47\xc9\xd1\xd3\xf0\x18\x7e\x7c\x89\x9e\x5a\xe2\x46\x44\xb3\xd1\x55\xd5\xcd\x60\x32\x95\x46\xa6\x77\x77\xc2\x59\xf6\x4e\x57\x1a\xed\x71\xb2\x16\xb1\x20\x2c\x4d\xb7\xee\x00\x67\xbf\x65\xff\x57\xb6\x54\x76\xf6\x31\xce\x15\x45\x8b\x1e\x62\xf1\x4c\x92\x5b\x13\xf7\x8e\x65\x52\xbe\x77\x1c\x0f\x0f\x64\xa2\x9b\xd0\x34\x11\xac\x1b\xfb\xc6\x41\x32\x80\xdd\xe1\x3d\x62\x12\x77\x1e\x99\x02\x59\xce\x47\xaf\x4f\xcd\xb5\x77\xf6\xd6\x15\x87\x9a\x74\xc4\x03\x03\xa0\xb1\x20\xdd\x74\x0c\x60\xdb\x11\xa3\xb5\xaa\x15\xf8\xe9\x49\x38\xcd\x42\x45\x32\x65\x07\x39\xa7\x32\x6a\xca\x61\x30\x86\xb3\xd1\x25\x9c\xa5\x6f\x90\x01\x48\x67\x65\xf4\x9e\xac\x5c\x3e\x38\xad\xe4\x32\x87\x09\x55\x3a\x3d\xb2\x00\x53\x54\x9a\xca\x5b\x57\x84\x1b\x15\xd8\xf9\xe5\x17\x65\x14\xe7\x70\x31\xce\x00\x9e\x5d\xf1\xd4\x3a\x7b\x25\x68\x4d\x96\x16\x6f\x44\xb6\xa3\xcd\xff\xc6\x55\xef\xb9\xe1\xd5\x42\xc9\xaa\xa6\x4f\x84\xa5\x56\x96\x1e\x49\x3a\x5b\x86\x1c\xfe\xb9\x1c\x8f\x3b\x59\xc9\xf6\xa8\x2c\xf9\xd0\xa5\x49\xde\xea\xef\x76\x77\x29\x83\xb3\x75\xcf\x36\x23\xd2\x19\x83\xb6\xcc\xb7\xb6\x13\xec\xe8\x2d\xc4\x94\x20\x44\x70\xd6\x15\xd1\x2d\x93\xe9\x45\x85\x3c\xbf\x1a\xad\xb7\x7a\xf3\x93\x3f\x52\xa3\x49\x44\xaf\xaf\x4e\x42\x6f\xd2\xca\x04\xbb\x51\xb2\xf5\xb6\xcc\xf5\xd9\x6f\x9e\x9e\x1e\x7e\x3d\x4c\xee\xbf\x7d\xdf\x41\xac\x51\xc7\xfe\xa3\x77\x8b\x1f\xff\xba\xfa\xee\xfe\x5d\xa5\x9e\x02\xa3\xe7\xb5\x17\xef\xa8\x26\x9f\xfd\x09\x00\x00\xff\xff\x6c\x1a\xb7\x24\x12\x08\x00\x00")

func bindataTkgManifestTelemetryConfigazureYamlBytes() ([]byte, error) {
	return bindataRead(
		_bindataTkgManifestTelemetryConfigazureYaml,
		"tkg/manifest/telemetry/config-azure.yaml",
	)
}

func bindataTkgManifestTelemetryConfigazureYaml() (*asset, error) {
	bytes, err := bindataTkgManifestTelemetryConfigazureYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name:        "tkg/manifest/telemetry/config-azure.yaml",
		size:        2066,
		md5checksum: "",
		mode:        os.FileMode(420),
		modTime:     time.Unix(1, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTkgManifestTelemetryConfigdockerYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x54\xdd\x6a\x1b\x3d\x10\xbd\xdf\xa7\x18\x0c\x21\x10\x22\xdb\xe1\x83\xf0\xb1\x90\x8b\x36\x85\x86\x50\x92\xe0\x84\xd2\x52\x4a\x99\xd5\x8e\x6d\xc5\xfa\x59\xa4\xd1\x52\xf7\xe9\x8b\x76\xd7\xc1\x3f\xeb\xd8\x29\x44\x7b\xb5\x9a\x99\x73\x8e\x46\x67\x24\x84\xc8\xb0\x52\x5f\xc9\x07\xe5\x6c\x0e\xf5\x45\xb6\x50\xb6\xcc\xe1\x0e\x0d\x85\x0a\x25\x65\x86\x18\x4b\x64\xcc\x33\x00\x8b\x86\x72\xe0\xc5\x4c\x84\x65\x60\x32\x82\x49\x93\x21\xf6\xcb\x2c\xdb\x0b\xf5\x48\xbe\x56\x92\x3e\x48\xe9\xa2\xe5\x3d\x78\x2f\x40\x22\x60\x17\x68\xe8\x5f\x63\x6b\xe1\xaf\x75\x0c\x4c\x7e\xe2\x34\x6d\xf0\xfb\x02\xe5\x10\x23\xcf\x9d\x57\x7f\x90\x95\xb3\xc3\xc5\xff\x61\xa8\xdc\xa8\xbe\x38\x28\x42\xb6\xa0\xc2\x27\x54\x1f\x35\x85\x94\x29\x00\x2b\xf5\xd9\xbb\x58\x85\x1c\x7e\x0c\x06\x3f\x33\x00\x00\x4f\xc1\x45\x2f\xa9\xd9\x0b\x24\x3d\x71\x18\x9c\xc3\xe0\xe5\x0c\xcd\x9f\x74\x76\xaa\x66\x06\xab\xd0\x95\xd5\xe4\x8b\xa6\x64\x46\x9c\x12\xb4\x0a\xdc\x84\xb6\x58\x3a\x29\xc3\xdf\xa2\x95\xdf\xc7\xda\xe5\xbc\x2b\xf4\x39\x0c\x0c\xca\xb9\xb2\x54\x52\xa5\xdd\xd2\x90\xe5\xb7\x12\x2a\x3b\xf5\x18\xd8\x47\xc9\xd1\xd3\xf0\x18\xfe\xd2\xc9\x05\xf9\x8e\xb9\x51\xd1\xee\xac\xeb\xda\xc8\x61\x32\x95\x46\xa6\x37\x37\xc3\x59\xf6\x4e\x57\x1a\xed\x71\xca\x16\xb1\x20\x2c\xcd\x7a\xdd\x01\xce\x7e\xd7\x7e\x54\xb6\x54\x76\xf6\x3e\xe6\x15\x45\x87\x1e\x62\xf1\x4c\x92\x3b\x1f\xf7\x4e\x66\x52\xbe\x77\x22\x0f\xcf\x64\xa2\x9b\xd0\x34\x11\xac\x1a\xfb\xca\x41\x32\x80\xdd\xf9\x3d\x62\x18\x77\xde\x99\x02\x59\xce\x47\x2f\xaf\xcd\xb5\x77\xf6\xd6\x15\x87\x9a\x74\xc4\x1b\x03\xa0\xb1\x20\xdd\x74\x0c\x60\xdb\x11\xa3\x95\xaa\x16\xfc\xf4\x24\x9c\x66\xa1\x22\x99\xb2\x83\x9c\x53\x19\x35\xe5\x30\x18\xc3\xd9\xe8\x12\xce\xd2\x37\xc8\x00\xa4\xb3\x32\x7a\x4f\x56\x2e\x1f\x9c\x56\x72\x99\xc3\x84\x2a\x9d\xde\x59\x80\x29\x2a\x4d\xe5\xad\x2b\xc2\x8d\x0a\xec\xfc\xf2\x8b\x32\x8a\x73\xb8\x18\x67\x00\xcf\xae\x78\xea\x9c\xdd\x0a\x5a\x91\xa5\xc5\x1b\x91\xed\x68\xf3\xbf\x71\xd5\x7b\x6e\xb8\x5d\x28\x59\xd5\xf4\x89\xb0\xd4\xca\xd2\x23\x49\x67\xcb\x90\xc3\x7f\x97\xe3\xf1\x5a\x56\xb2\x3d\x2a\x4b\x3e\xac\xd3\x24\x6f\xf5\x77\x7b\x7d\x29\x83\xb3\x55\xcf\x36\x23\xd2\x19\x83\xb6\xcc\xb7\xb6\x13\xec\xe8\x35\xc4\x94\x20\x44\x70\xd6\x15\xd1\x2d\x93\xe9\x45\x85\x3c\xbf\x1a\xad\xb6\x7a\xf3\x93\x3f\x52\xa3\x49\x44\xaf\xaf\x4e\x42\x6f\x52\x6b\x82\xdd\x28\xd9\x7a\x5b\xe6\xea\xec\x37\x4f\x4f\x0f\xbf\x1e\x26\xf7\xdf\xbe\xef\x20\xd6\xa8\x63\xff\xd1\xd7\x8b\x1f\xff\xb9\xfa\xee\xfe\x4d\xa5\x9e\x02\xa3\xe7\x95\x17\xef\xa8\x26\x9f\xfd\x0d\x00\x00\xff\xff\x34\xb1\xf5\xd5\x15\x08\x00\x00")

func bindataTkgManifestTelemetryConfigdockerYamlBytes() ([]byte, error) {
	return bindataRead(
		_bindataTkgManifestTelemetryConfigdockerYaml,
		"tkg/manifest/telemetry/config-docker.yaml",
	)
}

func bindataTkgManifestTelemetryConfigdockerYaml() (*asset, error) {
	bytes, err := bindataTkgManifestTelemetryConfigdockerYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name:        "tkg/manifest/telemetry/config-docker.yaml",
		size:        2069,
		md5checksum: "",
		mode:        os.FileMode(420),
		modTime:     time.Unix(1, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTkgManifestTelemetryConfigvsphereYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x54\xdd\x6a\x1b\x3d\x10\xbd\xdf\xa7\x18\x0c\x21\x10\x22\xdb\xe1\x83\xf0\xb1\x90\x8b\x36\x85\x86\x50\x92\xe0\x84\xd2\x52\x4a\x99\xd5\x8e\x6d\xc5\xfa\x59\xa4\xd1\x52\xf7\xe9\x8b\xd6\xbb\xc1\x3f\xeb\xd8\x29\x44\x7b\xb5\x9a\x99\x73\x8e\x46\x67\x24\x84\xc8\xb0\x52\x5f\xc9\x07\xe5\x6c\x0e\xf5\x45\xb6\x50\xb6\xcc\xe1\x0e\x0d\x85\x0a\x25\x65\x86\x18\x4b\x64\xcc\x33\x00\x8b\x86\x72\xe0\xc5\x4c\x84\x65\x60\x32\x82\x49\x93\x21\xf6\xcb\x2c\xdb\x0b\xf5\x48\xbe\x56\x92\x3e\x48\xe9\xa2\xe5\x3d\x78\x2f\x40\x22\x60\x1b\x68\xe8\x5f\x63\x5b\xc1\x5f\xeb\x18\x98\xfc\xc4\x69\xda\xe0\xf7\x05\xca\x21\x46\x9e\x3b\xaf\xfe\x20\x2b\x67\x87\x8b\xff\xc3\x50\xb9\x51\x7d\x71\x50\x84\x5c\x81\x0a\x9f\x50\x7d\xd4\x14\x52\xa6\x00\xac\xd4\x67\xef\x62\x15\x72\xf8\x31\x18\xfc\xcc\x00\x00\x3c\x05\x17\xbd\xa4\x66\x2f\x90\xf4\xc4\x61\x70\x0e\x83\x97\x33\x34\x7f\xd2\xd9\xa9\x9a\x19\xac\x42\x5b\x56\x93\x2f\x9a\x92\x19\x71\x4a\xd0\x2a\x70\x13\xda\x62\x69\xa5\x0c\x7f\x8b\x95\xfc\x3e\xd6\x36\xe7\x5d\xa1\xcf\x61\x60\x50\xce\x95\xa5\x92\x2a\xed\x96\x86\x2c\xbf\x95\x50\xd9\xa9\xc7\xc0\x3e\x4a\x8e\x9e\x86\xc7\xf0\xd7\xa1\x9a\x93\xa7\x96\xba\x91\xd1\x6e\xad\x2b\xdb\xcc\x62\x32\x95\x46\xa6\x37\xf7\xc3\x59\xf6\x4e\x57\x1a\xed\x71\xe2\x16\xb1\x20\x2c\xcd\x7a\xdd\x01\xce\x7e\xe3\x7e\x54\xb6\x54\x76\xf6\x3e\xfe\x15\x45\x8b\x1e\x62\xf1\x4c\x92\x5b\x2b\xf7\x0e\x67\x52\xbe\x77\x28\x0f\x8f\x65\xa2\x9b\xd0\x34\x11\x74\x8d\x7d\xe5\x20\x19\xc0\xee\x08\x1f\x31\x8f\x3b\x4f\x4d\x81\x2c\xe7\xa3\x97\x07\xe7\xda\x3b\x7b\xeb\x8a\x43\x4d\x3a\xe2\x99\x01\xd0\x58\x90\x6e\x3a\x06\xb0\xed\x88\x51\xa7\x6a\x05\x7e\x7a\x12\x4e\xb3\x50\x91\x4c\xd9\x41\xce\xa9\x8c\x9a\x72\x18\x8c\xe1\x6c\x74\x09\x67\xe9\x1b\x64\x00\xd2\x59\x19\xbd\x27\x2b\x97\x0f\x4e\x2b\xb9\xcc\x61\x42\x95\x4e\x4f\x2d\xc0\x14\x95\xa6\xf2\xd6\x15\xe1\x46\x05\x76\x7e\xf9\x45\x19\xc5\x39\x5c\x8c\x33\x80\x67\x57\x3c\xb5\xce\x5e\x09\xea\xc8\xd2\xe2\x8d\xc8\x76\xb4\xf9\xdf\xb8\xea\x3d\x37\xbc\x5a\x28\x59\xd5\xf4\x89\xb0\xd4\xca\xd2\x23\x49\x67\xcb\x90\xc3\x7f\x97\xe3\xf1\x5a\x56\xb2\x3d\x2a\x4b\x3e\xac\xd3\x24\x6f\xf5\x77\x7b\x7d\x29\x83\xb3\xae\x67\x9b\x11\xe9\x8c\x41\x5b\xe6\x5b\xdb\x09\x76\xf4\x1a\x62\x4a\x10\x22\x38\xeb\x8a\xe8\x96\xc9\xf4\xa2\x42\x9e\x5f\x8d\xba\xad\xde\xfc\xe4\x8f\xd4\x68\x12\xd1\xeb\xab\x93\xd0\x9b\xb4\x32\xc1\x6e\x94\x6c\xbd\x2d\xb3\x3b\xfb\xcd\xd3\xd3\xc3\xaf\x87\xc9\xfd\xb7\xef\x3b\x88\x35\xea\xd8\x7f\xf4\xf5\xe2\xc7\x7f\xae\xbe\xbb\x7f\x53\xa9\xa7\xc0\xe8\xb9\xf3\xe2\x1d\xd5\xe4\xb3\xbf\x01\x00\x00\xff\xff\xeb\x0f\x9c\x9a\x18\x08\x00\x00")

func bindataTkgManifestTelemetryConfigvsphereYamlBytes() ([]byte, error) {
	return bindataRead(
		_bindataTkgManifestTelemetryConfigvsphereYaml,
		"tkg/manifest/telemetry/config-vsphere.yaml",
	)
}

func bindataTkgManifestTelemetryConfigvsphereYaml() (*asset, error) {
	bytes, err := bindataTkgManifestTelemetryConfigvsphereYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name:        "tkg/manifest/telemetry/config-vsphere.yaml",
		size:        2072,
		md5checksum: "",
		mode:        os.FileMode(420),
		modTime:     time.Unix(1, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataTkgManifestTelemetryZzgeneratedBindataGo = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x9a\x5b\x6f\x23\x47\x92\x85\x9f\xc9\x5f\x51\x2b\x60\x06\xe4\x42\x23\xd5\xfd\x22\xc0\x2f\xd3\xf6\x02\x7e\xb0\x07\xd8\xed\x7d\x58\x6c\x2e\x8c\xac\xaa\x2c\x99\x68\x49\xd4\x92\x94\x9d\x2d\xc3\xff\x7d\xf1\x45\x44\x49\x6c\xf5\x4d\xea\xe9\x19\x60\xb0\x63\x80\x6e\x91\x55\x95\x19\x19\x71\x22\xe2\x9c\xcc\x3a\x3f\x4f\x5e\x6d\xc7\x90\x5c\x86\x9b\xb0\xf3\x87\x30\x26\xfd\xdb\xe4\x72\xfb\xa7\x7e\x73\x33\xfa\x83\x3f\x4b\xbe\xfd\x4b\xf2\xe3\x5f\x5e\x27\xdf\x7d\xfb\xfd\xeb\xb3\xe5\xf9\x79\xb2\xdf\xde\xed\x86\xb0\xbf\xe0\xef\xc3\x9b\xcb\xf3\x6b\x7f\xb3\x99\xc2\xfe\x70\x7e\x08\x57\xe1\x3a\x1c\x76\x6f\xcf\x87\xed\xcd\xb4\xb9\xfc\x93\xff\x75\x7f\xf6\xd6\x5f\x5f\x3d\xe3\xce\xfb\xbb\x5d\x78\xe6\xbd\xe3\x76\x78\x13\x76\xcf\xbc\xf9\x97\xfd\xed\xcf\xe1\xf3\x43\xdf\xdf\xff\xf4\xb0\xfe\xb3\x79\xe5\x97\xdb\xe5\xf2\xd6\x0f\x6f\xfc\x65\x48\x1e\x6e\x5d\x2e\x97\x9b\xeb\xdb\xed\xee\x90\xac\x96\x8b\x93\xfe\xed\x21\xec\x4f\x96\x8b\x93\x61\x7b\x7d\xbb\x0b\xfb\xfd\xf9\xe5\xfd\xe6\x96\x1f\xa6\xeb\x03\xff\x6c\xb6\xfa\xff\xf3\xcd\xf6\xee\xb0\xb9\xe2\xcb\x56\x1e\xb8\xf5\x87\x9f\xcf\xa7\xcd\x55\xe0\x0f\x7e\xd8\x1f\x76\x9b\x9b\x4b\xb9\x76\xd8\x5c\x87\x93\xe5\x7a\xb9\x9c\xee\x6e\x86\xc4\xcc\xf9\xf7\xe0\xc7\x15\x7f\x24\xff\xfd\x3f\x4c\x7b\x9a\xdc\xf8\xeb\x90\xe8\x63\xeb\x64\x35\xff\x1a\x76\xbb\xed\x6e\x9d\xfc\xb6\x5c\x5c\xde\xcb\xb7\xe4\xe2\x9b\x04\xab\xce\x7e\x0c\xbf\x32\x48\xd8\xad\xc4\x6c\xbe\xff\xf9\x6e\x9a\xc2\x4e\x86\x5d\xaf\x97\x8b\xcd\x24\x0f\xfc\xcb\x37\xc9\xcd\xe6\x8a\x21\x16\xbb\x70\xb8\xdb\xdd\xf0\xf5\x34\x99\xae\x0f\x67\xdf\x31\xfa\xb4\x3a\x61\xa0\xe4\x0f\xff\x7b\x91\xfc\xe1\x97\x13\xb5\x44\xe6\x5a\x2f\x17\xbf\x2f\x97\x8b\x5f\xfc\x2e\xe9\xef\xa6\x44\xe7\xd1\x49\x96\x8b\x9f\xd4\x9c\x6f\x92\xcd\xf6\xec\xd5\xf6\xf6\xed\xea\x8f\xfd\xdd\x74\x9a\x5c\xde\xaf\x97\x8b\xe1\xea\xbb\xd9\xd2\xb3\x57\x57\xdb\x7d\x58\xad\x97\x5f\xcb\x1e\x86\xd1\xf1\x3f\x32\x50\xd8\xed\xd4\x6e\xfb\xb1\xbf\x9b\xce\xfe\x8c\xe9\xab\xf5\x29\x77\x2c\x7f\x5f\x2e\x97\x87\xb7\xb7\x21\xf1\xfb\x7d\x38\xe0\xf3\xbb\xe1\xc0\x30\xb2\x40\x0b\xc8\x72\xb1\xb9\x99\xb6\x49\x42\x50\xbf\xbf\x99\xb6\xdf\x45\x9e\x93\xc7\x1e\x7f\x4a\x36\x37\x87\xb0\x9b\xfc\x10\x78\x7c\xbb\x3f\xfb\x37\xbb\xb4\x5c\xfc\xf0\x6d\xf5\xea\xe7\x30\xbc\xd9\xdf\x5d\xaf\xd6\x16\xd7\x87\x11\x0c\x04\xf3\xdd\x47\x26\x08\x0a\xec\x3f\x7b\x68\xb1\xdf\xdc\x3f\xfc\xb6\xb9\x39\xd4\xe5\x72\x71\x4d\x8e\xdb\x7f\x36\xed\x0f\xdb\x31\xc8\x85\xd7\x1b\x1b\x02\xe0\x9d\xf1\x6d\xb9\xb8\x1e\xab\xc1\xac\x39\xb2\x45\x00\xb9\x9a\x36\x4f\xed\x59\x27\x3f\xfa\xeb\xf0\x60\x36\x76\x99\x2f\xa7\xcd\x19\x16\x2e\x7f\xff\xc4\xb3\xff\xb1\xb9\xe7\x59\xb1\xf4\xdd\x47\x59\xc8\x27\x1f\x65\x0d\xab\xf5\xf1\x8a\xde\x1d\x80\x65\x7f\x6e\x00\x16\xbc\x5a\x3f\x2e\xfe\xbd\x11\xc4\x23\x9f\x1c\xe4\x03\xa1\x7b\x32\xca\xa3\x3b\x3f\x39\xd2\xf7\xfb\x6f\x37\xbb\xd5\x3a\xe9\xb7\xdb\xab\xe3\x11\xfc\xd5\xfe\x33\x3e\x7c\xbb\x57\x17\x2a\xba\x7e\xfb\xfd\xe8\x69\x83\x30\x59\xf9\x93\x3d\xf8\xfa\xcd\xe5\x0f\x56\x06\x5f\xcf\xa5\xed\x95\x94\x4c\xff\xeb\xfe\xbf\xfc\xf5\x55\xf2\x8d\xc1\x7a\x75\xe2\x62\x36\xb9\xd8\xf6\x2e\xa6\xad\x8b\x69\xfa\xe1\xcf\x34\xb9\xd8\x97\x2e\x56\xa5\x8b\xe3\xe8\x62\xed\x5d\xcc\x7a\x17\xcb\xd1\xc5\x74\x70\xb1\x1f\x5d\x1c\x27\x17\x7d\xe3\x62\x96\xba\xd8\xd6\xfa\x2f\x63\xf2\xfb\xe8\x5d\x4c\x83\xcd\x55\xba\xd8\xe7\x2e\x76\xa9\xce\x5b\xd4\x2e\xb6\x95\x3e\x53\xa5\x2e\x76\xb9\x8b\x21\xd5\xfb\xc6\xdc\xc5\x2a\x77\xb1\xf4\x2e\x8e\x99\x8b\x43\x70\x71\xf0\x2e\x4e\xb5\x8b\x43\xe9\xe2\x54\xb8\xd8\x17\x2e\xb6\x83\x8b\x45\xe9\x62\xd7\xbb\xd8\x7b\x17\xcb\xc9\xc5\x6a\x72\xb1\xae\xd5\x36\xe6\x9e\x82\x8b\xed\xe8\x62\x37\xba\x58\xe6\x2e\x0e\x8d\x8b\x55\xe3\x62\xd1\xeb\x9c\x8c\x5d\x74\x3a\x67\x36\xe8\xdf\xc3\xa4\x73\xb0\xc6\xba\x77\x71\x1a\xd5\xd6\x34\xd7\xef\x61\x72\x31\xf5\x2e\xd6\xad\x3e\x8f\x2d\x5d\xe7\x62\xc3\x5a\xb8\x2f\x73\xb1\x6f\xf4\x7a\xdb\xe9\x33\xf8\xb0\xc3\x97\x9d\x8b\x59\xee\xa2\x1f\x74\xdd\x7d\xea\xe2\x60\x31\x60\xfe\xca\xbb\x98\xb3\x96\x54\xd7\xd1\x31\xf6\xe0\x62\x68\x5c\xcc\xb1\xa7\x74\xb1\x29\x5c\xcc\x33\x17\x4b\xae\xd5\x2e\xe6\xa5\x8b\xc3\xa0\x76\xd4\x83\x8b\x39\xbe\x2f\x5d\x4c\x0b\xb5\xbb\xe4\x9e\xde\xc5\xc6\xbb\xe8\x27\xb5\xa3\x9d\x5c\x2c\x83\xfa\x14\xbf\x8b\xbf\x52\x17\x43\xa5\x7e\xad\x46\x17\x7d\xea\xa2\xaf\x5c\x6c\x26\xf5\x61\x3d\xea\xda\xf8\x5e\x57\x2e\xe6\xb9\x8b\x69\xaf\x73\x74\x99\x8b\x45\xa3\x78\x19\x6a\x8d\x39\xf1\x63\xdd\xcd\x68\xb1\xca\x5d\x9c\xb0\x27\x53\x6c\x61\x17\xf1\x1d\x83\xfe\x5e\x37\xea\xdb\xbe\x55\x3b\x27\xb3\x81\x98\x14\x93\x8b\x79\xab\x31\xcc\x47\xf5\x0f\x31\x02\x4f\xf8\x11\x5f\x80\x05\xf0\x91\x73\x3f\xf1\x6b\xd4\xcf\x8c\x0d\x46\x59\x67\x5d\xb8\x38\xb0\x0e\xd6\x35\xe8\x7c\x5c\x6f\x82\xfa\x76\xc6\x7e\x8b\x5d\xad\xe2\x45\xfc\xd0\xaa\x8f\x78\x1e\x6c\x72\x1d\xac\x36\xbd\xad\x25\xd3\xb1\xf0\x57\xdb\xa8\x5f\x4b\x7c\xde\xa9\x5f\x46\x7c\x52\x18\xf6\x99\xaf\x51\x1f\xa6\x86\xad\xde\xd6\x05\x8e\x58\x93\xc7\x96\xd2\x45\x5f\x28\x66\xc8\xa5\x11\xdb\x6c\xdc\xbc\xd1\xb9\x98\x9b\xd8\x82\xfb\xc1\xf2\x6e\xa8\xf4\x7e\x62\x49\x3c\xca\xd2\xc5\x0c\x5f\x57\x1a\x0f\xb0\xef\xbd\xae\x89\x3c\x94\xb5\x93\x63\x9d\x7e\x2a\xea\x42\xe7\x62\x5f\x69\xbe\xe3\x47\xf1\x0d\xd7\xc0\x5d\xa7\xb9\x01\xc6\x88\x4b\x6b\xf1\xee\x26\x5d\x6f\x65\xb9\x52\x04\xc5\x2a\xbe\x97\xdc\x6b\xd5\x0f\xa9\xe5\x3e\x71\x1f\x07\xb5\x11\x7f\x90\x17\xfd\xe0\xa2\xef\xd4\x2f\xe4\x74\x28\x5c\xcc\x0c\xdf\x75\xae\x73\x57\x8c\xd3\xe9\x7a\xb8\x27\xb5\x1c\xe9\x7b\xf5\xe1\x54\x6a\x3c\xf8\x0e\x06\xf9\x97\x75\xf5\xd8\x97\x69\xfc\xaa\xa0\xf9\x1a\x5a\xcd\xbb\xc1\x7c\xca\x35\x6a\x01\x71\xa0\xde\x91\x77\x9d\xd7\x1c\xc3\x9f\xc4\x7b\x6c\xd4\xae\x09\x9c\xe4\x7a\x0d\x3b\xb2\x4c\xf3\x84\x38\x0f\xe4\x49\xae\xf9\x0d\x96\xa5\x4e\xe5\x9a\x3f\x0f\x39\x5a\xe8\x33\x52\xbf\x32\xf5\x37\xb6\xe7\x60\xc8\xeb\xbc\xd4\x30\x6c\xad\x6a\x8b\xb5\x77\xb1\x69\xd5\x5f\xd8\x80\xdd\x60\x5f\xf0\xef\xd5\x47\x21\xd3\x79\x88\x11\xf5\x81\xd8\x81\x07\x6a\xd1\x3c\xff\x7c\x2f\xf6\x81\x09\xe6\x06\xf7\xc4\xb3\x18\x75\x8c\xd1\x6a\xac\xd4\xaf\x5c\xf3\x0a\x5b\xf1\x3b\x35\xc4\x5b\x0d\x04\x47\x6d\xaa\xf1\xa4\xfe\x53\x13\x89\x71\xd6\xb8\xd8\xa4\x6a\x03\x18\xeb\xad\x6f\x90\xbb\xe0\x04\x5c\x60\x03\xfd\x22\x94\x8a\xa7\x69\xb0\x71\xc0\x28\x76\xd4\x9a\xab\xd8\x31\xe3\x35\x74\x5a\xeb\x53\xea\x50\xaa\xd7\x58\x1f\x75\xa0\xa9\xd4\x2f\xe4\x48\x1f\x14\x6f\xd4\x4d\xb0\x08\x86\x58\x1f\xeb\xf1\xbd\xe5\xaa\xd5\x91\x3a\x53\x9b\x89\x29\xb6\xe0\x57\x72\x9f\xb5\x80\x35\xea\x1c\x98\x92\x71\x73\x17\x9b\x52\xfd\x0a\xde\xf8\x9e\x75\x5a\x1f\xc1\x18\xf8\xe8\x82\xd6\x5b\x7a\x09\xcf\x12\x6f\x62\x99\x5b\x4f\x03\xdf\xd4\x45\xb0\x43\x8d\x62\x0d\xe0\x51\xb0\x9b\x6a\xbf\x03\xb7\xe0\x80\xe7\xa7\x56\xe3\x8f\xbf\x8a\xd6\xc5\x6a\x50\x4c\x0b\xbe\xa9\x09\xe0\x61\xd4\x1a\xc9\x1a\xa4\x87\x90\x0b\xb9\xfa\x20\x90\xe3\x9d\x62\x31\x0b\x6a\x2b\xb9\x20\x3d\x79\xd0\x35\x62\x2f\xb5\xa2\xb1\x9e\x85\x9d\xe4\x32\xf5\x84\x9a\x8e\x6f\xc1\x7b\x36\x6a\x9e\x92\xf7\x5d\xab\x79\x3b\xd9\x38\xc4\x58\xfc\x69\x58\x96\xbf\xf9\xd7\xeb\x33\xd4\x7c\x72\x0b\x4c\xe3\x0f\xb0\x31\xf7\x63\xa9\xe7\x85\xc6\x88\x1c\xa4\xee\x82\x39\x7a\x15\xf9\xc4\x9a\xf1\xeb\x64\x18\xa3\x46\x4a\x5f\xee\xd4\x07\xd4\x4a\x62\xd7\xd9\xda\x99\x87\x38\x82\x27\x59\xb7\x7d\xa8\xd3\xf4\x30\x7e\x67\x6c\x7c\x03\x56\xc0\xe8\xd8\x6b\xad\x22\x47\xa8\xb7\x83\x71\x1f\xb0\x28\x63\xd4\x6a\x1f\x38\xc5\x1f\xf4\x0e\xfc\xce\x3d\x9d\xd5\x1b\xea\x11\xf9\x0c\x36\x3b\xeb\xa1\xe4\x42\x3d\x69\x6d\xf3\x99\xfe\x4e\x7f\xc3\xf6\x63\xbe\xc5\x47\x7c\x6a\x35\x5b\xea\x7c\xae\x35\x58\xef\x3b\x79\xa2\x22\x3f\xcb\xfb\x4c\xf7\x7c\x48\x4f\xce\xea\xe8\x48\x8f\x2e\x17\x8b\x67\x33\xca\xd3\xe5\x62\x71\xf2\xbc\x0d\x83\x93\xd3\xe5\x62\x2d\xaa\xeb\x65\xc6\x63\xf7\xbf\x8a\x46\x3b\xb6\x5b\x44\xda\x83\x14\x7e\xa1\x23\x3e\x27\x42\x1f\xb4\xa3\x88\xbf\xc7\xf1\x67\x66\xce\xfd\x08\xa0\x8b\xe4\x05\x6b\x17\x01\x77\x91\x64\x5d\x55\xf2\xed\x48\x3c\x5c\x24\x27\x72\x03\xba\xe6\xe2\x58\xf6\xac\xca\x3c\x5d\xdb\x15\xf4\xca\x85\xea\x99\xff\xbc\xd9\xc4\x55\x76\x9a\xc8\x35\xcc\xf4\xd8\xf8\x47\xf1\xd1\x6f\xe2\x98\x8b\xc4\xfc\xc3\x02\x2e\xe4\xff\x47\x52\xd8\x9f\xbe\x50\x3d\xdc\xdf\xed\xc2\x5f\xad\x1f\xe0\x93\xd4\x14\xea\x3a\xb9\x4e\x9f\x68\x4c\x3f\xc0\x5d\x8a\xcc\xc5\x6c\xe6\x77\xa9\xd6\x21\x72\x7a\x30\x0e\x4b\x1f\x81\x07\xc1\x21\xf8\x8d\x3a\x23\x3c\xaf\x54\x1e\x26\xbd\xb5\xd4\x7c\xa1\x37\x49\x5d\xc9\xb5\x8e\x7a\xab\xad\xd2\x77\x7b\xad\x87\xd4\x76\xf8\xab\x68\x83\x5a\xfb\x21\xb5\x9b\x3a\x50\xb7\x8f\xfa\x01\x9b\xe9\xf9\xd4\xbc\xc2\xb8\x2e\xd7\xe1\x3a\xf4\x64\x78\x26\x1c\x47\xfa\x44\xa6\x7c\x82\x5a\x8e\x4d\xac\x91\xbe\xc0\x7d\xc1\xeb\xdc\x53\xa7\x75\x8d\x5a\x09\x0f\xa5\x16\x30\x27\xeb\x22\xdf\xb1\x57\xec\xe9\x94\x1b\x71\x9d\x0f\xbd\x4d\xfe\x36\x2e\x94\x19\x97\xa5\xa7\x4b\xef\xe8\xb4\x7f\x51\x73\x46\xbb\x1f\xfb\xa9\xb9\x9d\xe9\x10\x7a\xbd\xf4\x5f\xab\xc3\x45\xaa\x3d\x01\x5e\x4e\x9d\xa3\xfe\x09\xef\x31\x9e\xc3\x7c\x99\xf1\x42\x7a\x9a\xe8\xb1\x5c\x6b\x27\xbe\xed\xad\xff\xcb\xdc\x56\x4f\x59\x63\x37\x6b\xc0\xd6\x78\x80\xb7\xfa\xdf\xe8\x6f\x95\xf5\x59\xc6\xc4\xef\xdc\x93\x5b\x5d\xa5\x27\x51\xbf\xe9\xf3\xf4\x3c\xd1\x0e\xc6\xa5\xda\x56\x75\x46\xb0\xbe\x4c\x0f\xcf\x8c\x9b\xc2\xcd\xc2\xa0\x38\x99\x32\xad\xb5\xe0\x89\x3e\x08\xff\x86\xab\xcc\x3d\x58\xb8\x55\x6a\xbc\x39\x55\xae\x03\x67\xe3\x1e\xfa\x81\xf4\x0f\xc3\x0e\x75\x18\x8e\xd0\x55\xba\x86\x30\x6b\xa3\x41\xef\x61\x3c\xec\x16\x4e\x5b\x28\x2e\x32\xab\xe9\x3c\x37\xeb\x0e\x6a\x78\xd7\x28\xc7\x61\x0c\x38\x0d\x98\xa5\x9f\xd1\xcf\x89\x0b\x6b\x82\x87\xe0\x1b\x7a\x31\xfd\x15\xfe\x03\xd7\xc1\x4f\xe0\x8f\x7e\x0c\xe6\x98\xab\x32\x2c\xd2\xe3\xe0\xaa\xf4\xe0\x12\x8e\x5f\x68\x4c\xf3\x42\x35\x9e\xf0\x76\xaf\x73\xc2\xf3\xb1\x1b\x8e\x0f\x2f\xa2\xaf\x91\x2f\xc1\x7c\x46\xdf\xc2\x2e\xd1\x11\x8d\xe9\x88\x5a\xb5\x92\x70\x8b\x52\xb1\x89\xbd\x62\x33\x71\xed\x94\x47\x48\x9e\xe7\xda\xef\xe8\x87\xa3\xe9\x7d\x62\x33\x19\xbf\xcf\x6d\x1d\xf4\xcd\xcc\x30\x02\x77\x11\x8c\x19\xe7\x64\xfd\x5c\x03\x8f\xcc\x0b\x76\xf1\x13\x6b\x08\xb9\x62\x57\xea\x43\xf1\xc8\x51\xe0\xa4\xe4\x13\xf8\x67\xed\x9d\xe9\x5e\x5f\x2b\x57\xe6\x33\xe4\x5a\x43\x44\x9b\x05\xe5\x1d\x60\x33\x1c\x71\x21\x72\x10\x1f\x50\x83\xf0\x0f\x3c\x02\xdc\xc2\x11\x88\x29\x1a\xb5\x37\xfd\x4b\x7d\x63\x0c\xea\x94\xe8\xdc\x5a\xc7\x2e\x6c\xcf\x22\xab\x34\x1e\xd4\x49\xe1\x5c\xb9\x62\x93\xfc\xeb\x2c\x06\xe0\x0e\xdf\xb4\x41\xc7\x94\xbc\x0a\xa6\x87\x06\xcd\x2b\xe1\xb9\xa5\x61\xa8\x57\x5e\x42\xdd\x43\x3b\x80\x43\x6a\xea\xd4\x2b\x47\x83\xa3\xb2\x7e\xe1\xe3\xc6\xe3\xe1\xf3\xcc\x8d\xff\x65\x5c\xcb\xf3\xd4\xfc\x81\xad\xc4\x00\x7e\x03\xd6\xf0\x63\x6f\x1a\x95\xf5\x88\x76\x35\x5d\x4b\x6e\x34\xa6\xb3\xe1\x39\x7c\xa8\x67\xb3\xcf\x98\x93\xeb\xb2\x77\x92\x2a\x9f\x12\x3b\x0b\x1d\x93\x1c\x63\x2c\x72\x5b\x78\x7c\xa7\xf8\x1f\x6c\x7f\x06\x7f\x61\xbf\x68\xe3\x46\xb1\x46\x0e\xc9\x6f\x5e\xf1\x46\x8e\x14\xb6\xdf\x34\x63\x00\xcd\x44\x4d\x05\x7b\xe0\x58\x6a\x10\xbe\x6e\x35\x66\x33\x1f\x05\x47\xb2\xf7\x62\x9a\x80\xb9\x45\x67\x94\xe6\xb7\xc6\xf2\xa5\xb2\x9c\x33\x7e\x4f\xec\xc9\xab\xca\xf8\x25\x76\x10\x3f\xe2\x2f\xb5\xbe\xd7\x75\x81\x67\x6a\x7c\x98\x39\x5f\xa5\x35\x24\x98\x16\x9a\xb1\xc1\x78\x82\xe9\x51\x31\x21\x5a\xda\x6b\xdd\xcd\x2d\xaf\x4b\xeb\x4d\x60\x92\x35\xcd\x3d\xd1\xdb\x1e\x06\xf8\x1b\x6c\xcf\x49\xf6\x4d\x3a\xd3\x6d\x95\xd5\xaf\xce\x74\xbe\xd5\x10\xec\xc4\x3f\xd4\x5f\x19\x37\x98\xff\x6d\x2d\xa2\xe3\x07\xdd\x33\x63\x2c\xfe\xa6\x1e\x4a\xdd\xb1\xbd\x01\xf0\x80\x0e\x23\x37\x58\xe7\x68\xfc\x96\x71\xa8\xff\xf4\x39\xe1\xd7\x96\x77\xc2\xdf\x6d\xbf\x0e\xbe\x8c\x1e\xa2\xb6\x53\x9b\x46\xd3\x4d\x72\xbd\xb6\x3d\x2c\xcb\x27\xf2\x82\x3a\x49\x3f\x45\x63\x0d\xb6\x7f\x36\x73\x03\x7c\x39\xe3\x52\xe2\x3f\xaa\x76\x10\x4c\x55\xda\xa7\x8b\x41\x73\x95\x1c\xf7\x96\xc7\xd4\xd5\x63\x9d\x21\xf5\xaa\x50\xfb\x19\x97\x1c\xf0\xb6\x37\x26\x3a\xc0\xb4\x08\xf1\x86\x5f\xa0\xc9\x2a\xab\x15\xd4\xb0\xd9\x6e\xfe\x46\x43\x09\xf6\x82\x7e\x32\xd3\x14\x60\x88\x9c\x17\x9d\x31\x3d\xf2\x9e\xd2\xf2\x97\xf1\xa8\xb5\xf8\x1d\x5b\xf0\x9b\xd4\x0c\xd3\x03\xbd\xf5\xcf\x60\xfb\x99\xd2\x7f\x2a\x9d\x23\x35\xfb\xc1\x57\x6f\x98\x93\x7d\xa8\x56\x7b\x6a\x37\x7d\x5c\x67\xc0\x87\x32\xeb\xcb\xe8\xd0\x2c\x3f\xe6\x73\x2f\xd0\x19\x33\x43\xfc\x5b\x28\x8d\x79\xec\xe7\x68\x8d\x87\x23\xc7\x97\xab\x8d\x79\x9a\xaf\xa3\x37\x9e\x38\xe4\xef\xa7\x38\xde\xf1\x80\x69\x8e\x3c\xad\xeb\x7f\x30\xcd\xa1\x07\xc2\x5f\xf5\xd0\x82\x24\xcc\xd2\x27\x87\x16\x46\xbc\xf2\x4c\xaf\x21\x08\x68\x58\x14\xd1\xd6\x88\x48\x9f\xbd\xfc\xd0\xa2\xb3\xcd\xc9\xd6\x36\x5a\x06\x23\x8f\x14\x51\x08\xfc\x68\x9b\x90\x34\x7e\x0a\x16\x63\x53\x30\x21\xc4\x43\xa6\x85\x29\xf4\xda\x4c\x72\x2b\x4c\x34\x06\x12\xbc\xb3\xf1\x11\x1d\x8c\x0f\xc9\xa9\x2d\x81\xb1\x03\x21\x21\x9b\x7a\xb9\x6d\x0a\x77\xda\x48\x69\x48\x24\x7c\x6a\x9b\xed\xa5\x15\x2f\x39\x84\xb0\x22\x26\x1b\x87\xe1\x71\x53\x3a\xf5\xd6\x8c\x6c\xcd\x22\xd0\x7a\x25\x23\xb2\x99\x56\x18\xd1\xb5\xe6\xd1\x98\x3f\x58\x6f\x69\x22\x88\x67\x67\x52\xd6\xda\x06\x6e\x67\x07\x18\x72\xa0\x63\x44\x14\xbf\xa7\x46\x0c\x69\xda\xbd\x6d\x7a\x76\xb6\x91\x5c\x1a\x61\x80\x04\xb1\x16\xe2\xd9\xb4\x4a\xea\x4b\x8b\x1d\xf3\x64\x76\xf0\x42\xe3\x90\xc3\xa0\xc2\x0e\x6d\x32\x2d\xec\x34\xa1\x59\x04\x51\x9c\x21\x94\x90\x29\xe2\x44\xbc\x21\x52\xa9\xcd\x21\x98\x28\xb4\x69\x75\xa3\x36\xfb\xc6\xc8\x05\x04\x93\xa6\x86\x18\x1d\x8c\x0c\x42\x4c\x28\xbc\xa3\x6d\x56\xf6\x46\x30\xe6\x83\x0b\x69\xde\xa9\x12\x51\xd9\x6c\x2e\x15\xd3\x85\x35\xac\xce\x0e\x2f\x52\x13\x27\x73\x53\x91\xa6\xd3\xea\x3d\x99\xd9\x2d\x04\xdc\x44\xd3\x1c\x07\x3e\xa5\x6d\x96\x11\xdf\xdc\x36\x9c\xe6\x0d\x4c\x30\x52\x18\xd9\x25\x96\x22\x42\x0b\x8d\x59\x30\xbf\x40\x2c\x64\x23\xdf\x08\xad\x6c\xf2\xda\xa6\xbd\x37\xf2\xd9\x19\xe1\x86\x34\x11\x7b\x6f\x9b\xcc\xb2\xb9\x37\xd8\x46\x7e\xa9\xf8\x19\x0d\x07\x82\xb5\x56\xc9\x06\x6b\x64\x5e\xc9\x43\x23\xeb\xef\x08\x8e\x4a\x85\xab\xf8\xe0\x0b\x05\x47\xff\x05\x82\x83\xe6\x3a\x1a\x71\x42\x2c\x41\x70\x5a\x23\x43\x34\x69\x11\xf8\x46\x20\x25\x3f\xbd\xe6\x5b\x9d\xbd\x2f\x3a\x10\x9d\x7c\x86\xe2\xd3\xa2\x43\x44\xe6\x13\xd1\x41\x0c\x20\x2a\xa3\x1d\xb8\x0d\xf3\xe6\x64\xa3\xff\x56\xe5\x07\x44\x47\x50\xa1\xf6\x20\x3a\x9a\x0f\x8b\x8e\x62\x26\x16\x8d\xda\x53\xd7\xef\x8a\x0e\x39\xd4\x9b\x14\xf3\x9f\x14\x1d\xde\x44\x87\x7f\x5f\x74\xe0\xb7\x07\xd1\x61\x87\x15\xe3\x47\x44\x47\x6e\x84\xeb\xb9\xa2\x43\x0e\xa4\x5e\x20\x3a\xba\x59\x74\xd4\xba\x69\x2b\x07\x2c\xa5\xe2\x94\xb5\xfb\x4c\xd7\xdc\x19\xa6\x5a\x3b\x4c\x82\xec\x43\x4e\xc9\x4d\xa9\x37\x86\x11\xc1\x6d\x67\x07\x47\x76\x80\x0c\x66\x24\xd6\xd6\x97\xbc\x89\x53\x7c\xd2\xcc\x07\xd3\x90\x5e\x3b\x0c\x97\xf1\xad\x5e\x76\x96\x0b\xe4\x08\x58\xa5\xce\xe0\x2b\x7c\xd4\xda\x06\xf8\xbc\x99\x8f\x1d\x10\x59\xf2\x62\xb0\xcd\xab\xd4\x36\x25\x64\x6d\xb9\xf5\xa9\x5c\x37\x39\xe8\x11\x7d\xab\x76\xd7\xb6\x81\x4c\x5c\x7d\xd0\xda\x09\x59\x97\xcd\x7f\x1b\x07\x81\xdd\x65\x9a\x4f\x83\x1d\xfe\xc9\x86\xcb\xa8\x35\x55\x0e\xde\x6d\x93\x66\xb4\x4d\xa2\x3a\xe8\x1c\x95\x1d\x3e\x54\xd6\xfb\xf0\xf3\xbc\xe9\x84\xb8\x19\x6c\xf3\x28\x37\xb2\x8d\x3d\x83\x6d\x16\xb1\x7e\xa9\xab\xf3\xc1\x5a\xa6\xb6\xe0\x23\x62\x02\x96\xc4\x07\xbd\xe2\x1c\xfc\x91\xb7\xc4\xb1\x9e\x0f\x70\x1b\xb5\xa7\x19\xdf\x15\x1d\xd5\x17\x8a\x0e\xe2\xfe\x55\x44\x47\xfd\x65\xa2\xa3\xb2\x97\x0d\xe8\x99\xa3\xf9\x76\x3e\xa4\x41\xac\x05\xdb\x54\xc4\x5e\xb0\x57\x4e\x8f\x87\x62\xbd\x1d\xf2\x20\x00\xc8\xf3\xd1\x36\x36\x11\x06\xb9\xad\x81\xfe\x44\x6f\xa4\x6f\x49\x1d\x7c\x2a\x3a\xba\x77\x45\x47\xf9\x11\xd1\xd1\x7d\x4a\x74\x8c\xef\x8b\x0e\x6a\x08\x79\x3b\x59\x0f\xc1\x97\x5f\x22\x3a\x1e\x29\xe2\xdf\x40\x75\x3c\x0e\xfe\x0c\xd9\x71\xf4\xf6\xe2\x8b\x75\xc7\xe3\x44\x5f\x45\x78\xbc\xe7\x94\xbf\x9b\xf2\x78\xe2\x84\x47\xe9\xd1\xfd\x83\x49\x0f\x7b\xbd\xf4\xff\x95\xf6\xa0\x1e\xc8\x4b\x12\xff\xd4\x1e\x7f\x9d\xf6\xe8\xdf\xd7\x1e\x95\xf9\xb5\x32\xbf\x0a\xef\x6e\xb5\x5f\xd3\x5f\x6a\x7b\xb9\x07\xdb\x89\x09\xbc\x93\x5e\x29\x2f\x78\xd5\x76\x40\xd0\xdb\x0b\x51\xc1\x38\x62\xaf\x7e\x6f\xad\x66\xcb\x46\x60\x61\xfe\x34\xae\xda\x18\xbf\xac\x0c\x23\xd5\xa0\x07\x29\xf8\x55\x0e\xaf\xe6\x03\xb8\x52\x37\x28\xa5\x27\xd5\xc6\x5f\xec\xe5\x9e\xf9\x85\x2c\xe1\x71\x95\xfa\x78\x6a\x1f\x0f\xeb\x64\x0d\xc3\x23\xfe\x0b\x3b\xd8\xa0\x1f\x80\x75\xf0\x06\xc7\xca\x4d\x33\x4d\xc6\xe1\xe0\xcb\x8d\x1d\x50\x09\xdf\x6a\x34\x9e\xf4\x86\xc6\x36\xdc\xe8\x75\x68\x17\x30\x47\xbf\x9e\x0f\x5c\x64\xd3\xd5\x5e\x8c\xcb\xed\xb0\xa6\xb3\x9e\xce\x7a\x4b\x7b\x51\x4a\x5e\x94\x49\xb5\xa7\x8f\xf6\xb2\x8b\xbc\x00\x61\x3a\xa7\xb1\x7c\x96\x83\x1d\xdb\x0c\x26\x0f\xe4\x05\x93\xc6\x36\xd9\x32\xe3\xc1\x76\x98\x5f\xd9\x8b\x07\xc2\x29\xed\x00\x27\x37\x2e\xdf\xd5\x8a\xc3\xdc\xfa\x3f\xf7\x70\xff\xbc\x0e\x79\x99\x24\x35\x3c\x79\xcd\x1f\x62\xdb\xda\x61\xe4\x7c\x20\xe6\xed\x00\xac\xb3\x17\xa7\x18\xb7\x37\x5e\x46\xee\xc2\xb7\x66\x0e\x08\x6e\xe7\x17\xa3\x8e\xf4\xc7\xff\x05\x00\x00\xff\xff\x8a\xbf\x41\x88\x00\x30\x00\x00")

func bindataTkgManifestTelemetryZzgeneratedBindataGoBytes() ([]byte, error) {
	return bindataRead(
		_bindataTkgManifestTelemetryZzgeneratedBindataGo,
		"tkg/manifest/telemetry/zz_generated.bindata.go",
	)
}

func bindataTkgManifestTelemetryZzgeneratedBindataGo() (*asset, error) {
	bytes, err := bindataTkgManifestTelemetryZzgeneratedBindataGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name:        "tkg/manifest/telemetry/zz_generated.bindata.go",
		size:        28672,
		md5checksum: "",
		mode:        os.FileMode(420),
		modTime:     time.Unix(1, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

//
// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
//
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}

//
// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
// nolint: deadcode
//
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

//
// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or could not be loaded.
//
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}

//
// AssetNames returns the names of the assets.
// nolint: deadcode
//
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

//
// _bindata is a table, holding each asset generator, mapped to its name.
//
var _bindata = map[string]func() (*asset, error){
	"tkg/manifest/telemetry/config-aws.yaml":         bindataTkgManifestTelemetryConfigawsYaml,
	"tkg/manifest/telemetry/config-azure.yaml":       bindataTkgManifestTelemetryConfigazureYaml,
	"tkg/manifest/telemetry/config-docker.yaml":      bindataTkgManifestTelemetryConfigdockerYaml,
	"tkg/manifest/telemetry/config-vsphere.yaml":     bindataTkgManifestTelemetryConfigvsphereYaml,
	"tkg/manifest/telemetry/zz_generated.bindata.go": bindataTkgManifestTelemetryZzgeneratedBindataGo,
}

//
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
//
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, &os.PathError{
					Op:   "open",
					Path: name,
					Err:  os.ErrNotExist,
				}
			}
		}
	}
	if node.Func != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: name,
			Err:  os.ErrNotExist,
		}
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{Func: nil, Children: map[string]*bintree{
	"tkg": {Func: nil, Children: map[string]*bintree{
		"manifest": {Func: nil, Children: map[string]*bintree{
			"telemetry": {Func: nil, Children: map[string]*bintree{
				"config-aws.yaml":         {Func: bindataTkgManifestTelemetryConfigawsYaml, Children: map[string]*bintree{}},
				"config-azure.yaml":       {Func: bindataTkgManifestTelemetryConfigazureYaml, Children: map[string]*bintree{}},
				"config-docker.yaml":      {Func: bindataTkgManifestTelemetryConfigdockerYaml, Children: map[string]*bintree{}},
				"config-vsphere.yaml":     {Func: bindataTkgManifestTelemetryConfigvsphereYaml, Children: map[string]*bintree{}},
				"zz_generated.bindata.go": {Func: bindataTkgManifestTelemetryZzgeneratedBindataGo, Children: map[string]*bintree{}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
