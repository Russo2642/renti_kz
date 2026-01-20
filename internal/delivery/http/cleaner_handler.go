package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type CleanerHandler struct {
	cleanerUseCase domain.CleanerUseCase
}

func NewCleanerHandler(cleanerUseCase domain.CleanerUseCase) *CleanerHandler {
	return &CleanerHandler{
		cleanerUseCase: cleanerUseCase,
	}
}

// @Summary Создать уборщицу
// @Description Создание новой уборщицы (только для админов)
// @Tags Admin - Cleaners
// @Accept json
// @Produce json
// @Param request body domain.CreateCleanerRequest true "Данные для создания уборщицы"
// @Success 201 {object} domain.SuccessResponse{data=domain.Cleaner}
// @Security ApiKeyAuth
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/cleaners [post]
func (h *CleanerHandler) AdminCreateCleaner(c *gin.Context) {
	var request domain.CreateCleanerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный формат данных: " + err.Error(),
		})
		return
	}

	adminID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Success: false,
			Error:   "Необходима авторизация",
		})
		return
	}

	cleaner, err := h.cleanerUseCase.CreateCleaner(&request, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка создания уборщицы: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.SuccessResponse{
		Success: true,
		Message: "Уборщица успешно создана",
		Data:    cleaner,
	})
}

// @Summary Получить всех уборщиц
// @Description Получение списка всех уборщиц с фильтрацией (только для админов)
// @Tags Admin - Cleaners
// @Accept json
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Param is_active query bool false "Фильтр по активности"
// @Param user_id query int false "Фильтр по ID пользователя"
// @Param apartment_id query int false "Фильтр по ID квартиры"
// @Security ApiKeyAuth
// @Success 200 {object} domain.PaginatedResponse{data=[]domain.Cleaner}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/cleaners [get]
func (h *CleanerHandler) AdminGetAllCleaners(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filters := make(map[string]interface{})
	if isActive := c.Query("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filters["is_active"] = active
		}
	}
	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.Atoi(userID); err == nil {
			filters["user_id"] = id
		}
	}
	if apartmentID := c.Query("apartment_id"); apartmentID != "" {
		if id, err := strconv.Atoi(apartmentID); err == nil {
			filters["apartment_id"] = id
		}
	}

	cleaners, total, err := h.cleanerUseCase.GetAllCleaners(filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка получения уборщиц: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.PaginatedResponse{
		Success: true,
		Data:    cleaners,
		Page:    page,
		PerPage: pageSize,
		Total:   total,
	})
}

// @Summary Получить уборщицу по ID
// @Description Получение информации об уборщице по ID (только для админов)
// @Tags Admin - Cleaners
// @Accept json
// @Produce json
// @Param id path int true "ID уборщицы"
// @Success 200 {object} domain.SuccessResponse{data=domain.Cleaner}
// @Security ApiKeyAuth
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/cleaners/{id} [get]
func (h *CleanerHandler) AdminGetCleanerByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный ID уборщицы",
		})
		return
	}

	cleaner, err := h.cleanerUseCase.GetCleanerByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{
			Success: false,
			Error:   "Уборщица не найдена: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Data:    cleaner,
	})
}

// @Summary Обновить уборщицу
// @Description Обновление информации об уборщице (только для админов)
// @Tags Admin - Cleaners
// @Accept json
// @Produce json
// @Param id path int true "ID уборщицы"
// @Param request body domain.UpdateCleanerRequest true "Данные для обновления"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/cleaners/{id} [put]
func (h *CleanerHandler) AdminUpdateCleaner(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный ID уборщицы",
		})
		return
	}

	var request domain.UpdateCleanerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный формат данных: " + err.Error(),
		})
		return
	}

	adminID, _ := utils.GetUserIDFromContext(c)
	err = h.cleanerUseCase.UpdateCleaner(id, &request, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка обновления уборщицы: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Уборщица успешно обновлена",
	})
}

// @Summary Удалить уборщицу
// @Description Удаление уборщицы (только для админов)
// @Tags Admin - Cleaners
// @Accept json
// @Produce json
// @Param id path int true "ID уборщицы"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/cleaners/{id} [delete]
func (h *CleanerHandler) AdminDeleteCleaner(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный ID уборщицы",
		})
		return
	}

	adminID, _ := utils.GetUserIDFromContext(c)
	err = h.cleanerUseCase.DeleteCleaner(id, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка удаления уборщицы: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Уборщица успешно удалена",
	})
}

// @Summary Назначить уборщицу на квартиру
// @Description Назначение уборщицы на квартиру (только для админов)
// @Tags Admin - Cleaners
// @Accept json
// @Produce json
// @Param request body domain.AssignCleanerToApartmentRequest true "Данные для назначения"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/cleaners/assign [post]
func (h *CleanerHandler) AdminAssignCleanerToApartment(c *gin.Context) {
	var request domain.AssignCleanerToApartmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный формат данных: " + err.Error(),
		})
		return
	}

	adminID, _ := utils.GetUserIDFromContext(c)
	err := h.cleanerUseCase.AssignCleanerToApartment(request.CleanerID, request.ApartmentID, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка назначения уборщицы: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Уборщица успешно назначена на квартиру",
	})
}

