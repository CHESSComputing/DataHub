package main

// server module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"embed"
	"log"
	"net/http"

	srvConfig "github.com/CHESSComputing/golib/config"
	server "github.com/CHESSComputing/golib/server"
	"github.com/gin-gonic/gin"
)

// content is our static web server content.
//
//go:embed static
var StaticFs embed.FS

var Verbose int
var StaticDir, StorageDir string

// helper function to setup our router
func setupRouter() *gin.Engine {
	routes := []server.Route{
		{Method: "GET", Path: "/:did", Handler: DataHandler, Authorized: false},
		{Method: "POST", Path: "/", Handler: UploadHandler, Authorized: true, Scope: "write"},
		{Method: "DELETE", Path: "/", Handler: DeleteHandler, Authorized: true, Scope: "delete"},
	}

	r := server.Router(routes, nil, "static", srvConfig.Config.DataHub.WebServer)
	r.StaticFS("/datastore", http.Dir(StorageDir))
	return r
}

// Server defines our HTTP server
func Server() {
	Verbose = srvConfig.Config.DataHub.WebServer.Verbose
	StaticDir = srvConfig.Config.DataHub.WebServer.StaticDir
	StorageDir = srvConfig.Config.DataHub.StorageDir
	log.Println("storage dir", StorageDir)

	// setup web router and start the service
	r := setupRouter()
	webServer := srvConfig.Config.DataHub.WebServer
	log.Printf("### webServer %+v", webServer)
	server.StartServer(r, webServer)
}
