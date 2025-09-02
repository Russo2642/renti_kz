package http

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"math"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/services"
	"github.com/russo2642/renti_kz/internal/utils"
)

type UserHandler struct {
	userUseCase          domain.UserUseCase
	propertyOwnerUseCase domain.PropertyOwnerUseCase
	renterUseCase        domain.RenterUseCase
	renterRepo           domain.RenterRepository
	apartmentUseCase     domain.ApartmentUseCase
	bookingUseCase       domain.BookingUseCase
	otpUseCase           domain.OTPUseCase
	responseCacheService *services.ResponseCacheService
}

func NewUserHandler(
	userUseCase domain.UserUseCase,
	propertyOwnerUseCase domain.PropertyOwnerUseCase,
	renterUseCase domain.RenterUseCase,
	renterRepo domain.RenterRepository,
	apartmentUseCase domain.ApartmentUseCase,
	bookingUseCase domain.BookingUseCase,
	otpUseCase domain.OTPUseCase,
	responseCacheService *services.ResponseCacheService,
) *UserHandler {
	return &UserHandler{
		userUseCase:          userUseCase,
		propertyOwnerUseCase: propertyOwnerUseCase,
		renterUseCase:        renterUseCase,
		renterRepo:           renterRepo,
		apartmentUseCase:     apartmentUseCase,
		bookingUseCase:       bookingUseCase,
		otpUseCase:           otpUseCase,
		responseCacheService: responseCacheService,
	}
}

func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.GET("/me", CacheMiddlewareWithTTL(h.responseCacheService, 1*time.Minute), h.GetMe)
		users.PUT("/me", h.UpdateProfile)
		users.POST("/documents-base64", h.UploadDocumentsBase64)
		users.POST("/change-phone", h.ChangePhone)
		users.POST("/confirm-phone", h.ConfirmPhoneChange)
		users.DELETE("/me", h.DeleteAccount)
		users.GET("/apartments/me", CacheMiddlewareWithTTL(h.responseCacheService, 2*time.Minute), h.GetMyApartments)
		users.GET("/my-bookings", h.GetMyBookings)
		users.GET("/property-bookings", h.GetPropertyBookings)
	}

	admin := router.Group("/admin")
	{
		admin.GET("/users", CacheMiddlewareWithTTL(h.responseCacheService, 3*time.Minute), h.AdminGetAllUsers)
		admin.GET("/users/:id", CacheMiddlewareWithTTL(h.responseCacheService, 5*time.Minute), h.AdminGetUserByID)
		admin.PUT("/users/:id/role", h.AdminUpdateUserRole)
		admin.PUT("/users/:id/status", h.AdminUpdateUserStatus)
		admin.DELETE("/users/:id", h.DeleteUserByAdmin)
		admin.GET("/users/statistics", CacheMiddlewareWithTTL(h.responseCacheService, 10*time.Minute), h.AdminGetUserStatistics)
		admin.GET("/users/:id/booking-history", h.AdminGetUserBookingHistory)

		admin.PUT("/renters/:id/verification-status", h.UpdateRenterVerificationStatus)
	}
}

// @Summary Получение данных текущего пользователя
// @Description Получает профиль авторизованного пользователя с информацией о всех ролях
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	user, err := h.userUseCase.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	responseData := gin.H{
		"user": user,
	}

	renter, err := h.renterUseCase.GetByUserID(userID)
	if err == nil && renter != nil {
		responseData["renter_info"] = gin.H{
			"verification_status": renter.VerificationStatus,
			"document_url":        renter.DocumentURL,
			"photo_with_doc_url":  renter.PhotoWithDocURL,
			"document_type":       renter.DocumentType,
		}
	}

	if user.Role == domain.RoleOwner {
		propertyOwner, err := h.propertyOwnerUseCase.GetByUserID(userID)
		if err == nil && propertyOwner != nil {
			responseData["property_owner_info"] = gin.H{
				"id": propertyOwner.ID,
			}
		}
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("данные пользователя получены успешно", responseData))
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	CityID    int    `json:"city_id"`
}

