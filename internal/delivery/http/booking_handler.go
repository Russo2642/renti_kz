package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/services"
	"github.com/russo2642/renti_kz/internal/utils"
)

type BookingHandler struct {
	bookingUseCase       domain.BookingUseCase
	userUseCase          domain.UserUseCase
	lockUseCase          domain.LockUseCase
	responseCacheService *services.ResponseCacheService
}

func NewBookingHandler(bookingUseCase domain.BookingUseCase, userUseCase domain.UserUseCase, lockUseCase domain.LockUseCase, responseCacheService *services.ResponseCacheService) *BookingHandler {
	return &BookingHandler{
		bookingUseCase:       bookingUseCase,
		userUseCase:          userUseCase,
		lockUseCase:          lockUseCase,
		responseCacheService: responseCacheService,
	}
}

func (h *BookingHandler) RegisterRoutes(router *gin.RouterGroup) {

	bookings := router.Group("/bookings")
	bookings.Use()
	{

		bookings.POST("", h.CreateBooking)
		bookings.POST("/:id/confirm", h.ConfirmBooking)
		bookings.POST("/:id/payment", h.ProcessPayment)
		bookings.GET("/:id", h.GetBookingByID)
		bookings.GET("/number/:number", h.GetBookingByNumber)

		bookings.POST("/:id/approve", h.ApproveBooking)
		bookings.POST("/:id/reject", h.RejectBooking)
		bookings.POST("/:id/cancel", h.CancelBooking)
		bookings.GET("/:id/receipt", h.GetPaymentReceipt)
		bookings.POST("/:id/finish", h.FinishSession)
		bookings.POST("/:id/extend", h.ExtendBooking)

		bookings.GET("/:id/extensions", h.GetBookingExtensions)
		bookings.GET("/:id/available-extensions", h.GetAvailableExtensions)
		bookings.POST("/:id/extensions/:extensionId/payment", h.ProcessExtensionPayment)
		bookings.POST("/:id/extensions/:extensionId/approve", h.ApproveExtension)
		bookings.POST("/:id/extensions/:extensionId/reject", h.RejectExtension)

		bookings.GET("/my/lock-access", h.GetMyBookingsLockAccess)
		bookings.GET("/:id/lock-access", h.GetBookingLockAccess)
		bookings.POST("/:id/generate-password", h.GeneratePasswordForBooking)

	}

}

func (h *BookingHandler) RegisterAdminRoutes(router *gin.RouterGroup) {
	router.GET("/bookings", h.AdminGetAllBookings)
	router.GET("/bookings/:id", h.AdminGetBookingByID)
	router.PUT("/bookings/:id/status", h.AdminUpdateBookingStatus)
	router.DELETE("/bookings/:id", h.AdminCancelBooking)
	router.GET("/bookings/statistics", CacheMiddlewareWithTTL(h.responseCacheService, 10*time.Minute), h.AdminGetBookingStatistics)
}

// @Summary Создание бронирования
// @Description Создает новое бронирование квартиры
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body domain.CreateBookingRequest true "Данные бронирования"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {

	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	var request domain.CreateBookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return
	}

	if err := utils.ValidateCreateBookingRequest(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	booking, err := h.bookingUseCase.CreateBooking(userID, &request)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	response := utils.ConvertBookingToResponse(booking)

	c.JSON(http.StatusCreated, domain.NewSuccessResponse("бронирование успешно создано", response))
}

// @Summary Подтверждение бронирования
// @Description Подтверждает бронирование со статусом created и переводит его в pending
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Param request body domain.ConfirmBookingRequest true "Данные подтверждения"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/confirm [post]
func (h *BookingHandler) ConfirmBooking(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var request domain.ConfirmBookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return
	}

	booking, err := h.bookingUseCase.ConfirmBooking(bookingID, userID, &request)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	response := utils.ConvertBookingToResponse(booking)

	c.JSON(http.StatusOK, domain.NewSuccessResponse("бронирование успешно подтверждено", response))
}

