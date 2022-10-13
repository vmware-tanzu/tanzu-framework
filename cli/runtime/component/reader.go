// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/aunum/log"
	"github.com/pkg/errors"
)

const (
	stdInIdentifier = "-"
)

// ReadInput reads from file or std input and return a byte array.
func ReadInput(filePath string) ([]byte, error) {
	if filePath == stdInIdentifier {
		return readFromStdInput()
	}
	return readFromFile(filePath)
}

// readFromFile opens the file at path and reads the file into a byte array.
func readFromFile(filePath string) ([]byte, error) {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("Error opening the input file %s.", filePath))
	}
	defer func() {
		if err = inputFile.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, inputFile)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("error reading from input file %s", filePath))
	}
	log.Debugf("read object --> \n---\n%s\n---\n", buf.String())
	return buf.Bytes(), nil
}

// readFromStdInput reads the incoming stream of bytes from console and return a byte array.
func readFromStdInput() ([]byte, error) {
	inBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, errors.WithMessage(err, "error reading from stdin")
	}
	return inBytes, nil
}