// @Summary Обновление профиля
// @Description Обновляет данные профиля пользователя
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body UpdateProfileRequest true "Данные для обновления"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /users/me [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	user, err := h.userUseCase.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	updated := false

	if req.FirstName != "" {
		user.FirstName = req.FirstName
		updated = true
	}

	if req.LastName != "" {
		user.LastName = req.LastName
		updated = true
	}

	if req.Email != "" {

		if err := utils.ValidateEmail(req.Email); err != nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
			return
		}
		user.Email = req.Email
		updated = true
	}

	if req.CityID > 0 {
		user.CityID = req.CityID
		updated = true
	}

	if !updated {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("не указаны поля для обновления"))
		return
	}

	if err := h.userUseCase.UpdateProfile(user); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении профиля"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("профиль успешно обновлен", user))
}

type UploadDocumentsBase64Request struct {
	UserID              int      `json:"user_id" binding:"required"`
	DocumentType        string   `json:"document_type" binding:"required,oneof=udv passport"`
	DocumentsBase64     []string `json:"documents_base64" binding:"required"`
	SelfieWithDocBase64 string   `json:"selfie_with_doc_base64" binding:"required"`
}

// @Summary Загрузка документов в формате base64
// @Description Загружает документы и селфи с документом для верификации пользователя в формате base64 без авторизации
// @Tags users
// @Accept json
// @Produce json
// @Param request body UploadDocumentsBase64Request true "Данные документов"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /users/documents-base64 [post]
func (h *UserHandler) UploadDocumentsBase64(c *gin.Context) {
	var req UploadDocumentsBase64Request

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	user, err := h.userUseCase.GetByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	if user.Role != domain.RoleUser && user.Role != domain.RoleOwner {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("только арендаторы и владельцы недвижимости могут загружать документы для верификации"))
		return
	}

	documentType := domain.DocumentType(req.DocumentType)

	if documentType == domain.DocTypeID && len(req.DocumentsBase64) != 2 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("для удостоверения личности требуется 2 фото документа"))
		return
	}

	if documentType == domain.DocTypePassport && len(req.DocumentsBase64) != 1 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("для паспорта требуется 1 фото документа"))
		return
	}

	documentsData := make([][]byte, len(req.DocumentsBase64))
	for i, b64 := range req.DocumentsBase64 {
		data, err := h.convertBase64ToBytes(b64)
		if err != nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(fmt.Sprintf("ошибка при декодировании документа %d: %v", i+1, err)))
			return
		}
		documentsData[i] = data
	}

	selfieData, err := h.convertBase64ToBytes(req.SelfieWithDocBase64)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("ошибка при декодировании селфи: "+err.Error()))
		return
	}

	renter, err := h.renterUseCase.GetByUserID(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных арендатора"))
		return
	}

	if renter == nil {
		emptyDocURLs := make(map[string]string)
		newRenter := &domain.Renter{
			UserID:             req.UserID,
			DocumentType:       documentType,
			DocumentURL:        emptyDocURLs,
			PhotoWithDocURL:    "",
			VerificationStatus: domain.VerificationPending,
		}

		if err := h.renterRepo.Create(newRenter); err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при создании записи арендатора"))
			return
		}

		renter = newRenter
	}

	documentURLs, err := h.renterUseCase.UploadDocumentsParallel(renter.ID, documentType, documentsData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при загрузке документов: "+err.Error()))
		return
	}

	renter.DocumentURL = documentURLs
	renter.DocumentType = documentType

	if err := h.renterRepo.Update(renter); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении данных арендатора"))
		return
	}

	photoURL, err := h.renterUseCase.UploadPhotoWithDoc(renter.ID, selfieData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при загрузке селфи: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("документы и селфи успешно загружены", gin.H{
		"document_urls": documentURLs,
		"selfie_url":    photoURL,
	}))
}

func (h *UserHandler) convertBase64ToBytes(b64String string) ([]byte, error) {
	if strings.Contains(b64String, ",") {
		parts := strings.Split(b64String, ",")
		if len(parts) == 2 {
			b64String = parts[1]
		}
	}

	b64String = strings.ReplaceAll(b64String, " ", "")
	b64String = strings.ReplaceAll(b64String, "\n", "")
	b64String = strings.ReplaceAll(b64String, "\r", "")
	b64String = strings.ReplaceAll(b64String, "\t", "")

	data, err := base64.StdEncoding.DecodeString(b64String)
	if err != nil {
		return nil, fmt.Errorf("некорректные данные base64: %w", err)
	}
	return data, nil
}