// @Summary Обработать оплату бронирования
// @Description Обрабатывает оплату для бронирования со статусом awaiting_payment. Можно передать либо payment_id, либо order_id
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Param request body domain.ProcessPaymentRequest true "Данные платежа (payment_id или order_id)"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /bookings/{id}/payment [post]
func (h *BookingHandler) ProcessPayment(c *gin.Context) {
	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var request domain.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return
	}

	if request.PaymentID == "" && request.OrderID == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("необходимо указать либо payment_id, либо order_id"))
		return
	}

	if request.PaymentID != "" && request.OrderID != "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("нельзя указывать одновременно payment_id и order_id"))
		return
	}

	var booking *domain.Booking
	var err error

	if request.OrderID != "" {
		booking, err = h.bookingUseCase.ProcessPaymentWithOrder(bookingID, request.OrderID)
	} else {
		booking, err = h.bookingUseCase.ProcessPayment(bookingID, request.PaymentID)
	}

	if err != nil {
		if strings.Contains(err.Error(), "уже использован") || strings.Contains(err.Error(), "уже привязан") {
			c.JSON(http.StatusConflict, domain.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	response := utils.ConvertBookingToResponse(booking)

	c.JSON(http.StatusOK, domain.NewSuccessResponse("оплата успешно обработана", response))
}

// @Summary Получить бронирование по ID
// @Description Получает информацию о бронировании по его идентификатору
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id} [get]
func (h *BookingHandler) GetBookingByID(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	canAccess, err := h.bookingUseCase.CanUserAccessBooking(bookingID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке прав доступа"))
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("нет прав для просмотра этого бронирования"))
		return
	}

	booking, err := h.bookingUseCase.GetBookingByID(bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("бронирование не найдено"))
		return
	}

	response := utils.ConvertBookingToResponse(booking)

	c.JSON(http.StatusOK, domain.NewSuccessResponse("бронирование получено успешно", response))
}

// @Summary Получить бронирование по номеру
// @Description Получает информацию о бронировании по его номеру
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param number path string true "Номер бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/number/{number} [get]
func (h *BookingHandler) GetBookingByNumber(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingNumber := c.Param("number")
	if bookingNumber == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("номер бронирования обязателен"))
		return
	}

	booking, err := h.bookingUseCase.GetBookingByNumber(bookingNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("бронирование не найдено"))
		return
	}

	canAccess, err := h.bookingUseCase.CanUserAccessBooking(booking.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке прав доступа"))
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("нет прав для просмотра этого бронирования"))
		return
	}

	response := utils.ConvertBookingToResponse(booking)

	c.JSON(http.StatusOK, domain.NewSuccessResponse("бронирование получено успешно", response))
}

// @Summary Подтвердить бронирование
// @Description Подтверждает бронирование (только владелец или админ)
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/approve [post]
func (h *BookingHandler) ApproveBooking(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	err := h.bookingUseCase.ApproveBooking(bookingID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("бронирование успешно подтверждено", nil))
}

// @Summary Отклонить бронирование
// @Description Отклоняет бронирование (только владелец или админ)
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Param request body domain.RejectBookingRequest true "Комментарий к отклонению"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/reject [post]
func (h *BookingHandler) RejectBooking(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var request domain.RejectBookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return
	}

	err := h.bookingUseCase.RejectBooking(bookingID, userID, request.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("бронирование отклонено", nil))
}

// @Summary Отменить бронирование
// @Description Отменяет бронирование (для арендатора). Возврат платежа зависит от статуса бронирования и времени до начала: если бронь approved/pending и до начала менее 6 часов - деньги не возвращаются
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Param request body domain.CancelBookingRequest true "Причина отмены"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/cancel [post]
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var request domain.CancelBookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return
	}

	err := h.bookingUseCase.CancelBooking(bookingID, userID, request.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("бронирование отменено", nil))
}

// @Summary Получить чек об оплате
// @Description Возвращает чек об оплате бронирования в JSON формате
// @Tags bookings
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} domain.PaymentReceipt
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/receipt [get]
func (h *BookingHandler) GetPaymentReceipt(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	receipt, err := h.bookingUseCase.GetPaymentReceipt(bookingID, userID)
	if err != nil {
		errorMsg := err.Error()

		if strings.Contains(errorMsg, "не найдено") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(errorMsg))
		} else if strings.Contains(errorMsg, "отказано в доступе") ||
			strings.Contains(errorMsg, "недостаточно прав") ||
			strings.Contains(errorMsg, "не являетесь владельцем") {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse(errorMsg))
		} else if strings.Contains(errorMsg, "пользователь не найден") {
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(errorMsg))
		} else if strings.Contains(errorMsg, "платеж") {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(errorMsg))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка получения чека: "+errorMsg))
		}
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// @Summary Продление аренды
// @Description Создает запрос на продление бронирования
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Param request body domain.ExtendBookingRequest true "Данные продления"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/extend [post]
func (h *BookingHandler) ExtendBooking(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("необходима авторизация"))
		return
	}

	userIDInt := userID.(int)

	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID бронирования"))
		return
	}

	var request domain.ExtendBookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных"))
		return
	}

	if request.Duration <= 0 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("продолжительность должна быть больше 0"))
		return
	}

	err = h.bookingUseCase.RequestExtension(bookingID, userIDInt, &request)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("ваша заявка на продление создана, перейдите в раздел оплаты", nil))
}

