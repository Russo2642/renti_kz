package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type LockHandler struct {
	lockUseCase         domain.LockUseCase
	userUseCase         domain.UserUseCase
	bookingUseCase      domain.BookingUseCase
	notificationUseCase domain.NotificationUseCase
	apartmentRepo       domain.ApartmentRepository
}

func NewLockHandler(lockUseCase domain.LockUseCase, userUseCase domain.UserUseCase, bookingUseCase domain.BookingUseCase, notificationUseCase domain.NotificationUseCase, apartmentRepo domain.ApartmentRepository) *LockHandler {
	return &LockHandler{
		lockUseCase:         lockUseCase,
		userUseCase:         userUseCase,
		bookingUseCase:      bookingUseCase,
		notificationUseCase: notificationUseCase,
		apartmentRepo:       apartmentRepo,
	}
}

func (h *LockHandler) RegisterRoutes(router *gin.RouterGroup, middleware *Middleware) {
	locks := router.Group("/locks")
	locks.Use(middleware.AuthMiddleware())
	{

		locks.POST("", h.CreateLock)
		locks.GET("", h.GetAllLocks)
		locks.GET("/:id", h.GetLockByID)
		locks.GET("/unique/:uniqueId", h.GetLockByUniqueID)
		locks.GET("/apartment/:apartmentId", h.GetLockByApartmentID)
		locks.PUT("/:id", h.UpdateLock)
		locks.DELETE("/:id", h.DeleteLock)

		locks.GET("/password/:uniqueId/owner", h.GetOwnerPassword)
		locks.DELETE("/password/booking/:bookingId", h.DeactivatePasswordForBooking)
		locks.GET("/status/:uniqueId", h.GetLockStatus)
		locks.GET("/history/:uniqueId", h.GetLockHistory)
	}

	deviceAPI := router.Group("/device/locks")
	{
		deviceAPI.POST("/status", h.UpdateLockStatus)
		deviceAPI.POST("/heartbeat", h.ProcessHeartbeat)
	}

}

func (h *LockHandler) RegisterAdminRoutes(router *gin.RouterGroup) {
	router.GET("/locks", h.AdminGetAllLocks)
	router.GET("/locks/:id", h.AdminGetLockByID)
	router.POST("/locks/:id/bind-apartment", h.AdminBindLockToApartment)
	router.DELETE("/locks/:id/unbind-apartment", h.AdminUnbindLockFromApartment)
	router.PUT("/locks/:id/emergency-reset", h.AdminEmergencyResetLock)
	router.GET("/locks/statistics", h.AdminGetLocksStatistics)

	router.POST("/locks/by-unique-id/:uniqueId/passwords", h.AdminGeneratePassword)
	router.GET("/locks/by-unique-id/:uniqueId/passwords", h.AdminGetAllLockPasswords)
	router.POST("/passwords/:passwordId/deactivate", h.AdminDeactivatePassword)
}

// @Summary Создание замка
// @Description Создает новый умный замок для квартиры
// @Tags locks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body domain.CreateLockRequest true "Данные замка"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /locks [post]
func (h *LockHandler) CreateLock(c *gin.Context) {

	_, _, ok := utils.RequireAnyRole(c, domain.RoleAdmin, domain.RoleModerator)
	if !ok {
		return
	}

	var request domain.CreateLockRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных"))
		return
	}

	lock, err := h.lockUseCase.CreateLock(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, domain.NewSuccessResponse("замок создан", lock))
}

// @Summary Получение всех замков
// @Description Получает список всех умных замков (только для админов и модераторов)
// @Tags locks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /locks [get]
func (h *LockHandler) GetAllLocks(c *gin.Context) {

	_, _, ok := utils.RequireAnyRole(c, domain.RoleAdmin, domain.RoleModerator)
	if !ok {
		return
	}

	locks, err := h.lockUseCase.GetAllLocks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("список замков", locks))
}

