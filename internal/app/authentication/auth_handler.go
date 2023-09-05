package authentication

import (
	"errors"
	"github.com/gin-gonic/gin"
	"idp-automations-hub/internal/app/dto"
	"net/http"
	"strings"
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
	response := dto.TokenDto{
		AccessToken:  tokenDetails.AccessToken,
		RefreshToken: tokenDetails.RefreshToken,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Logout(c *gin.Context) {
	accessToken, err := ExtractTokenFromHeader(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.authService.Logout(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var tokenDto dto.TokenDto

	if err := c.ShouldBindJSON(&tokenDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	tokenDetails, err := h.authService.RefreshToken(tokenDto.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	response := dto.TokenDto{
		AccessToken: tokenDetails.AccessToken,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) IsUserAuthenticated(c *gin.Context) {
	var tokenDto dto.TokenDto
	if err := c.ShouldBindJSON(&tokenDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	isAuthenticated, err := h.authService.IsUserAuthenticated(tokenDto.AccessToken)
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

	resetToken, _, err := h.authService.RequestPasswordReset(email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resetToken": resetToken,
	})
}

func (h *Handler) ConfirmPasswordReset(c *gin.Context) {
	token := c.Query("reset-token")
	newPassword := c.PostForm("newPassword")

	err := h.authService.ConfirmPasswordReset(token, newPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	accessToken, err := ExtractTokenFromHeader(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newPassword := c.PostForm("newPassword")

	err = h.authService.ChangePassword(accessToken, newPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func ExtractTokenFromHeader(header string) (string, error) {
	splitted := strings.Split(header, " ")
	if len(splitted) != 2 {
		return "", errors.New("invalid or malformed auth token")
	}

	return splitted[1], nil
}
