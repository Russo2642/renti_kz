package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type ConciergeHandler struct {
	conciergeUseCase domain.ConciergeUseCase
}

func NewConciergeHandler(
	conciergeUseCase domain.ConciergeUseCase,
) *ConciergeHandler {
	return &ConciergeHandler{
		conciergeUseCase: conciergeUseCase,
	}
}

func (h *ConciergeHandler) RegisterRoutes(router *gin.RouterGroup) {
	concierges := router.Group("/concierges")
	{
		concierges.POST("", h.CreateConcierge)
		concierges.GET("", h.GetAllConcierges)
		concierges.GET("/:id", h.GetConciergeByID)
		concierges.PUT("/:id", h.UpdateConcierge)
		concierges.DELETE("/:id", h.DeleteConcierge)

		concierges.GET("/apartment/:apartmentId", h.GetConciergesByApartment)
		concierges.POST("/assign", h.AssignConciergeToApartment)
		concierges.DELETE("/:id/apartments/:apartmentId/remove", h.RemoveConciergeFromApartment)

		concierges.GET("/owner/:ownerId", h.GetConciergesByOwner)

		concierges.GET("/user/:userId/status", h.IsUserConcierge)
	}
}

// @Summary Создание консьержа
// @Description Создает нового консьержа и привязывает его к квартире (только для админов)
// @Tags concierge
// @Accept json
// @Produce json
// @Param request body domain.CreateConciergeRequest true "Данные для создания консьержа"
// @Success 201 {object} domain.SuccessResponse{data=domain.Concierge} "Консьерж успешно создан"
// @Failure 400 {object} domain.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Security ApiKeyAuth
// @Router /admin/concierges [post]
func (h *ConciergeHandler) CreateConcierge(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	var request domain.CreateConciergeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, domain.NewErrorResponse("Invalid request body"))
		return
	}

	if request.UserID == 0 || len(request.ApartmentIDs) == 0 {
		c.JSON(400, domain.NewErrorResponse("User ID and at least one Apartment ID are required"))
		return
	}

	concierge, err := h.conciergeUseCase.CreateConcierge(&request, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(201, domain.NewSuccessResponse("Concierge created successfully", concierge))
}

// @Summary Список консьержей
// @Description Получает список всех консьержей с фильтрацией и пагинацией (только для админов)
// @Tags concierge
// @Produce json
// @Param is_active query boolean false "Фильтр по активности"
// @Param apartment_id query integer false "Фильтр по ID квартиры"
// @Param page query integer false "Номер страницы" default(1)
// @Param page_size query integer false "Размер страницы" default(20)
// @Success 200 {object} object{success=boolean,data=[]domain.Concierge,pagination=object} "Список консьержей"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Failure 500 {object} domain.ErrorResponse "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /admin/concierges [get]
func (h *ConciergeHandler) GetAllConcierges(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	filters := make(map[string]interface{})

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filters["is_active"] = isActive
		}
	}

	if apartmentIDStr := c.Query("apartment_id"); apartmentIDStr != "" {
		if apartmentID, err := strconv.Atoi(apartmentIDStr); err == nil {
			filters["apartment_id"] = apartmentID
		}
	}

	page, pageSize := utils.ParsePagination(c)

	concierges, total, err := h.conciergeUseCase.GetAllConcierges(filters, page, pageSize)
	if err != nil {
		c.JSON(500, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    concierges,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})

	_ = userID
}

