package authentication

import (
	"github.com/gin-gonic/gin"
	"idp-automations-hub/internal/app/dto"
	"net/http"
)

type Handler struct {
	authService IService
}

func NewHandler(authService IService) *Handler {
	return &Handler{
		authService: authService,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var userDTO dto.UserDTO
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.Register(userDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	tokenDetails, err := h.authService.Login(email, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenDetails)
}

func (h *Handler) Logout(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")

	err := h.authService.Logout(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *Handler) RefreshToken(c *gin.Context) {
	refreshToken := c.PostForm("refreshToken")

	tokenDetails, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenDetails)
}

func (h *Handler) IsUserAuthenticated(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")

	isAuthenticated, err := h.authService.IsUserAuthenticated(accessToken)
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

func (h *Handler) RequestPasswordReset(c *gin.Context) {
	email := c.PostForm("email")

	resetToken, resetTokenExpires, err := h.authService.RequestPasswordReset(email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resetToken":        resetToken,
		"resetTokenExpires": resetTokenExpires,
	})
}

func (h *Handler) ConfirmPasswordReset(c *gin.Context) {
	token := c.PostForm("token")
	newPassword := c.PostForm("newPassword")

	err := h.authService.ConfirmPasswordReset(token, newPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	email := c.PostForm("email")
	newPassword := c.PostForm("newPassword")

	err := h.authService.ChangePassword(email, newPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
