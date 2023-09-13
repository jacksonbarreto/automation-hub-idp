package authentication

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func AuthMiddleware(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, _ := c.Cookie("access_token")
		refreshToken, _ := c.Cookie("refresh_token")

		if isValid, _ := h.authService.IsUserAuthenticated(accessToken); isValid {
			c.Next()
			return
		}

		newAccessToken, err := h.authService.RefreshToken(refreshToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Please login again"})
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

		c.Next()
	}
}