// @Summary Убрать уборщицу с квартиры
// @Description Удаление назначения уборщицы с квартиры (только для админов)
// @Tags Admin - Cleaners
// @Accept json
// @Produce json
// @Param request body domain.RemoveCleanerFromApartmentRequest true "Данные для удаления назначения"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/cleaners/remove [post]
func (h *CleanerHandler) AdminRemoveCleanerFromApartment(c *gin.Context) {
	var request domain.RemoveCleanerFromApartmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный формат данных: " + err.Error(),
		})
		return
	}

	adminID, _ := utils.GetUserIDFromContext(c)
	err := h.cleanerUseCase.RemoveCleanerFromApartment(request.CleanerID, request.ApartmentID, adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка удаления назначения уборщицы: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Уборщица успешно удалена с квартиры",
	})
}

// @Summary Получить квартиры, нуждающиеся в уборке
// @Description Получение списка всех квартир, нуждающихся в уборке (только для админов)
// @Tags Admin - Cleaners
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse{data=[]domain.ApartmentForCleaning}
// @Security ApiKeyAuth
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/cleaners/apartments-needing-cleaning [get]
func (h *CleanerHandler) AdminGetApartmentsNeedingCleaning(c *gin.Context) {
	apartments, err := h.cleanerUseCase.GetApartmentsNeedingCleaning()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка получения квартир для уборки: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Data:    apartments,
	})
}

// @Summary Получить профиль уборщицы
// @Description Получение профиля уборщицы с расписанием
// @Tags cleaner
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse{data=domain.Cleaner}
// @Security ApiKeyAuth
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /cleaner/profile [get]
func (h *CleanerHandler) GetCleanerProfile(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Success: false,
			Error:   "Необходима авторизация",
		})
		return
	}

	cleaner, err := h.cleanerUseCase.GetCleanerByUserID(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка получения профиля: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Data:    cleaner,
	})
}

// @Summary Получить мои квартиры
// @Description Получение списка квартир, назначенных уборщице
// @Tags cleaner
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse{data=[]domain.Apartment}
// @Security ApiKeyAuth
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /cleaner/apartments [get]
func (h *CleanerHandler) GetCleanerApartments(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Success: false,
			Error:   "Необходима авторизация",
		})
		return
	}

	apartments, err := h.cleanerUseCase.GetCleanerApartments(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка получения квартир: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Data:    apartments,
	})
}

// @Summary Получить квартиры для уборки
// @Description Получение списка квартир, которые нуждаются в уборке
// @Tags cleaner
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse{data=[]domain.ApartmentForCleaning}
// @Security ApiKeyAuth
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /cleaner/apartments-for-cleaning [get]
func (h *CleanerHandler) GetApartmentsForCleaning(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Success: false,
			Error:   "Необходима авторизация",
		})
		return
	}

	apartments, err := h.cleanerUseCase.GetApartmentsForCleaning(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка получения квартир для уборки: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Data:    apartments,
	})
}

// @Summary Получить статистику уборщицы
// @Description Получение статистики работы уборщицы
// @Tags cleaner
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse{data=map[string]interface{}}
// @Security ApiKeyAuth
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /cleaner/stats [get]
func (h *CleanerHandler) GetCleanerStats(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Success: false,
			Error:   "Необходима авторизация",
		})
		return
	}

	stats, err := h.cleanerUseCase.GetCleanerStats(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка получения статистики: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Data:    stats,
	})
}

// @Summary Начать уборку
// @Description Отметить начало уборки квартиры
// @Tags cleaner
// @Accept json
// @Produce json
// @Param request body domain.StartCleaningRequest true "Данные для начала уборки"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /cleaner/start-cleaning [post]
func (h *CleanerHandler) StartCleaning(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Success: false,
			Error:   "Необходима авторизация",
		})
		return
	}

	var request domain.StartCleaningRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный формат данных: " + err.Error(),
		})
		return
	}

	err := h.cleanerUseCase.StartCleaning(userID, &request)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка начала уборки: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Уборка начата",
	})
}

// @Summary Завершить уборку
// @Description Отметить завершение уборки квартиры и освободить её
// @Tags cleaner
// @Accept json
// @Produce json
// @Param request body domain.CompleteCleaningRequest true "Данные для завершения уборки"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /cleaner/complete-cleaning [post]
func (h *CleanerHandler) CompleteCleaning(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Success: false,
			Error:   "Необходима авторизация",
		})
		return
	}

	var request domain.CompleteCleaningRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный формат данных: " + err.Error(),
		})
		return
	}

	err := h.cleanerUseCase.CompleteCleaning(userID, &request)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка завершения уборки: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Уборка завершена, квартира освобождена",
	})
}