// ProcessExtensionPayment обрабатывает оплату продления
// @Summary Оплатить продление бронирования
// @Description Обрабатывает оплату для продления со статусом awaiting_payment. Можно передать либо payment_id, либо order_id
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Param extensionId path int true "ID продления"
// @Param request body domain.ProcessPaymentRequest true "Данные платежа (payment_id или order_id)"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/extensions/{extensionId}/payment [post]
func (h *BookingHandler) ProcessExtensionPayment(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	extensionID, err := strconv.Atoi(c.Param("extensionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID продления"))
		return
	}

	var request domain.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return
	}

	if request.PaymentID == "" && request.OrderID == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("необходимо указать либо payment_id, либо order_id"))
		return
	}

	if request.PaymentID != "" && request.OrderID != "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("нельзя указывать одновременно payment_id и order_id"))
		return
	}

	canAccess, err := h.bookingUseCase.CanUserAccessBooking(bookingID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка проверки доступа"))
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("нет доступа к данному бронированию"))
		return
	}

	var extension *domain.BookingExtension

	if request.OrderID != "" {
		extension, err = h.bookingUseCase.ProcessExtensionPaymentWithOrder(extensionID, request.OrderID)
	} else {
		extension, err = h.bookingUseCase.ProcessExtensionPayment(extensionID, request.PaymentID)
	}

	if err != nil {
		if strings.Contains(err.Error(), "уже использован") || strings.Contains(err.Error(), "уже привязан") {
			c.JSON(http.StatusConflict, domain.NewErrorResponse(err.Error()))
		} else {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("оплата продления обработана", extension))
}

// GetBookingExtensions получает продления бронирования
// @Summary Получить продления бронирования
// @Description Получает список продлений для указанного бронирования
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/extensions [get]
func (h *BookingHandler) GetBookingExtensions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("необходима авторизация"))
		return
	}

	userIDInt := userID.(int)

	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID бронирования"))
		return
	}

	canAccess, err := h.bookingUseCase.CanUserAccessBooking(bookingID, userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("нет прав доступа к этому бронированию"))
		return
	}

	extensions, err := h.bookingUseCase.GetBookingExtensions(bookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("продления получены", extensions))
}

// @Summary Получить доступные варианты продления
// @Description Получает список доступных вариантов продления для указанного бронирования с учетом следующих бронирований и времени уборки
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/available-extensions [get]
func (h *BookingHandler) GetAvailableExtensions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("необходима авторизация"))
		return
	}

	userIDInt := userID.(int)

	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID бронирования"))
		return
	}

	availableExtensions, err := h.bookingUseCase.GetAvailableExtensions(bookingID, userIDInt)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("доступные варианты продления получены", availableExtensions))
}

// @Summary Подтверждение продления
// @Description Подтверждает запрос на продление бронирования
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Param extensionId path int true "ID продления"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/extensions/{extensionId}/approve [post]
func (h *BookingHandler) ApproveExtension(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("необходима авторизация"))
		return
	}

	userIDInt := userID.(int)

	extensionID, err := strconv.Atoi(c.Param("extensionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID продления"))
		return
	}

	err = h.bookingUseCase.ApproveExtension(extensionID, userIDInt)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("продление подтверждено", nil))
}

// @Summary Отклонение продления
// @Description Отклоняет запрос на продление бронирования
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Param extensionId path int true "ID продления"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/extensions/{extensionId}/reject [post]
func (h *BookingHandler) RejectExtension(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("необходима авторизация"))
		return
	}

	userIDInt := userID.(int)

	extensionID, err := strconv.Atoi(c.Param("extensionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID продления"))
		return
	}

	err = h.bookingUseCase.RejectExtension(extensionID, userIDInt)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("продление отклонено", nil))
}

