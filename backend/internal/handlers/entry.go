package handlers

import (
	"backend/pkg/logging"
	"bytes"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"io"
	"net/http"
	"strings"
)

const (
	readinessProbeName = "/healthcheck"
	apiEndpoint        = "/api/v2"
)

type HTTPController interface {
	RegisterHandlers(router *gin.RouterGroup)
}

func Register(router *gin.RouterGroup, host string, controllers ...HTTPController) {
	router.Use(gin.Recovery(), corsMiddleware())
	router.Use(gin.Recovery(), logMiddleware())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.GET(readinessProbeName, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
		})
	})

	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(204)
	})

	api := router.Group(apiEndpoint)

	for _, controller := range controllers {
		controller.RegisterHandlers(api)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, PUT, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func logMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		defer c.Next()

		if strings.Contains(c.Request.RequestURI, "/swagger") {
			return
		}

		var mpValue map[string][]string

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logging.Errorf("Error reading request body: %v", err)
			return
		}

		if c.Request.MultipartForm != nil {
			mpValue = c.Request.MultipartForm.Value
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		entry := logging.WithFields(logging.Fields{
			"client_ip":      c.ClientIP(),
			"method":         c.Request.Method,
			"path":           c.Request.RequestURI,
			"form":           c.Request.Form,
			"post_form":      c.Request.PostForm,
			"multipart_form": mpValue,
			"body":           string(body),
		})

		if c.Writer.Status() >= 500 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("")
		}
	}
}
