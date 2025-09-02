package http

import (
	"net/http"

	"github.com/russo2642/renti_kz/internal/domain"

	"github.com/gin-gonic/gin"
)

type PlatformSettingsHandler struct {
	settingsUseCase domain.PlatformSettingsUseCase
	middleware      *Middleware
}

func NewPlatformSettingsHandler(
	settingsUseCase domain.PlatformSettingsUseCase,
	middleware *Middleware,
) *PlatformSettingsHandler {
	return &PlatformSettingsHandler{
		settingsUseCase: settingsUseCase,
		middleware:      middleware,
	}
}

func (h *PlatformSettingsHandler) RegisterRoutes(router *gin.RouterGroup) {
	settings := router.Group("/settings")

	{
		settings.GET("/service-fee", h.GetServiceFeePercentage)
	}

	adminOnly := settings.Group("/", h.middleware.AuthMiddleware(), h.middleware.RoleMiddleware(domain.RoleAdmin))
	{
		adminOnly.GET("", h.GetAllSettings)
		adminOnly.GET("/:key", h.GetSettingByKey)
		adminOnly.POST("", h.CreateSetting)
		adminOnly.PUT("/:key", h.UpdateSetting)
		adminOnly.DELETE("/:key", h.DeleteSetting)
	}
}

type CreateSettingRequest struct {
	SettingKey   string `json:"setting_key" binding:"required"`
	SettingValue string `json:"setting_value" binding:"required"`
	Description  string `json:"description"`
	DataType     string `json:"data_type" binding:"required,oneof=string integer decimal boolean"`
	IsActive     *bool  `json:"is_active"`
}

type UpdateSettingRequest struct {
	SettingValue string `json:"setting_value" binding:"required"`
	Description  string `json:"description"`
	DataType     string `json:"data_type" binding:"required,oneof=string integer decimal boolean"`
	IsActive     *bool  `json:"is_active"`
}

// @Summary Получение процента сервисного сбора
// @Description Возвращает текущий процент сервисного сбора платформы
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /settings/service-fee [get]
func (h *PlatformSettingsHandler) GetServiceFeePercentage(c *gin.Context) {
	percentage, err := h.settingsUseCase.GetServiceFeePercentage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка получения настроек сервисного сбора"))
		return
	}

	response := gin.H{
		"service_fee_percentage": percentage,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("процент сервисного сбора получен", response))
}

// @Summary Получение всех настроек
// @Description Возвращает все настройки платформы (только для администраторов)
// @Tags settings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /settings [get]
func (h *PlatformSettingsHandler) GetAllSettings(c *gin.Context) {
	settings, err := h.settingsUseCase.GetAllActive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка получения настроек"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("настройки получены", settings))
}

// @Summary Получение настройки по ключу
// @Description Возвращает конкретную настройку по ключу (только для администраторов)
// @Tags settings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param key path string true "Ключ настройки"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /settings/{key} [get]
func (h *PlatformSettingsHandler) GetSettingByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("ключ настройки обязателен"))
		return
	}

	setting, err := h.settingsUseCase.GetByKey(key)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("настройка не найдена"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("настройка получена", setting))
}

// @Summary Создание новой настройки
// @Description Создает новую настройку платформы (только для администраторов)
// @Tags settings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body CreateSettingRequest true "Данные настройки"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /settings [post]
func (h *PlatformSettingsHandler) CreateSetting(c *gin.Context) {
	var req CreateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверные данные запроса: "+err.Error()))
		return
	}

	// Проверка на дублирующиеся ключи
	if _, err := h.settingsUseCase.GetByKey(req.SettingKey); err == nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("настройка с таким ключом уже существует"))
		return
	}

	setting := &domain.PlatformSetting{
		SettingKey:   req.SettingKey,
		SettingValue: req.SettingValue,
		Description:  req.Description,
		DataType:     req.DataType,
		IsActive:     true,
	}

	if req.IsActive != nil {
		setting.IsActive = *req.IsActive
	}

	err := h.settingsUseCase.Create(setting)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка создания настройки: "+err.Error()))
		return
	}

	c.JSON(http.StatusCreated, domain.NewSuccessResponse("настройка создана", setting))
}

// @Summary Обновление настройки
// @Description Обновляет существующую настройку платформы (только для администраторов)
// @Tags settings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param key path string true "Ключ настройки"
// @Param request body UpdateSettingRequest true "Данные для обновления"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /settings/{key} [put]
func (h *PlatformSettingsHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("ключ настройки обязателен"))
		return
	}

	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверные данные запроса: "+err.Error()))
		return
	}

	existingSetting, err := h.settingsUseCase.GetByKey(key)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("настройка не найдена"))
		return
	}

	existingSetting.SettingValue = req.SettingValue
	existingSetting.Description = req.Description
	existingSetting.DataType = req.DataType

	if req.IsActive != nil {
		existingSetting.IsActive = *req.IsActive
	}

	err = h.settingsUseCase.Update(existingSetting)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка обновления настройки: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("настройка обновлена", existingSetting))
}

// @Summary Удаление настройки
// @Description Удаляет настройку платформы (только для администраторов)
// @Tags settings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param key path string true "Ключ настройки"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /settings/{key} [delete]
func (h *PlatformSettingsHandler) DeleteSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("ключ настройки обязателен"))
		return
	}

	criticalSettings := []string{
		domain.SettingKeyServiceFeePercentage,
		domain.SettingKeyMinBookingDurationHours,
		domain.SettingKeyMaxBookingDurationHours,
	}

	for _, criticalKey := range criticalSettings {
		if key == criticalKey {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("критические настройки нельзя удалять"))
			return
		}
	}

	err := h.settingsUseCase.Delete(key)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("настройка не найдена"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("настройка удалена", nil))
}