// @Summary Получение замка по ID
// @Description Получает информацию о замке по его идентификатору
// @Tags locks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /locks/{id} [get]
func (h *LockHandler) GetLockByID(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	lockID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	lock, err := h.lockUseCase.GetLockByID(lockID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("замок не найден"))
		return
	}

	canAccess, err := h.canUserAccessLock(userID, lock.UniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("нет прав для просмотра этого замка"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("информация о замке", lock))
}

// @Summary Получение замка по уникальному ID
// @Description Получает информацию о замке по его уникальному идентификатору
// @Tags locks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uniqueId path string true "Уникальный ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /locks/unique/{uniqueId} [get]
func (h *LockHandler) GetLockByUniqueID(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	uniqueID := c.Param("uniqueId")

	lock, err := h.lockUseCase.GetLockByUniqueID(uniqueID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("замок не найден"))
		return
	}

	canAccess, err := h.canUserAccessLock(userID, uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("нет прав для просмотра этого замка"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("информация о замке", lock))
}

// @Summary Получение замка квартиры
// @Description Получает информацию о замке для указанной квартиры
// @Tags locks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param apartmentId path int true "ID квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /locks/apartment/{apartmentId} [get]
func (h *LockHandler) GetLockByApartmentID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("необходима авторизация"))
		return
	}

	apartmentID, err := strconv.Atoi(c.Param("apartmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	lock, err := h.lockUseCase.GetLockByApartmentID(apartmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("замок не найден"))
		return
	}

	canAccess, err := h.canUserAccessLock(userID.(int), lock.UniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("нет прав для просмотра этого замка"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("информация о замке", lock))
}

// @Summary Обновление замка
// @Description Обновляет информацию об умном замке (только для админов и модераторов)
// @Tags locks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID замка"
// @Param request body domain.UpdateLockRequest true "Данные для обновления"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /locks/{id} [put]
func (h *LockHandler) UpdateLock(c *gin.Context) {

	_, _, ok := utils.RequireAnyRole(c, domain.RoleAdmin, domain.RoleModerator)
	if !ok {
		return
	}

	lockID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var request domain.UpdateLockRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных"))
		return
	}

	err := h.lockUseCase.UpdateLock(lockID, &request)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("замок обновлен", nil))
}

// @Summary Удаление замка
// @Description Удаляет умный замок (только для админов)
// @Tags locks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /locks/{id} [delete]
func (h *LockHandler) DeleteLock(c *gin.Context) {

	_, _, ok := utils.RequireAnyRole(c, domain.RoleAdmin)
	if !ok {
		return
	}

	lockID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	err := h.lockUseCase.DeleteLock(lockID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("замок удален", nil))
}

// @Summary Получение пароля владельца
// @Description Получает постоянный пароль владельца замка
// @Tags locks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uniqueId path string true "Уникальный ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /locks/password/{uniqueId}/owner [get]
func (h *LockHandler) GetOwnerPassword(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	uniqueID := c.Param("uniqueId")

	password, err := h.lockUseCase.GetOwnerPassword(uniqueID, userID)
	if err != nil {
		if err.Error() == "только владелец квартиры может получить постоянный пароль" {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse(err.Error()))
		} else if err.Error() == "замок не найден" {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
		} else {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		}
		return
	}

	response := map[string]interface{}{
		"password": password,
		"message":  "Постоянный пароль владельца",
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("пароль владельца", response))
}

// @Summary Деактивация пароля для бронирования
// @Description Деактивирует временный пароль для указанного бронирования
// @Tags locks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param bookingId path int true "ID бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /locks/password/booking/{bookingId} [delete]
func (h *LockHandler) DeactivatePasswordForBooking(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "bookingId")
	if !ok {
		return
	}

	user, err := h.userUseCase.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка получения пользователя"))
		return
	}

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator && user.Role != domain.RoleOwner {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("нет прав для деактивации паролей"))
		return
	}

	err = h.lockUseCase.DeactivatePasswordForBooking(bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("пароль для бронирования деактивирован", nil))
}

