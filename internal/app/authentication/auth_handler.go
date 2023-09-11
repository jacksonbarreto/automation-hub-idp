package authentication

import (
	"automation-hub-idp/internal/app/dto"
	"errors"
	"github.com/gin-gonic/gin"
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

// Register
// @Summary Register a new user
// @Description Register a new user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body dto.UserDTO true "User object"
// @Success 200 {object} dto.UserDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var userDTO dto.UserDTO
	var errorResponse dto.ErrorResponse
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		errorResponse.Message = "Invalid request body"
		errorResponse.ErrorCode = http.StatusBadRequest
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	response, err := h.authService.Register(userDTO)
	if err != nil {
		errorResponse.Message = err.Error()
		errorResponse.ErrorCode = http.StatusInternalServerError
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Login
// @Summary Login
// @Description Login
// @Tags Authentication
// @Accept json
// @Produce json
// @Param email formData string true "Email"
// @Param password formData string true "Password"
// @Success 200 {object} dto.TokenDto
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var errorResponse dto.ErrorResponse
	email := c.PostForm("email")
	password := c.PostForm("password")

	tokenDetails, err := h.authService.Login(email, password)
	if err != nil {
		errorResponse.Message = err.Error()
		errorResponse.ErrorCode = http.StatusUnauthorized
		c.JSON(http.StatusUnauthorized, errorResponse)
		return
	}
	response := dto.TokenDto{
		AccessToken:  tokenDetails.AccessToken,
		RefreshToken: tokenDetails.RefreshToken,
	}

	c.JSON(http.StatusOK, response)
}

// Logout
// @Summary Logout
// @Description Logout
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization"
// @Success 200 {object} string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/logout [get]
func (h *Handler) Logout(c *gin.Context) {
	var errorResponse dto.ErrorResponse
	accessToken, err := ExtractTokenFromHeader(c.GetHeader("Authorization"))
	if err != nil {
		errorResponse.Message = err.Error()
		errorResponse.ErrorCode = http.StatusBadRequest
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	err = h.authService.Logout(accessToken)
	if err != nil {
		errorResponse.Message = err.Error()
		errorResponse.ErrorCode = http.StatusInternalServerError
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}
	response := dto.SuccessResponse{
		Message:    "Logged out successfully",
		StatusCode: http.StatusOK,
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken
// @Summary RefreshToken
// @Description RefreshToken
// @Tags Authentication
// @Accept json
// @Produce json
// @Param token body dto.TokenDto true "refreshToken"
// @Success 200 {object} string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/refresh-token [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var tokenDto dto.TokenDto
	var errorResponse dto.ErrorResponse

	if err := c.ShouldBindJSON(&tokenDto); err != nil {
		errorResponse.Message = "Invalid request body"
		errorResponse.ErrorCode = http.StatusBadRequest
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	tokenDetails, err := h.authService.RefreshToken(tokenDto.RefreshToken)
	if err != nil {
		errorResponse.Message = err.Error()
		errorResponse.ErrorCode = http.StatusUnauthorized
		c.JSON(http.StatusUnauthorized, errorResponse)
		return
	}
	response := dto.TokenDto{
		AccessToken: tokenDetails.AccessToken,
	}

	c.JSON(http.StatusOK, response)
}

// IsUserAuthenticated
// @Summary IsUserAuthenticated
// @Description IsUserAuthenticated
// @Tags Authentication
// @Accept json
// @Produce json
// @Param token body dto.TokenDto true "accesstoken"
// @Success 200 {object} dto.AuthenticateStatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/is-user-authenticated [post]
func (h *Handler) IsUserAuthenticated(c *gin.Context) {
	var tokenDto dto.TokenDto
	var errorResponse dto.ErrorResponse
	if err := c.ShouldBindJSON(&tokenDto); err != nil {
		errorResponse.Message = "Invalid request body"
		errorResponse.ErrorCode = http.StatusBadRequest
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	isAuthenticated, err := h.authService.IsUserAuthenticated(tokenDto.AccessToken)
	if err != nil {
		errorResponse.Message = err.Error()
		errorResponse.ErrorCode = http.StatusInternalServerError
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}
	authenticateStatusResponse := dto.AuthenticateStatusResponse{
		IsAuthenticated: isAuthenticated,
	}

	if isAuthenticated {
		authenticateStatusResponse.StatusCode = http.StatusOK
		c.JSON(http.StatusOK, authenticateStatusResponse)
	} else {
		authenticateStatusResponse.StatusCode = http.StatusUnauthorized
		c.JSON(http.StatusUnauthorized, authenticateStatusResponse)
	}
}

// RequestPasswordReset
// @Summary RequestPasswordReset
// @Description RequestPasswordReset
// @Tags Authentication
// @Accept json
// @Produce json
// @Param email formData string true "Email"
// @Success 200 {object} string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/request-password-reset [post]
func (h *Handler) RequestPasswordReset(c *gin.Context) {
	var errorResponse dto.ErrorResponse
	email := c.PostForm("email")

	_, _, err := h.authService.RequestPasswordReset(email)
	if err != nil {
		errorResponse.Message = err.Error()
		errorResponse.ErrorCode = http.StatusInternalServerError
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}
	response := dto.SuccessResponse{
		Message:    "Password reset token sent successfully",
		StatusCode: http.StatusOK,
	}
	c.JSON(http.StatusOK, response)
}

// ConfirmPasswordReset
// @Summary ConfirmPasswordReset
// @Description ConfirmPasswordReset
// @Tags Authentication
// @Accept json
// @Produce json
// @Param reset-token path string true "reset-token"
// @Param newPassword formData string true "newPassword"
// @Success 200 {object} string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/confirm-password-reset [post]
func (h *Handler) ConfirmPasswordReset(c *gin.Context) {
	var errorResponse dto.ErrorResponse
	token := c.Query("reset-token")
	newPassword := c.PostForm("newPassword")

	err := h.authService.ConfirmPasswordReset(token, newPassword)
	if err != nil {
		errorResponse.Message = err.Error()
		errorResponse.ErrorCode = http.StatusBadRequest
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}
	response := dto.SuccessResponse{
		Message:    "Password reset successfully",
		StatusCode: http.StatusOK,
	}
	c.JSON(http.StatusOK, response)
}

// ChangePassword
// @Summary ChangePassword
// @Description ChangePassword
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization"
// @Param newPassword formData string true "newPassword"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	var errorResponse dto.ErrorResponse
	accessToken, err := ExtractTokenFromHeader(c.GetHeader("Authorization"))
	if err != nil {
		errorResponse.Message = err.Error()
		errorResponse.ErrorCode = http.StatusBadRequest
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	newPassword := c.PostForm("newPassword")

	err = h.authService.ChangePassword(accessToken, newPassword)
	if err != nil {
		errorResponse.Message = err.Error()
		errorResponse.ErrorCode = http.StatusInternalServerError
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}
	response := dto.SuccessResponse{
		Message:    "Password changed successfully",
		StatusCode: http.StatusOK,
	}

	c.JSON(http.StatusOK, response)
}

func ExtractTokenFromHeader(header string) (string, error) {
	splitted := strings.Split(header, " ")
	if len(splitted) != 2 {
		return "", errors.New("invalid or malformed auth token")
	}

	return splitted[1], nil
}
