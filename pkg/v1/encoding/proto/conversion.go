// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb" // nolint
	"github.com/golang/protobuf/proto"  // nolint
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
)

// InputFileToProto reads a json/yaml/stdin and converts it to protobuf format.
func InputFileToProto(filePath string, outResource proto.Message) error {
	b, err := component.ReadInput(filePath)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	err = BufferToProto(buf, outResource, strings.TrimPrefix(filepath.Ext(filePath), "."))
	if err != nil {
		return err
	}
	return nil
}

// BufferToProto tries to unmarshal a json/yaml buffer of bytes to a protocol buffer message.
func BufferToProto(buf *bytes.Buffer, message proto.Message, format string) error {
	switch format {
	case "json":
		return jsonUnmarshal(buf, message)
	case "yaml", "yml":
		return yamlUnmarshal(buf, message)
	default:
		if yamlUnmarshal(buf, message) != nil {
			if jsonUnmarshal(buf, message) != nil {
				return fmt.Errorf("format %q unknown, options are json|yaml", format)
			}
		}
		return nil
	}
}

// jsonUnmarshal will unmarshal json bytes to a protocol buffer message.
func jsonUnmarshal(buf *bytes.Buffer, message proto.Message) error {
	err := jsonpb.Unmarshal(buf, message)
	if err != nil {
		return err
	}
	return nil
}

// yamlUnmarshal will unmarshal yaml bytes to a protocol buffer message.
func yamlUnmarshal(buf *bytes.Buffer, message proto.Message) error {
	jsonBytes, err := yaml.YAMLToJSON(buf.Bytes())
	if err != nil {
		return err
	}
	err = jsonpb.UnmarshalString(string(jsonBytes), message)
	if err != nil {
		return err
	}
	return nil
}
