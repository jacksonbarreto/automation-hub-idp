package router

import (
	"github.com/gin-gonic/gin"
	"idp-automations-hub/internal/app/config"
)

func Initialize() {
	// initialize Router
	router := gin.Default()

	// initialize routes
	initializeRoutes(router)

	// run server
	port := ":" + config.ServerConfig.Port
	err := router.Run(port)
	if err != nil {
		return
	}
}
