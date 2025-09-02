package utils

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
)

func GetUserIDFromContext(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(int), true
}

func GetUserRoleFromContext(c *gin.Context) (domain.UserRole, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	return role.(domain.UserRole), true
}

func RequireAuth(c *gin.Context) (int, bool) {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return 0, false
	}
	return userID, true
}

func RequireRole(c *gin.Context, role domain.UserRole) (int, bool) {
	userID, ok := RequireAuth(c)
	if !ok {
		return 0, false
	}

	userRole, exists := GetUserRoleFromContext(c)
	if !exists || userRole != role {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав"))
		return 0, false
	}

	return userID, true
}

func RequireAnyRole(c *gin.Context, roles ...domain.UserRole) (int, domain.UserRole, bool) {
	userID, ok := RequireAuth(c)
	if !ok {
		return 0, "", false
	}

	userRole, exists := GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав"))
		return 0, "", false
	}

	for _, role := range roles {
		if userRole == role {
			return userID, userRole, true
		}
	}

	c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав"))
	return 0, "", false
}

func ParseIDParam(c *gin.Context, paramName string) (int, bool) {
	idParam := c.Param(paramName)
	if idParam == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("отсутствует обязательный параметр: "+paramName))
		return 0, false
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный "+paramName))
		return 0, false
	}

	return id, true
}

func ParsePagination(c *gin.Context) (page, pageSize int) {
	pageParam := c.DefaultQuery("page", "1")
	pageSizeParam := c.DefaultQuery("page_size", "10")

	var err error
	page, err = strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err = strconv.Atoi(pageSizeParam)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return page, pageSize
}