type ChangePhoneRequest struct {
	NewPhone string `json:"new_phone" binding:"required"`
}

// @Summary Запрос на смену номера телефона
// @Description Отправляет OTP на новый номер телефона для подтверждения смены
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body ChangePhoneRequest true "Новый номер телефона"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /users/change-phone [post]
func (h *UserHandler) ChangePhone(c *gin.Context) {
	var req ChangePhoneRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	if err := utils.ValidatePhone(req.NewPhone); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	user, err := h.userUseCase.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	if user.Phone == req.NewPhone {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("новый номер телефона должен отличаться от текущего"))
		return
	}

	otpResponse, err := h.otpUseCase.RequestOTP(req.NewPhone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при отправке OTP: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("OTP отправлен на новый номер телефона", gin.H{
		"phone":  req.NewPhone,
		"otp_id": otpResponse.ID,
	}))
}

type ConfirmPhoneChangeRequest struct {
	NewPhone string `json:"new_phone" binding:"required"`
	OTP      string `json:"otp" binding:"required"`
	OTPID    string `json:"otp_id" binding:"required"`
}

// @Summary Подтверждение смены номера телефона
// @Description Подтверждает смену номера телефона через OTP
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body ConfirmPhoneChangeRequest true "Новый номер, OTP и OTP ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /users/confirm-phone [post]
func (h *UserHandler) ConfirmPhoneChange(c *gin.Context) {
	var req ConfirmPhoneChangeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	if err := utils.ValidatePhone(req.NewPhone); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	isValid, err := h.otpUseCase.VerifyOTP(req.NewPhone, req.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке OTP: "+err.Error()))
		return
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный OTP код"))
		return
	}

	user, err := h.userUseCase.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	user.Phone = req.NewPhone

	if err := h.userUseCase.UpdateProfile(user); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении номера телефона"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("номер телефона успешно изменён", gin.H{
		"new_phone": req.NewPhone,
	}))
}

// @Summary Получение моих бронирований
// @Description Получает список квартир, которые пользователь бронирует как арендатор с пагинацией
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Param status query []string false "Фильтр по статусу"
// @Param date_from query string false "Дата От (по created_at) в формате YYYY-MM-DD"
// @Param date_to query string false "Дата До (по created_at) в формате YYYY-MM-DD"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /users/my-bookings [get]
func (h *UserHandler) GetMyBookings(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	renter, err := h.renterUseCase.GetByUserID(userID)
	if err != nil || renter == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("профиль арендатора не найден"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	statuses := utils.ParseStatusFilter(c)

	dateFrom, dateTo := parseDateFilters(c)

	bookings, total, err := h.bookingUseCase.GetRenterBookings(userID, statuses, dateFrom, dateTo, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении бронирований: "+err.Error()))
		return
	}

	enrichedBookings := h.enrichBookingsWithPriceDetails(bookings)

	response := gin.H{
		"bookings": enrichedBookings,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"pages":     (total + pageSize - 1) / pageSize,
		},
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("мои бронирования получены успешно", response))
}