// @Summary Полное обновление расписания
// @Description Полностью заменяет расписание работы уборщицы. ВНИМАНИЕ: Все дни, которые не переданы в запросе, будут очищены! Для частичного обновления используйте PATCH /cleaner/schedule/patch
// @Tags cleaner
// @Accept json
// @Produce json
// @Param schedule body domain.CleanerSchedule true "Полное расписание на всю неделю"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /cleaner/schedule [put]
func (h *CleanerHandler) UpdateCleanerSchedule(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Success: false,
			Error:   "Необходима авторизация",
		})
		return
	}

	var schedule domain.CleanerSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный формат данных: " + err.Error(),
		})
		return
	}

	err := h.cleanerUseCase.UpdateCleanerSchedule(userID, &schedule)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка обновления расписания: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Расписание успешно обновлено",
	})
}

// UpdateCleanerSchedulePatch Частичное обновление расписания уборщицы (patch - только переданные дни)
// @Summary Частичное обновление расписания уборщицы
// @Description Обновляет только переданные дни недели в расписании, остальные дни остаются без изменений. Для очистки дня передайте пустой массив [], для игнорирования дня не передавайте поле вообще
// @Tags cleaner
// @Accept json
// @Produce json
// @Param schedule body domain.CleanerSchedulePatch true "Данные расписания (только нужные дни). Используйте null для игнорирования дня, [] для очистки дня"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /cleaner/schedule/patch [patch]
func (h *CleanerHandler) UpdateCleanerSchedulePatch(c *gin.Context) {
	userID, _ := utils.GetUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Success: false,
			Error:   "Необходима авторизация",
		})
		return
	}

	var schedulePatch domain.CleanerSchedulePatch
	if err := c.ShouldBindJSON(&schedulePatch); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный формат данных: " + err.Error(),
		})
		return
	}

	err := h.cleanerUseCase.UpdateCleanerSchedulePatch(userID, &schedulePatch)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка частичного обновления расписания: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Расписание частично обновлено",
	})
}

// @Summary Частичное обновление расписания уборщицы (админ)
// @Description Обновляет только переданные дни недели в расписании уборщицы (только для админов)
// @Tags Admin - Cleaners
// @Accept json
// @Produce json
// @Param id path int true "ID уборщицы"
// @Param schedule body domain.CleanerSchedulePatch true "Данные расписания (только нужные дни). Используйте null для игнорирования дня, [] для очистки дня"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/cleaners/{id}/schedule [patch]
func (h *CleanerHandler) AdminUpdateCleanerSchedulePatch(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный ID уборщицы",
		})
		return
	}

	var schedulePatch domain.CleanerSchedulePatch
	if err := c.ShouldBindJSON(&schedulePatch); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Неверный формат данных: " + err.Error(),
		})
		return
	}

	// Получаем cleaner по ID и берем userID
	cleaner, err := h.cleanerUseCase.GetCleanerByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{
			Success: false,
			Error:   "Уборщица не найдена: " + err.Error(),
		})
		return
	}

	err = h.cleanerUseCase.UpdateCleanerSchedulePatch(cleaner.UserID, &schedulePatch)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "Ошибка частичного обновления расписания: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Message: "Расписание уборщицы частично обновлено",
	})
}

func (h *CleanerHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.POST("/cleaners", h.AdminCreateCleaner)
	group.GET("/cleaners", h.AdminGetAllCleaners)
	group.GET("/cleaners/:id", h.AdminGetCleanerByID)
	group.PUT("/cleaners/:id", h.AdminUpdateCleaner)
	group.DELETE("/cleaners/:id", h.AdminDeleteCleaner)
	group.POST("/cleaners/assign", h.AdminAssignCleanerToApartment)
	group.POST("/cleaners/remove", h.AdminRemoveCleanerFromApartment)
	group.GET("/cleaners/apartments-needing-cleaning", h.AdminGetApartmentsNeedingCleaning)
	group.PATCH("/cleaners/:id/schedule", h.AdminUpdateCleanerSchedulePatch)
}

func (h *CleanerHandler) RegisterCleanerRoutes(group *gin.RouterGroup) {
	group.GET("/profile", h.GetCleanerProfile)
	group.GET("/apartments", h.GetCleanerApartments)
	group.GET("/apartments-for-cleaning", h.GetApartmentsForCleaning)
	group.GET("/stats", h.GetCleanerStats)
	group.POST("/start-cleaning", h.StartCleaning)
	group.POST("/complete-cleaning", h.CompleteCleaning)
	group.PUT("/schedule", h.UpdateCleanerSchedule)
	group.PATCH("/schedule/patch", h.UpdateCleanerSchedulePatch)
}
