// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const nanoSecondsPerMicroSecond = 1000

// logEntry defines the information that can be used for composing a log line.
type logEntry struct {
	// Prefix of the log line, composed of the hierarchy of log.WithName values.
	Prefix string

	// Level of the LogEntry.
	Level int32

	// Values of the log line, composed of the concatenation of log.WithValues and KeyValue pairs passed to log.Info.
	Values []interface{}
}

// NewLogger returns a new instance of the clusterctl.
func NewLogger() LoggerImpl {
	return &logger{}
}

// logger defines a clusterctl friendly logr.Logger
type logger struct {
	threshold *int32
	level     int32
	prefix    string
	values    []interface{}
	callDepth int
}

var _ LoggerImpl = &logger{}

// SetThreshold implements a New Option that allows to set the threshold level for a logger.
// The logger will write only log messages with a level/V(x) equal or higher to the threshold.
func (l *logger) SetThreshold(threshold *int32) {
	l.threshold = threshold
}

// WithCallDepth implements a New Option that allows to set the callDepth level for a logger.
func (l *logger) WithCallDepth(callDepth int) LoggerImpl {
	nl := l.clone()
	nl.callDepth = callDepth
	return nl
}

// Enabled tests whether this Logger is enabled.
func (l *logger) Enabled() bool {
	if l.threshold == nil {
		return true
	}
	return l.level <= *l.threshold
}

// Info logs a non-error message with the given key/value pairs as context.
func (l *logger) Info(msg string, kvs ...interface{}) {
	l.Print(msg, nil, "INFO", kvs...)
}

// Infof logs a non-error messages with the given message format with format specifier and arguments.
func (l *logger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Print(msg, nil, "INFO")
}

// Error logs an error message with the given key/value pairs as context.
func (l *logger) Error(err error, msg string, kvs ...interface{}) {
	l.Print(msg, err, "ERROR", kvs...)
}

// Warning logs a warning messages with the given key/value pairs as context.
func (l *logger) Warning(msg string, kvs ...interface{}) {
	l.Print(msg, nil, "WARN", kvs...)
}

// Warningf logs a warning messages with the given message format with format specifier and arguments.
func (l *logger) Warningf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Print(msg, nil, "WARN")
}

// Fatal logs a fatal message with the given key/value pairs as context and returns with os.exit(1)
func (l *logger) Fatal(err error, msg string, kvs ...interface{}) {
	l.Print(msg, err, "ERROR", kvs...)
	os.Exit(1)
}

// Outputf writes a message to stdout
func (l *logger) Outputf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Print(msg, nil, "OUTPUT")
}

// V returns an InfoLogger value for a specific verbosity level.
func (l *logger) V(level int) LoggerImpl {
	nl := l.clone()
	nl.level = int32(level)
	return nl
}

// WithName adds a new element to the logger's name.
func (l *logger) WithName(name string) LoggerImpl {
	nl := l.clone()
	if len(l.prefix) > 0 {
		nl.prefix = l.prefix + "/"
	}
	nl.prefix += name
	return nl
}

// WithValues adds some key-value pairs of context to a logger.
func (l *logger) WithValues(kvList ...interface{}) LoggerImpl {
	nl := l.clone()
	nl.values = append(nl.values, kvList...)
	return nl
}

func (l *logger) clone() *logger {
	return &logger{
		threshold: l.threshold,
		level:     l.level,
		prefix:    l.prefix,
		values:    copySlice(l.values),
		callDepth: l.callDepth,
	}
}

func (l *logger) Clone() LoggerImpl {
	return l.clone()
}

func (l *logger) CloneWithLevel(level int) LoggerImpl {
	return &logger{
		threshold: l.threshold,
		level:     int32(level),
		prefix:    l.prefix,
		values:    copySlice(l.values),
		callDepth: l.callDepth,
	}
}

func (l *logger) Print(msg string, err error, logType string, kvs ...interface{}) {
	values := copySlice(l.values)
	values = append(values, kvs...)
	values = append(values, "msg", msg)
	if err != nil {
		values = append(values, "error", err)
	}
	header := []byte(l.header(logType, l.callDepth))
	_, _ = logWriter.Write(header, []byte(l.getLogString(values)), l.Enabled(), l.level, logType)
}

func (l *logger) getLogString(values []interface{}) string {
	entry := logEntry{
		Prefix: l.prefix,
		Level:  l.level,
		Values: values,
	}
	f, err := flatten(entry)
	if err != nil {
		_, _ = logWriter.Write([]byte{}, []byte(err.Error()), l.Enabled(), 0, "WARN")
		return ""
	}
	return f
}

func copySlice(in []interface{}) []interface{} {
	out := make([]interface{}, len(in))
	copy(out, in)
	return out
}

// flatten returns a human readable/machine parsable text representing the LogEntry.
// Most notable difference with the klog implementation are:
//   - The message is printed at the beginning of the line, without the Msg= variable name e.g.
//     "Msg"="This is a message" --> This is a message
//   - Variables name are not quoted, eg.
//     This is a message "Var1"="value" --> This is a message Var1="value"
//   - Variables are not sorted, thus allowing full control to the developer on the output.
func flatten(entry logEntry) (string, error) { // nolint:gocyclo
	var msgValue string
	var errorValue error
	if len(entry.Values)%2 == 1 {
		return "", errors.New("log entry cannot have odd number off keyAndValues")
	}

	keys := make([]string, 0, len(entry.Values)/2)
	values := make(map[string]interface{}, len(entry.Values)/2)
	for i := 0; i < len(entry.Values); i += 2 {
		k, ok := entry.Values[i].(string)
		if !ok {
			return "", errors.Errorf("WARNING: key is not a string: %s", entry.Values[i])
		}
		var v interface{}
		if i+1 < len(entry.Values) {
			v = entry.Values[i+1]
		}
		switch k {
		case "msg":
			msgValue, ok = v.(string)
			if !ok {
				return "", errors.Errorf("WARNING: the msg value is not of type string: %s", v)
			}
		case "error":
			errorValue, ok = v.(error)
			if !ok {
				return "", errors.Errorf("WARNING: the error value is not of type error: %s", v)
			}
		default:
			if _, ok := values[k]; !ok {
				keys = append(keys, k)
			}
			values[k] = v
		}
	}
	str := ""
	if entry.Prefix != "" {
		str += fmt.Sprintf("[%s] ", entry.Prefix)
	}
	str += msgValue
	if errorValue != nil {
		if msgValue != "" {
			str += ": "
		}
		str += errorValue.Error()
	}
	for _, k := range keys {
		prettyValue, err := pretty(values[k])
		if err != nil {
			return "", err
		}
		str += fmt.Sprintf(" %s=%s", k, prettyValue)
	}
	if str[len(str)-1] != '\n' {
		str += "\n"
	}

	return str, nil
}

func pretty(value interface{}) (string, error) {
	jb, err := json.Marshal(value)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to marshal %s", value)
	}
	return string(jb), nil
}

func (l *logger) header(logType string, depth int) string {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	} else if slash := strings.LastIndex(file, "/"); slash >= 0 {
		path := file
		file = path[slash+1:]
	}

	now := time.Now()
	_, month, day := now.Date()
	hour, minute, second := now.Clock()

	// Lmmdd hh:mm:ss.uuuuuu file:line]
	return fmt.Sprintf("%c%02d%02d %02d:%02d:%02d.%06d %s:%d] ", logType[0], int(month), day, hour, minute, second, now.Nanosecond()/nanoSecondsPerMicroSecond, file, line)
}
