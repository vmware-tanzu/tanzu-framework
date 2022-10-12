// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"log"
	"os"
	"path"
)

// Writer defines methods to write and configure tkg writer
type Writer interface {

	// Write writes message to stdout/stderr, logfile and sends to the channel if channel is set
	//
	// header is message header to append as a message prefix
	// msg is actual message to write
	// logEnabled is used to decide whether to write this message to stdout/stderr or not
	// logVerbosity is used to decide which message to write for different output types
	// logType used to decide should write to stdout or stderr
	Write(header []byte, msg []byte, logEnabled bool, logVerbosity int32, logType string) (n int, err error)

	// SetFile sets the logFile to writer
	// if the non-empty file name is used, writer will also
	// write the logs to this file
	SetFile(fileName string)

	// SetAuditLog sets the log file that should capture all logging activity. This
	// file will contain all logging regardless of set verbosity level.
	SetAuditLog(fileName string)

	// SetChannel sets the channel to writer
	// if channel is set, writer will forward log messages to this log channel
	SetChannel(channel chan<- []byte)

	// QuietMode sets the logging mode to quiet
	// If this mode is set, writer will not write anything to stderr
	QuietMode(quiet bool)

	// SetVerbosity sets verbosity level and also updates default verbosity level
	SetVerbosity(verbosity int32)

	// SendProgressUpdate sends the progress to the listening logChannel
	SendProgressUpdate(status string, step string, totalSteps []string)
}

var (
	originalStdout *os.File
	originalStderr *os.File
)

var defaultVerbosity int32 = 4

// NewWriter returns new custom writter for tkg-cli
func NewWriter() Writer {
	return &writer{}
}

type writer struct {
	logFile    string
	logChannel chan<- []byte
	verbosity  int32
	quiet      bool
	auditFile  string
}

// SetFile sets the logFile to writer
// if the non-empty file name is used, writer will also
// write the logs to this file
func (w *writer) SetFile(fileName string) {
	w.logFile = fileName
}

// SetAuditLog sets the log file that should capture all logging activity. This
// file will contain all logging regardless of set verbosity level.
func (w *writer) SetAuditLog(fileName string) {
	w.auditFile = fileName
}

// SetChannel sets the channel to writer
// if channel is set, writer will forward log messages to this log channel
func (w *writer) SetChannel(channel chan<- []byte) {
	w.logChannel = channel
}

// QuietMode sets the logging mode to quiet
// If this mode is set, writer will not write anything to stderr
func (w *writer) QuietMode(quiet bool) {
	w.quiet = quiet
}

// SetVerbosity sets verbosity level and also updates default verbosity level
func (w *writer) SetVerbosity(verbosity int32) {
	w.verbosity = verbosity
	if verbosity > defaultVerbosity {
		defaultVerbosity = verbosity
	}
}

// Write writes message to stdout/stderr, logfile and sends to the channel if channel is set
//
// header is message header to append as a message prefix
// msg is actual message to write
// logEnabled is used to decide whether to write this message to stdout/stderr or not
// logVerbosity is used to decide which message to write for different output types
// logType used to decide should write to stdout or stderr
func (w *writer) Write(header, msg []byte, logEnabled bool, logVerbosity int32, logType string) (n int, err error) {
	fullMsg := append(header, msg...) //nolint:gocritic

	// Always write to the audit log so it captures everything
	if w.auditFile != "" {
		fileWriter(w.auditFile, fullMsg)
	}

	// write to logfile, channel only if verbosityLevel is <= default VerbosityLevel
	if logVerbosity <= defaultVerbosity {
		if w.logFile != "" {
			fileWriter(w.logFile, fullMsg)
		}
		if w.logChannel != nil {
			w.logChannel <- convertLogMsgToJSONBytes(fullMsg)
		}
	}

	// write to stdout/stderr if quiet mode is not set and logEnabled is true
	if !w.quiet && logEnabled {
		if logType == "OUTPUT" {
			stdoutWriter(msg)
		} else {
			stderrWriter(msg)
		}
	}

	return len(msg), nil
}

func (w *writer) SendProgressUpdate(status, currentPhase string, totalPhases []string) {
	if w.logChannel == nil {
		return
	}

	msgData := logData{
		Status:       status,
		CurrentPhase: currentPhase,
		TotalPhases:  totalPhases,
	}
	w.logChannel <- convertProgressMsgToJSONBytes(&msgData)
}

// UnsetStdoutStderr intercept the actual stdout and stderr
// this will ensure no other external library prints to stdout/stderr
// and use actual stdout/stderr through tkg writer only
// Note: Should not use this functions for normal cli commands like
//
//	tkg get regions, tkg set regions
//
// As it will stop any libraries like table.pretty to print on stdout
func UnsetStdoutStderr() {
	originalStdout = os.Stdout
	originalStderr = os.Stderr
	os.Stdout = nil
	os.Stderr = nil
	log.SetOutput(nil)
}

func fileWriter(logFileName string, msg []byte) {
	basePath := path.Dir(logFileName)
	filePath := path.Base(logFileName)

	if os.MkdirAll(basePath, 0o600) != nil {
		msg := "Unable to create log directory: " + basePath
		stderrWriter([]byte(msg))
		os.Exit(1)
	}

	logFileName = path.Join(basePath, filePath)
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		stderrWriter([]byte(err.Error()))
		os.Exit(1)
	}
	defer file.Close()

	_, fileWriteErr := file.Write(msg)
	if fileWriteErr != nil {
		stderrWriter([]byte(fileWriteErr.Error()))
		os.Exit(1) //nolint:gocritic
	}
}

func stdoutWriter(msg []byte) {
	if originalStdout != nil {
		_, _ = originalStdout.Write(msg)
	} else {
		_, _ = os.Stdout.Write(msg)
	}
}

func stderrWriter(msg []byte) {
	if originalStderr != nil {
		_, _ = originalStderr.Write(msg)
	} else {
		_, _ = os.Stderr.Write(msg)
	}
}

// ForceWriteToStdErr writes to stdErr
func ForceWriteToStdErr(msg []byte) {
	stderrWriter(msg)
}
