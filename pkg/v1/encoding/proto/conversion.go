// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aunum/log"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// InputFileToProto reads a json/yaml input file and converts it to protobuf format
func InputFileToProto(filePath string, outResource proto.Message) error {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return errors.WithMessage(err, "Error opening the input file.")
	}
	defer inputFile.Close()

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, inputFile)
	log.Debugf("read object --> \n---\n%s\n---\n", string(buf.Bytes()))

	err = BufferToProto(buf, outResource, strings.TrimPrefix(filepath.Ext(filePath), "."))
	if err != nil {
		return err
	}

	return nil
}

// BufferToProto tries to unmarshal a json|yaml bytes buffer to a proto message.
func BufferToProto(buf *bytes.Buffer, message proto.Message, format string) error {
	switch format {
	case "json":
		err := jsonpb.Unmarshal(buf, message)
		if err != nil {
			return err
		}
	case "yaml":
		jsonBytes, err := yaml.YAMLToJSON(buf.Bytes())
		if err != nil {
			return err
		}
		err = jsonpb.UnmarshalString(string(jsonBytes), message)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("format %q unknown, options are json|yaml", format)
	}
	return nil
}
