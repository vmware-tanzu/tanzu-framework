// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package log

import "github.com/go-logr/logr"

// GetLogr returns logr logger object that can be used to attach to some external logr logger
func GetLogr() logr.Logger {
	lr := logr.Logger{}.WithSink(&logrWrapper{logger: l.WithCallDepth(2)})
	return lr
}

// logrWrapper is a logr.LogSink implementation wrapper using LoggerImpl interface
type logrWrapper struct {
	logger LoggerImpl
}

var _ logr.LogSink = &logrWrapper{}

// Init receives optional information about the logr library for LogSink
// implementations that need it.
func (lr *logrWrapper) Init(info logr.RuntimeInfo) {
	lr.logger = lr.logger.WithCallDepth(info.CallDepth)
}

// Enabled tests whether this LogSink is enabled at the specified V-level.
// For example, commandline flags might be used to set the logging
// verbosity and disable some info logs.
func (lr *logrWrapper) Enabled(level int) bool {
	return lr.logger.V(level).Enabled()
}

// Info logs a non-error message with the given key/value pairs as context.
// The level argument is provided for optional logging.  This method will
// only be called when Enabled(level) is true. See Logger.Info for more
// details.
func (lr *logrWrapper) Info(level int, msg string, keysAndValues ...interface{}) {
	lr.logger.V(level).Info(msg, keysAndValues...)
}

// Error logs an error, with the given message and key/value pairs as
// context.  See Logger.Error for more details.
func (lr *logrWrapper) Error(err error, msg string, keysAndValues ...interface{}) {
	lr.logger.Error(err, msg, keysAndValues...)
}

// WithValues returns a new LogSink with additional key/value pairs.  See
// Logger.WithValues for more details.
func (lr *logrWrapper) WithValues(keysAndValues ...interface{}) logr.LogSink {
	nlr := lr.clone()
	nlr.logger.WithValues(keysAndValues...)
	return nlr
}

// WithName returns a new LogSink with the specified name appended.  See
// Logger.WithName for more details.
func (lr *logrWrapper) WithName(name string) logr.LogSink {
	nlr := lr.clone()
	nlr.logger.WithName(name)
	return nlr
}

func (lr *logrWrapper) clone() *logrWrapper {
	return &logrWrapper{
		logger: lr.logger.Clone(),
	}
}
