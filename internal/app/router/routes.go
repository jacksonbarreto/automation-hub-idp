package router

import (
	"automation-hub-idp/docs"
	"automation-hub-idp/internal/app/authentication"
	"automation-hub-idp/internal/app/config"
	"automation-hub-idp/internal/app/users"
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
	authHandler := authentication.NewHandler(authService)
	authMiddleware := authentication.AuthMiddleware(authHandler)

	userService, err := users.GetDefaultUserService()
	if err != nil {
		return err
	}
	userHandler := users.NewHandler(userService)

	auth := apiVersion.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/logout", authMiddleware, authHandler.Logout)
		auth.POST("/request-password-reset", authHandler.RequestPasswordReset)
		auth.POST("/confirm-password-reset/:reset-token", authHandler.ConfirmPasswordReset)
		auth.GET("/is-user-authenticated", authHandler.IsUserAuthenticated)
	}

	user := apiVersion.Group("/user")
	{
		user.GET("/", authMiddleware, userHandler.GetCurrentUser)
		user.PATCH("/", authMiddleware, userHandler.Update)
	}
	return nil
}
