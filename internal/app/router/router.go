package router

import (
	"automation-hub-idp/internal/app/config"
	"github.com/gin-gonic/gin"
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
