package handlers

import (
	docsSwag "backend/docs"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

const (
	readinessProbeName = "/healthcheck"
	apiEndpoint        = "/api/v1"
)

type HTTPController interface {
	RegisterHandlers(router *gin.RouterGroup)
}

func Register(router *gin.RouterGroup, host string, controllers ...HTTPController) {
	router.Use(gin.Recovery(), corsMiddleware())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.GET(readinessProbeName, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
		})
	})

	docsSwag.SwaggerInfo.Host = host
	docsSwag.SwaggerInfo.BasePath = apiEndpoint

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
