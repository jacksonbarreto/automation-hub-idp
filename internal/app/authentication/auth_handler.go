package authentication

import (
	"automation-hub-idp/internal/app/dto"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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
// @Accept application/json
// @Param body dto.UserLoginDTO
// @Param password formData string true "Password"
// @Success 200 "Successfully logged in"
// @Failure 400 "Unauthorized"
// @Failure 500 "Internal Server Error"
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var userLoginDTO dto.UserLoginDTO
	if err := c.ShouldBindJSON(&userLoginDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenDetails, err := h.authService.Login(userLoginDTO.Email, userLoginDTO.Password)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	atExpiresTime := time.Unix(tokenDetails.AtExpires, 0)
	rtExpiresTime := time.Unix(tokenDetails.RtExpires, 0)

	// Set the access token as a cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    tokenDetails.AccessToken,
		Expires:  atExpiresTime,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	// Set the refresh token as a cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokenDetails.RefreshToken,
		Expires:  rtExpiresTime,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	c.Status(http.StatusOK)
}

// Logout
// @Summary Logout
// @Description Logout
// @Tags Authentication
// @Success 200 "OK"
// @Failure 400 "Unauthorized"
// @Failure 500 "Internal Server Error"
// @Router /auth/logout [get]
func (h *Handler) Logout(c *gin.Context) {
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	err = h.authService.Logout(accessToken)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// IsUserAuthenticated
// @Summary IsUserAuthenticated
// @Description IsUserAuthenticated
// @Tags Authentication
// @Success 200 "OK"
// @Failure 400 "Unauthorized"
// @Failure 500 "Internal Server Error"
// @Router /auth/is-user-authenticated [get]
func (h *Handler) IsUserAuthenticated(c *gin.Context) {
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	isAuthenticated, err := h.authService.IsUserAuthenticated(accessToken)
	if err != nil || !isAuthenticated {
		// If the access token is not valid, try to refresh it
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.Status(http.StatusUnauthorized)
			return
		}

		newAccessToken, err := h.authService.RefreshToken(refreshToken)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			return
		}

		atExpiresTime := time.Unix(newAccessToken.AtExpires, 0)

		// Set the new access token as a cookie
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "access_token",
			Value:    newAccessToken.AccessToken,
			Expires:  atExpiresTime,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
		})

		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusOK)
}

// RequestPasswordReset
// @Summary RequestPasswordReset
// @Description RequestPasswordReset
// @Tags Authentication
// @Accept application/x-www-form-urlencoded
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
// @Accept application/x-www-form-urlencoded
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
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param newPassword formData string true "newPassword"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	var errorResponse dto.ErrorResponse
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