// @Summary Получение статуса замка
// @Description Получает текущий статус умного замка
// @Tags locks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uniqueId path string true "Уникальный ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /locks/status/{uniqueId} [get]
func (h *LockHandler) GetLockStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("необходима авторизация"))
		return
	}

	uniqueID := c.Param("uniqueId")

	canAccess, err := h.canUserAccessLock(userID.(int), uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("нет прав для просмотра статуса этого замка"))
		return
	}

	lock, err := h.lockUseCase.GetLockStatus(uniqueID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("замок не найден"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статус замка", gin.H{
		"unique_id":          lock.UniqueID,
		"current_status":     lock.CurrentStatus,
		"last_status_update": lock.LastStatusUpdate,
		"last_heartbeat":     lock.LastHeartbeat,
		"is_online":          lock.IsOnline,
		"battery_level":      lock.BatteryLevel,
		"signal_strength":    lock.SignalStrength,
	}))
}

func (h *LockHandler) GetLockHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("необходима авторизация"))
		return
	}

	uniqueID := c.Param("uniqueId")

	canAccess, err := h.canUserAccessLock(userID.(int), uniqueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("нет прав для просмотра истории этого замка"))
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	history, err := h.lockUseCase.GetLockHistory(uniqueID, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("замок не найден"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("история замка", history))
}

// @Summary Обновление статуса замка устройством
// @Description Обновляет статус умного замка от самого устройства (для интеграции с Tuya)
// @Tags device-locks
// @Accept json
// @Produce json
// @Param request body domain.LockStatusUpdateRequest true "Данные статуса замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Router /device/locks/status [post]
func (h *LockHandler) UpdateLockStatus(c *gin.Context) {
	var request domain.LockStatusUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных"))
		return
	}

	if request.Timestamp.IsZero() {
		request.Timestamp = utils.GetCurrentTimeUTC()
	}

	err := h.lockUseCase.UpdateLockStatus(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статус обновлен", nil))
}

// @Summary Heartbeat от замка
// @Description Обновляет heartbeat и статус замка от устройства (для мониторинга онлайн/офлайн)
// @Tags device-locks
// @Accept json
// @Produce json
// @Param request body domain.LockHeartbeatRequest true "Данные heartbeat замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Router /device/locks/heartbeat [post]
func (h *LockHandler) ProcessHeartbeat(c *gin.Context) {
	var request domain.LockHeartbeatRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных"))
		return
	}

	if request.Timestamp.IsZero() {
		request.Timestamp = utils.GetCurrentTimeUTC()
	}

	err := h.lockUseCase.ProcessHeartbeat(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("heartbeat обработан", nil))
}

func (h *LockHandler) GetBookingUseCase() (domain.BookingUseCase, bool) {
	return h.bookingUseCase, true
}

func (h *LockHandler) canUserAccessLock(userID int, uniqueID string) (bool, error) {
	user, err := h.userUseCase.GetByID(userID)
	if err != nil {
		return false, err
	}

	if user.Role == domain.RoleAdmin || user.Role == domain.RoleModerator {
		return true, nil
	}

	return h.lockUseCase.CanUserControlLock(uniqueID, userID)
}

