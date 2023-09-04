package router

import (
	"github.com/gin-gonic/gin"
	"idp-automations-hub/internal/app/config"
)

func Initialize() error {
	// initialize Router
	router := gin.Default()

	// initialize routes
	err := initializeRoutes(router)
	if err != nil {
		return err
	}

	// run server
	port := ":" + config.ServerConfig.Port
	err = router.Run(port)
	if err != nil {
		return err
	}

	return nil
}
