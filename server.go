package main

// server module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"log"

	srvConfig "github.com/CHESSComputing/golib/config"
	server "github.com/CHESSComputing/golib/server"
	"github.com/gin-gonic/gin"
)

// Verbose level
var Verbose int

// StorageDir defines location of storage directory
var StorageDir string

// helper function to setup our router
func setupRouter() *gin.Engine {
	routes := []server.Route{
		{Method: "GET", Path: "/datahub", Handler: DataHandler, Authorized: false, Scope: "read"},
		{Method: "GET", Path: "/datahub/:didhash/*filepath", Handler: DataHandler, Authorized: false, Scope: "read"},
		{Method: "POST", Path: "/datahub", Handler: UploadHandler, Authorized: true, Scope: "write"},
		{Method: "DELETE", Path: "/datahub/:didhash", Handler: DeleteHandler, Authorized: true, Scope: "delete"},
	}

	r := server.Router(routes, nil, "static", srvConfig.Config.DataHub.WebServer)
	return r
}

// Server defines our HTTP server
func Server() {
	Verbose = srvConfig.Config.DataHub.WebServer.Verbose
	StorageDir = srvConfig.Config.DataHub.StorageDir
	if StorageDir == "" {
		log.Fatal("DataHub cannot be started with empty storage dir")
	}
	log.Println("storage dir", StorageDir)

	// setup web router and start the service
	r := setupRouter()
	webServer := srvConfig.Config.DataHub.WebServer
	log.Printf("### webServer %+v", webServer)
	server.StartServer(r, webServer)
}
