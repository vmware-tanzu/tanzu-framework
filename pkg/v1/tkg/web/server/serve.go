// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package server providers backend api for UI
package server

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/loads"
	"github.com/pkg/errors"

	assetfs "github.com/elazarl/go-bindata-assetfs"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/handlers"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/ws"
	servermanifest "github.com/vmware-tanzu/tanzu-framework/tkg/manifest/server"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/types"
)

// Serve serve the kickstart UI
// nolint:gocritic
func Serve(initOptions client.InitRegionOptions, appConfig types.AppConfig, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter,
	tkgTimeOut time.Duration, bind string, browser string) error {

	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "2.0")
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewKickstartUIAPI(swaggerSpec)
	server := restapi.NewServer(api)

	server.EnabledListeners = []string{"http"}
	host, port, err := net.SplitHostPort(bind)
	if err != nil {
		return errors.Errorf("Invalid binding address provided. Please use address in the the form '127.0.0.1:8080'")
	}
	server.Port, err = strconv.Atoi(port)
	if err != nil {
		return errors.Errorf("Invalid binding port provided. Please provide a valid number (e.g. 8080).")
	}
	server.Host = host
	server.Browser = browser

	ws.InitWebsocketUpgrader(server.Host)

	app := &handlers.App{InitOptions: initOptions, AppConfig: appConfig, TKGTimeout: tkgTimeOut, TKGConfigReaderWriter: tkgConfigReaderWriter}
	app.ConfigureHandlers(api)
	server.SetAPI(api)
	server.SetHandler(api.Serve(FileServerMiddleware))

	// check if the port is already in use, if so exit gracefully
	l, err := net.Listen("tcp", server.Host+":"+strconv.Itoa(server.Port))
	if err != nil {
		server.Logf("Failed to start the kickstart UI Server[Host:%s, Port:%d], error: %s\n", server.Host, server.Port, err)
		os.Exit(1)
	}
	l.Close()

	defer func() {
		err := server.Shutdown()
		if err != nil {
			log.Fatalln(err)
		}
	}()
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
	return nil
}

// FileServerMiddleware serves ui resource
func FileServerMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/ws") {
			ws.HandleWebsocketRequest(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/api/providers") {
			handler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/api/features") {
			handler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/api/edition") {
			handler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/api/integration") {
			handler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/api/avi") {
			handler.ServeHTTP(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/api/ldap") {
			handler.ServeHTTP(w, r)
		} else {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")

			if strings.HasSuffix(r.URL.Path, ".css") {
				w.Header().Add("Content-Type", "text/css")
			}

			fs := &assetfs.AssetFS{Asset: servermanifest.Asset, AssetDir: servermanifest.AssetDir, AssetInfo: servermanifest.AssetInfo, Prefix: "pkg/v1/tkg/web/dist/tkg-kickstart-ui", Fallback: "index.html"}
			http.FileServer(fs).ServeHTTP(w, r)
		}
	})
}