// @Summary Завершение сеанса бронирования
// @Description Позволяет арендатору завершить активное бронирование досрочно, деактивирует пароли замка и освобождает квартиру
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/finish [post]
func (h *BookingHandler) FinishSession(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	err := h.bookingUseCase.FinishSession(bookingID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("сеанс успешно завершен", nil))
}

// @Summary Получение всех бронирований (админ)
// @Description Возвращает список всех бронирований в системе с фильтрами (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "Статус бронирования (created, pending, approved, active, completed, canceled, rejected)"
// @Param apartment_id query int false "ID квартиры"
// @Param renter_id query int false "ID арендатора"
// @Param date_from query string false "Дата начала (YYYY-MM-DD)"
// @Param date_to query string false "Дата окончания (YYYY-MM-DD)"
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/bookings [get]
func (h *BookingHandler) AdminGetAllBookings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := make(map[string]interface{})

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	if apartmentIDStr := c.Query("apartment_id"); apartmentIDStr != "" {
		if apartmentID, err := strconv.Atoi(apartmentIDStr); err == nil {
			filters["apartment_id"] = apartmentID
		}
	}

	if renterIDStr := c.Query("renter_id"); renterIDStr != "" {
		if renterID, err := strconv.Atoi(renterIDStr); err == nil {
			filters["renter_id"] = renterID
		}
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters["date_from"] = dateFrom
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		filters["date_to"] = dateTo
	}

	bookings, total, err := h.bookingUseCase.AdminGetAllBookings(filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка получения бронирований: "+err.Error()))
		return
	}

	statusStats, err := h.bookingUseCase.GetStatusStatistics()
	if err != nil {
		statusStats = make(map[string]int)
	}

	bookingResponses := make([]*domain.BookingResponse, len(bookings))
	for i, booking := range bookings {
		bookingResponses[i] = utils.ConvertBookingToResponse(booking)
	}

	response := gin.H{
		"bookings": bookingResponses,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"pages":     (total + pageSize - 1) / pageSize,
		},
		"statistics": gin.H{
			"total_bookings": total,
			"by_status":      statusStats,
		},
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("список бронирований получен", response))
}

// @Summary Получение бронирования по ID (админ)
// @Description Возвращает детальную информацию о бронировании с расширенными данными (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/bookings/{id} [get]
func (h *BookingHandler) AdminGetBookingByID(c *gin.Context) {
	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	booking, err := h.bookingUseCase.GetBookingByID(bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("бронирование не найдено"))
		return
	}

	response := utils.ConvertBookingToResponse(booking)

	c.JSON(http.StatusOK, domain.NewSuccessResponse("бронирование получено", response))
}

type AdminUpdateBookingStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=created pending approved active completed canceled rejected"`
	Reason string `json:"reason,omitempty"`
}

// @Summary Изменение статуса бронирования (админ)
// @Description Изменяет статус бронирования (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Param request body AdminUpdateBookingStatusRequest true "Новый статус"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/bookings/{id}/status [put]
func (h *BookingHandler) AdminUpdateBookingStatus(c *gin.Context) {
	adminID, _, ok := utils.RequireAnyRole(c, domain.RoleAdmin)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req AdminUpdateBookingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверные данные запроса: "+err.Error()))
		return
	}

	status := domain.BookingStatus(req.Status)
	err := h.bookingUseCase.AdminUpdateBookingStatus(bookingID, status, req.Reason, adminID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse("бронирование не найдено"))
		} else if strings.Contains(err.Error(), "only admins") {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав"))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка изменения статуса: "+err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статус бронирования успешно изменен", gin.H{
		"booking_id": bookingID,
		"new_status": status,
		"reason":     req.Reason,
	}))
}