// @Summary Получение консьержа по ID
// @Description Получает информацию о консьерже по его идентификатору (только для админов)
// @Tags concierge
// @Produce json
// @Param id path integer true "ID консьержа"
// @Success 200 {object} domain.SuccessResponse{data=domain.Concierge} "Консьерж найден"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Failure 404 {object} domain.ErrorResponse "Консьерж не найден"
// @Security ApiKeyAuth
// @Router /admin/concierges/{id} [get]
func (h *ConciergeHandler) GetConciergeByID(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	conciergeID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	concierge, err := h.conciergeUseCase.GetConciergeByID(conciergeID)
	if err != nil {
		c.JSON(404, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Concierge retrieved successfully", concierge))

	_ = userID
}

// @Summary Получение консьержа по квартире
// @Description Получает консьержа, назначенного на определенную квартиру (только для админов)
// @Tags concierge
// @Produce json
// @Param apartmentId path integer true "ID квартиры"
// @Success 200 {object} domain.SuccessResponse{data=domain.Concierge} "Консьерж найден"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Failure 404 {object} domain.ErrorResponse "Консьерж не найден"
// @Security ApiKeyAuth
// @Router /admin/concierges/apartment/{apartmentId} [get]
func (h *ConciergeHandler) GetConciergesByApartment(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	apartmentID, ok := utils.ParseIDParam(c, "apartmentId")
	if !ok {
		return
	}

	concierges, err := h.conciergeUseCase.GetConciergesByApartment(apartmentID)
	if err != nil {
		c.JSON(404, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Concierges retrieved successfully", concierges))

	_ = userID
}

// @Summary Обновление консьержа
// @Description Обновляет информацию о консьерже (только для админов)
// @Tags concierge
// @Accept json
// @Produce json
// @Param id path integer true "ID консьержа"
// @Param request body domain.UpdateConciergeRequest true "Данные для обновления"
// @Success 200 {object} domain.SuccessResponse "Консьерж обновлен"
// @Failure 400 {object} domain.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Failure 404 {object} domain.ErrorResponse "Консьерж не найден"
// @Security ApiKeyAuth
// @Router /admin/concierges/{id} [put]
func (h *ConciergeHandler) UpdateConcierge(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	conciergeID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var request domain.UpdateConciergeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, domain.NewErrorResponse("Invalid request body"))
		return
	}

	err := h.conciergeUseCase.UpdateConcierge(conciergeID, &request, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Concierge updated successfully", nil))
}

// @Summary Удаление консьержа
// @Description Удаляет консьержа (только для админов)
// @Tags concierge
// @Produce json
// @Param id path integer true "ID консьержа"
// @Success 200 {object} domain.SuccessResponse "Консьерж удален"
// @Failure 400 {object} domain.ErrorResponse "Ошибка при удалении"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Failure 404 {object} domain.ErrorResponse "Консьерж не найден"
// @Security ApiKeyAuth
// @Router /admin/concierges/{id} [delete]
func (h *ConciergeHandler) DeleteConcierge(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	conciergeID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	err := h.conciergeUseCase.DeleteConcierge(conciergeID, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Concierge deleted successfully", nil))
}

// @Summary Назначение консьержа на квартиру
// @Description Назначает консьержа для обслуживания определенной квартиры (только для админов)
// @Tags concierge
// @Accept json
// @Produce json
// @Param request body object{user_id=integer,apartment_id=integer} true "ID пользователя и квартиры"
// @Success 201 {object} domain.SuccessResponse{data=domain.Concierge} "Консьерж назначен"
// @Failure 400 {object} domain.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Security ApiKeyAuth
// @Router /admin/concierges/assign [post]
func (h *ConciergeHandler) AssignConciergeToApartment(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	var request domain.AssignConciergeToApartmentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, domain.NewErrorResponse("Invalid request body"))
		return
	}

	if request.ConciergeID == 0 || request.ApartmentID == 0 {
		c.JSON(400, domain.NewErrorResponse("Concierge ID and Apartment ID are required"))
		return
	}

	err := h.conciergeUseCase.AssignConciergeToApartment(request.ConciergeID, request.ApartmentID, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(201, domain.NewSuccessResponse("Concierge assigned to apartment successfully", nil))
}

// @Summary Отзыв консьержа с квартиры
// @Description Отзывает консьержа с обслуживания квартиры (только для админов)
// @Tags concierge
// @Produce json
// @Param apartmentId path integer true "ID квартиры"
// @Success 200 {object} domain.SuccessResponse "Консьерж отозван"
// @Failure 400 {object} domain.ErrorResponse "Ошибка при отзыве"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Failure 404 {object} domain.ErrorResponse "Консьерж не найден"
// @Security ApiKeyAuth
// @Router /admin/concierges/{id}/apartments/{apartmentId}/remove [delete]
func (h *ConciergeHandler) RemoveConciergeFromApartment(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	conciergeID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	apartmentID, ok := utils.ParseIDParam(c, "apartmentId")
	if !ok {
		return
	}

	err := h.conciergeUseCase.RemoveConciergeFromApartment(conciergeID, apartmentID, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Concierge removed from apartment successfully", nil))
}

// @Summary Получение консьержей владельца
// @Description Получает список консьержей, назначенных на квартиры определенного владельца (только для админов)
// @Tags concierge
// @Produce json
// @Param ownerId path integer true "ID владельца"
// @Success 200 {object} domain.SuccessResponse{data=[]domain.Concierge} "Список консьержей"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Failure 500 {object} domain.ErrorResponse "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /admin/concierges/owner/{ownerId} [get]
func (h *ConciergeHandler) GetConciergesByOwner(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	ownerID, ok := utils.ParseIDParam(c, "ownerId")
	if !ok {
		return
	}

	concierges, err := h.conciergeUseCase.GetConciergesByOwner(ownerID)
	if err != nil {
		c.JSON(500, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Concierges retrieved successfully", concierges))

	_ = userID
}

// @Summary Проверка статуса консьержа
// @Description Проверяет, является ли пользователь консьержем (только для админов)
// @Tags concierge
// @Produce json
// @Param userId path integer true "ID пользователя"
// @Success 200 {object} domain.SuccessResponse{data=object{is_concierge=boolean,user_id=integer}} "Статус пользователя"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Failure 500 {object} domain.ErrorResponse "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /admin/concierges/user/{userId}/status [get]
func (h *ConciergeHandler) IsUserConcierge(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	checkUserID, ok := utils.ParseIDParam(c, "userId")
	if !ok {
		return
	}

	isConcierge, err := h.conciergeUseCase.IsUserConcierge(checkUserID)
	if err != nil {
		c.JSON(500, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Status retrieved successfully", gin.H{
		"is_concierge": isConcierge,
		"user_id":      checkUserID,
	}))

	_ = userID
}