func (h *UserHandler) enrichBookingsWithPriceDetails(bookings []*domain.Booking) []gin.H {
	enrichedBookings := make([]gin.H, len(bookings))

	for i, booking := range bookings {
		apartment := booking.Apartment

		enrichedBooking := gin.H{
			"id":                   booking.ID,
			"renter_id":            booking.RenterID,
			"renter":               booking.Renter,
			"apartment_id":         booking.ApartmentID,
			"apartment":            booking.Apartment,
			"contract_id":          booking.ContractID,
			"start_date":           booking.StartDate,
			"end_date":             booking.EndDate,
			"duration":             booking.Duration,
			"cleaning_duration":    booking.CleaningDuration,
			"status":               booking.Status,
			"total_price":          booking.TotalPrice,
			"service_fee":          booking.ServiceFee,
			"final_price":          booking.FinalPrice,
			"is_contract_accepted": booking.IsContractAccepted,
			"payment_id":           booking.PaymentID,
			"cancellation_reason":  booking.CancellationReason,
			"owner_comment":        booking.OwnerComment,
			"booking_number":       booking.BookingNumber,
			"door_status":          booking.DoorStatus,
			"last_door_action":     booking.LastDoorAction,
			"can_extend":           booking.CanExtend,
			"extension_requested":  booking.ExtensionRequested,
			"extension_end_date":   booking.ExtensionEndDate,
			"extension_duration":   booking.ExtensionDuration,
			"extension_price":      booking.ExtensionPrice,
			"created_at":           booking.CreatedAt,
			"updated_at":           booking.UpdatedAt,
		}

		if apartment != nil {
			timeInfo := utils.GetRentalTimeInfo(booking.StartDate)

			var basePrice int
			var calculationInfo string
			if booking.Duration == 24 && apartment.RentalTypeDaily {
				basePrice = apartment.DailyPrice
				calculationInfo = fmt.Sprintf("Посуточная аренда: %d тг", apartment.DailyPrice)
			} else {
				basePrice = utils.CalculateHourlyPrice(apartment.Price, booking.Duration)
				if booking.Duration == utils.RentalDuration6Hours {
					discountPercent := utils.Discount6Hours
					originalPrice := apartment.Price * booking.Duration
					calculationInfo = fmt.Sprintf("Почасовая аренда: %d тг × %d ч = %d тг (скидка %d%%, было %d тг)", apartment.Price, booking.Duration, basePrice, discountPercent, originalPrice)
				} else if booking.Duration == utils.RentalDuration12Hours {
					discountPercent := utils.Discount12Hours
					originalPrice := apartment.Price * booking.Duration
					calculationInfo = fmt.Sprintf("Почасовая аренда: %d тг × %d ч = %d тг (скидка %d%%, было %d тг)", apartment.Price, booking.Duration, basePrice, discountPercent, originalPrice)
				} else {
					calculationInfo = fmt.Sprintf("Почасовая аренда: %d тг × %d ч = %d тг", apartment.Price, booking.Duration, basePrice)
				}
			}

			var serviceFeeInfo string
			if booking.Duration == 24 {
				serviceFeeInfo = "Фиксированный сервисный сбор: 3000 тг"
			} else {
				serviceFeePercentage := domain.ServiceFeePercentage
				if h.apartmentUseCase != nil {
					// Получаем настройки через apartmentUseCase, если доступно
					// Можно добавить метод для получения процента сервисного сбора
				}
				serviceFeeInfo = fmt.Sprintf("Сервисный сбор (%d%%): %d тг", serviceFeePercentage, booking.ServiceFee)
			}

			enrichedBooking["price_details"] = gin.H{
				"base_price":   basePrice,
				"hourly_price": apartment.Price,
				"daily_price":  apartment.DailyPrice,
				"time_info":    timeInfo,
				"price_breakdown": gin.H{
					"calculation":      calculationInfo,
					"service_fee_info": serviceFeeInfo,
				},
			}
		}

		enrichedBookings[i] = enrichedBooking
	}

	return enrichedBookings
}

// @Summary Получение бронирований моих квартир
// @Description Получает список бронирований на квартиры пользователя как владельца с пагинацией
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Param status query []string false "Фильтр по статусу"
// @Param date_from query string false "Дата От в формате YYYY-MM-DD"
// @Param date_to query string false "Дата До в формате YYYY-MM-DD"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /users/property-bookings [get]
func (h *UserHandler) GetPropertyBookings(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	user, err := h.userUseCase.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	if user.Role != domain.RoleOwner {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("доступ разрешен только владельцам недвижимости"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	statuses := utils.ParseStatusFilter(c)

	dateFrom, dateTo := parseDateFilters(c)

	bookings, total, err := h.bookingUseCase.GetOwnerBookings(userID, statuses, dateFrom, dateTo, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении бронирований: "+err.Error()))
		return
	}

	response := gin.H{
		"bookings": bookings,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"pages":     (total + pageSize - 1) / pageSize,
		},
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("бронирования моих квартир получены успешно", response))
}

type UpdateVerificationStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending approved rejected"`
}