// @Summary Отмена бронирования (админ)
// @Description Принудительная отмена бронирования (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/bookings/{id} [delete]
func (h *BookingHandler) AdminCancelBooking(c *gin.Context) {
	adminID, _, ok := utils.RequireAnyRole(c, domain.RoleAdmin)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var requestBody map[string]string
	reason := ""
	if err := c.ShouldBindJSON(&requestBody); err == nil {
		reason = requestBody["reason"]
	}

	err := h.bookingUseCase.AdminCancelBooking(bookingID, reason, adminID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse("бронирование не найдено"))
		} else if strings.Contains(err.Error(), "only admins") {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав"))
		} else if strings.Contains(err.Error(), "already canceled") {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("бронирование уже отменено"))
		} else if strings.Contains(err.Error(), "cannot cancel completed") {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("нельзя отменить завершенное бронирование"))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка отмены бронирования: "+err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("бронирование успешно отменено", gin.H{
		"booking_id": bookingID,
		"reason":     reason,
	}))
}

// @Summary Детальная статистика бронирований (админ)
// @Description Возвращает детальную статистику по бронированиям с разбивкой по статусам, месяцам и другим параметрам
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Номер страницы для детальных данных" default(1)
// @Param page_size query int false "Размер страницы для детальных данных" default(1000)
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/bookings/statistics [get]
func (h *BookingHandler) AdminGetBookingStatistics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "1000"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 1000
	}

	statusStats, err := h.bookingUseCase.GetStatusStatistics()
	if err != nil {
		statusStats = make(map[string]int)
	}

	bookings, total, err := h.bookingUseCase.AdminGetAllBookings(map[string]interface{}{}, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении бронирований: "+err.Error()))
		return
	}

	monthlyStats := make(map[string]int)
	durationStats := make(map[string]int)
	doorStatusStats := make(map[string]int)
	totalPrice := 0
	totalDuration := 0
	bookingCount := len(bookings)

	for _, booking := range bookings {
		month := booking.StartDate.Format("2006-01")
		monthlyStats[month]++

		if booking.Duration <= 2 {
			durationStats["short"]++
		} else if booking.Duration <= 6 {
			durationStats["medium"]++
		} else if booking.Duration <= 12 {
			durationStats["long"]++
		} else {
			durationStats["extended"]++
		}

		doorStatusStats[string(booking.DoorStatus)]++

		totalPrice += booking.FinalPrice
		totalDuration += booking.Duration
	}

	var avgPrice, avgDuration float64
	if bookingCount > 0 {
		avgPrice = float64(totalPrice) / float64(bookingCount)
		avgDuration = float64(totalDuration) / float64(bookingCount)
	}

	response := gin.H{
		"summary": gin.H{
			"total_bookings_in_sample": bookingCount,
			"total_bookings_system":    total,
			"total_revenue_in_sample":  totalPrice,
			"avg_price":                avgPrice,
			"avg_duration":             avgDuration,
		},
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
			"pages":     (total + pageSize - 1) / pageSize,
		},
		"by_status":      statusStats,
		"by_month":       monthlyStats,
		"by_duration":    durationStats,
		"by_door_status": doorStatusStats,
		"note":           "Детальная статистика (by_month, by_duration, by_door_status) рассчитана на основе выборки. Статистика by_status показывает данные по всей системе.",
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("детальная статистика бронирований получена", response))
}

// @Summary Получение информации о доступе к замкам для всех активных бронирований
// @Description Возвращает статус доступа к паролям замков для всех активных бронирований пользователя
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/my/lock-access [get]
func (h *BookingHandler) GetMyBookingsLockAccess(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	response, err := h.bookingUseCase.GetMyBookingsLockAccess(userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("информация о доступе к замкам получена", response))
}

// @Summary Получение информации о доступе к замку для конкретного бронирования
// @Description Возвращает детальную информацию о статусе доступа к паролю замка для указанного бронирования
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/lock-access [get]
func (h *BookingHandler) GetBookingLockAccess(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID бронирования"))
		return
	}

	response, err := h.bookingUseCase.GetBookingLockAccess(bookingID, userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("информация о доступе к замку получена", response))
}

// @Summary Генерация пароля для бронирования
// @Description Генерирует временный пароль для замка квартиры по ID бронирования
// @Tags bookings
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/generate-password [post]
func (h *BookingHandler) GeneratePasswordForBooking(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID бронирования"))
		return
	}

	password, err := h.lockUseCase.GeneratePasswordForBookingByID(bookingID, userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("пароль сгенерирован", gin.H{
		"password":     password,
		"booking_id":   bookingID,
		"message":      "Пароль успешно создан",
		"instructions": "Введите пароль на клавиатуре замка и нажмите #",
	}))
}
