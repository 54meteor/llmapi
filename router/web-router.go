package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/controller"
	"strings"
)

func SetWebRouter(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.RequestURI, "/v1") || strings.HasPrefix(c.Request.RequestURI, "/api") {
			controller.RelayNotFound(c)
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("Frontend not available. Please deploy frontend separately or set FRONTEND_BASE_URL."))
	})
}
