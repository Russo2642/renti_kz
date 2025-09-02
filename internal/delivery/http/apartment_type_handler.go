package http

import (
	"net/http"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"

	"github.com/gin-gonic/gin"
)

type ApartmentTypeHandler struct {
	apartmentTypeUseCase domain.ApartmentTypeUseCase
}

func NewApartmentTypeHandler(apartmentTypeUseCase domain.ApartmentTypeUseCase) *ApartmentTypeHandler {
	return &ApartmentTypeHandler{
		apartmentTypeUseCase: apartmentTypeUseCase,
	}
}

func (h *ApartmentTypeHandler) RegisterRoutes(router *gin.RouterGroup) {
	apartmentTypes := router.Group("/apartment-types")
	{
		apartmentTypes.GET("", h.GetAll)
		apartmentTypes.GET("/:id", h.GetByID)
	}
}

func (h *ApartmentTypeHandler) RegisterAdminRoutes(router *gin.RouterGroup) {
	apartmentTypes := router.Group("/apartment-types")
	{
		apartmentTypes.POST("", h.Create)
		apartmentTypes.GET("", h.AdminGetAll)
		apartmentTypes.GET("/:id", h.AdminGetByID)
		apartmentTypes.PUT("/:id", h.Update)
		apartmentTypes.DELETE("/:id", h.Delete)
	}
}

// @Summary Получить все типы квартир
// @Description Возвращает список всех типов квартир (включая неактивные)
// @Tags apartment-types
// @Produce json
// @Success 200 {object} domain.SuccessResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartment-types [get]
func (h *ApartmentTypeHandler) GetAll(c *gin.Context) {
	apartmentTypes, err := h.apartmentTypeUseCase.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка получения типов квартир: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("типы квартир получены", apartmentTypes))
}

// @Summary Получить тип квартиры по ID
// @Description Возвращает информацию о типе квартиры по его ID
// @Tags apartment-types
// @Produce json
// @Param id path int true "ID типа квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartment-types/{id} [get]
func (h *ApartmentTypeHandler) GetByID(c *gin.Context) {
	id, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	apartmentType, err := h.apartmentTypeUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("тип квартиры не найден: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("тип квартиры получен", apartmentType))
}

// @Summary Создать новый тип квартиры (админ)
// @Description Создает новый тип квартиры (только для админов и модераторов)
// @Tags admin-apartment-types
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body domain.CreateApartmentTypeRequest true "Данные нового типа квартиры"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartment-types [post]
func (h *ApartmentTypeHandler) Create(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	var request domain.CreateApartmentTypeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return
	}

	apartmentType, err := h.apartmentTypeUseCase.Create(&request, userID)
	if err != nil {
		if err.Error() == "недостаточно прав для создания типов квартир" {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, domain.NewSuccessResponse("тип квартиры создан", apartmentType))
}

// @Summary Получить все типы квартир (админ)
// @Description Возвращает список всех типов квартир для админа
// @Tags admin-apartment-types
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartment-types [get]
func (h *ApartmentTypeHandler) AdminGetAll(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	_ = userID

	apartmentTypes, err := h.apartmentTypeUseCase.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка получения типов квартир: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("типы квартир получены", apartmentTypes))
}

// @Summary Получить тип квартиры по ID (админ)
// @Description Возвращает информацию о типе квартиры по его ID (для админа)
// @Tags admin-apartment-types
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID типа квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartment-types/{id} [get]
func (h *ApartmentTypeHandler) AdminGetByID(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	_ = userID

	id, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	apartmentType, err := h.apartmentTypeUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("тип квартиры не найден: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("тип квартиры получен", apartmentType))
}

// @Summary Обновить тип квартиры (админ)
// @Description Обновляет существующий тип квартиры (только для админов и модераторов)
// @Tags admin-apartment-types
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID типа квартиры"
// @Param request body domain.UpdateApartmentTypeRequest true "Данные для обновления"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartment-types/{id} [put]
func (h *ApartmentTypeHandler) Update(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	id, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var request domain.UpdateApartmentTypeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return
	}

	apartmentType, err := h.apartmentTypeUseCase.Update(id, &request, userID)
	if err != nil {
		if err.Error() == "недостаточно прав для обновления типов квартир" {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse(err.Error()))
			return
		}
		if err.Error() == "тип квартиры не найден" {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("тип квартиры обновлен", apartmentType))
}

// @Summary Удалить тип квартиры (админ)
// @Description Полностью удаляет тип квартиры (только для админов)
// @Tags admin-apartment-types
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID типа квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartment-types/{id} [delete]
func (h *ApartmentTypeHandler) Delete(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	id, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	err := h.apartmentTypeUseCase.Delete(id, userID)
	if err != nil {
		if err.Error() == "недостаточно прав для удаления типов квартир" {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse(err.Error()))
			return
		}
		if err.Error() == "тип квартиры не найден" {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("тип квартиры удален", nil))
}
