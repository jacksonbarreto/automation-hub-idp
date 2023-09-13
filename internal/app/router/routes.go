package router

import (
	"automation-hub-idp/docs"
	"automation-hub-idp/internal/app/authentication"
	"automation-hub-idp/internal/app/config"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func initializeRoutes(router *gin.Engine) error {
	relativePathV1 := config.ServerConfig.BaseURL + "/v1"
	docs.SwaggerInfo.BasePath = relativePathV1
	v1 := router.Group(relativePathV1)
	{
		// initialize auth routes
		err := initializeAuthRoutes(v1)
		if err != nil {
			return err
		}
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
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
		auth.GET("/logout", authentication.AuthMiddleware, authHandler.Logout)
		auth.POST("/request-password-reset", authHandler.RequestPasswordReset)
		auth.POST("/confirm-password-reset/:reset-token", authHandler.ConfirmPasswordReset)
		auth.POST("/change-password", authentication.AuthMiddleware, authHandler.ChangePassword)
		auth.GET("/is-user-authenticated", authHandler.IsUserAuthenticated)
	}

	return nil
}
