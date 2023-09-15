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