// @Description Обновляет статус верификации арендатора (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID пользователя"
// @Param request body UpdateVerificationStatusRequest true "Новый статус верификации"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/renters/{id}/verification-status [put]
func (h *UserHandler) UpdateRenterVerificationStatus(c *gin.Context) {
	var req UpdateVerificationStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("идентификатор пользователя обязателен"))
		return
	}

	role, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	if role.(domain.UserRole) != domain.RoleAdmin {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("доступ запрещен"))
		return
	}

	userIDInt, _ := strconv.Atoi(userID)
	user, err := h.userUseCase.GetByID(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("пользователь не найден"))
		return
	}

	if user.Role != domain.RoleUser {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("пользователь не является арендатором"))
		return
	}

	renter, err := h.renterUseCase.GetByUserID(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных арендатора"))
		return
	}
	if renter == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("данные арендатора не найдены"))
		return
	}

	status := domain.VerificationStatus(req.Status)
	if err := h.renterUseCase.UpdateVerificationStatus(renter.ID, status); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении статуса верификации"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статус верификации арендатора успешно обновлен", gin.H{
		"renter_id": renter.ID,
		"user_id":   userIDInt,
		"status":    status,
	}))
}

// @Summary Удаление своего аккаунта
// @Description Позволяет пользователю удалить свой собственный аккаунт
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /users/me [delete]
func (h *UserHandler) DeleteAccount(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	userIDInt := userID.(int)

	if err := h.userUseCase.DeleteOwnAccount(userIDInt); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при удалении аккаунта"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("аккаунт успешно удален", nil))
}

// @Summary Удаление аккаунта пользователя
// @Description Удаляет аккаунт пользователя (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID пользователя"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/users/{id} [delete]
func (h *UserHandler) DeleteUserByAdmin(c *gin.Context) {

	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	role, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	if role.(domain.UserRole) != domain.RoleAdmin {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("доступ запрещен"))
		return
	}

	userIDParam := c.Param("id")
	if userIDParam == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("идентификатор пользователя обязателен"))
		return
	}

	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный идентификатор пользователя"))
		return
	}

	if err := h.userUseCase.DeleteUser(userID, adminID.(int)); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse("пользователь не найден"))
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при удалении пользователя"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("пользователь успешно удален", nil))
}

// @Summary Получение квартир пользователя
// @Description Получает список квартир, принадлежащих пользователю
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /users/apartments/me [get]
func (h *UserHandler) GetMyApartments(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	userIDInt := userID.(int)

	user, err := h.userUseCase.GetByID(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	if user.Role != domain.RoleOwner {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("доступ разрешен только владельцам недвижимости"))
		return
	}

	owner, err := h.propertyOwnerUseCase.GetByUserID(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении профиля владельца"))
		return
	}
	if owner == nil {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("профиль владельца не найден"))
		return
	}

	pageParam := c.DefaultQuery("page", "1")
	pageSizeParam := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeParam)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filters := make(map[string]interface{})
	filters["owner_id"] = owner.ID

	if status := c.Query("status"); status != "" {

		switch status {
		case string(domain.AptStatusPending), string(domain.AptStatusApproved),
			string(domain.AptStatusNeedsRevision), string(domain.AptStatusRejected):
			filters["status"] = status
		default:
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный статус. Допустимые значения: 'pending', 'approved', 'needs_revision', 'rejected'"))
			return
		}
	}

	apartments, total, err := h.apartmentUseCase.GetAll(filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении списка квартир"))
		return
	}

	response := gin.H{
		"apartments": apartments,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"pages":     (total + pageSize - 1) / pageSize,
		},
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("список собственных квартир получен успешно", response))
}

func parseDateFilters(c *gin.Context) (*time.Time, *time.Time) {
	var dateFrom, dateTo *time.Time

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			start := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location())
			dateFrom = &start
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateToStr); err == nil {
			end := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
			dateTo = &end
		}
	}

	return dateFrom, dateTo
}