// @Summary Получение всех замков (админ)
// @Description Возвращает список всех замков в системе с расширенной информацией и фильтрами (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "Статус замка (closed, open)"
// @Param min_battery_level query int false "Минимальный уровень батареи (0-100)"
// @Param max_battery_level query int false "Максимальный уровень батареи (0-100)"
// @Param is_online query bool false "Статус онлайн (true - онлайн, false - офлайн)"
// @Param apartment_id query int false "ID квартиры"
// @Param unbound query bool false "Замки без привязки к квартире (true - только непривязанные)"
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/locks [get]
func (h *LockHandler) AdminGetAllLocks(c *gin.Context) {
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

	if minBatteryLevelStr := c.Query("min_battery_level"); minBatteryLevelStr != "" {
		if minBatteryLevel, err := strconv.Atoi(minBatteryLevelStr); err == nil && minBatteryLevel >= 0 {
			filters["min_battery_level"] = minBatteryLevel
		}
	}

	if maxBatteryLevelStr := c.Query("max_battery_level"); maxBatteryLevelStr != "" {
		if maxBatteryLevel, err := strconv.Atoi(maxBatteryLevelStr); err == nil && maxBatteryLevel >= 0 {
			filters["max_battery_level"] = maxBatteryLevel
		}
	}

	if isOnlineStr := c.Query("is_online"); isOnlineStr != "" {
		if isOnline, err := strconv.ParseBool(isOnlineStr); err == nil {
			filters["is_online"] = isOnline
		}
	}

	if apartmentIDStr := c.Query("apartment_id"); apartmentIDStr != "" {
		if apartmentID, err := strconv.Atoi(apartmentIDStr); err == nil && apartmentID > 0 {
			filters["apartment_id"] = apartmentID
		}
	}

	if unboundStr := c.Query("unbound"); unboundStr != "" {
		if unbound, err := strconv.ParseBool(unboundStr); err == nil && unbound {
			filters["unbound"] = true
		}
	}

	locks, total, err := h.lockUseCase.GetAllWithFilters(filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	statusStats := make(map[string]int)
	batteryStats := make(map[string]int)

	enrichedLocks := make([]map[string]interface{}, 0, len(locks))
	for _, lock := range locks {
		if lock.IsOnline {
			statusStats["online"]++
		} else {
			statusStats["offline"]++
		}

		if lock.BatteryLevel != nil {
			batteryLevel := *lock.BatteryLevel
			if batteryLevel >= 80 {
				batteryStats["high"]++
			} else if batteryLevel >= 40 {
				batteryStats["medium"]++
			} else if batteryLevel >= 20 {
				batteryStats["low"]++
			} else {
				batteryStats["critical"]++
			}
		} else {
			batteryStats["unknown"]++
		}

		lockData := map[string]interface{}{
			"id":              lock.ID,
			"unique_id":       lock.UniqueID,
			"name":            lock.Name,
			"apartment_id":    lock.ApartmentID,
			"status":          lock.CurrentStatus,
			"battery_level":   lock.BatteryLevel,
			"last_heartbeat":  lock.LastHeartbeat,
			"signal_strength": lock.SignalStrength,
			"is_online":       lock.IsOnline,
			"created_at":      lock.CreatedAt,
			"updated_at":      lock.UpdatedAt,
		}

		if lock.ApartmentID != nil {
			apartment, err := h.apartmentRepo.GetByID(*lock.ApartmentID)
			if err == nil {
				lockData["apartment_info"] = gin.H{
					"street":           apartment.Street,
					"building":         apartment.Building,
					"apartment_number": apartment.ApartmentNumber,
				}
			}
		}

		if tempPasswords, err := h.lockUseCase.GetTempPasswordsByLockID(lock.ID); err == nil {
			activeCount := 0
			now := utils.GetCurrentTimeUTC()
			for _, pwd := range tempPasswords {
				if pwd.IsActive && pwd.ValidUntil.After(now) {
					activeCount++
				}
			}
			lockData["active_passwords_count"] = activeCount
			lockData["total_passwords_count"] = len(tempPasswords)
		}

		enrichedLocks = append(enrichedLocks, lockData)
	}

	response := gin.H{
		"locks": enrichedLocks,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
			"pages":     (total + pageSize - 1) / pageSize,
		},
		"filters": filters,
		"statistics": gin.H{
			"total_locks": total,
			"by_status":   statusStats,
			"by_battery":  batteryStats,
		},
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("список замков получен", response))
}

