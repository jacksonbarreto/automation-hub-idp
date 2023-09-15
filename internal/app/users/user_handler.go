package users

import (
	"automation-hub-idp/internal/app/authentication"
	"automation-hub-idp/internal/app/dto"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	userService UserService
	authService authentication.IService
}

func NewHandler(userService UserService, authService authentication.IService) *Handler {
	return &Handler{
		userService: userService,
		authService: authService,
	}
}

func (h *Handler) Update(c *gin.Context) {
	var user dto.UserRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	accessToken, err := c.Cookie("access_token")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	// check if userRequest.password is not empty
	if user.Password != "" {
		err = h.authService.ChangePassword(accessToken, user.Password)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	userID, err := h.authService.GetIdFromToken(accessToken)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	userToUpdate, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	userToUpdate.Email = user.Email

	updatedUser, err := h.userService.UpdateUser(*userToUpdate)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	userResponse := dto.UserResponse{
		ID:    updatedUser.ID,
		Email: updatedUser.Email,
	}
	c.JSON(http.StatusOK, userResponse)
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	userID, err := h.authService.GetIdFromToken(accessToken)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	userResponse := dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
	}
	c.JSON(http.StatusOK, userResponse)
}