// @Summary Получение всех пользователей (админ)
// @Description Возвращает список всех пользователей в системе с фильтрами (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role query string false "Роль пользователя (user, owner, moderator, admin)"
// @Param city_id query int false "ID города"
// @Param verification_status query string false "Статус верификации арендатора (pending, approved, rejected)"
// @Param search query string false "Поиск по ФИО, телефону или email"
// @Param is_active query bool false "Активность пользователя (true - активен, false - неактивен)"
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/users [get]
func (h *UserHandler) AdminGetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := make(map[string]interface{})

	if role := c.Query("role"); role != "" {
		filters["role"] = domain.UserRole(role)
	}

	if cityIDStr := c.Query("city_id"); cityIDStr != "" {
		if cityID, err := strconv.Atoi(cityIDStr); err == nil {
			filters["city_id"] = cityID
		}
	}

	if verificationStatus := c.Query("verification_status"); verificationStatus != "" {
		filters["verification_status"] = verificationStatus
	}

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filters["is_active"] = isActive
		}
	}

	users, total, err := h.userUseCase.GetAllUsers(filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка получения пользователей: "+err.Error()))
		return
	}

	roleStats, err := h.userUseCase.GetRoleStatistics()
	if err != nil {
		roleStats = make(map[string]int)
	}

	statusStats, err := h.userUseCase.GetStatusStatistics()
	if err != nil {
		statusStats = make(map[string]int)
	}

	response := gin.H{
		"users": users,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"pages":     (total + pageSize - 1) / pageSize,
		},
		"statistics": gin.H{
			"total_users": total,
			"by_role":     roleStats,
			"by_status":   statusStats,
		},
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("список пользователей получен", response))
}

// @Summary Получение пользователя по ID (админ)
// @Description Возвращает детальную информацию о пользователе с расширенными данными (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID пользователя"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/users/{id} [get]
func (h *UserHandler) AdminGetUserByID(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID пользователя"))
		return
	}

	user, err := h.userUseCase.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("пользователь не найден"))
		return
	}

	// Получаем дополнительную информацию
	response := gin.H{
		"user": user,
	}

	// Если пользователь является арендатором
	if renter, err := h.renterUseCase.GetByUserID(userID); err == nil && renter != nil {
		response["renter_info"] = gin.H{
			"id":                  renter.ID,
			"verification_status": renter.VerificationStatus,
			"document_url":        renter.DocumentURL,
			"photo_with_doc_url":  renter.PhotoWithDocURL,
			"document_type":       renter.DocumentType,
			"created_at":          renter.CreatedAt,
			"updated_at":          renter.UpdatedAt,
		}
	}

	// Если пользователь является владельцем
	if user.Role == domain.RoleOwner {
		if owner, err := h.propertyOwnerUseCase.GetByUserID(userID); err == nil && owner != nil {
			response["property_owner_info"] = gin.H{
				"id":         owner.ID,
				"created_at": owner.CreatedAt,
				"updated_at": owner.UpdatedAt,
			}

			// Получаем статистику квартир владельца
			if apartments, err := h.apartmentUseCase.GetByOwnerID(owner.ID); err == nil {
				response["apartments_count"] = len(apartments)
			}
		}
	}

	// Получаем статистику бронирований
	if bookings, _, err := h.bookingUseCase.GetRenterBookings(userID, nil, nil, nil, 1, 1000); err == nil {
		response["bookings_count"] = len(bookings)
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("пользователь получен", response))
}

type AdminUpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user owner moderator admin concierge cleaner"`
}

// @Summary Изменение роли пользователя (админ)
// @Description Изменяет роль пользователя (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID пользователя"
// @Param request body AdminUpdateUserRoleRequest true "Новая роль"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/users/{id}/role [put]
func (h *UserHandler) AdminUpdateUserRole(c *gin.Context) {
	adminID, _, ok := utils.RequireAnyRole(c, domain.RoleAdmin)
	if !ok {
		return
	}

	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID пользователя"))
		return
	}

	var req AdminUpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверные данные запроса: "+err.Error()))
		return
	}

	newRole := domain.UserRole(req.Role)
	err = h.userUseCase.UpdateUserRole(userID, newRole, adminID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse("пользователь не найден"))
		} else if strings.Contains(err.Error(), "only admins") {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав"))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка изменения роли: "+err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("роль пользователя успешно изменена", gin.H{
		"user_id":  userID,
		"new_role": newRole,
	}))
}

