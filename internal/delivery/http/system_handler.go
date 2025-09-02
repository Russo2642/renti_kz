package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/services"
)

type SystemHandler struct {
	userCacheService *services.UserCacheService
}

func NewSystemHandler(userCacheService *services.UserCacheService) *SystemHandler {
	return &SystemHandler{
		userCacheService: userCacheService,
	}
}

// @Summary Очистка кэша токенов
// @Description ЭКСТРЕННАЯ ОЧИСТКА: удаляет все токены из Redis кэша (только для админов)
// @Tags system
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} domain.SuccessResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/system/clear-token-cache [post]
func (h *SystemHandler) ClearTokenCache(c *gin.Context) {
	if h.userCacheService == nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("сервис кэша недоступен"))
		return
	}

	err := h.userCacheService.InvalidateAllTokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при очистке кэша токенов: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("кэш токенов успешно очищен", nil))
}

// @Summary Очистка конкретного токена
// @Description Удаляет конкретный токен из кэша (только для админов)
// @Tags system
// @Security ApiKeyAuth
// @Produce json
// @Param token query string true "Токен для удаления"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/system/invalidate-token [post]
func (h *SystemHandler) InvalidateToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("токен не указан"))
		return
	}

	if h.userCacheService == nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("сервис кэша недоступен"))
		return
	}

	err := h.userCacheService.InvalidateToken(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при инвалидации токена: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("токен успешно инвалидирован", nil))
}

func (h *SystemHandler) RegisterAdminRoutes(router *gin.RouterGroup) {
	systemRoutes := router.Group("/system")
	{
		systemRoutes.POST("/clear-token-cache", h.ClearTokenCache)
		systemRoutes.POST("/invalidate-token", h.InvalidateToken)
	}
}
