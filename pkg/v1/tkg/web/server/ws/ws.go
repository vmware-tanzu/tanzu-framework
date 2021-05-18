// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package ws implements websocket used with UI communications
package ws

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
)

var upgrader = websocket.Upgrader{}

var (
	wsConnections []*websocket.Conn
	logs          = [][]byte{}
)

const (
	localhostIP = "127.0.0.1"
	localhost   = "localhost"
)

// InitWebsocketUpgrader initializes the upgrader and configures the
// CheckOrigin function using the provided host
func InitWebsocketUpgrader(hostBind string) {
	isLocal := hostBind == localhostIP || hostBind == localhost
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 1. Get the request origin
			// 2. Verify the origin is in allowed origins
			// (Allowed origins are 'localhost' '127.0.0.1' and 'user-specified host with --bind flag')
			originURL, err := url.Parse(r.Header.Get("Origin"))
			if err != nil {
				log.ForceWriteToStdErr([]byte(fmt.Sprintf("Error verifying websocket origin url: %s", err.Error())))
				return false
			}

			origin, _, err := net.SplitHostPort(originURL.Host)
			if err != nil {
				log.ForceWriteToStdErr([]byte(fmt.Sprintf("Error verifying websocket origin: %s", err.Error())))
				return false
			}

			// (Allowed origins are 'localhost' '127.0.0.1' and 'user-specified hostname with --bind flag')
			if (isLocal && (origin == localhostIP || origin == localhost)) || origin == hostBind {
				return true
			}
			return false
		},
	}
}

// HandleWebsocketRequest handles the websocket request coming from clients
// upgrade normal http request to websocket request and stores the connection
func HandleWebsocketRequest(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.ForceWriteToStdErr([]byte(fmt.Sprintf("web socket upgrade error: %s", err.Error())))
		return
	}
	wsConnections = append(wsConnections, ws)

	log.ForceWriteToStdErr([]byte("web socket connection established\n"))

	sendPendingLogsOnConnection(ws)

	ws.SetCloseHandler(func(code int, text string) error {
		deleteWSConnection(ws)
		ws.Close()
		return nil
	})

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func sendPendingLogsOnConnection(ws *websocket.Conn) {
	log.ForceWriteToStdErr([]byte(fmt.Sprintf("sending pending %v logs to UI\n", len(logs))))
	for _, logMsg := range logs {
		err := ws.WriteMessage(1, logMsg)
		if err != nil {
			log.ForceWriteToStdErr([]byte("fail to write log message to web socket"))
			break
		}
	}
}

func deleteWSConnection(conn *websocket.Conn) {
	index := -1
	for i, ws := range wsConnections {
		if conn == ws {
			index = i
			break
		}
	}
	if index != -1 {
		wsConnections = append(wsConnections[:index], wsConnections[index+1:]...)
	}
}

// SendLog send the log message to all the connected websocket clients
func SendLog(logMsg []byte) {
	var err error

	if len(wsConnections) == 0 {
		logs = append(logs, logMsg)
		return
	}

	for _, ws := range wsConnections {
		err = ws.WriteMessage(1, logMsg)
		if err != nil {
			log.ForceWriteToStdErr([]byte("fail to write log message to web socket"))
			break
		}
	}
}
