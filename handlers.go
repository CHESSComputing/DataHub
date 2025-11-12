package main

// handlers module holds all HTTP handlers functions
//
// Copyright (c) 2025 - Valentin Kuznetsov <vkuznet@gmail.com>
//

import (
	"net/http"

	services "github.com/CHESSComputing/golib/services"
	"github.com/gin-gonic/gin"
)

// DocParam defines parameters for uri binding
type DocParams struct {
	Name string `uri:"name" binding:"required"`
}

// DataHandler handles GET HTTP requests
func DataHandler(c *gin.Context) {
	var err error
	resp := services.Response("MLHub", http.StatusBadRequest, services.ServiceError, err)
	c.JSON(http.StatusBadRequest, resp)
}

// UploadHandler handles GET HTTP requests
func UploadHandler(c *gin.Context) {
	var err error
	resp := services.Response("MLHub", http.StatusBadRequest, services.ServiceError, err)
	c.JSON(http.StatusBadRequest, resp)
}

// DeleteHandler handles GET HTTP requests
func DeleteHandler(c *gin.Context) {
	var err error
	resp := services.Response("MLHub", http.StatusBadRequest, services.ServiceError, err)
	c.JSON(http.StatusBadRequest, resp)
}