type AdminUpdateUserStatusRequest struct {
	IsActive bool   `json:"is_active"`
	Reason   string `json:"reason,omitempty"`
}

// @Summary Изменение статуса пользователя (админ)
// @Description Блокировка/разблокировка пользователя (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID пользователя"
// @Param request body AdminUpdateUserStatusRequest true "Новый статус"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/users/{id}/status [put]
func (h *UserHandler) AdminUpdateUserStatus(c *gin.Context) {
	adminID, _, ok := utils.RequireAnyRole(c, domain.RoleAdmin)
	if !ok {
		return
	}

	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID пользователя"))
		return
	}

	var req AdminUpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверные данные запроса: "+err.Error()))
		return
	}

	err = h.userUseCase.UpdateUserStatus(userID, req.IsActive, req.Reason, adminID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse("пользователь не найден"))
		} else if strings.Contains(err.Error(), "only admins") {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав"))
		} else if strings.Contains(err.Error(), "not yet implemented") {
			c.JSON(http.StatusNotImplemented, domain.NewErrorResponse("функция изменения статуса пользователя пока не реализована"))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка изменения статуса: "+err.Error()))
		}
		return
	}

	statusText := "активирован"
	if !req.IsActive {
		statusText = "заблокирован"
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статус пользователя успешно изменен", gin.H{
		"user_id":   userID,
		"is_active": req.IsActive,
		"status":    statusText,
	}))
}

// @Summary Детальная статистика пользователей (админ)
// @Description Возвращает детальную статистику по пользователям с разбивкой по ролям, статусам, городам и другим параметрам
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/users/statistics [get]
func (h *UserHandler) AdminGetUserStatistics(c *gin.Context) {
	users, _, err := h.userUseCase.GetAllUsers(map[string]interface{}{}, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении пользователей: "+err.Error()))
		return
	}

	roleStats, err := h.userUseCase.GetRoleStatistics()
	if err != nil {
		roleStats = make(map[string]int)
	}

	statusStats, err := h.userUseCase.GetStatusStatistics()
	if err != nil {
		statusStats = make(map[string]int)
	}

	cityStats := make(map[string]int)
	monthlyRegistrations := make(map[string]int)
	verificationStats := make(map[string]int)
	userCount := len(users)

	for _, user := range users {
		if user.City != nil {
			cityStats[user.City.Name]++
		}

		month := user.CreatedAt.Format("2006-01")
		monthlyRegistrations[month]++

		if user.Role == domain.RoleUser {
			renter, err := h.renterUseCase.GetByUserID(user.ID)
			if err == nil && renter != nil {
				verificationStats[string(renter.VerificationStatus)]++
			} else {
				verificationStats["not_registered"]++
			}
		}
	}

	ownersCount := 0
	for _, user := range users {
		if user.Role == domain.RoleOwner {
			ownersCount++
		}
	}

	response := gin.H{
		"summary": gin.H{
			"total_users":    userCount,
			"total_owners":   ownersCount,
			"total_renters":  userCount - ownersCount,
			"active_users":   statusStats["active"],
			"inactive_users": statusStats["inactive"],
		},
		"by_role":               roleStats,
		"by_status":             statusStats,
		"by_city":               cityStats,
		"by_month_registration": monthlyRegistrations,
		"renter_verification":   verificationStats,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("детальная статистика пользователей получена", response))
}

