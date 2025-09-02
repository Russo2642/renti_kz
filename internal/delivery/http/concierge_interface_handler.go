package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type ConciergeInterfaceHandler struct {
	conciergeUseCase domain.ConciergeUseCase
	apartmentUseCase domain.ApartmentUseCase
	bookingUseCase   domain.BookingUseCase
	chatUseCase      domain.ChatUseCase
}

func NewConciergeInterfaceHandler(
	conciergeUseCase domain.ConciergeUseCase,
	apartmentUseCase domain.ApartmentUseCase,
	bookingUseCase domain.BookingUseCase,
	chatUseCase domain.ChatUseCase,
) *ConciergeInterfaceHandler {
	return &ConciergeInterfaceHandler{
		conciergeUseCase: conciergeUseCase,
		apartmentUseCase: apartmentUseCase,
		bookingUseCase:   bookingUseCase,
		chatUseCase:      chatUseCase,
	}
}

func (h *ConciergeInterfaceHandler) RegisterRoutes(router *gin.RouterGroup) {
	concierge := router.Group("")
	{
		concierge.GET("/profile", h.GetProfile)
		concierge.GET("/apartments", h.GetApartments)
		concierge.GET("/bookings", h.GetBookings)
		concierge.GET("/chat/rooms", h.GetChatRooms)
		concierge.GET("/stats", h.GetStats)
		concierge.PUT("/schedule", h.UpdateSchedule)
	}
}

// @Summary Получение профиля консьержа
// @Description Получает профиль текущего консьержа
// @Tags concierge-interface
// @Produce json
// @Success 200 {object} domain.SuccessResponse{data=domain.Concierge} "Профиль консьержа"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Failure 404 {object} domain.ErrorResponse "Консьерж не найден"
// @Security ApiKeyAuth
// @Router /concierge/profile [get]
func (h *ConciergeInterfaceHandler) GetProfile(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	concierge, err := h.conciergeUseCase.GetConciergeByUserID(userID)
	if err != nil {
		c.JSON(404, domain.NewErrorResponse("Concierge profile not found"))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Profile retrieved successfully", concierge))
}

// @Summary Получение квартир консьержа
// @Description Получает список квартир, закрепленных за консьержем
// @Tags concierge-interface
// @Produce json
// @Success 200 {object} domain.SuccessResponse{data=[]domain.Apartment} "Квартиры консьержа"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Security ApiKeyAuth
// @Router /concierge/apartments [get]
func (h *ConciergeInterfaceHandler) GetApartments(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	apartments, err := h.conciergeUseCase.GetConciergeApartments(userID)
	if err != nil {
		c.JSON(500, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Apartments retrieved successfully", apartments))
}

// @Summary Получение броней консьержа
// @Description Получает список бронирований для квартир консьержа
// @Tags concierge-interface
// @Produce json
// @Param status query string false "Фильтр по статусу"
// @Param active query boolean false "Только активные брони"
// @Param page query integer false "Номер страницы" default(1)
// @Param page_size query integer false "Размер страницы" default(20)
// @Success 200 {object} object{success=boolean,data=object{bookings=[]domain.Booking,total=integer}} "Брони консьержа"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Security ApiKeyAuth
// @Router /concierge/bookings [get]
func (h *ConciergeInterfaceHandler) GetBookings(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	// Параметры фильтрации
	filters := make(map[string]interface{})

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	if activeStr := c.Query("active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			filters["active"] = active
		}
	}

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	if startDateFrom := c.Query("start_date_from"); startDateFrom != "" {
		filters["start_date_from"] = startDateFrom
	}

	if startDateTo := c.Query("start_date_to"); startDateTo != "" {
		filters["start_date_to"] = startDateTo
	}

	page, pageSize := utils.ParsePagination(c)

	bookings, total, err := h.conciergeUseCase.GetConciergeBookings(userID, filters, page, pageSize)
	if err != nil {
		c.JSON(500, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Bookings retrieved successfully", gin.H{
		"bookings": bookings,
		"total":    total,
	}))
}

// @Summary Получение чатов консьержа
// @Description Получает список чат-комнат для консьержа
// @Tags concierge-interface
// @Produce json
// @Param status query string false "Фильтр по статусу"
// @Param page query integer false "Номер страницы" default(1)
// @Param page_size query integer false "Размер страницы" default(20)
// @Success 200 {object} object{success=boolean,data=object{rooms=[]domain.ChatRoom,total=integer}} "Чаты консьержа"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Security ApiKeyAuth
// @Router /concierge/chat/rooms [get]
func (h *ConciergeInterfaceHandler) GetChatRooms(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	// Получаем ID консьержа по userID
	concierge, err := h.conciergeUseCase.GetConciergeByUserID(userID)
	if err != nil {
		c.JSON(404, domain.NewErrorResponse("Concierge not found"))
		return
	}

	// Параметры фильтрации
	var statusFilter []domain.ChatRoomStatus
	if statusStr := c.Query("status"); statusStr != "" {
		statusFilter = append(statusFilter, domain.ChatRoomStatus(statusStr))
	}

	page, pageSize := utils.ParsePagination(c)

	rooms, total, err := h.chatUseCase.GetConciergeChatRooms(concierge.ID, statusFilter, page, pageSize)
	if err != nil {
		c.JSON(500, domain.NewErrorResponse(err.Error()))
		return
	}

	// Автоматически активируем pending чаты для консьержа
	for _, room := range rooms {
		if room.Status == domain.ChatRoomStatusPending {
			err := h.chatUseCase.ActivateChat(room.ID, userID)
			if err == nil {
				room.Status = domain.ChatRoomStatusActive // Обновляем статус в ответе
			}
		}
	}

	c.JSON(200, domain.NewSuccessResponse("Chat rooms retrieved successfully", gin.H{
		"rooms": rooms,
		"total": total,
	}))
}

// @Summary Получение статистики консьержа
// @Description Получает статистику работы консьержа
// @Tags concierge-interface
// @Produce json
// @Success 200 {object} domain.SuccessResponse{data=object} "Статистика консьержа"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Security ApiKeyAuth
// @Router /concierge/stats [get]
func (h *ConciergeInterfaceHandler) GetStats(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	stats, err := h.conciergeUseCase.GetConciergeStats(userID)
	if err != nil {
		c.JSON(500, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Statistics retrieved successfully", stats))
}

// @Summary Обновление расписания консьержа
// @Description Обновляет рабочее расписание консьержа
// @Tags concierge-interface
// @Accept json
// @Produce json
// @Param request body domain.ConciergeSchedule true "Новое расписание"
// @Success 200 {object} domain.SuccessResponse "Расписание обновлено"
// @Failure 400 {object} domain.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Недостаточно прав"
// @Security ApiKeyAuth
// @Router /concierge/schedule [put]
func (h *ConciergeInterfaceHandler) UpdateSchedule(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	var schedule domain.ConciergeSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(400, domain.NewErrorResponse("Invalid request body"))
		return
	}

	err := h.conciergeUseCase.UpdateConciergeSchedule(userID, &schedule)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Schedule updated successfully", nil))
}
