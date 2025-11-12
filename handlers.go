package main

// handlers module holds all HTTP handlers functions
//
// Copyright (c) 2025 - Valentin Kuznetsov <vkuznet@gmail.com>
//

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	services "github.com/CHESSComputing/golib/services"
	"github.com/gin-gonic/gin"
)

// UploadPayload define structure of HTTP POST request
type UploadPayload struct {
	DID  string `json:"did"`
	File string `json:"file"`
}

// DataDirsHandler handles GET HTTP requests
func DataDirsHandler(c *gin.Context) {
	// we return all dids registered in data hub
	dirs, err := ListDirs(StorageDir)
	if err != nil {
		resp := services.Response("DataHub", http.StatusBadRequest, services.ServiceError, err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	c.JSON(http.StatusOK, dirs)
	return
}

// DataHandler handles GET HTTP requests
func DataHandler(c *gin.Context) {
	didhash := c.Param("didhash")
	if didhash == "" {
		// we return all dids registered in data hub
		dirs, err := ListDirs(StorageDir)
		if err != nil {
			resp := services.Response("DataHub", http.StatusBadRequest, services.ServiceError, err)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		c.JSON(http.StatusOK, dirs)
		return
	}
	filepathParam := c.Param("filepath")
	path := filepath.Join(StorageDir, didhash, filepath.Clean(filepathParam))

	// get info about our path
	_, err := os.Stat(path)
	if err != nil {
		resp := services.Response("DataHub", http.StatusBadRequest, services.ServiceError, err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	if path == "" {
		resp := services.Response("DataHub", http.StatusBadRequest, services.ServiceError, err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Serve file content if it's a file
	http.ServeFile(c.Writer, c.Request, path)
}

// UploadHandler handles GET HTTP requests
func UploadHandler(c *gin.Context) {
	var payload UploadPayload

	// Parse incoming JSON
	if err := c.ShouldBindJSON(&payload); err != nil {
		resp := services.Response("DataHub", http.StatusBadRequest, services.ServiceError, err)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if payload.DID == "" || payload.File == "" {
		resp := services.Response("DataHub", http.StatusBadRequest, services.ServiceError, errors.New("missing did or file"))
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	// we need to convert did into md5 hash
	sum := md5.Sum([]byte(payload.DID))
	hash := hex.EncodeToString(sum[:])

	targetDir := filepath.Join(StorageDir, hash)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		resp := services.Response("DataHub", http.StatusInternalServerError, services.ServiceError, err)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Determine if `file` is base64 or path to file
	var tmpFilePath string
	if _, err := os.Stat(payload.File); err == nil {
		tmpFilePath = payload.File // user passed existing path
	} else {
		// assume base64-encoded content
		tmpFile, err := os.CreateTemp("", "upload-*")
		if err != nil {
			resp := services.Response("DataHub", http.StatusInternalServerError, services.ServiceError, err)
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
		defer tmpFile.Close()
		decoded, err := base64.StdEncoding.DecodeString(payload.File)
		if err != nil {
			resp := services.Response("DataHub", http.StatusBadRequest, services.ServiceError, err)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		if _, err = tmpFile.Write(decoded); err != nil {
			resp := services.Response("DataHub", http.StatusInternalServerError, services.ServiceError, err)
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
		tmpFilePath = tmpFile.Name()
	}

	// Detect file type by extension
	switch {
	case strings.HasSuffix(tmpFilePath, ".zip"):
		if err := extractZIP(tmpFilePath, targetDir); err != nil {
			resp := services.Response("DataHub", http.StatusInternalServerError, services.ServiceError, err)
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
	case strings.HasSuffix(tmpFilePath, ".tar"):
		if err := extractTAR(tmpFilePath, targetDir); err != nil {
			resp := services.Response("DataHub", http.StatusInternalServerError, services.ServiceError, err)
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
	case strings.HasSuffix(tmpFilePath, ".tar.gz") || strings.HasSuffix(tmpFilePath, ".tgz"):
		if err := extractTARGZ(tmpFilePath, targetDir); err != nil {
			resp := services.Response("DataHub", http.StatusInternalServerError, services.ServiceError, err)
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
	default:
		// regular file — copy to target dir
		dest := filepath.Join(targetDir, filepath.Base(tmpFilePath))
		if err := copyFile(tmpFilePath, dest); err != nil {
			resp := services.Response("DataHub", http.StatusInternalServerError, services.ServiceError, err)
			c.JSON(http.StatusInternalServerError, resp)
			return
		}
	}

	resp := services.Response("DataHub", http.StatusOK, services.OK, nil)
	c.JSON(http.StatusOK, resp)
}

// DeleteHandler handles DELETE /delete/:did requests
func DeleteHandler(c *gin.Context) {
	var err error
	didhash := c.Param("didhash")

	if didhash == "" {
		resp := services.Response("DataHub", http.StatusBadRequest, services.ServiceError, errMissingParam("didhash"))
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	targetDir := filepath.Join(StorageDir, didhash)

	// Check if directory exists
	if _, err = os.Stat(targetDir); os.IsNotExist(err) {
		resp := services.Response("DataHub", http.StatusNotFound, services.ServiceError, errNotFoundDir(targetDir))
		c.JSON(http.StatusNotFound, resp)
		return
	} else if err != nil {
		resp := services.Response("DataHub", http.StatusInternalServerError, services.ServiceError, err)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Remove directory recursively
	if err = os.RemoveAll(targetDir); err != nil {
		resp := services.Response("DataHub", http.StatusInternalServerError, services.ServiceError, err)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := services.Response("DataHub", http.StatusOK, services.OK, nil)
	c.JSON(http.StatusOK, resp)
}