// @Summary История бронирований пользователя (админ)
// @Description Возвращает детальную историю всех бронирований конкретного пользователя с аналитикой
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID пользователя"
// @Param role query string false "Роль пользователя (renter, owner)" default(renter)
// @Param status query []string false "Фильтр по статусу бронирования"
// @Param date_from query string false "Дата начала периода (YYYY-MM-DD)"
// @Param date_to query string false "Дата окончания периода (YYYY-MM-DD)"
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/users/{id}/booking-history [get]
func (h *UserHandler) AdminGetUserBookingHistory(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный ID пользователя"))
		return
	}

	user, err := h.userUseCase.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("пользователь не найден"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	role := c.DefaultQuery("role", "renter")

	filters := make(map[string]interface{})

	if role == "owner" {
		filters["owner_user_id"] = userID
	} else {
		filters["renter_user_id"] = userID
	}

	if statusParam := c.QueryArray("status"); len(statusParam) > 0 {
		filters["status"] = statusParam
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

	analytics := h.calculateUserAnalytics(bookings, user, role)

	bookingResponses := make([]gin.H, len(bookings))
	for i, booking := range bookings {
		bookingResponses[i] = gin.H{
			"id":             booking.ID,
			"booking_number": booking.BookingNumber,
			"apartment":      booking.Apartment,
			"renter":         booking.Renter,
			"start_date":     booking.StartDate,
			"end_date":       booking.EndDate,
			"duration":       booking.Duration,
			"status":         booking.Status,
			"total_price":    booking.TotalPrice,
			"service_fee":    booking.ServiceFee,
			"final_price":    booking.FinalPrice,
			"door_status":    booking.DoorStatus,
			"can_extend":     booking.CanExtend,
			"created_at":     booking.CreatedAt,
			"updated_at":     booking.UpdatedAt,
		}
	}

	response := gin.H{
		"user": gin.H{
			"id":         user.ID,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"phone":      user.Phone,
			"role":       user.Role,
			"is_active":  user.IsActive,
		},
		"search_role": role,
		"bookings":    bookingResponses,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"pages":     (total + pageSize - 1) / pageSize,
		},
		"analytics": analytics,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("история бронирований пользователя получена", response))
}

func (h *UserHandler) calculateUserAnalytics(bookings []*domain.Booking, user *domain.User, role string) gin.H {
	var totalSpent, totalEarned, totalDuration int
	var completedBookings, canceledBookings, activeBookings int
	statusCounts := make(map[string]int)
	monthlyStats := make(map[string]int)
	apartmentFrequency := make(map[int]int)
	apartmentNames := make(map[int]string)

	for _, booking := range bookings {
		statusCounts[string(booking.Status)]++

		switch booking.Status {
		case domain.BookingStatusCompleted:
			completedBookings++
			if role == "renter" {
				totalSpent += booking.FinalPrice
			} else {
				totalEarned += booking.FinalPrice
			}
		case domain.BookingStatusCanceled:
			canceledBookings++
		case domain.BookingStatusActive:
			activeBookings++
			if role == "renter" {
				totalSpent += booking.FinalPrice
			} else {
				totalEarned += booking.FinalPrice
			}
		}

		totalDuration += booking.Duration

		month := booking.StartDate.Format("2006-01")
		monthlyStats[month]++

		if booking.Apartment != nil {
			apartmentFrequency[booking.Apartment.ID]++
			apartmentNames[booking.Apartment.ID] = booking.Apartment.Description
		}
	}

	type apartmentStat struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Bookings int    `json:"bookings"`
	}
	var topApartments []apartmentStat
	for aptID, count := range apartmentFrequency {
		topApartments = append(topApartments, apartmentStat{
			ID:       aptID,
			Name:     apartmentNames[aptID],
			Bookings: count,
		})
	}

	var avgSpentEarned, avgDuration float64

	if len(bookings) > 0 {
		avgDuration = float64(totalDuration) / float64(len(bookings))
		if role == "renter" && (completedBookings+activeBookings) > 0 {
			avgSpentEarned = float64(totalSpent) / float64(completedBookings+activeBookings)
		} else if role == "owner" && (completedBookings+activeBookings) > 0 {
			avgSpentEarned = float64(totalEarned) / float64(completedBookings+activeBookings)
		}
	}

	analytics := gin.H{
		"total_bookings":     len(bookings),
		"completed_bookings": completedBookings,
		"canceled_bookings":  canceledBookings,
		"active_bookings":    activeBookings,
		"avg_duration":       avgDuration,
		"status_breakdown":   statusCounts,
		"monthly_stats":      monthlyStats,
		"top_apartments":     topApartments,
	}

	if role == "renter" {
		analytics["total_spent"] = totalSpent
		analytics["avg_spent"] = avgSpentEarned
		analytics["loyalty_score"] = float64(completedBookings) / math.Max(float64(len(bookings)), 1) * 100
	} else {
		analytics["total_earned"] = totalEarned
		analytics["avg_earned"] = avgSpentEarned
		analytics["apartments_count"] = len(apartmentFrequency)
	}

	return analytics
}