// @Summary Получение замка по ID (админ)
// @Description Возвращает детальную информацию о замке с расширенными данными (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/locks/{id} [get]
func (h *LockHandler) AdminGetLockByID(c *gin.Context) {
	lockID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	lock, err := h.lockUseCase.GetLockByID(lockID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("замок не найден"))
		return
	}
	response := gin.H{
		"lock": lock,
	}

	if lock.ApartmentID != nil {
		apartment, err := h.apartmentRepo.GetByID(*lock.ApartmentID)
		if err == nil {
			response["apartment"] = apartment
		}
	}

	// Активные пароли
	if tempPasswords, err := h.lockUseCase.GetTempPasswordsByLockID(lock.ID); err == nil {
		response["temp_passwords"] = tempPasswords
	}

	// История событий (ограничиваем до 50 записей)
	if history, err := h.lockUseCase.GetLockHistory(lock.UniqueID, 50); err == nil {
		response["history"] = history
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("замок получен", response))
}

type AdminBindLockRequest struct {
	ApartmentID int `json:"apartment_id" binding:"required"`
}

// @Summary Привязка замка к квартире (админ)
// @Description Привязывает замок к квартире (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID замка"
// @Param request body AdminBindLockRequest true "ID квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/locks/{id}/bind-apartment [post]
func (h *LockHandler) AdminBindLockToApartment(c *gin.Context) {
	lockID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req AdminBindLockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверные данные запроса: "+err.Error()))
		return
	}

	err := h.lockUseCase.BindLockToApartment(lockID, req.ApartmentID)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
		} else {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("замок успешно привязан к квартире", gin.H{
		"lock_id":      lockID,
		"apartment_id": req.ApartmentID,
	}))
}

// @Summary Отвязка замка от квартиры (админ)
// @Description Отвязывает замок от квартиры (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/locks/{id}/unbind-apartment [delete]
func (h *LockHandler) AdminUnbindLockFromApartment(c *gin.Context) {
	lockID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	err := h.lockUseCase.UnbindLockFromApartment(lockID)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
		} else {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("замок успешно отвязан от квартиры", gin.H{
		"lock_id": lockID,
	}))
}

// @Summary Экстренный сброс замка (админ)
// @Description Экстренный сброс замка с деактивацией всех паролей (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/locks/{id}/emergency-reset [put]
func (h *LockHandler) AdminEmergencyResetLock(c *gin.Context) {
	lockID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	err := h.lockUseCase.EmergencyResetLock(lockID)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("экстренный сброс замка выполнен", gin.H{
		"lock_id": lockID,
		"message": "Все временные пароли деактивированы, замок переведен в состояние 'закрыто'",
	}))
}

// @Summary Статистика замков (админ)
// @Description Возвращает общую статистику по замкам в системе (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/locks/statistics [get]
func (h *LockHandler) AdminGetLocksStatistics(c *gin.Context) {
	locks, err := h.lockUseCase.GetAllLocks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	statistics := gin.H{
		"total_locks":            len(locks),
		"online_locks":           0,
		"offline_locks":          0,
		"error_locks":            0,
		"bound_locks":            0,
		"unbound_locks":          0,
		"total_active_passwords": 0,
		"low_battery_locks":      0,
	}

	now := utils.GetCurrentTimeUTC()
	for _, lock := range locks {
		if lock.IsOnline {
			statistics["online_locks"] = statistics["online_locks"].(int) + 1
		} else {
			statistics["offline_locks"] = statistics["offline_locks"].(int) + 1
		}

		if lock.ApartmentID != nil {
			statistics["bound_locks"] = statistics["bound_locks"].(int) + 1
		} else {
			statistics["unbound_locks"] = statistics["unbound_locks"].(int) + 1
		}

		if lock.BatteryLevel != nil && *lock.BatteryLevel < 20 {
			statistics["low_battery_locks"] = statistics["low_battery_locks"].(int) + 1
		}

		if tempPasswords, err := h.lockUseCase.GetTempPasswordsByLockID(lock.ID); err == nil {
			for _, pwd := range tempPasswords {
				if pwd.IsActive && pwd.ValidUntil.After(now) {
					statistics["total_active_passwords"] = statistics["total_active_passwords"].(int) + 1
				}
			}
		}
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статистика замков", statistics))
}

// @Summary Генерация временного пароля для замка (админ)
// @Description Создает временный пароль для замка с указанным периодом действия (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uniqueId path string true "Уникальный ID замка"
// @Param request body domain.AdminGeneratePasswordRequest true "Параметры пароля"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/locks/by-unique-id/{uniqueId}/passwords [post]
func (h *LockHandler) AdminGeneratePassword(c *gin.Context) {
	uniqueID := c.Param("uniqueId")

	var request domain.AdminGeneratePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверные данные запроса: "+err.Error()))
		return
	}

	password, err := h.lockUseCase.AdminGeneratePassword(uniqueID, &request)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
		} else {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusCreated, domain.NewSuccessResponse("пароль создан", password))
}

