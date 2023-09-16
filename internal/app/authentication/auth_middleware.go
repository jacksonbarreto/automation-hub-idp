package authentication

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func AuthMiddleware(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Please login again"})
			return
		}
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Please login again"})
			return
		}

		userID, err := h.authService.GetIdFromToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
			return
		}
		c.Set("userID", userID)

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
