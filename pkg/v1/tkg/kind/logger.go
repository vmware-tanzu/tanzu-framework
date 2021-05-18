// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kind

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kind/pkg/log"

	logt "github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
)

// NewLogger creates logger to retrieve kind cluster provider logs
func NewLogger(logLevel log.Level) log.Logger {
	return &Logger{
		level: logLevel,
	}
}

// NewLoggerWithChannel creates logger to retrieve kind cluster provider logs with channel
func NewLoggerWithChannel(ch chan string, logLevel log.Level) log.Logger {
	return &Logger{
		ch: ch,
	}
}

// Logger implements interface to retrieve logs from kind cluster provider
type Logger struct {
	ch    chan string
	level log.Level
}

// Warn should be used to write user facing warnings
func (l *Logger) Warn(message string) {
	write(message, l.ch, l.level)
}

// Warnf should be used to write Printf style user facing warnings
func (l *Logger) Warnf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	write(message, l.ch, l.level)
}

// Error may be used to write an error message when it occurs
// Prefer returning an error instead in most cases
func (l *Logger) Error(message string) {
	write(message, l.ch, l.level)
}

// Errorf may be used to write a Printf style error message when it occurs
// Prefer returning an error instead in most cases
func (l *Logger) Errorf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	write(message, l.ch, l.level)
}

// V returns an InfoLogger for a given verbosity Level
func (l *Logger) V(level log.Level) log.InfoLogger {
	return &InfoLogger{
		ch:      l.ch,
		level:   level,
		enabled: level <= l.level,
	}
}

// InfoLogger defines the info logging interface kind uses
type InfoLogger struct {
	ch      chan string
	level   log.Level
	enabled bool
}

// Info is used to write a user facing status message
func (l *InfoLogger) Info(message string) {
	if l.enabled {
		write(message, l.ch, l.level)
	}
}

// Infof is used to write a Printf style user facing status message
func (l *InfoLogger) Infof(format string, args ...interface{}) {
	if l.enabled {
		message := fmt.Sprintf(format, args...)
		write(message, l.ch, l.level)
	}
}

// Enabled should return true if this verbosity level is enabled
func (l *InfoLogger) Enabled() bool {
	return false
}

// TODO: check if we should include other characters or not
func write(message string, ch chan string, logLevel log.Level) {
	// Skip status message which prints the same message again with success/failure
	// character at the start of the log.
	if strings.HasPrefix(message, " ✗ ") || strings.HasPrefix(message, " ✓ ") {
		return
	}

	// Also skip all not ascii characters from the logs
	cleanMessage := []rune{}
	lastchar := rune(32) //nolint:gomnd
	for _, c := range message {
		if int(c) < 32 || int(c) > 126 || (lastchar == ' ' && c == lastchar) {
			continue
		}
		lastchar = c
		cleanMessage = append(cleanMessage, c)
	}
	if ch == nil {
		if logLevel <= 1 {
			logt.V(3).Info(string(cleanMessage))
		} else {
			logt.V(7).Info(string(cleanMessage))
		}
	} else {
		ch <- string(cleanMessage)
	}
}
