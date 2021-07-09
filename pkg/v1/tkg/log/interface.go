// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package log provides logging functionalities
package log

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
)

// LoggerImpl represents the ability to log messages, both errors and not.
type LoggerImpl interface {
	logr.Logger

	// Infof logs a non-error messages with the given message format with format specifier and arguments.
	Infof(format string, args ...interface{})
	// Warning logs a warning messages with the given key/value pairs as context.
	Warning(msg string, kvs ...interface{})
	// Warningf logs a warning messages with the given message format with format specifier and arguments.
	Warningf(format string, args ...interface{})
	// Fatal logs a fatal message with the given key/value pairs as context and returns with os.Exit(1)
	Fatal(err error, msg string, kvs ...interface{})
	// Print logs a message of generic type
	Print(msg string, err error, logType string, kvs ...interface{})
	// Output writes a message to stdout
	Outputf(msg string, kvs ...interface{})
	// SetThreshold implements a New Option that allows to set the threshold level for a logger.
	// The logger will write only log messages with a level/V(x) equal or higher to the threshold.
	SetThreshold(threshold *int32)

	CloneWithLevel(level int) LoggerImpl
}

var l = NewLogger()

// Info logs a non-error message with the given key/value pairs as context.
func Info(msg string, kvs ...interface{}) {
	l.Print(msg, nil, "INFO", kvs...)
}

// Infof logs a non-error message with the given key/value pairs as context.
func Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Print(msg, nil, "INFO")
}

// Error logs an error message with the given key/value pairs as context.
func Error(err error, msg string, kvs ...interface{}) {
	l.Print(msg, err, "ERROR", kvs...)
}

// Warning logs a warning messages with the given key/value pairs as context.
func Warning(msg string, kvs ...interface{}) {
	l.Print(msg, nil, "WARN", kvs...)
}

// Warningf logs a warning messages with the given message format with format specifier and arguments.
func Warningf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Print(msg, nil, "WARN")
}

// Fatal logs a fatal message with the given key/value pairs as context and returns with os.Exit(255)
func Fatal(err error, msg string, kvs ...interface{}) {
	l.Print(msg, err, "ERROR", kvs...)
	os.Exit(1)
}

// Outputf writes a message to stdout
func Outputf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Print(msg, nil, "OUTPUT")
}

// V returns an InfoLogger value for a specific verbosity level.
func V(level int) LoggerImpl {
	return l.CloneWithLevel(level)
}

// WithName adds a new element to the logger's name.
func WithName(name string) logr.Logger {
	return l.WithName(name)
}

// WithValues adds some key-value pairs of context to a logger.
func WithValues(kvList ...interface{}) logr.Logger {
	return l.WithValues(kvList...)
}

// GetLogr logs a warning messages with the given message format with format specifier and arguments.
func GetLogr() logr.Logger {
	return l
}

var logWriter = NewWriter()

// SetFile sets the logFile to writer
// if the non-empty file name is used, writer will also
// write the logs to this file
func SetFile(fileName string) {
	logWriter.SetFile(fileName)
}

// SetChannel sets the channel to writer
// if channel is set, writer will forward log messages to this log channel
func SetChannel(channel chan<- []byte) {
	logWriter.SetChannel(channel)
}

// QuietMode sets the logging mode to quiet
// If this mode is set, writer will not write anything to stderr
func QuietMode(quiet bool) {
	logWriter.QuietMode(quiet)
}

// SetVerbosity sets verbosity level and also updates default verbosity level
func SetVerbosity(verbosity int32) {
	l.SetThreshold(&verbosity)
	logWriter.SetVerbosity(verbosity)
}

// SendProgressUpdate sends the progress to the listening logChannel
func SendProgressUpdate(status, currentPhase string, totalPhases []string) {
	logWriter.SendProgressUpdate(status, currentPhase, totalPhases)
}
