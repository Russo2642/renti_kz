package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
)

func ExtractBearerToken(c *gin.Context) (string, bool) {
	header := c.GetHeader("Authorization")
	if header != "" {
		headerParts := strings.Split(header, " ")
		if len(headerParts) == 2 && headerParts[0] == "Bearer" {
			return headerParts[1], true
		}
	}

	token := c.Query("token")
	if token != "" {
		return token, true
	}

	c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("необходим заголовок авторизации или токен в query параметре"))
	return "", false
}

func TryExtractBearerToken(c *gin.Context) (string, bool) {
	header := c.GetHeader("Authorization")
	if header != "" {
		headerParts := strings.Split(header, " ")
		if len(headerParts) == 2 && headerParts[0] == "Bearer" {
			return headerParts[1], true
		}
	}

	token := c.Query("token")
	if token != "" {
		return token, true
	}

	return "", false
}

func SetUserContext(c *gin.Context, userID int, userRole domain.UserRole) {
	c.Set("user_id", userID)
	c.Set("user_role", userRole)
}

func CheckUserRole(c *gin.Context, allowedRoles ...domain.UserRole) bool {
	role, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return false
	}

	userRole := role.(domain.UserRole)
	for _, allowedRole := range allowedRoles {
		if userRole == allowedRole {
			return true
		}
	}

	c.JSON(http.StatusForbidden, domain.NewErrorResponse("доступ запрещен"))
	return false
}

func AbortWithUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(message))
	c.Abort()
}

func AbortWithForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, domain.NewErrorResponse(message))
	c.Abort()
}

func SetCORSHeaders(c *gin.Context) {
	origin := c.GetHeader("Origin")

	allowedOrigins := map[string]bool{
		"http://localhost:5173": true,
		"http://localhost:3000": true,
		"http://localhost:8080": true,
		"https://renti.kz":      true,
		"https://www.renti.kz":  true,
	}

	if allowedOrigins[origin] {
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	} else if origin == "" {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	}

	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
	c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
}
