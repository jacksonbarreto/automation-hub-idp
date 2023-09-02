package controllers

import (
	"github.com/gin-gonic/gin"
	"idp-automations-hub/internal/app/dto"
	"idp-automations-hub/internal/app/services/iservice"
	"net/http"
)

type AuthController struct {
	authService iservice.AuthService
}

func NewAuthController(authService iservice.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (ac *AuthController) Register(c *gin.Context) {
	var userDTO dto.UserDTO
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := ac.authService.Register(userDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (ac *AuthController) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	tokenDetails, err := ac.authService.Login(email, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenDetails)
}

func (ac *AuthController) Logout(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")

	err := ac.authService.Logout(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (ac *AuthController) RefreshToken(c *gin.Context) {
	refreshToken := c.PostForm("refreshToken")

	tokenDetails, err := ac.authService.RefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenDetails)
}

func (ac *AuthController) IsUserAuthenticated(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")

	isAuthenticated, err := ac.authService.IsUserAuthenticated(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if isAuthenticated {
		c.JSON(http.StatusOK, gin.H{"authenticated": true})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
	}
}

func (ac *AuthController) RequestPasswordReset(c *gin.Context) {
	email := c.PostForm("email")

	resetToken, resetTokenExpires, err := ac.authService.RequestPasswordReset(email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resetToken":        resetToken,
		"resetTokenExpires": resetTokenExpires,
	})
}

func (ac *AuthController) ConfirmPasswordReset(c *gin.Context) {
	token := c.PostForm("token")
	newPassword := c.PostForm("newPassword")

	err := ac.authService.ConfirmPasswordReset(token, newPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (ac *AuthController) ChangePassword(c *gin.Context) {
	email := c.PostForm("email")
	newPassword := c.PostForm("newPassword")

	err := ac.authService.ChangePassword(email, newPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
