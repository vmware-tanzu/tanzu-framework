// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package log

import "encoding/json"

const (
	msgTypeLog      = "log"
	msgTypeProgress = "progress"
)

const (
	logTypeINFO    = "INFO"
	logTypeWARN    = "WARN"
	logTypeERROR   = "ERROR"
	logTypeFATAL   = "FATAL"
	logTypeUNKNOWN = "UNKNOWN"
)

// Update is a structure to be used for sending websocket message
type logUpdate struct {
	Type string  `json:"type"`
	Data logData `json:"data"`
}

// Data is a structure to be used for describing message data
type logData struct {
	Message string `json:"message,omitempty"`
	LogType string `json:"logType,omitempty"`

	Status       string   `json:"status,omitempty"`
	CurrentPhase string   `json:"currentPhase,omitempty"`
	TotalPhases  []string `json:"totalPhases,omitempty"`
}

func convertLogMsgToJSONBytes(logMsg []byte) []byte {
	data := logData{
		Message: string(logMsg),
		LogType: getLogType(logMsg),
	}
	update := logUpdate{Type: msgTypeLog, Data: data}

	updateBytes, err := json.Marshal(update)
	if err != nil {
		ForceWriteToStdErr([]byte("unable unmarshal log message"))
		return []byte{}
	}
	return updateBytes
}

func convertProgressMsgToJSONBytes(data *logData) []byte {
	update := logUpdate{Type: msgTypeProgress, Data: *data}

	updateBytes, err := json.Marshal(update)
	if err != nil {
		ForceWriteToStdErr([]byte("unable unmarshal progress message"))
		return []byte{}
	}
	return updateBytes
}

func getLogType(logMsg []byte) string {
	logType := logTypeUNKNOWN
	if len(msgTypeLog) > 0 {
		switch logMsg[0] {
		case 'I':
			logType = logTypeINFO
		case 'W':
			logType = logTypeWARN
		case 'E':
			logType = logTypeERROR
		case 'F':
			logType = logTypeFATAL
		}
	}
	return logType
}
