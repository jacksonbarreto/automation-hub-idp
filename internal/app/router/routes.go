package router

import (
	"github.com/gin-gonic/gin"
	"idp-automations-hub/internal/app/authentication"
	"idp-automations-hub/internal/app/config"
)

func initializeRoutes(router *gin.Engine) error {
	relativePathV1 := config.ServerConfig.BaseURL + "/v1"
	v1 := router.Group(relativePathV1)
	{
		// initialize auth routes
		err := initializeAuthRoutes(v1)
		if err != nil {
			return err
		}
	}

	return nil
}

func initializeAuthRoutes(apiVersion *gin.RouterGroup) error {
	authService, err := authentication.GetDefaultAuthService()
	if err != nil {
		return err
	}
	auth := apiVersion.Group("/auth")
	{
		// initialize auth routes
		authHandler := authentication.NewHandler(authService)
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/refresh-token", authHandler.RefreshToken)
		auth.POST("/request-password-reset", authHandler.RequestPasswordReset)
		auth.POST("/confirm-password-reset/:reset-token", authHandler.ConfirmPasswordReset)
		auth.POST("/change-password", authHandler.ChangePassword)
		auth.POST("/is-user-authenticated", authHandler.IsUserAuthenticated)
	}

	return nil
}