// @Summary Получение всех паролей замка (админ)
// @Description Возвращает все временные пароли замка с информацией о привязке (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uniqueId path string true "Уникальный ID замка"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/locks/by-unique-id/{uniqueId}/passwords [get]
func (h *LockHandler) AdminGetAllLockPasswords(c *gin.Context) {
	uniqueID := c.Param("uniqueId")

	passwords, err := h.lockUseCase.AdminGetAllLockPasswords(uniqueID)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
		} else {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		}
		return
	}

	enrichedPasswords := make([]gin.H, 0, len(passwords))
	for _, pwd := range passwords {
		passwordInfo := gin.H{
			"id":           pwd.ID,
			"lock_id":      pwd.LockID,
			"password":     pwd.Password,
			"name":         pwd.Name,
			"valid_from":   pwd.ValidFrom,
			"valid_until":  pwd.ValidUntil,
			"is_active":    pwd.IsActive,
			"created_at":   pwd.CreatedAt,
			"updated_at":   pwd.UpdatedAt,
			"binding_type": "manual",
			"binding_info": nil,
		}

		if pwd.BookingID != nil {
			passwordInfo["binding_type"] = "booking"
			if bookingUC, ok := h.GetBookingUseCase(); ok {
				if booking, err := bookingUC.GetBookingByID(*pwd.BookingID); err == nil {
					passwordInfo["binding_info"] = gin.H{
						"booking_id":    booking.ID,
						"booking_dates": fmt.Sprintf("%s - %s", booking.StartDate.Format("2006-01-02"), booking.EndDate.Format("2006-01-02")),
						"renter_id":     booking.RenterID,
					}
				}
			}
		} else if pwd.UserID != nil {
			passwordInfo["binding_type"] = "user"
			if user, err := h.userUseCase.GetByID(*pwd.UserID); err == nil {
				passwordInfo["binding_info"] = gin.H{
					"user_id":    user.ID,
					"user_name":  user.FirstName + " " + user.LastName,
					"user_email": user.Email,
				}
			}
		}

		enrichedPasswords = append(enrichedPasswords, passwordInfo)
	}

	response := gin.H{
		"lock_unique_id": uniqueID,
		"passwords":      enrichedPasswords,
		"total_count":    len(passwords),
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("пароли замка", response))
}

// @Summary Деактивация пароля (админ)
// @Description Деактивирует временный пароль замка (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param passwordId path int true "ID пароля"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/passwords/{passwordId}/deactivate [post]
func (h *LockHandler) AdminDeactivatePassword(c *gin.Context) {
	passwordID, ok := utils.ParseIDParam(c, "passwordId")
	if !ok {
		return
	}

	err := h.lockUseCase.AdminDeactivatePassword(passwordID)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
		} else {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("пароль деактивирован", gin.H{
		"password_id": passwordID,
	}))
}
