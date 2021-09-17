# Logging

This section describes how the logging is implemented in tkgctl library.

tkgctl library has implemented custom `logr.Logger`. This custom logger is super set of `logr.Logger` which includes header formatting and verbosity level management and passes all the information about the logs with verbosity levels to writer. The actual write to stdout/stderr is done through this writer implementation which is logger itself.

Logger can be configured with the following options when creating tkgctl client:

```go
// LoggingOptions options to configure logging with tkgctl client
type LoggingOptions struct {
    // File log file name where you want logs to be written
    File string
    // Quietly if set logs will not be written to stdout/stderr
    Quietly bool
    // Verbosity number for the log level verbosity
    Verbosity int32
    // LogChannel if channel is set, writer will forward log messages to this log channel
    // The result of this will be in the format of 'LogData' struct mentioned in `pkg/log/type.go`
    LogChannel chan<- []byte
}
```

## writer

This writer is implemented to do 3 things when log arrives.

* Write log message to log-file if the log file is given
* Send the log message through the channel if channel is set (This is used to send the logs to the UI through websockets)
* Print the log to stdout/stderr if the quiet mode is not set. If quiet flag is passed skip printing to stdout/stderr

## Clusterctl logger

To retrive logs from the clusterctl library which is used as part of tkgctl client, the library implements and passes the subset of custom logger as `logr.Logger`.
