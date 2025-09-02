package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
)

type CancellationRuleHandler struct {
	cancellationRuleUseCase domain.CancellationRuleUseCase
}

func NewCancellationRuleHandler(
	cancellationRuleUseCase domain.CancellationRuleUseCase,
) *CancellationRuleHandler {
	return &CancellationRuleHandler{
		cancellationRuleUseCase: cancellationRuleUseCase,
	}
}

func (h *CancellationRuleHandler) RegisterRoutes(router *gin.RouterGroup) {
	cancellationRules := router.Group("/cancellation-rules")
	{
		cancellationRules.GET("", h.GetCancellationRules)
		cancellationRules.GET("/type/:type", h.GetCancellationRulesByType)
	}
}

// @Summary Получение всех активных правил отмены
// @Description Получает список всех активных правил отмены бронирования
// @Tags cancellation-rules
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /cancellation-rules [get]
func (h *CancellationRuleHandler) GetCancellationRules(c *gin.Context) {
	rules, err := h.cancellationRuleUseCase.GetActiveCancellationRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка получения правил отмены: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("правила отмены получены", rules))
}

// @Summary Получение правил отмены по типу
// @Description Получает список правил отмены определенного типа (general, refund, conditions)
// @Tags cancellation-rules
// @Accept json
// @Produce json
// @Param type path string true "Тип правил (general, refund, conditions)"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /cancellation-rules/type/{type} [get]
func (h *CancellationRuleHandler) GetCancellationRulesByType(c *gin.Context) {
	ruleTypeStr := c.Param("type")
	ruleType := domain.CancellationRuleType(ruleTypeStr)

	if ruleType != domain.CancellationRuleTypeGeneral &&
		ruleType != domain.CancellationRuleTypeRefund &&
		ruleType != domain.CancellationRuleTypeConditions {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("недопустимый тип правила. Доступные типы: general, refund, conditions"))
		return
	}

	rules, err := h.cancellationRuleUseCase.GetCancellationRulesByType(ruleType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка получения правил отмены: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("правила отмены получены", rules))
}
