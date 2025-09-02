package http

import (
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/services"
	"github.com/russo2642/renti_kz/internal/utils"
	"github.com/russo2642/renti_kz/pkg/logger"
)

type ApartmentHandler struct {
	apartmentUseCase     domain.ApartmentUseCase
	userUseCase          domain.UserUseCase
	ownerUseCase         domain.PropertyOwnerUseCase
	locationUseCase      domain.LocationUseCase
	notificationUseCase  domain.NotificationUseCase
	settingsUseCase      domain.PlatformSettingsUseCase
	bookingUseCase       domain.BookingUseCase
	lockUseCase          domain.LockUseCase
	middleware           *Middleware
	userRepo             domain.UserRepository
	ownerRepo            domain.PropertyOwnerRepository
	roleRepo             domain.RoleRepository
	responseCacheService *services.ResponseCacheService
}

func NewApartmentHandler(
	apartmentUseCase domain.ApartmentUseCase,
	userUseCase domain.UserUseCase,
	ownerUseCase domain.PropertyOwnerUseCase,
	locationUseCase domain.LocationUseCase,
	notificationUseCase domain.NotificationUseCase,
	settingsUseCase domain.PlatformSettingsUseCase,
	bookingUseCase domain.BookingUseCase,
	lockUseCase domain.LockUseCase,
	middleware *Middleware,
	userRepo domain.UserRepository,
	ownerRepo domain.PropertyOwnerRepository,
	roleRepo domain.RoleRepository,
	responseCacheService *services.ResponseCacheService,
) *ApartmentHandler {
	return &ApartmentHandler{
		apartmentUseCase:     apartmentUseCase,
		userUseCase:          userUseCase,
		ownerUseCase:         ownerUseCase,
		locationUseCase:      locationUseCase,
		notificationUseCase:  notificationUseCase,
		settingsUseCase:      settingsUseCase,
		bookingUseCase:       bookingUseCase,
		lockUseCase:          lockUseCase,
		middleware:           middleware,
		userRepo:             userRepo,
		ownerRepo:            ownerRepo,
		roleRepo:             roleRepo,
		responseCacheService: responseCacheService,
	}
}

func (h *ApartmentHandler) RegisterRoutes(router *gin.RouterGroup) {
	apartments := router.Group("/apartments")
	{

		apartments.GET("", h.middleware.OptionalAuthMiddleware(), h.GetAll)
		apartments.GET("/search/geo", h.GetByCoordinates)
		apartments.GET("/:id", h.middleware.OptionalAuthMiddleware(), h.GetByID)
		apartments.GET("/:id/photos", h.GetPhotosByApartmentID)
		apartments.GET("/:id/location", h.GetLocationByApartmentID)
		apartments.GET("/:id/available-durations", h.GetAvailableDurations)
		apartments.GET("/:id/can-book-now", h.CanBookNow)
		apartments.GET("/:id/calculate-price", h.CalculatePrice)
		apartments.GET("/:id/booked-dates", h.GetBookedDates)
		apartments.GET("/:id/availability", h.CheckApartmentAvailability)
		apartments.GET("/:id/available-slots", h.GetAvailableTimeSlots)

		authorized := apartments.Group("/", h.middleware.AuthMiddleware())
		{

			authorized.POST("", h.Create)
			authorized.PUT("/:id", h.Update)
			authorized.DELETE("/:id", h.Delete)

			authorized.POST("/:id/photos", h.AddPhotos)
			authorized.DELETE("/photos/:photoId", h.DeletePhoto)

			authorized.POST("/:id/location", h.AddLocation)
			authorized.PUT("/:id/location", h.UpdateLocation)

			authorized.GET("/:id/documents", h.GetDocumentsByApartmentID)
			authorized.POST("/:id/documents", h.AddDocuments)
			authorized.DELETE("/documents/:documentId", h.DeleteDocument)

			authorized.POST("/:id/confirm-agreement", h.ConfirmApartmentAgreement)

			authorized.GET("/owner/statistics", h.GetOwnerStatistics)
		}

		adminModerator := authorized.Group("/", h.middleware.RoleMiddleware(domain.RoleAdmin, domain.RoleModerator))
		{

			adminModerator.GET("/dashboard", h.GetDashboardStats)
		}
	}
}

type CreateApartmentRequest struct {
	CityID             int     `json:"city_id" binding:"required"`
	DistrictID         int     `json:"district_id" binding:"required"`
	MicrodistrictID    *int    `json:"microdistrict_id"`
	Street             string  `json:"street" binding:"required"`
	Building           string  `json:"building" binding:"required"`
	ApartmentNumber    int     `json:"apartment_number" binding:"required"`
	ResidentialComplex string  `json:"residential_complex"`
	RoomCount          int     `json:"room_count" binding:"required,min=1"`
	TotalArea          float64 `json:"total_area" binding:"required,gt=0"`
	KitchenArea        float64 `json:"kitchen_area" binding:"required,gt=0"`
	Floor              int     `json:"floor" binding:"required,min=1"`
	TotalFloors        int     `json:"total_floors" binding:"required,min=1"`
	ConditionID        int     `json:"condition_id" binding:"required"`
	Price              int     `json:"price" binding:"min=0"`
	DailyPrice         int     `json:"daily_price" binding:"min=0"`
	RentalTypeHourly   bool    `json:"rental_type_hourly"`
	RentalTypeDaily    bool    `json:"rental_type_daily"`
	Description        string  `json:"description"`
	ListingType        string  `json:"listing_type" binding:"required,oneof=owner realtor"`
	HouseRuleIDs       []int   `json:"house_rule_ids"`
	AmenityIDs         []int   `json:"amenity_ids"`

	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	PhotosBase64 []string `json:"photos_base64"`
}

type UpdateApartmentRequest struct {
	CityID             *int     `json:"city_id"`
	DistrictID         *int     `json:"district_id"`
	MicrodistrictID    *int     `json:"microdistrict_id"`
	Street             *string  `json:"street"`
	Building           *string  `json:"building"`
	ApartmentNumber    *int     `json:"apartment_number"`
	ResidentialComplex *string  `json:"residential_complex"`
	RoomCount          *int     `json:"room_count"`
	TotalArea          *float64 `json:"total_area"`
	KitchenArea        *float64 `json:"kitchen_area"`
	Floor              *int     `json:"floor"`
	TotalFloors        *int     `json:"total_floors"`
	ConditionID        *int     `json:"condition_id"`
	Price              *int     `json:"price"`
	DailyPrice         *int     `json:"daily_price"`
	RentalTypeHourly   *bool    `json:"rental_type_hourly"`
	RentalTypeDaily    *bool    `json:"rental_type_daily"`
	Description        *string  `json:"description"`
	ListingType        *string  `json:"listing_type"`
	HouseRuleIDs       []int    `json:"house_rule_ids"`
	AmenityIDs         []int    `json:"amenity_ids"`

	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	PhotosBase64 []string `json:"photos_base64"`
}

type LocationRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type UpdateStatusRequest struct {
	Status          string `json:"status" binding:"required"`
	Comment         string `json:"comment"`
	ApartmentTypeID *int   `json:"apartment_type_id"`
}

type GetByCoordinatesRequest struct {
	MinLat           float64  `form:"min_lat" binding:"required"`
	MaxLat           float64  `form:"max_lat" binding:"required"`
	MinLng           float64  `form:"min_lng" binding:"required"`
	MaxLng           float64  `form:"max_lng" binding:"required"`
	CityID           *int     `form:"city_id"`
	DistrictID       *int     `form:"district_id"`
	MicrodistrictID  *int     `form:"microdistrict_id"`
	RoomCount        *int     `form:"room_count"`
	MinArea          *float64 `form:"min_area"`
	MaxArea          *float64 `form:"max_area"`
	MinPrice         *int     `form:"min_price"`
	MaxPrice         *int     `form:"max_price"`
	RentalTypeHourly *bool    `form:"rental_type_hourly"`
	RentalTypeDaily  *bool    `form:"rental_type_daily"`
	IsFree           *bool    `form:"is_free"`
	Status           *string  `form:"status"`
	ListingType      *string  `form:"listing_type"`
	ApartmentTypeID  *int     `form:"apartment_type_id"`
}

// @Summary Создание новой квартиры
// @Description Создает новую квартиру для владельца
// @Tags apartments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body CreateApartmentRequest true "Данные квартиры"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments [post]
func (h *ApartmentHandler) Create(c *gin.Context) {
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

	if user.Role != domain.RoleUser && user.Role != domain.RoleOwner && user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для создания квартиры"))
		return
	}

	var ownerID int
	if user.Role == domain.RoleOwner {
		logger.Debug("user has owner role, getting owner profile", slog.Int("user_id", userIDInt))
		owner, err := h.ownerUseCase.GetByUserID(userIDInt)
		if err != nil {
			logger.Error("failed to get owner profile",
				slog.Int("user_id", userIDInt),
				slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных владельца"))
			return
		}
		if owner == nil {
			logger.Warn("owner profile not found", slog.Int("user_id", userIDInt))
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("профиль владельца не найден"))
			return
		}
		logger.Debug("owner profile retrieved", slog.Int("owner_id", owner.ID))
		ownerID = owner.ID
	} else if user.Role == domain.RoleUser {
		logger.Info("upgrading user to owner role", slog.Int("user_id", userIDInt))

		ownerRole, err := h.roleRepo.GetByName(string(domain.RoleOwner))
		if err != nil {
			logger.Error("failed to get owner role", slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении роли owner"))
			return
		}

		user.Role = domain.RoleOwner
		user.RoleID = ownerRole.ID
		if err := h.userRepo.Update(user); err != nil {
			logger.Error("failed to update user role",
				slog.Int("user_id", userIDInt),
				slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении роли пользователя"))
			return
		}

		propertyOwner := &domain.PropertyOwner{
			UserID: userIDInt,
		}

		if err := h.ownerRepo.Create(propertyOwner); err != nil {
			logger.Error("failed to create owner profile",
				slog.Int("user_id", userIDInt),
				slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при создании записи владельца"))
			return
		}

		logger.Info("owner profile created", slog.Int("owner_id", propertyOwner.ID))
		ownerID = propertyOwner.ID
	} else {
		ownerIDParam := c.Query("owner_id")
		if ownerIDParam == "" {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("требуется указать owner_id для создания квартиры"))
			return
		}
		var err error
		ownerID, err = strconv.Atoi(ownerIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный формат owner_id"))
			return
		}
	}

	var req CreateApartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректные данные запроса"))
		return
	}

	city, err := h.locationUseCase.GetCityByID(req.CityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке города"))
		return
	}
	if city == nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("указанный город не найден"))
		return
	}

	district, err := h.locationUseCase.GetDistrictByID(req.DistrictID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке района"))
		return
	}
	if district == nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("указанный район не найден"))
		return
	}

	if req.MicrodistrictID != nil {
		microdistrict, err := h.locationUseCase.GetMicrodistrictByID(*req.MicrodistrictID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке микрорайона"))
			return
		}
		if microdistrict == nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("указанный микрорайон не найден"))
			return
		}
	}

	condition, err := h.apartmentUseCase.GetConditionByID(req.ConditionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке состояния квартиры"))
		return
	}
	if condition == nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("указанное состояние квартиры не найдено"))
		return
	}

	apartment := &domain.Apartment{
		OwnerID:            ownerID,
		CityID:             req.CityID,
		DistrictID:         req.DistrictID,
		MicrodistrictID:    req.MicrodistrictID,
		Street:             req.Street,
		Building:           req.Building,
		ApartmentNumber:    req.ApartmentNumber,
		ResidentialComplex: utils.PreprocessResidentialComplexPointer(req.ResidentialComplex),
		RoomCount:          req.RoomCount,
		TotalArea:          req.TotalArea,
		KitchenArea:        req.KitchenArea,
		Floor:              req.Floor,
		TotalFloors:        req.TotalFloors,
		ConditionID:        req.ConditionID,
		Price:              req.Price,
		DailyPrice:         req.DailyPrice,
		RentalTypeHourly:   req.RentalTypeHourly,
		RentalTypeDaily:    req.RentalTypeDaily,
		IsFree:             true,
		Status:             domain.AptStatusPending,
		Description:        req.Description,
		ListingType:        req.ListingType,
	}

	if len(req.HouseRuleIDs) > 0 {
		apartment.HouseRules = make([]*domain.HouseRules, 0, len(req.HouseRuleIDs))
		for _, ruleID := range req.HouseRuleIDs {
			rule := &domain.HouseRules{ID: ruleID}
			apartment.HouseRules = append(apartment.HouseRules, rule)
		}
	}

	if len(req.AmenityIDs) > 0 {
		apartment.Amenities = make([]*domain.PopularAmenities, 0, len(req.AmenityIDs))
		for _, amenityID := range req.AmenityIDs {
			amenity := &domain.PopularAmenities{ID: amenityID}
			apartment.Amenities = append(apartment.Amenities, amenity)
		}
	}

	if err := h.apartmentUseCase.Create(apartment); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при создании квартиры: "+err.Error()))
		return
	}

	if h.notificationUseCase != nil {
		apartmentTitle := fmt.Sprintf("%s, д. %s, кв. %d", apartment.Street, apartment.Building, apartment.ApartmentNumber)
		err = h.notificationUseCase.NotifyApartmentCreated(user.ID, apartment.ID, apartmentTitle)
		if err != nil {
			logger.Warn("failed to send apartment creation notification",
				slog.Int("apartment_id", apartment.ID),
				slog.String("error", err.Error()))
		}
	}

	if req.Latitude != nil && req.Longitude != nil {
		location := &domain.ApartmentLocation{
			ApartmentID: apartment.ID,
			Latitude:    *req.Latitude,
			Longitude:   *req.Longitude,
		}
		if err := h.apartmentUseCase.AddLocation(location); err != nil {
			logger.Warn("failed to add coordinates to apartment",
				slog.Int("apartment_id", apartment.ID),
				slog.String("error", err.Error()))
		}
	}

	if len(req.PhotosBase64) > 0 {
		filesData := make([][]byte, 0, len(req.PhotosBase64))
		for i, photoBase64 := range req.PhotosBase64 {
			if strings.Contains(photoBase64, ",") {
				parts := strings.Split(photoBase64, ",")
				if len(parts) == 2 {
					photoBase64 = parts[1]
				}
			}

			photoData, err := base64.StdEncoding.DecodeString(photoBase64)
			if err != nil {
				logger.Warn("failed to decode photo",
					slog.Int("photo_index", i+1),
					slog.String("error", err.Error()))
				continue
			}

			if len(photoData) > 10*1024*1024 {
				logger.Warn("photo size exceeds limit",
					slog.Int("photo_index", i+1),
					slog.Int("size_mb", len(photoData)/(1024*1024)))
				continue
			}

			filesData = append(filesData, photoData)
		}

		if len(filesData) > 0 {
			if _, err := h.apartmentUseCase.AddPhotos(apartment.ID, filesData); err != nil {
				fmt.Printf("⚠️ Ошибка при добавлении фотографий к квартире %d: %v\n", apartment.ID, err)
			}
		}
	}

	fullApartment, err := h.apartmentUseCase.GetByID(apartment.ID)
	if err != nil {
		c.JSON(http.StatusCreated, domain.NewSuccessResponse("квартира успешно создана", apartment))
		return
	}

	enrichedApartment := h.enrichApartmentWithLocationData(fullApartment)
	c.JSON(http.StatusCreated, domain.NewSuccessResponse("квартира успешно создана", enrichedApartment))
}

// @Summary Получение квартиры по ID
// @Description Получает информацию о квартире по её идентификатору
// @Tags apartments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id} [get]
func (h *ApartmentHandler) GetByID(c *gin.Context) {

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	userID, exists := c.Get("user_id")
	var userIDPtr *int
	if exists {
		userIDInt := userID.(int)
		userIDPtr = &userIDInt
	}

	apartment, err := h.apartmentUseCase.GetByIDWithUserContext(id, userIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	if !exists {
		if apartment.Status != domain.AptStatusApproved {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
			return
		}
	} else {

		user, err := h.userUseCase.GetByID(userID.(int))
		if err != nil || user == nil {

			if apartment.Status != domain.AptStatusApproved {
				c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
				return
			}
		} else {

			if user.Role == domain.RoleAdmin || user.Role == domain.RoleModerator {
			} else {
				isOwner := false
				if user.Role == domain.RoleOwner {
					owner, err := h.ownerUseCase.GetByUserID(userID.(int))
					if err == nil && owner != nil && owner.ID == apartment.OwnerID {
						isOwner = true
					}
				}

				if !isOwner && apartment.Status != domain.AptStatusApproved {
					c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
					return
				}
			}
		}
	}

	go func() {
		if err := h.apartmentUseCase.IncrementViewCount(id); err != nil {
			// Логируем ошибку, но не прерываем выполнение
			// TODO: добавить логирование
		}
	}()

	enrichedApartment := h.enrichApartmentWithLocationData(apartment)

	c.JSON(http.StatusOK, domain.NewSuccessResponse("квартира успешно получена", enrichedApartment))
}

// @Summary Получение списка квартир
// @Description Получает список квартир с возможностью фильтрации и пагинации
// @Tags apartments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param city_id query int false "ID города"
// @Param district_id query int false "ID района"
// @Param microdistrict_id query int false "ID микрорайона"
// @Param room_count query int false "Количество комнат"
// @Param min_area query number false "Минимальная площадь"
// @Param max_area query number false "Максимальная площадь"
// @Param min_price query int false "Минимальная цена"
// @Param max_price query int false "Максимальная цена"
// @Param rental_type_hourly query bool false "Поддержка почасовой аренды"
// @Param rental_type_daily query bool false "Поддержка посуточной аренды"
// @Param apartment_type_id query int false "Тип квартиры (ID)"
// @Param is_free query bool false "Доступность квартиры (true - свободная, false - занятая)"
// @Param status query string false "Статус квартиры"
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(10)
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments [get]
func (h *ApartmentHandler) GetAll(c *gin.Context) {

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

	if cityID := c.Query("city_id"); cityID != "" {
		if id, err := strconv.Atoi(cityID); err == nil && id > 0 {
			filters["city_id"] = id
		}
	}

	if districtID := c.Query("district_id"); districtID != "" {
		if id, err := strconv.Atoi(districtID); err == nil && id > 0 {
			filters["district_id"] = id
		}
	}

	if microdistrictID := c.Query("microdistrict_id"); microdistrictID != "" {
		if id, err := strconv.Atoi(microdistrictID); err == nil && id > 0 {
			filters["microdistrict_id"] = id
		}
	}

	if roomCount := c.Query("room_count"); roomCount != "" {
		if count, err := strconv.Atoi(roomCount); err == nil && count > 0 {
			filters["room_count"] = count
		}
	}

	if minArea := c.Query("min_area"); minArea != "" {
		if area, err := strconv.ParseFloat(minArea, 64); err == nil && area > 0 {
			filters["min_area"] = area
		}
	}

	if maxArea := c.Query("max_area"); maxArea != "" {
		if area, err := strconv.ParseFloat(maxArea, 64); err == nil && area > 0 {
			filters["max_area"] = area
		}
	}

	if minPrice := c.Query("min_price"); minPrice != "" {
		if price, err := strconv.Atoi(minPrice); err == nil && price >= 0 {
			filters["min_price"] = price
		}
	}

	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if price, err := strconv.Atoi(maxPrice); err == nil && price >= 0 {
			filters["max_price"] = price
		}
	}

	if isFree := c.Query("is_free"); isFree != "" {
		if free, err := strconv.ParseBool(isFree); err == nil {
			filters["is_free"] = free
		}
	}

	if rentalTypeHourly := c.Query("rental_type_hourly"); rentalTypeHourly != "" {
		if hourly, err := strconv.ParseBool(rentalTypeHourly); err == nil {
			filters["rental_type_hourly"] = hourly
		}
	}

	if rentalTypeDaily := c.Query("rental_type_daily"); rentalTypeDaily != "" {
		if daily, err := strconv.ParseBool(rentalTypeDaily); err == nil {
			filters["rental_type_daily"] = daily
		}
	}

	if listingType := c.Query("listing_type"); listingType != "" {
		if listingType == domain.ListingTypeOwner || listingType == domain.ListingTypeRealtor {
			filters["listing_type"] = listingType
		}
	}

	if apartmentTypeID := c.Query("apartment_type_id"); apartmentTypeID != "" {
		if id, err := strconv.Atoi(apartmentTypeID); err == nil && id > 0 {
			filters["apartment_type_id"] = id
		}
	}

	userID, exists := c.Get("user_id")
	var userIDPtr *int
	if exists {
		userIDInt := userID.(int)
		userIDPtr = &userIDInt
	}

	var isAdminOrModerator bool = false

	if exists {

		user, err := h.userUseCase.GetByID(userID.(int))
		if err == nil && user != nil {

			if user.Role == domain.RoleAdmin || user.Role == domain.RoleModerator {
				isAdminOrModerator = true

				if status := c.Query("status"); status != "" {
					filters["status"] = status
				}
			}
		}
	}

	_, hasStatus := filters["status"]
	if !hasStatus {
		if !isAdminOrModerator {
			filters["status"] = domain.AptStatusApproved
		}
	}

	apartments, total, err := h.apartmentUseCase.GetAllWithUserContext(filters, page, pageSize, userIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении списка квартир"))
		return
	}

	enrichedApartments := h.enrichApartmentsWithLocationData(apartments)

	responseData := gin.H{
		"apartments": enrichedApartments,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"pages":     (total + pageSize - 1) / pageSize,
		},
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("список квартир успешно получен", responseData))
}

// @Summary Обновление квартиры
// @Description Обновляет информацию о квартире
// @Tags apartments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Param request body UpdateApartmentRequest true "Данные для обновления"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id} [put]
func (h *ApartmentHandler) Update(c *gin.Context) {

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

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator {

		if user.Role == domain.RoleOwner {
			owner, err := h.ownerUseCase.GetByUserID(userIDInt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных владельца"))
				return
			}
			if owner == nil || owner.ID != apartment.OwnerID {
				c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для обновления квартиры"))
				return
			}
		} else {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для обновления квартиры"))
			return
		}
	}

	var req UpdateApartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректные данные запроса"))
		return
	}

	if req.CityID == nil && req.DistrictID == nil && req.MicrodistrictID == nil &&
		req.Street == nil && req.Building == nil && req.ApartmentNumber == nil && req.ResidentialComplex == nil &&
		req.RoomCount == nil && req.TotalArea == nil && req.KitchenArea == nil &&
		req.Floor == nil && req.TotalFloors == nil && req.ConditionID == nil && req.Price == nil &&
		req.RentalTypeHourly == nil && req.RentalTypeDaily == nil &&
		req.Description == nil && req.ListingType == nil && req.HouseRuleIDs == nil && req.AmenityIDs == nil &&
		req.Latitude == nil && req.Longitude == nil && len(req.PhotosBase64) == 0 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("не указаны поля для обновления"))
		return
	}

	if req.CityID != nil {

		city, err := h.locationUseCase.GetCityByID(*req.CityID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке города"))
			return
		}
		if city == nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("указанный город не найден"))
			return
		}
		apartment.CityID = *req.CityID
	}

	if req.DistrictID != nil {

		district, err := h.locationUseCase.GetDistrictByID(*req.DistrictID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке района"))
			return
		}
		if district == nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("указанный район не найден"))
			return
		}
		apartment.DistrictID = *req.DistrictID
	}

	if req.MicrodistrictID != nil {
		microdistrict, err := h.locationUseCase.GetMicrodistrictByID(*req.MicrodistrictID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке микрорайона"))
			return
		}
		if microdistrict == nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("указанный микрорайон не найден"))
			return
		}
		apartment.MicrodistrictID = req.MicrodistrictID
	}

	if req.ConditionID != nil {

		condition, err := h.apartmentUseCase.GetConditionByID(*req.ConditionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке состояния квартиры"))
			return
		}
		if condition == nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("указанное состояние квартиры не найдено"))
			return
		}
		apartment.ConditionID = *req.ConditionID
	}

	if req.Street != nil {
		apartment.Street = *req.Street
	}
	if req.Building != nil {
		apartment.Building = *req.Building
	}
	if req.ApartmentNumber != nil {
		apartment.ApartmentNumber = *req.ApartmentNumber
	}
	if req.ResidentialComplex != nil {
		apartment.ResidentialComplex = utils.PreprocessResidentialComplexPointer(*req.ResidentialComplex)
	}
	if req.RoomCount != nil {
		if *req.RoomCount < 1 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("количество комнат должно быть больше 0"))
			return
		}
		apartment.RoomCount = *req.RoomCount
	}
	if req.TotalArea != nil {
		if *req.TotalArea <= 0 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("общая площадь должна быть больше 0"))
			return
		}
		apartment.TotalArea = *req.TotalArea
	}
	if req.KitchenArea != nil {
		if *req.KitchenArea <= 0 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("площадь кухни должна быть больше 0"))
			return
		}
		apartment.KitchenArea = *req.KitchenArea
	}
	if req.Floor != nil {
		if *req.Floor < 1 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("этаж должен быть больше 0"))
			return
		}
		apartment.Floor = *req.Floor
	}
	if req.TotalFloors != nil {
		if *req.TotalFloors < 1 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("количество этажей должно быть больше 0"))
			return
		}
		apartment.TotalFloors = *req.TotalFloors
	}
	if req.Price != nil {
		if *req.Price < 0 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("цена не может быть отрицательной"))
			return
		}
		apartment.Price = *req.Price
	}
	if req.DailyPrice != nil {
		if *req.DailyPrice < 0 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("цена за сутки не может быть отрицательной"))
			return
		}
		apartment.DailyPrice = *req.DailyPrice
	}

	if req.RentalTypeHourly != nil {
		apartment.RentalTypeHourly = *req.RentalTypeHourly
	}
	if req.RentalTypeDaily != nil {
		apartment.RentalTypeDaily = *req.RentalTypeDaily
	}

	if req.Description != nil {
		apartment.Description = *req.Description
	}

	if req.ListingType != nil {
		if *req.ListingType != domain.ListingTypeOwner && *req.ListingType != domain.ListingTypeRealtor {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный тип размещения объявления"))
			return
		}
		apartment.ListingType = *req.ListingType
	}

	if req.HouseRuleIDs != nil {
		apartment.HouseRules = make([]*domain.HouseRules, 0, len(req.HouseRuleIDs))
		for _, ruleID := range req.HouseRuleIDs {
			rule := &domain.HouseRules{ID: ruleID}
			apartment.HouseRules = append(apartment.HouseRules, rule)
		}
	}

	if req.AmenityIDs != nil {
		apartment.Amenities = make([]*domain.PopularAmenities, 0, len(req.AmenityIDs))
		for _, amenityID := range req.AmenityIDs {
			amenity := &domain.PopularAmenities{ID: amenityID}
			apartment.Amenities = append(apartment.Amenities, amenity)
		}
	}

	if req.Latitude != nil || req.Longitude != nil {
		existingLocation, err := h.apartmentUseCase.GetLocationByApartmentID(apartment.ID)
		if err != nil {
			if req.Latitude != nil && req.Longitude != nil {
				location := &domain.ApartmentLocation{
					ApartmentID: apartment.ID,
					Latitude:    *req.Latitude,
					Longitude:   *req.Longitude,
				}
				if err := h.apartmentUseCase.AddLocation(location); err != nil {
					fmt.Printf("⚠️ Ошибка при добавлении координат к квартире %d: %v\n", apartment.ID, err)
				}
			}
		} else if existingLocation != nil {
			if req.Latitude != nil {
				existingLocation.Latitude = *req.Latitude
			}
			if req.Longitude != nil {
				existingLocation.Longitude = *req.Longitude
			}
			if err := h.apartmentUseCase.UpdateLocation(existingLocation); err != nil {
				fmt.Printf("⚠️ Ошибка при обновлении координат квартиры %d: %v\n", apartment.ID, err)
			}
		}
	}

	if len(req.PhotosBase64) > 0 {
		filesData := make([][]byte, 0, len(req.PhotosBase64))
		for i, photoBase64 := range req.PhotosBase64 {
			if strings.Contains(photoBase64, ",") {
				parts := strings.Split(photoBase64, ",")
				if len(parts) == 2 {
					photoBase64 = parts[1]
				}
			}

			photoData, err := base64.StdEncoding.DecodeString(photoBase64)
			if err != nil {
				fmt.Printf("⚠️ Ошибка при декодировании фотографии %d: %v\n", i+1, err)
				continue
			}

			if len(photoData) > 10*1024*1024 {
				fmt.Printf("⚠️ Размер фотографии %d превышает 10MB\n", i+1)
				continue
			}

			filesData = append(filesData, photoData)
		}

		if len(filesData) > 0 {
			if _, err := h.apartmentUseCase.AddPhotos(apartment.ID, filesData); err != nil {
				fmt.Printf("⚠️ Ошибка при добавлении фотографий к квартире %d: %v\n", apartment.ID, err)
			}
		}
	}

	apartment.Status = domain.AptStatusPending

	if err := h.apartmentUseCase.Update(apartment); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении квартиры: "+err.Error()))
		return
	}

	if h.notificationUseCase != nil {
		apartmentTitle := fmt.Sprintf("%s, д. %s, кв. %d", apartment.Street, apartment.Building, apartment.ApartmentNumber)
		err = h.notificationUseCase.NotifyApartmentUpdated(user.ID, apartment.ID, apartmentTitle)
		if err != nil {
			logger.Warn("failed to send apartment update notification",
				slog.Int("apartment_id", apartment.ID),
				slog.String("error", err.Error()))
		}
	}

	fullApartment, err := h.apartmentUseCase.GetByID(apartment.ID)
	if err != nil {
		c.JSON(http.StatusOK, domain.NewSuccessResponse("квартира успешно обновлена", apartment))
		return
	}

	enrichedApartment := h.enrichApartmentWithLocationData(fullApartment)
	c.JSON(http.StatusOK, domain.NewSuccessResponse("квартира успешно обновлена", enrichedApartment))
}

// @Summary Удаление квартиры
// @Description Удаляет квартиру и все связанные файлы
// @Tags apartments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id} [delete]
func (h *ApartmentHandler) Delete(c *gin.Context) {

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

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator {

		if user.Role == domain.RoleOwner {
			owner, err := h.ownerUseCase.GetByUserID(userIDInt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных владельца"))
				return
			}
			if owner == nil || owner.ID != apartment.OwnerID {
				c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для удаления квартиры"))
				return
			}
		} else {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для удаления квартиры"))
			return
		}
	}

	if err := h.apartmentUseCase.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при удалении квартиры: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("квартира успешно удалена", nil))
}

func (h *ApartmentHandler) GetPhotosByApartmentID(c *gin.Context) {

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	userID, exists := c.Get("user_id")

	if !exists {
		if apartment.Status != domain.AptStatusApproved {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
			return
		}
	} else {

		user, err := h.userUseCase.GetByID(userID.(int))
		if err != nil || user == nil {

			if apartment.Status != domain.AptStatusApproved {
				c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
				return
			}
		} else {

			if user.Role == domain.RoleAdmin || user.Role == domain.RoleModerator {

			} else {

				if apartment.Status != domain.AptStatusApproved {
					c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
					return
				}
			}
		}
	}

	photos, err := h.apartmentUseCase.GetPhotosByApartmentID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении фотографий квартиры"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("фотографии квартиры успешно получены", photos))
}

func (h *ApartmentHandler) GetLocationByApartmentID(c *gin.Context) {

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	userID, exists := c.Get("user_id")

	if !exists {
		if apartment.Status != domain.AptStatusApproved {
			c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
			return
		}
	} else {

		user, err := h.userUseCase.GetByID(userID.(int))
		if err != nil || user == nil {

			if apartment.Status != domain.AptStatusApproved {
				c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
				return
			}
		} else {

			if user.Role == domain.RoleAdmin || user.Role == domain.RoleModerator {

			} else {

				if apartment.Status != domain.AptStatusApproved {
					c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
					return
				}
			}
		}
	}

	location, err := h.apartmentUseCase.GetLocationByApartmentID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении координат квартиры"))
		return
	}
	if location == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("координаты для данной квартиры не найдены"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("координаты квартиры успешно получены", location))
}

// @Summary Добавление координат квартиры
// @Description Добавляет географические координаты к указанной квартире
// @Tags apartments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Param request body LocationRequest true "Координаты"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id}/location [post]
func (h *ApartmentHandler) AddLocation(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	userIDInt := userID.(int)

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	user, err := h.userUseCase.GetByID(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator {

		if user.Role == domain.RoleOwner {
			owner, err := h.ownerUseCase.GetByUserID(userIDInt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных владельца"))
				return
			}
			if owner == nil || owner.ID != apartment.OwnerID {
				c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для добавления координат"))
				return
			}
		} else {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для добавления координат"))
			return
		}
	}

	var req LocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректные данные запроса"))
		return
	}

	if req.Latitude == nil || req.Longitude == nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("не указаны обе координаты"))
		return
	}

	location := &domain.ApartmentLocation{
		ApartmentID: id,
		Latitude:    *req.Latitude,
		Longitude:   *req.Longitude,
	}

	if err := h.apartmentUseCase.AddLocation(location); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при добавлении координат: "+err.Error()))
		return
	}

	c.JSON(http.StatusCreated, domain.NewSuccessResponse("координаты успешно добавлены", location))
}

// @Summary Обновление координат квартиры
// @Description Обновляет географические координаты указанной квартиры
// @Tags apartments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Param request body LocationRequest true "Координаты"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id}/location [put]
func (h *ApartmentHandler) UpdateLocation(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	userIDInt := userID.(int)

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	user, err := h.userUseCase.GetByID(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator {

		if user.Role == domain.RoleOwner {
			owner, err := h.ownerUseCase.GetByUserID(userIDInt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных владельца"))
				return
			}
			if owner == nil || owner.ID != apartment.OwnerID {
				c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для обновления координат"))
				return
			}
		} else {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для обновления координат"))
			return
		}
	}

	existingLocation, err := h.apartmentUseCase.GetLocationByApartmentID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении координат"))
		return
	}
	if existingLocation == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("координаты для данной квартиры не найдены"))
		return
	}

	var req LocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректные данные запроса"))
		return
	}

	if req.Latitude == nil && req.Longitude == nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("не указаны поля для обновления"))
		return
	}

	if req.Latitude != nil {
		existingLocation.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		existingLocation.Longitude = *req.Longitude
	}

	if err := h.apartmentUseCase.UpdateLocation(existingLocation); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении координат: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("координаты успешно обновлены", existingLocation))
}

// @Summary Обновление статуса квартиры
// @Description Обновляет статус квартиры (только для админов и модераторов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Param request body UpdateStatusRequest true "Новый статус и комментарий"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartments/{id}/status [put]
func (h *ApartmentHandler) UpdateStatus(c *gin.Context) {

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

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для обновления статуса"))
		return
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректные данные запроса"))
		return
	}

	fmt.Printf("DEBUG: Получен запрос на изменение статуса: %+v\n", req)
	fmt.Printf("DEBUG: Текущий статус квартиры: %s\n", apartment.Status)

	oldStatus := string(apartment.Status)

	if err := h.apartmentUseCase.UpdateStatus(id, domain.ApartmentStatus(req.Status), req.Comment); err != nil {
		fmt.Printf("DEBUG: Ошибка при обновлении статуса: %v\n", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении статуса: "+err.Error()))
		return
	}

	if req.ApartmentTypeID != nil {
		if err := h.apartmentUseCase.UpdateApartmentType(id, *req.ApartmentTypeID); err != nil {
			fmt.Printf("DEBUG: Ошибка при обновлении типа квартиры: %v\n", err)
		}
	}

	updatedApartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		fmt.Printf("DEBUG: Ошибка при получении обновленной квартиры: %v\n", err)
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении обновленных данных квартиры"))
		return
	}

	fmt.Printf("DEBUG: Статус квартиры после обновления: %s\n", updatedApartment.Status)

	if h.notificationUseCase != nil && oldStatus != req.Status {
		owner, err := h.ownerUseCase.GetByID(updatedApartment.OwnerID)
		if err == nil && owner != nil {
			apartmentTitle := fmt.Sprintf("%s, д. %s, кв. %d", updatedApartment.Street, updatedApartment.Building, updatedApartment.ApartmentNumber)

			switch req.Status {
			case "approved":
				err = h.notificationUseCase.NotifyApartmentApproved(owner.UserID, updatedApartment.ID, apartmentTitle)
			case "rejected":
				err = h.notificationUseCase.NotifyApartmentRejected(owner.UserID, updatedApartment.ID, apartmentTitle, req.Comment)
			default:
				err = h.notificationUseCase.NotifyApartmentStatusChanged(owner.UserID, updatedApartment.ID, apartmentTitle, oldStatus, req.Status)
			}

			if err != nil {
				fmt.Printf("⚠️ Ошибка отправки уведомления о смене статуса квартиры: %v\n", err)
			}
		} else {
			fmt.Printf("⚠️ Ошибка получения владельца квартиры для уведомления: %v\n", err)
		}
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статус квартиры успешно обновлен", updatedApartment))
}

// @Summary Добавление фотографий
// @Description Загружает фотографии для квартиры
// @Tags apartments
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Param files formData file true "Файлы фотографий"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id}/photos [post]
func (h *ApartmentHandler) AddPhotos(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	userIDInt := userID.(int)

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	user, err := h.userUseCase.GetByID(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator {

		if user.Role == domain.RoleOwner {
			owner, err := h.ownerUseCase.GetByUserID(userIDInt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных владельца"))
				return
			}
			if owner == nil || owner.ID != apartment.OwnerID {
				c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для добавления фотографий"))
				return
			}
		} else {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для добавления фотографий"))
			return
		}
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("ошибка при получении формы: "+err.Error()))
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("не загружено ни одной фотографии"))
		return
	}

	filesData := make([][]byte, 0, len(files))

	for _, file := range files {

		if file.Size > 10*1024*1024 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(fmt.Sprintf("размер файла %s превышает допустимый предел (10 MB)", file.Filename)))
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(fmt.Sprintf("ошибка при открытии файла %s: %v", file.Filename, err)))
			return
		}

		fileData, err := io.ReadAll(src)
		src.Close()

		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(fmt.Sprintf("ошибка при чтении файла %s: %v", file.Filename, err)))
			return
		}

		filesData = append(filesData, fileData)
	}

	urls, err := h.apartmentUseCase.AddPhotosParallel(id, filesData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при добавлении фотографий: "+err.Error()))
		return
	}

	c.JSON(http.StatusCreated, domain.NewSuccessResponse(fmt.Sprintf("успешно добавлено %d фотографий", len(urls)), gin.H{
		"urls":  urls,
		"count": len(urls),
	}))
}

// @Summary Удаление фотографии квартиры
// @Description Удаляет указанную фотографию квартиры
// @Tags apartments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param photoId path int true "ID фотографии"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/photos/{photoId} [delete]
func (h *ApartmentHandler) DeletePhoto(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	userIDInt := userID.(int)

	photoIDParam := c.Param("photoId")
	photoID, err := strconv.Atoi(photoIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID фотографии"))
		return
	}

	_, err = h.userUseCase.GetByID(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	if err := h.apartmentUseCase.DeletePhoto(photoID); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при удалении фотографии: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("фотография успешно удалена", nil))
}

// @Summary Добавление документов квартиры
// @Description Добавляет документы к указанной квартире
// @Tags apartments
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Param type formData string true "Тип документа" Enums(owner,realtor)
// @Param files formData file true "Файлы документов"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id}/documents [post]
func (h *ApartmentHandler) AddDocuments(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	user, err := h.userUseCase.GetByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	owner, err := h.ownerUseCase.GetByID(apartment.OwnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных владельца"))
		return
	}

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator && user.ID != owner.UserID {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для добавления документов к этой квартире"))
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("ошибка при получении формы: "+err.Error()))
		return
	}

	documentTypes := form.Value["type"]
	if len(documentTypes) == 0 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("не указан тип документа"))
		return
	}

	documentType := documentTypes[0]

	if documentType != domain.DocumentTypeOwner && documentType != domain.DocumentTypeRealtor {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный тип документа. Допустимые типы: '"+domain.DocumentTypeOwner+"', '"+domain.DocumentTypeRealtor+"'"))
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("не выбраны файлы для загрузки"))
		return
	}

	if len(files) < 1 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("необходимо загрузить минимум 1 документ"))
		return
	}

	currentDocs, err := h.apartmentUseCase.GetDocumentsByApartmentID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении текущих документов: "+err.Error()))
		return
	}

	if len(currentDocs)+len(files) > 10 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("превышено максимальное количество документов (10)"))
		return
	}

	filesData := make([][]byte, 0, len(files))

	for _, file := range files {

		if file.Size > 10*1024*1024 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse(fmt.Sprintf("размер файла %s превышает допустимый предел (10 MB)", file.Filename)))
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(fmt.Sprintf("ошибка при открытии файла %s: %v", file.Filename, err)))
			return
		}

		fileData, err := io.ReadAll(src)
		src.Close()

		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(fmt.Sprintf("ошибка при чтении файла %s: %v", file.Filename, err)))
			return
		}

		filesData = append(filesData, fileData)
	}

	urls, err := h.apartmentUseCase.AddDocumentsWithType(id, filesData, documentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при добавлении документов: "+err.Error()))
		return
	}

	c.JSON(http.StatusCreated, domain.NewSuccessResponse(fmt.Sprintf("успешно добавлено %d документов", len(urls)), gin.H{
		"urls":  urls,
		"count": len(urls),
		"type":  documentType,
	}))
}

func (h *ApartmentHandler) GetDocumentsByApartmentID(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("требуется авторизация"))
		return
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	user, err := h.userUseCase.GetByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	var canViewDocuments bool = false

	if user.Role == domain.RoleAdmin || user.Role == domain.RoleModerator {
		canViewDocuments = true
	} else if user.Role == domain.RoleOwner {

		owner, err := h.ownerUseCase.GetByUserID(userID.(int))
		if err == nil && owner != nil && owner.ID == apartment.OwnerID {
			canViewDocuments = true
		}
	}

	if !canViewDocuments {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для просмотра документов"))
		return
	}

	documentType := c.Query("type")
	var documents []*domain.ApartmentDocument

	if documentType != "" {

		if documentType != domain.DocumentTypeOwner && documentType != domain.DocumentTypeRealtor {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный тип документа для фильтрации. Допустимые типы: '"+domain.DocumentTypeOwner+"', '"+domain.DocumentTypeRealtor+"'"))
			return
		}

		documents, err = h.apartmentUseCase.GetDocumentsByApartmentIDAndType(id, documentType)
	} else {

		documents, err = h.apartmentUseCase.GetDocumentsByApartmentID(id)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении документов квартиры"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("документы квартиры успешно получены", documents))
}

func (h *ApartmentHandler) DeleteDocument(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	userIDInt := userID.(int)

	documentIDParam := c.Param("documentId")
	documentID, err := strconv.Atoi(documentIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID документа"))
		return
	}

	document, err := h.apartmentUseCase.GetDocumentByID(documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных документа"))
		return
	}
	if document == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("документ не найден"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(document.ApartmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}

	user, err := h.userUseCase.GetByID(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return
	}

	owner, err := h.ownerUseCase.GetByID(apartment.OwnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных владельца"))
		return
	}

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator && user.ID != owner.UserID {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для удаления документа"))
		return
	}

	if err := h.apartmentUseCase.DeleteDocument(documentID); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при удалении документа: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("документ успешно удален", nil))
}

func (h *ApartmentHandler) GetDashboardStats(c *gin.Context) {

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

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator && user.Role != domain.RoleOwner {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для просмотра статистики"))
		return
	}

	var response gin.H

	if user.Role == domain.RoleOwner {

		owner, err := h.ownerUseCase.GetByUserID(userIDInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных владельца"))
			return
		}
		if owner == nil {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("профиль владельца не найден"))
			return
		}

		filtersApproved := map[string]interface{}{
			"owner_id": owner.ID,
			"status":   domain.AptStatusApproved,
		}
		filtersPending := map[string]interface{}{
			"owner_id": owner.ID,
			"status":   domain.AptStatusPending,
		}
		filtersRejected := map[string]interface{}{
			"owner_id": owner.ID,
			"status":   domain.AptStatusRejected,
		}
		filtersRevision := map[string]interface{}{
			"owner_id": owner.ID,
			"status":   domain.AptStatusNeedsRevision,
		}

		approvedApts, _, err := h.apartmentUseCase.GetAll(filtersApproved, 1, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении статистики одобренных квартир"))
			return
		}

		pendingApts, _, err := h.apartmentUseCase.GetAll(filtersPending, 1, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении статистики ожидающих квартир"))
			return
		}

		rejectedApts, _, err := h.apartmentUseCase.GetAll(filtersRejected, 1, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении статистики отклоненных квартир"))
			return
		}

		revisionApts, _, err := h.apartmentUseCase.GetAll(filtersRevision, 1, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении статистики квартир на доработке"))
			return
		}

		response = gin.H{
			"approved":       len(approvedApts),
			"pending":        len(pendingApts),
			"rejected":       len(rejectedApts),
			"needs_revision": len(revisionApts),
			"total":          len(approvedApts) + len(pendingApts) + len(rejectedApts) + len(revisionApts),
		}
	} else {

		filtersApproved := map[string]interface{}{
			"status": domain.AptStatusApproved,
		}
		filtersPending := map[string]interface{}{
			"status": domain.AptStatusPending,
		}
		filtersRejected := map[string]interface{}{
			"status": domain.AptStatusRejected,
		}
		filtersRevision := map[string]interface{}{
			"status": domain.AptStatusNeedsRevision,
		}

		_, totalApproved, err := h.apartmentUseCase.GetAll(filtersApproved, 1, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении статистики одобренных квартир"))
			return
		}

		_, totalPending, err := h.apartmentUseCase.GetAll(filtersPending, 1, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении статистики ожидающих квартир"))
			return
		}

		_, totalRejected, err := h.apartmentUseCase.GetAll(filtersRejected, 1, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении статистики отклоненных квартир"))
			return
		}

		_, totalRevision, err := h.apartmentUseCase.GetAll(filtersRevision, 1, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении статистики квартир на доработке"))
			return
		}

		response = gin.H{
			"approved":       totalApproved,
			"pending":        totalPending,
			"rejected":       totalRejected,
			"needs_revision": totalRevision,
			"total":          totalApproved + totalPending + totalRejected + totalRevision,
		}
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статистика квартир успешно получена", response))
}

// @Summary Получение доступных вариантов продолжительности бронирования
// @Description Возвращает массив доступных вариантов продолжительности бронирования для конкретной квартиры с учетом времени начала аренды
// @Tags apartments
// @Accept json
// @Produce json
// @Param id path int true "ID квартиры"
// @Param start_time query string false "Время начала аренды (формат: 2006-01-02T15:04:05)"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id}/available-durations [get]
func (h *ApartmentHandler) GetAvailableDurations(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	var startTime time.Time
	startTimeParam := c.Query("start_time")
	if startTimeParam != "" {
		if parsedDate, err := time.Parse("2006-01-02", startTimeParam); err == nil {
			nowUTC := utils.GetCurrentTimeUTC()
			nowLocal := utils.ConvertOutputFromUTC(nowUTC)
			todayLocal := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, utils.KazakhstanTZ)
			inputDateLocal := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, utils.KazakhstanTZ)

			if inputDateLocal.Equal(todayLocal) {
				if nowLocal.Hour() < 10 {
					startTime = time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 10, 0, 0, 0, utils.KazakhstanTZ).UTC()
				} else {
					startTime = nowUTC
				}
			} else {
				startTime = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 10, 0, 0, 0, utils.KazakhstanTZ).UTC()
			}
		} else {
			parsedTime, err := utils.ParseUserInput(startTimeParam)
			if err != nil {
				c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный формат времени. Используйте формат: 2006-01-02 или 2006-01-02T15:04:05"))
				return
			}
			startTime = parsedTime
		}
	} else {
		nowUTC := utils.GetCurrentTimeUTC()
		nowLocal := utils.ConvertOutputFromUTC(nowUTC)

		if nowLocal.Hour() < 10 {
			startTime = time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 10, 0, 0, 0, utils.KazakhstanTZ).UTC()
		} else {
			startTime = nowUTC
		}
	}

	durations := utils.GetAvailableRentalDurations(startTime, apartment.RentalTypeHourly, apartment.RentalTypeDaily)
	timeInfo := utils.GetRentalTimeInfo(startTime)

	if len(durations) > 0 {
		filteredDurations := []int{}
		startTimeLocal := startTime.In(utils.KazakhstanTZ)
		now := utils.GetCurrentTimeUTC().In(utils.KazakhstanTZ)

		for _, duration := range durations {
			hasAvailableSlot := false

			if duration < 24 {
				var effectiveStartTime time.Time
				today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, utils.KazakhstanTZ)
				selectedDay := time.Date(startTimeLocal.Year(), startTimeLocal.Month(), startTimeLocal.Day(), 0, 0, 0, 0, utils.KazakhstanTZ)

				if selectedDay.Equal(today) {
					if now.After(startTimeLocal) {
						effectiveStartTime = now
					} else {
						effectiveStartTime = startTimeLocal
					}
				} else {
					effectiveStartTime = startTimeLocal
				}

				checkStartHour := effectiveStartTime.Hour()
				if effectiveStartTime.Minute() > 0 {
					checkStartHour++
				}

				maxStartHour := 22 - duration

				if checkStartHour < 10 {
					checkStartHour = 10
				}

				maxChecks := 6
				checked := 0

				for hour := checkStartHour; hour <= maxStartHour && checked < maxChecks; hour++ {
					slotStart := time.Date(startTimeLocal.Year(), startTimeLocal.Month(), startTimeLocal.Day(), hour, 0, 0, 0, utils.KazakhstanTZ).UTC()
					slotEnd := slotStart.Add(time.Duration(duration) * time.Hour)

					if selectedDay.Equal(today) && slotStart.Before(utils.GetCurrentTimeUTC()) {
						checked++
						continue
					}

					isAvailable, err := h.bookingUseCase.CheckApartmentAvailability(apartment.ID, slotStart, slotEnd)
					if err == nil && isAvailable {
						hasAvailableSlot = true
						break
					}
					checked++
				}
			} else {
				dayStart := time.Date(startTimeLocal.Year(), startTimeLocal.Month(), startTimeLocal.Day(), 0, 0, 0, 0, utils.KazakhstanTZ).UTC()
				dayEnd := dayStart.Add(24 * time.Hour)

				isAvailable, err := h.bookingUseCase.CheckApartmentAvailability(apartment.ID, dayStart, dayEnd)
				if err == nil && isAvailable {
					hasAvailableSlot = true
				}
			}

			if hasAvailableSlot {
				filteredDurations = append(filteredDurations, duration)
			}
		}
		durations = filteredDurations
	}

	if len(durations) == 0 {
		if apartment.RentalTypeHourly && !apartment.RentalTypeDaily {
			response := gin.H{
				"apartment_id":        apartment.ID,
				"available_durations": []int{},
				"rental_type_hourly":  apartment.RentalTypeHourly,
				"rental_type_daily":   apartment.RentalTypeDaily,
				"start_time":          utils.FormatForUser(startTime),
				"time_info":           timeInfo,
				"message":             "Данная квартира доступна только для почасовой аренды с 10:00 до 22:00",
				"unavailable_today":   true,
			}
			c.JSON(http.StatusOK, domain.NewSuccessResponse("информация о доступности получена", response))
			return
		}
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("для данной квартиры не настроены типы аренды"))
		return
	}

	response := gin.H{
		"apartment_id":        apartment.ID,
		"available_durations": durations,
		"rental_type_hourly":  apartment.RentalTypeHourly,
		"rental_type_daily":   apartment.RentalTypeDaily,
		"start_time":          utils.FormatForUser(startTime),
		"time_info":           timeInfo,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("доступные варианты продолжительности получены", response))
}

// @Summary Проверка возможности немедленного бронирования
// @Description Проверяет можно ли забронировать квартиру прямо сейчас на указанное время
// @Tags apartments
// @Accept json
// @Produce json
// @Param id path int true "ID квартиры"
// @Param time query string true "Время бронирования (формат: HH:MM)"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /apartments/{id}/can-book-now [get]
func (h *ApartmentHandler) CanBookNow(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	timeParam := c.Query("time")
	if timeParam == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("параметр time обязателен (формат: HH:MM)"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	nowUTC := utils.GetCurrentTimeUTC()
	nowLocal := utils.ConvertOutputFromUTC(nowUTC)

	timeParts := strings.Split(timeParam, ":")
	if len(timeParts) != 2 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный формат времени, используйте HH:MM"))
		return
	}

	hour, err := strconv.Atoi(timeParts[0])
	if err != nil || hour < 0 || hour > 23 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный час"))
		return
	}

	minute, err := strconv.Atoi(timeParts[1])
	if err != nil || minute < 0 || minute > 59 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректная минута"))
		return
	}

	requestedTime := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), hour, minute, 0, 0, utils.KazakhstanTZ)

	canBookImmediately := false
	reason := ""
	message := ""

	if !apartment.IsFree {
		reason = "apartment_busy"
		message = "Квартира сейчас занята"
	} else {
		timeDiff := requestedTime.Sub(nowLocal)
		if timeDiff < 0 || timeDiff > 2*time.Hour {
			reason = "time_too_far"
			if timeDiff < 0 {
				message = "Указанное время уже прошло"
			} else {
				message = "Указанное время слишком далеко в будущем (больше 2 часов)"
			}
		} else {
			if hour >= 10 && hour < 22 && apartment.RentalTypeHourly {
				canBookImmediately = true
				reason = "time_available"
				message = "Можно забронировать прямо сейчас"
			} else if apartment.RentalTypeDaily && hour < 23 {
				canBookImmediately = true
				reason = "time_available"
				message = "Можно забронировать на сутки прямо сейчас"
			} else {
				reason = "outside_hours"
				if !apartment.RentalTypeHourly && !apartment.RentalTypeDaily {
					message = "Для данной квартиры не настроены типы аренды"
				} else if hour < 10 || hour >= 22 {
					message = "Почасовая аренда доступна только с 10:00 до 22:00"
				} else if hour >= 23 {
					message = "Суточная аренда недоступна после 23:00"
				}
			}
		}
	}

	response := gin.H{
		"apartment_id":         apartment.ID,
		"requested_time":       timeParam,
		"can_book_immediately": canBookImmediately,
		"reason":               reason,
		"message":              message,
		"is_free":              apartment.IsFree,
		"rental_type_hourly":   apartment.RentalTypeHourly,
		"rental_type_daily":    apartment.RentalTypeDaily,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("проверка возможности бронирования выполнена", response))
}

// @Summary Расчет цены бронирования
// @Description Рассчитывает стоимость бронирования для указанной квартиры и продолжительности с учетом времени начала аренды
// @Tags apartments
// @Accept json
// @Produce json
// @Param id path int true "ID квартиры"
// @Param duration query int true "Продолжительность бронирования в часах"
// @Param start_time query string false "Время начала аренды (формат: 2006-01-02T15:04:05)"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id}/calculate-price [get]
func (h *ApartmentHandler) CalculatePrice(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	durationParam := c.Query("duration")
	if durationParam == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("параметр duration обязателен"))
		return
	}

	duration, err := strconv.Atoi(durationParam)
	if err != nil || duration <= 0 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректное значение продолжительности"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	var startTime time.Time
	startTimeParam := c.Query("start_time")
	if startTimeParam != "" {
		parsedTime, err := utils.ParseUserInput(startTimeParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный формат времени. Используйте формат: 2006-01-02T15:04:05"))
			return
		}
		startTime = parsedTime
	} else {
		startTime = utils.GetCurrentTimeUTC()
	}

	if duration == 24 {
		if !apartment.RentalTypeDaily {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("данная квартира не поддерживает посуточную аренду"))
			return
		}
		if apartment.DailyPrice <= 0 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("для данной квартиры не установлена цена за сутки"))
			return
		}
	} else {
		if !apartment.RentalTypeHourly {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("данная квартира не поддерживает почасовую аренду"))
			return
		}
		if apartment.Price <= 0 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("для данной квартиры не установлена почасовая цена"))
			return
		}
	}

	if !utils.ValidateRentalTime(startTime, duration, apartment.RentalTypeHourly, apartment.RentalTypeDaily) {
		timeInfo := utils.GetRentalTimeInfo(startTime)
		isDaytime := timeInfo["is_daytime"].(bool)

		if duration < 24 && !isDaytime {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("почасовая аренда доступна только с 10:00 до 22:00 (местное время)"))
			return
		}

		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(fmt.Sprintf("продолжительность %d часов недоступна для данной квартиры в указанное время", duration)))
		return
	}

	var basePrice int
	if duration == 24 && apartment.RentalTypeDaily {
		basePrice = apartment.DailyPrice
	} else if duration < 24 && apartment.RentalTypeHourly {
		basePrice = utils.CalculateHourlyPrice(apartment.Price, duration)
	} else {
		if duration == 24 {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("посуточная аренда недоступна для данной квартиры"))
		} else {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("почасовая аренда недоступна для данной квартиры"))
		}
		return
	}
	var serviceFee int
	if duration == 24 {
		serviceFee = 3000
	} else {
		serviceFeePercentage := domain.ServiceFeePercentage
		if h.settingsUseCase != nil {
			if percentage, err := h.settingsUseCase.GetServiceFeePercentage(); err == nil {
				serviceFeePercentage = percentage
			}
		}
		serviceFee = basePrice * serviceFeePercentage / 100
	}
	finalPrice := basePrice + serviceFee

	timeInfo := utils.GetRentalTimeInfo(startTime)

	response := gin.H{
		"apartment_id": apartment.ID,
		"duration":     duration,
		"base_price":   basePrice,
		"service_fee":  serviceFee,
		"final_price":  finalPrice,
		"hourly_price": apartment.Price,
		"daily_price":  apartment.DailyPrice,
		"start_time":   utils.FormatForUser(startTime),
		"time_info":    timeInfo,
		"price_breakdown": gin.H{
			"calculation": func() string {
				if duration == 24 {
					return fmt.Sprintf("Посуточная аренда: %d тг", apartment.DailyPrice)
				}
				if duration == utils.RentalDuration6Hours {
					discountPercent := utils.Discount6Hours
					originalPrice := apartment.Price * duration
					return fmt.Sprintf("Почасовая аренда: %d тг × %d ч = %d тг (скидка %d%%, было %d тг)", apartment.Price, duration, basePrice, discountPercent, originalPrice)
				} else if duration == utils.RentalDuration12Hours {
					discountPercent := utils.Discount12Hours
					originalPrice := apartment.Price * duration
					return fmt.Sprintf("Почасовая аренда: %d тг × %d ч = %d тг (скидка %d%%, было %d тг)", apartment.Price, duration, basePrice, discountPercent, originalPrice)
				}
				return fmt.Sprintf("Почасовая аренда: %d тг × %d ч = %d тг", apartment.Price, duration, basePrice)
			}(),
			"service_fee_info": func() string {
				if duration == 24 {
					return "Фиксированный сервисный сбор: 3000 тг"
				}
				serviceFeePercentage := domain.ServiceFeePercentage
				if h.settingsUseCase != nil {
					if percentage, err := h.settingsUseCase.GetServiceFeePercentage(); err == nil {
						serviceFeePercentage = percentage
					}
				}
				return fmt.Sprintf("Сервисный сбор (%d%%): %d тг", serviceFeePercentage, serviceFee)
			}(),
		},
		"cancellation_policy": func() gin.H {
			freeCancellationDeadline := startTime.Add(-6 * time.Hour)
			freeCancellationLocal := utils.ConvertOutputFromUTC(freeCancellationDeadline)
			startTimeLocal := utils.ConvertOutputFromUTC(startTime)

			nowUTC := utils.GetCurrentTimeUTC()
			nowLocal := utils.ConvertOutputFromUTC(nowUTC)

			canCancelFree := nowLocal.Before(freeCancellationLocal)

			policy := gin.H{
				"free_cancellation_until": freeCancellationLocal.Format("2006-01-02 15:04:05"),
				"can_cancel_free_now":     canCancelFree,
				"booking_start":           startTimeLocal.Format("2006-01-02 15:04:05"),
				"policy_text":             "Бесплатная отмена за 6 часов до начала бронирования",
			}

			if canCancelFree {
				timeUntilDeadline := freeCancellationLocal.Sub(nowLocal)
				if timeUntilDeadline > 24*time.Hour {
					days := int(timeUntilDeadline.Hours() / 24)
					policy["time_remaining"] = fmt.Sprintf("Осталось %d дн. для бесплатной отмены", days)
				} else if timeUntilDeadline > time.Hour {
					hours := int(timeUntilDeadline.Hours())
					policy["time_remaining"] = fmt.Sprintf("Осталось %d ч. для бесплатной отмены", hours)
				} else {
					minutes := int(timeUntilDeadline.Minutes())
					policy["time_remaining"] = fmt.Sprintf("Осталось %d мин. для бесплатной отмены", minutes)
				}
				policy["full_refund_available"] = true
			} else {
				policy["time_remaining"] = "Время бесплатной отмены истекло"
				policy["full_refund_available"] = false
				policy["note"] = "После истечения времени бесплатной отмены деньги не возвращаются"
			}

			return policy
		}(),
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("стоимость бронирования рассчитана", response))
}

// @Summary Получение полностью забронированных дат
// @Description Возвращает список дат, когда все временные слоты квартиры заняты
// @Tags apartments
// @Accept json
// @Produce json
// @Param id path int true "ID квартиры"
// @Param days_ahead query int false "Количество дней вперед для проверки" default(30)
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id}/booked-dates [get]
func (h *ApartmentHandler) GetBookedDates(c *gin.Context) {
	apartmentID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	daysAhead := 30
	if daysStr := c.Query("days_ahead"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			daysAhead = d
		}
	}

	apartment, err := h.apartmentUseCase.GetByID(apartmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("Квартира не найдена"))
		return
	}

	if apartment.Status != domain.AptStatusApproved {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("Квартира недоступна для бронирования"))
		return
	}

	bookedDates, err := h.apartmentUseCase.GetBookedDates(apartmentID, daysAhead)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("Ошибка при получении забронированных дат"))
		return
	}

	nowUTC := utils.GetCurrentTimeUTC()
	nowLocal := utils.ConvertOutputFromUTC(nowUTC)
	today := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, utils.KazakhstanTZ)

	for i := 0; i < daysAhead; i++ {
		checkDate := today.AddDate(0, 0, i)
		shouldBlock := false

		if apartment.RentalTypeHourly && !apartment.RentalTypeDaily {
			if i == 0 && nowLocal.Hour() >= 22 {
				shouldBlock = true
			}
		}

		if apartment.RentalTypeDaily {
			if i == 0 && nowLocal.Hour() >= 23 {
				shouldBlock = true
			}
		}

		if shouldBlock {
			dateStr := checkDate.Format("2006-01-02")
			if !slices.Contains(bookedDates, dateStr) {
				bookedDates = append(bookedDates, dateStr)
			}
		}
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("Забронированные даты получены", bookedDates))
}

// @Summary Получение всех квартир (админ)
// @Description Возвращает список всех квартир в системе с расширенными фильтрами (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "Статус квартиры (pending, approved, needs_revision, rejected)"
// @Param city_id query int false "ID города"
// @Param owner_id query int false "ID владельца"
// @Param listing_type query string false "Тип объявления (owner, realtor)"
// @Param room_count query int false "Количество комнат"
// @Param apartment_type_id query int false "Тип квартиры (ID)"
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartments [get]
func (h *ApartmentHandler) AdminGetAllApartments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if pageSize > 100 {
		pageSize = 100
	}
	if pageSize < 1 {
		pageSize = 20
	}

	filters := make(map[string]interface{})

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	if cityIDStr := c.Query("city_id"); cityIDStr != "" {
		if cityID, err := strconv.Atoi(cityIDStr); err == nil {
			filters["city_id"] = cityID
		}
	}

	if ownerIDStr := c.Query("owner_id"); ownerIDStr != "" {
		if ownerID, err := strconv.Atoi(ownerIDStr); err == nil {
			filters["owner_id"] = ownerID
		}
	}

	if listingType := c.Query("listing_type"); listingType != "" {
		filters["listing_type"] = listingType
	}

	if roomCountStr := c.Query("room_count"); roomCountStr != "" {
		if roomCount, err := strconv.Atoi(roomCountStr); err == nil && roomCount > 0 {
			filters["room_count"] = roomCount
		}
	}

	if apartmentTypeID := c.Query("apartment_type_id"); apartmentTypeID != "" {
		if id, err := strconv.Atoi(apartmentTypeID); err == nil && id > 0 {
			filters["apartment_type_id"] = id
		}
	}

	apartments, total, err := h.apartmentUseCase.GetAll(filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении квартир"))
		return
	}

	statusStats, err := h.apartmentUseCase.GetStatusStatistics()
	if err != nil {
		statusStats = make(map[string]int)
	}

	enrichedApartments := h.enrichApartmentsWithLocationData(apartments)

	response := gin.H{
		"apartments": enrichedApartments,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
		"filters": filters,
		"statistics": gin.H{
			"total_apartments": total,
			"by_status":        statusStats,
		},
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("список квартир получен", response))
}

// @Summary Получение квартиры по ID (админ)
// @Description Возвращает детальную информацию о квартире с расширенными данными (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartments/{id} [get]
func (h *ApartmentHandler) AdminGetApartmentByID(c *gin.Context) {
	apartmentID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(apartmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	photos, _ := h.apartmentUseCase.GetPhotosByApartmentID(apartmentID)
	documents, _ := h.apartmentUseCase.GetDocumentsByApartmentID(apartmentID)

	enrichedApartment := h.enrichApartmentWithLocationData(apartment)

	response := gin.H{
		"apartment": enrichedApartment,
		"photos":    photos,
		"documents": documents,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("квартира получена", response))
}

// @Summary Удаление квартиры (админ)
// @Description Полное удаление квартиры из системы (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartments/{id} [delete]
func (h *ApartmentHandler) AdminDeleteApartment(c *gin.Context) {
	apartmentID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(apartmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	err = h.apartmentUseCase.Delete(apartmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при удалении квартиры"))
		return
	}

	go func() {
		var userID int
		if apartment.Owner != nil {
			userID = apartment.Owner.UserID
		} else {
			owner, err := h.ownerUseCase.GetByID(apartment.OwnerID)
			if err != nil || owner == nil {
				return
			}
			userID = owner.UserID
		}

		notification := &domain.Notification{
			UserID:  userID,
			Type:    domain.NotificationApartmentStatusChanged,
			Title:   "Квартира удалена",
			Message: fmt.Sprintf("Ваша квартира \"%s, %s\" была удалена администратором", apartment.Street, apartment.Building),
		}
		h.notificationUseCase.CreateNotification(notification)
	}()

	c.JSON(http.StatusOK, domain.NewSuccessResponse("квартира удалена", nil))
}

// @Summary Полная статистика дашборда (админ)
// @Description Возвращает комплексную статистику для админского дашборда с фильтрацией по датам (только для админов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param date_from query string false "Дата начала (YYYY-MM-DD)"
// @Param date_to query string false "Дата окончания (YYYY-MM-DD)"
// @Param period query string false "Группировка данных (day, week, month)" default(day)
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/dashboard/statistics [get]
func (h *ApartmentHandler) AdminGetFullDashboardStats(c *gin.Context) {
	var dateFrom, dateTo time.Time
	var err error

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		dateFrom, err = time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат date_from"))
			return
		}
	} else {
		dateFrom = time.Now().AddDate(0, -1, 0)
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		dateTo, err = time.Parse("2006-01-02", dateToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат date_to"))
			return
		}
	} else {
		dateTo = time.Now()
	}

	period := c.DefaultQuery("period", "day")
	if period != "day" && period != "week" && period != "month" {
		period = "day"
	}

	allApartments, _, _ := h.apartmentUseCase.GetAll(map[string]interface{}{"include_all_statuses": true}, 1, 10000)
	allUsers, _, _ := h.userRepo.GetAll(map[string]interface{}{}, 1, 10000)
	allBookings, _, _ := h.bookingUseCase.AdminGetAllBookings(map[string]interface{}{}, 1, 10000)
	allLocks, _ := h.lockUseCase.GetAllLocks()

	filteredBookings := filterBookingsByDate(allBookings, dateFrom, dateTo)
	filteredUsers := filterUsersByDate(allUsers, dateFrom, dateTo)
	filteredApartments := filterApartmentsByDate(allApartments, dateFrom, dateTo)

	totalStats := map[string]interface{}{
		"total_users":      len(allUsers),
		"total_apartments": len(allApartments),
		"total_bookings":   len(allBookings),
		"total_locks":      len(allLocks),
		"active_users":     countActiveUsers(allUsers),
		"online_locks":     countOnlineLocks(allLocks),
	}

	userStats := generateUserStatistics(allUsers, filteredUsers, period, dateFrom, dateTo)

	apartmentStats := generateApartmentStatistics(allApartments, filteredApartments, period, dateFrom, dateTo)

	bookingStats := generateBookingStatistics(allBookings, filteredBookings, period, dateFrom, dateTo)

	lockStats := generateLockStatistics(allLocks)

	financialStats := generateFinancialStatistics(filteredBookings, period, dateFrom, dateTo)

	timeSeriesData := generateTimeSeriesData(filteredBookings, filteredUsers, filteredApartments, period, dateFrom, dateTo)

	response := gin.H{
		"period":      period,
		"date_from":   dateFrom.Format("2006-01-02"),
		"date_to":     dateTo.Format("2006-01-02"),
		"totals":      totalStats,
		"users":       userStats,
		"apartments":  apartmentStats,
		"bookings":    bookingStats,
		"locks":       lockStats,
		"financial":   financialStats,
		"time_series": timeSeriesData,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статистика дашборда получена", response))
}

func filterBookingsByDate(bookings []*domain.Booking, dateFrom, dateTo time.Time) []*domain.Booking {
	var filtered []*domain.Booking
	for _, booking := range bookings {
		if booking.CreatedAt.After(dateFrom) && booking.CreatedAt.Before(dateTo.AddDate(0, 0, 1)) {
			filtered = append(filtered, booking)
		}
	}
	return filtered
}

func filterUsersByDate(users []*domain.User, dateFrom, dateTo time.Time) []*domain.User {
	var filtered []*domain.User
	for _, user := range users {
		if user.CreatedAt.After(dateFrom) && user.CreatedAt.Before(dateTo.AddDate(0, 0, 1)) {
			filtered = append(filtered, user)
		}
	}
	return filtered
}

func filterApartmentsByDate(apartments []*domain.Apartment, dateFrom, dateTo time.Time) []*domain.Apartment {
	var filtered []*domain.Apartment
	for _, apt := range apartments {
		if apt.CreatedAt.After(dateFrom) && apt.CreatedAt.Before(dateTo.AddDate(0, 0, 1)) {
			filtered = append(filtered, apt)
		}
	}
	return filtered
}

func countActiveUsers(users []*domain.User) int {
	count := 0
	for _, user := range users {
		if user.IsActive {
			count++
		}
	}
	return count
}

func countOnlineLocks(locks []*domain.Lock) int {
	count := 0
	for _, lock := range locks {
		if lock.IsOnline {
			count++
		}
	}
	return count
}

func generateUserStatistics(allUsers, filteredUsers []*domain.User, period string, dateFrom, dateTo time.Time) map[string]interface{} {
	roleStats := make(map[string]int)
	statusStats := make(map[string]int)
	cityStats := make(map[string]int)

	for _, user := range allUsers {
		roleStats[string(user.Role)]++
		if user.IsActive {
			statusStats["active"]++
		} else {
			statusStats["inactive"]++
		}
		if user.City != nil {
			cityStats[user.City.Name]++
		}
	}

	return map[string]interface{}{
		"total":         len(allUsers),
		"new_in_period": len(filteredUsers),
		"by_role":       roleStats,
		"by_status":     statusStats,
		"by_city":       cityStats,
		"growth_rate":   calculateGrowthRate(len(allUsers), len(filteredUsers), dateFrom, dateTo),
	}
}

func generateApartmentStatistics(allApartments, filteredApartments []*domain.Apartment, period string, dateFrom, dateTo time.Time) map[string]interface{} {
	statusStats := make(map[string]int)
	listingTypeStats := make(map[string]int)
	cityStats := make(map[string]int)
	roomCountStats := make(map[int]int)

	totalPrice := 0
	totalDailyPrice := 0

	for _, apt := range allApartments {
		statusStats[string(apt.Status)]++
		listingTypeStats[apt.ListingType]++
		roomCountStats[apt.RoomCount]++
		totalPrice += apt.Price
		totalDailyPrice += apt.DailyPrice

		if apt.City != nil {
			cityStats[apt.City.Name]++
		}
	}

	avgPrice := 0.0
	avgDailyPrice := 0.0
	if len(allApartments) > 0 {
		avgPrice = float64(totalPrice) / float64(len(allApartments))
		avgDailyPrice = float64(totalDailyPrice) / float64(len(allApartments))
	}

	return map[string]interface{}{
		"total":           len(allApartments),
		"new_in_period":   len(filteredApartments),
		"by_status":       statusStats,
		"by_listing_type": listingTypeStats,
		"by_city":         cityStats,
		"by_room_count":   roomCountStats,
		"avg_price":       avgPrice,
		"avg_daily_price": avgDailyPrice,
		"growth_rate":     calculateGrowthRate(len(allApartments), len(filteredApartments), dateFrom, dateTo),
	}
}

func generateBookingStatistics(allBookings, filteredBookings []*domain.Booking, period string, dateFrom, dateTo time.Time) map[string]interface{} {
	statusStats := make(map[string]int)
	durationStats := make(map[string]int)
	totalRevenue := 0
	filteredRevenue := 0

	for _, booking := range allBookings {
		statusStats[string(booking.Status)]++
		totalRevenue += booking.FinalPrice

		if booking.Duration <= 2 {
			durationStats["short"]++
		} else if booking.Duration <= 6 {
			durationStats["medium"]++
		} else if booking.Duration <= 12 {
			durationStats["long"]++
		} else {
			durationStats["extended"]++
		}
	}

	for _, booking := range filteredBookings {
		filteredRevenue += booking.FinalPrice
	}

	avgBookingValue := 0.0
	if len(allBookings) > 0 {
		avgBookingValue = float64(totalRevenue) / float64(len(allBookings))
	}

	return map[string]interface{}{
		"total":               len(allBookings),
		"new_in_period":       len(filteredBookings),
		"by_status":           statusStats,
		"by_duration":         durationStats,
		"total_revenue":       totalRevenue,
		"period_revenue":      filteredRevenue,
		"avg_booking_value":   avgBookingValue,
		"growth_rate":         calculateGrowthRate(len(allBookings), len(filteredBookings), dateFrom, dateTo),
		"revenue_growth_rate": calculateRevenueGrowthRate(totalRevenue, filteredRevenue, dateFrom, dateTo),
	}
}

func generateLockStatistics(locks []*domain.Lock) map[string]interface{} {
	statusStats := make(map[string]int)
	batteryStats := make(map[string]int)
	apartmentStats := make(map[string]int)

	for _, lock := range locks {
		statusStats[string(lock.CurrentStatus)]++

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

		if lock.ApartmentID != nil {
			apartmentStats["bound"]++
		} else {
			apartmentStats["unbound"]++
		}
	}

	return map[string]interface{}{
		"total":        len(locks),
		"by_status":    statusStats,
		"by_battery":   batteryStats,
		"by_apartment": apartmentStats,
	}
}

func generateFinancialStatistics(bookings []*domain.Booking, period string, dateFrom, dateTo time.Time) map[string]interface{} {
	totalRevenue := 0
	totalServiceFees := 0
	completedRevenue := 0
	pendingRevenue := 0

	for _, booking := range bookings {
		totalRevenue += booking.FinalPrice
		totalServiceFees += booking.ServiceFee

		if booking.Status == "completed" {
			completedRevenue += booking.FinalPrice
		} else if booking.Status == "pending" || booking.Status == "approved" || booking.Status == "active" {
			pendingRevenue += booking.FinalPrice
		}
	}

	commission := int(float64(totalServiceFees) * 0.85)
	netRevenue := totalRevenue - totalServiceFees

	return map[string]interface{}{
		"total_revenue":      totalRevenue,
		"completed_revenue":  completedRevenue,
		"pending_revenue":    pendingRevenue,
		"total_service_fees": totalServiceFees,
		"commission":         commission,
		"net_revenue":        netRevenue,
		"average_transaction": func() float64 {
			if len(bookings) > 0 {
				return float64(totalRevenue) / float64(len(bookings))
			}
			return 0.0
		}(),
	}
}

func generateTimeSeriesData(bookings []*domain.Booking, users []*domain.User, apartments []*domain.Apartment, period string, dateFrom, dateTo time.Time) map[string]interface{} {
	bookingsByPeriod := make(map[string]int)
	revenueByPeriod := make(map[string]int)
	usersByPeriod := make(map[string]int)
	apartmentsByPeriod := make(map[string]int)

	current := dateFrom
	for current.Before(dateTo) || current.Equal(dateTo) {
		var key string
		switch period {
		case "week":
			year, week := current.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
			current = current.AddDate(0, 0, 7)
		case "month":
			key = current.Format("2006-01")
			current = current.AddDate(0, 1, 0)
		default:
			key = current.Format("2006-01-02")
			current = current.AddDate(0, 0, 1)
		}

		bookingsByPeriod[key] = 0
		revenueByPeriod[key] = 0
		usersByPeriod[key] = 0
		apartmentsByPeriod[key] = 0
	}

	for _, booking := range bookings {
		var key string
		switch period {
		case "week":
			year, week := booking.CreatedAt.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
		case "month":
			key = booking.CreatedAt.Format("2006-01")
		default:
			key = booking.CreatedAt.Format("2006-01-02")
		}

		if _, exists := bookingsByPeriod[key]; exists {
			bookingsByPeriod[key]++
			revenueByPeriod[key] += booking.FinalPrice
		}
	}

	for _, user := range users {
		var key string
		switch period {
		case "week":
			year, week := user.CreatedAt.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
		case "month":
			key = user.CreatedAt.Format("2006-01")
		default:
			key = user.CreatedAt.Format("2006-01-02")
		}

		if _, exists := usersByPeriod[key]; exists {
			usersByPeriod[key]++
		}
	}

	for _, apartment := range apartments {
		var key string
		switch period {
		case "week":
			year, week := apartment.CreatedAt.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
		case "month":
			key = apartment.CreatedAt.Format("2006-01")
		default:
			key = apartment.CreatedAt.Format("2006-01-02")
		}

		if _, exists := apartmentsByPeriod[key]; exists {
			apartmentsByPeriod[key]++
		}
	}

	var bookingsTimeSeries, revenueTimeSeries, usersTimeSeries, apartmentsTimeSeries []map[string]interface{}

	current = dateFrom
	for current.Before(dateTo) || current.Equal(dateTo) {
		var key string
		switch period {
		case "week":
			year, week := current.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
			current = current.AddDate(0, 0, 7)
		case "month":
			key = current.Format("2006-01")
			current = current.AddDate(0, 1, 0)
		default:
			key = current.Format("2006-01-02")
			current = current.AddDate(0, 0, 1)
		}

		bookingsTimeSeries = append(bookingsTimeSeries, map[string]interface{}{
			"date":  key,
			"value": bookingsByPeriod[key],
		})

		revenueTimeSeries = append(revenueTimeSeries, map[string]interface{}{
			"date":  key,
			"value": revenueByPeriod[key],
		})

		usersTimeSeries = append(usersTimeSeries, map[string]interface{}{
			"date":  key,
			"value": usersByPeriod[key],
		})

		apartmentsTimeSeries = append(apartmentsTimeSeries, map[string]interface{}{
			"date":  key,
			"value": apartmentsByPeriod[key],
		})
	}

	return map[string]interface{}{
		"bookings":   bookingsTimeSeries,
		"revenue":    revenueTimeSeries,
		"users":      usersTimeSeries,
		"apartments": apartmentsTimeSeries,
	}
}

func calculateGrowthRate(total, newInPeriod int, dateFrom, dateTo time.Time) float64 {
	if total == 0 {
		return 0.0
	}

	daysDiff := dateTo.Sub(dateFrom).Hours() / 24
	if daysDiff <= 0 {
		return 0.0
	}

	dailyGrowth := float64(newInPeriod) / daysDiff
	monthlyGrowth := dailyGrowth * 30
	growthRate := (monthlyGrowth / float64(total)) * 100

	return growthRate
}

func calculateRevenueGrowthRate(totalRevenue, periodRevenue int, dateFrom, dateTo time.Time) float64 {
	if totalRevenue == 0 {
		return 0.0
	}

	daysDiff := dateTo.Sub(dateFrom).Hours() / 24
	if daysDiff <= 0 {
		return 0.0
	}

	dailyRevenue := float64(periodRevenue) / daysDiff
	monthlyRevenue := dailyRevenue * 30
	growthRate := (monthlyRevenue / float64(totalRevenue)) * 100

	return growthRate
}

// @Summary Поиск квартир по координатам
// @Description Возвращает квартиры в указанном прямоугольнике координат с возможностью фильтрации
// @Tags apartments
// @Accept json
// @Produce json
// @Param min_lat query float64 true "Минимальная широта"
// @Param max_lat query float64 true "Максимальная широта"
// @Param min_lng query float64 true "Минимальная долгота"
// @Param max_lng query float64 true "Максимальная долгота"
// @Param city_id query int false "ID города"
// @Param district_id query int false "ID района"
// @Param microdistrict_id query int false "ID микрорайона"
// @Param room_count query int false "Количество комнат"
// @Param min_area query number false "Минимальная площадь"
// @Param max_area query number false "Максимальная площадь"
// @Param min_price query int false "Минимальная цена"
// @Param max_price query int false "Максимальная цена"
// @Param rental_type_hourly query bool false "Поддержка почасовой аренды"
// @Param rental_type_daily query bool false "Поддержка посуточной аренды"
// @Param is_free query bool false "Доступность квартиры (true - свободная, false - занятая)"
// @Param status query string false "Статус квартиры"
// @Param listing_type query string false "Тип объявления (owner, realtor)"
// @Param apartment_type_id query int false "Тип квартиры (ID)"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /apartments/search/geo [get]
func (h *ApartmentHandler) GetByCoordinates(c *gin.Context) {
	var req GetByCoordinatesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	filters := make(map[string]interface{})

	if req.CityID != nil {
		filters["city_id"] = *req.CityID
	}
	if req.DistrictID != nil {
		filters["district_id"] = *req.DistrictID
	}
	if req.MicrodistrictID != nil {
		filters["microdistrict_id"] = *req.MicrodistrictID
	}
	if req.RoomCount != nil {
		filters["room_count"] = *req.RoomCount
	}
	if req.MinArea != nil {
		filters["min_area"] = *req.MinArea
	}
	if req.MaxArea != nil {
		filters["max_area"] = *req.MaxArea
	}
	if req.MinPrice != nil {
		filters["min_price"] = *req.MinPrice
	}
	if req.MaxPrice != nil {
		filters["max_price"] = *req.MaxPrice
	}
	if req.RentalTypeHourly != nil {
		filters["rental_type_hourly"] = *req.RentalTypeHourly
	}
	if req.RentalTypeDaily != nil {
		filters["rental_type_daily"] = *req.RentalTypeDaily
	}
	if req.IsFree != nil {
		filters["is_free"] = *req.IsFree
	}
	if req.Status != nil {
		filters["status"] = *req.Status
	}
	if req.ListingType != nil {
		filters["listing_type"] = *req.ListingType
	}
	if req.ApartmentTypeID != nil {
		filters["apartment_type_id"] = *req.ApartmentTypeID
	}

	apartments, err := h.apartmentUseCase.GetFullApartmentsByCoordinatesWithFilters(req.MinLat, req.MaxLat, req.MinLng, req.MaxLng, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	enrichedApartments := h.enrichApartmentsWithLocationData(apartments)

	c.JSON(http.StatusOK, domain.NewSuccessResponse("квартиры найдены успешно", gin.H{
		"apartments": enrichedApartments,
	}))
}

// @Summary Детальная статистика квартир (админ)
// @Description Возвращает детальную статистику по квартирам с разбивкой по статусам, городам и другим параметрам
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartments/statistics [get]
func (h *ApartmentHandler) AdminGetApartmentStatistics(c *gin.Context) {
	filters := map[string]interface{}{
		"include_all_statuses": true,
	}
	apartments, _, err := h.apartmentUseCase.GetAll(filters, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении квартир: "+err.Error()))
		return
	}

	statusStats, err := h.apartmentUseCase.GetStatusStatistics()
	if err != nil {
		statusStats = make(map[string]int)
	}

	cityStats, err := h.apartmentUseCase.GetCityStatistics()
	if err != nil {
		cityStats = make(map[string]int)
	}

	districtStats, err := h.apartmentUseCase.GetDistrictStatistics()
	if err != nil {
		districtStats = make(map[string]int)
	}

	listingTypeStats := make(map[string]int)
	roomCountStats := make(map[int]int)
	totalArea := 0.0
	totalPrice := 0
	totalDailyPrice := 0
	apartmentCount := len(apartments)

	for _, apt := range apartments {
		listingTypeStats[apt.ListingType]++
		roomCountStats[apt.RoomCount]++
		totalArea += apt.TotalArea
		totalPrice += apt.Price
		totalDailyPrice += apt.DailyPrice
	}

	var avgArea, avgPrice, avgDailyPrice float64
	if apartmentCount > 0 {
		avgArea = totalArea / float64(apartmentCount)
		avgPrice = float64(totalPrice) / float64(apartmentCount)
		avgDailyPrice = float64(totalDailyPrice) / float64(apartmentCount)
	}

	response := gin.H{
		"summary": gin.H{
			"total_apartments": apartmentCount,
			"avg_area":         avgArea,
			"avg_price":        avgPrice,
			"avg_daily_price":  avgDailyPrice,
		},
		"by_status":       statusStats,
		"by_city":         cityStats,
		"by_district":     districtStats,
		"by_listing_type": listingTypeStats,
		"by_room_count":   roomCountStats,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("детальная статистика квартир получена", response))
}

// @Summary Статистика для владельца квартир
// @Description Возвращает комплексную статистику по квартирам и бронированиям владельца
// @Tags apartments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param date_from query string false "Дата начала периода (YYYY-MM-DD)" default(30 дней назад)
// @Param date_to query string false "Дата конца периода (YYYY-MM-DD)" default(сегодня)
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/owner/statistics [get]
func (h *ApartmentHandler) GetOwnerStatistics(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("доступ разрешен только владельцам квартир"))
		return
	}

	owner, err := h.ownerUseCase.GetByUserID(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных владельца"))
		return
	}
	if owner == nil {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("профиль владельца не найден"))
		return
	}

	// Парсинг дат для периода анализа
	var dateFrom, dateTo time.Time
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		dateFrom, err = time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат date_from (используйте YYYY-MM-DD)"))
			return
		}
	} else {
		dateFrom = time.Now().AddDate(0, -1, 0) // 30 дней назад
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		dateTo, err = time.Parse("2006-01-02", dateToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат date_to (используйте YYYY-MM-DD)"))
			return
		}
	} else {
		dateTo = time.Now()
	}

	// Получение квартир владельца
	apartments, err := h.apartmentUseCase.GetByOwnerID(owner.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении квартир: "+err.Error()))
		return
	}

	// Статистика квартир
	apartmentStats := h.calculateOwnerApartmentStatistics(apartments)

	// Статистика бронирований
	bookingStats, err := h.calculateOwnerBookingStatistics(apartments, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при расчете статистики бронирований: "+err.Error()))
		return
	}

	// Финансовая статистика
	financialStats := h.calculateOwnerFinancialStatistics(bookingStats["bookings"].([]*domain.Booking))

	// Статистика эффективности квартир
	efficiencyStats := h.calculateApartmentEfficiencyStatistics(apartments, bookingStats["bookings"].([]*domain.Booking))

	response := gin.H{
		"period": gin.H{
			"date_from": dateFrom.Format("2006-01-02"),
			"date_to":   dateTo.Format("2006-01-02"),
		},
		"apartments": apartmentStats,
		"bookings":   bookingStats,
		"financial":  financialStats,
		"efficiency": efficiencyStats,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статистика владельца получена", response))
}

// @Summary Проверка доступности квартиры
// @Description Проверяет доступность квартиры на указанные даты
// @Tags apartments
// @Accept json
// @Produce json
// @Param id path int true "ID квартиры"
// @Param start_date query string true "Дата начала (формат RFC3339)"
// @Param end_date query string true "Дата окончания (формат RFC3339)"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id}/availability [get]
func (h *ApartmentHandler) CheckApartmentAvailability(c *gin.Context) {
	apartmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("необходимо указать start_date и end_date"))
		return
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат start_date (используйте RFC3339)"))
		return
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат end_date (используйте RFC3339)"))
		return
	}

	isAvailable, err := h.bookingUseCase.CheckApartmentAvailability(apartmentID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("проверка доступности выполнена", gin.H{"available": isAvailable}))
}

// @Summary Получение доступных временных слотов с учетом длительности
// @Description Получает список свободных временных слотов для бронирования квартиры на указанную дату с учетом времени уборки
// @Tags apartments
// @Accept json
// @Produce json
// @Param id path int true "ID квартиры"
// @Param date query string true "Дата в формате 2006-01-02"
// @Param duration query int true "Продолжительность бронирования в часах"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id}/available-slots [get]
func (h *ApartmentHandler) GetAvailableTimeSlots(c *gin.Context) {
	apartmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("необходимо указать параметр date"))
		return
	}

	durationStr := c.Query("duration")
	if durationStr == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("необходимо указать параметр duration"))
		return
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration < 1 {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректная продолжительность"))
		return
	}

	availableSlots, err := h.bookingUseCase.GetAvailableTimeSlots(apartmentID, date, duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("доступные временные слоты получены", gin.H{
		"date":            date,
		"duration":        duration,
		"available_slots": availableSlots,
		"cleaning_notice": "После каждого бронирования предусмотрен 1 час на уборку",
		"hourly_limits":   "Почасовая аренда доступна с 10:00 до 22:00",
	}))
}

func (h *ApartmentHandler) enrichApartmentWithLocationData(apartment *domain.Apartment) gin.H {
	result := gin.H{
		"id":                     apartment.ID,
		"owner_id":               apartment.OwnerID,
		"city_id":                apartment.CityID,
		"district_id":            apartment.DistrictID,
		"microdistrict_id":       apartment.MicrodistrictID,
		"apartment_type_id":      apartment.ApartmentTypeID,
		"street":                 apartment.Street,
		"building":               apartment.Building,
		"apartment_number":       apartment.ApartmentNumber,
		"residential_complex":    apartment.ResidentialComplex,
		"room_count":             apartment.RoomCount,
		"total_area":             apartment.TotalArea,
		"kitchen_area":           apartment.KitchenArea,
		"floor":                  apartment.Floor,
		"total_floors":           apartment.TotalFloors,
		"condition_id":           apartment.ConditionID,
		"price":                  apartment.Price,
		"daily_price":            apartment.DailyPrice,
		"service_fee_percentage": apartment.ServiceFeePercentage,
		"rental_type_hourly":     apartment.RentalTypeHourly,
		"rental_type_daily":      apartment.RentalTypeDaily,
		"is_free":                apartment.IsFree,
		"status":                 apartment.Status,
		"description":            apartment.Description,
		"moderator_comment":      apartment.ModeratorComment,
		"listing_type":           apartment.ListingType,
		"is_agreement_accepted":  apartment.IsAgreementAccepted,
		"agreement_accepted_at":  apartment.AgreementAcceptedAt,
		"contract_id":            apartment.ContractID,
		"view_count":             apartment.ViewCount,
		"booking_count":          apartment.BookingCount,
		"created_at":             apartment.CreatedAt,
		"updated_at":             apartment.UpdatedAt,
	}

	if apartment.Owner != nil {
		result["owner"] = apartment.Owner
	}
	if apartment.Photos != nil {
		result["photos"] = apartment.Photos
	}
	if apartment.HouseRules != nil {
		result["house_rules"] = apartment.HouseRules
	}
	if apartment.Amenities != nil {
		result["amenities"] = apartment.Amenities
	}
	if apartment.Condition != nil {
		result["condition"] = apartment.Condition
	}
	if apartment.Location != nil {
		result["location"] = apartment.Location
	}

	if apartment.CityID > 0 {
		if city, err := h.locationUseCase.GetCityByID(apartment.CityID); err == nil && city != nil {
			result["city_name"] = city.Name
			if region, err := h.locationUseCase.GetRegionByID(city.RegionID); err == nil && region != nil {
				result["region_name"] = region.Name
			}
		}
	}

	if apartment.DistrictID > 0 {
		if district, err := h.locationUseCase.GetDistrictByID(apartment.DistrictID); err == nil && district != nil {
			result["district_name"] = district.Name
		}
	}

	if apartment.MicrodistrictID != nil && *apartment.MicrodistrictID > 0 {
		if microdistrict, err := h.locationUseCase.GetMicrodistrictByID(*apartment.MicrodistrictID); err == nil && microdistrict != nil {
			result["microdistrict_name"] = microdistrict.Name
		}
	}

	return result
}

func (h *ApartmentHandler) enrichApartmentsWithLocationData(apartments []*domain.Apartment) []gin.H {
	if len(apartments) == 0 {
		return []gin.H{}
	}

	cityIDs := make(map[int]bool)
	districtIDs := make(map[int]bool)
	microdistrictIDs := make(map[int]bool)
	regionIDs := make(map[int]bool)

	for _, apartment := range apartments {
		if apartment.CityID > 0 {
			cityIDs[apartment.CityID] = true
		}
		if apartment.DistrictID > 0 {
			districtIDs[apartment.DistrictID] = true
		}
		if apartment.MicrodistrictID != nil && *apartment.MicrodistrictID > 0 {
			microdistrictIDs[*apartment.MicrodistrictID] = true
		}
	}

	cityMap := make(map[int]*domain.City)
	districtMap := make(map[int]*domain.District)
	microdistrictMap := make(map[int]*domain.Microdistrict)
	regionMap := make(map[int]*domain.Region)

	for cityID := range cityIDs {
		if city, err := h.locationUseCase.GetCityByID(cityID); err == nil && city != nil {
			cityMap[cityID] = city
			if city.RegionID > 0 {
				regionIDs[city.RegionID] = true
			}
		}
	}

	for regionID := range regionIDs {
		if region, err := h.locationUseCase.GetRegionByID(regionID); err == nil && region != nil {
			regionMap[regionID] = region
		}
	}

	for districtID := range districtIDs {
		if district, err := h.locationUseCase.GetDistrictByID(districtID); err == nil && district != nil {
			districtMap[districtID] = district
		}
	}

	for microdistrictID := range microdistrictIDs {
		if microdistrict, err := h.locationUseCase.GetMicrodistrictByID(microdistrictID); err == nil && microdistrict != nil {
			microdistrictMap[microdistrictID] = microdistrict
		}
	}

	result := make([]gin.H, len(apartments))
	for i, apartment := range apartments {
		result[i] = h.enrichApartmentWithLocationDataOptimized(apartment, cityMap, regionMap, districtMap, microdistrictMap)
	}
	return result
}

func (h *ApartmentHandler) enrichApartmentWithLocationDataOptimized(
	apartment *domain.Apartment,
	cityMap map[int]*domain.City,
	regionMap map[int]*domain.Region,
	districtMap map[int]*domain.District,
	microdistrictMap map[int]*domain.Microdistrict,
) gin.H {
	result := gin.H{
		"id":                     apartment.ID,
		"owner_id":               apartment.OwnerID,
		"city_id":                apartment.CityID,
		"district_id":            apartment.DistrictID,
		"microdistrict_id":       apartment.MicrodistrictID,
		"apartment_type_id":      apartment.ApartmentTypeID,
		"street":                 apartment.Street,
		"building":               apartment.Building,
		"apartment_number":       apartment.ApartmentNumber,
		"residential_complex":    apartment.ResidentialComplex,
		"room_count":             apartment.RoomCount,
		"total_area":             apartment.TotalArea,
		"kitchen_area":           apartment.KitchenArea,
		"floor":                  apartment.Floor,
		"total_floors":           apartment.TotalFloors,
		"condition_id":           apartment.ConditionID,
		"price":                  apartment.Price,
		"daily_price":            apartment.DailyPrice,
		"service_fee_percentage": apartment.ServiceFeePercentage,
		"rental_type_hourly":     apartment.RentalTypeHourly,
		"rental_type_daily":      apartment.RentalTypeDaily,
		"is_free":                apartment.IsFree,
		"status":                 apartment.Status,
		"description":            apartment.Description,
		"moderator_comment":      apartment.ModeratorComment,
		"listing_type":           apartment.ListingType,
		"is_agreement_accepted":  apartment.IsAgreementAccepted,
		"agreement_accepted_at":  apartment.AgreementAcceptedAt,
		"contract_id":            apartment.ContractID,
		"view_count":             apartment.ViewCount,
		"booking_count":          apartment.BookingCount,
		"created_at":             apartment.CreatedAt,
		"updated_at":             apartment.UpdatedAt,
	}

	if apartment.Owner != nil {
		result["owner"] = apartment.Owner
	}
	if apartment.Photos != nil {
		result["photos"] = apartment.Photos
	}
	if apartment.HouseRules != nil {
		result["house_rules"] = apartment.HouseRules
	}
	if apartment.Amenities != nil {
		result["amenities"] = apartment.Amenities
	}
	if apartment.Condition != nil {
		result["condition"] = apartment.Condition
	}
	if apartment.Location != nil {
		result["location"] = apartment.Location
	}
	if apartment.ApartmentType != nil {
		result["apartment_type"] = apartment.ApartmentType
	}

	if city, exists := cityMap[apartment.CityID]; exists {
		result["city_name"] = city.Name
		if region, regionExists := regionMap[city.RegionID]; regionExists {
			result["region_name"] = region.Name
		}
	}

	if district, exists := districtMap[apartment.DistrictID]; exists {
		result["district_name"] = district.Name
	}

	if apartment.MicrodistrictID != nil && *apartment.MicrodistrictID > 0 {
		if microdistrict, exists := microdistrictMap[*apartment.MicrodistrictID]; exists {
			result["microdistrict_name"] = microdistrict.Name
		}
	}

	return result
}

// @Summary Принятие договора публикации квартиры
// @Description Владелец квартиры принимает договор с платформой для публикации
// @Tags apartments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Param request body domain.ConfirmApartmentAgreementRequest true "Данные согласия"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id}/confirm-agreement [post]
func (h *ApartmentHandler) ConfirmApartmentAgreement(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("пользователь не авторизован"))
		return
	}

	idParam := c.Param("id")
	apartmentID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	var req domain.ConfirmApartmentAgreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return
	}

	apartment, err := h.apartmentUseCase.ConfirmApartmentAgreement(apartmentID, userID.(int), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	fullApartment, err := h.apartmentUseCase.GetByID(apartment.ID)
	if err != nil {
		c.JSON(http.StatusOK, domain.NewSuccessResponse("договор публикации принят", apartment))
		return
	}

	enrichedApartment := h.enrichApartmentWithLocationData(fullApartment)
	c.JSON(http.StatusOK, domain.NewSuccessResponse("договор публикации принят", enrichedApartment))
}

func (h *ApartmentHandler) calculateOwnerApartmentStatistics(apartments []*domain.Apartment) gin.H {
	if len(apartments) == 0 {
		return gin.H{
			"total_apartments": 0,
			"by_status":        map[string]int{},
			"by_room_count":    map[int]int{},
			"avg_area":         0,
			"avg_price":        0,
			"avg_daily_price":  0,
		}
	}

	statusStats := make(map[string]int)
	roomCountStats := make(map[int]int)
	totalArea := 0.0
	totalPrice := 0
	totalDailyPrice := 0

	for _, apt := range apartments {
		statusStats[string(apt.Status)]++
		roomCountStats[apt.RoomCount]++
		totalArea += apt.TotalArea
		totalPrice += apt.Price
		totalDailyPrice += apt.DailyPrice
	}

	apartmentCount := len(apartments)
	avgArea := totalArea / float64(apartmentCount)
	avgPrice := float64(totalPrice) / float64(apartmentCount)
	avgDailyPrice := float64(totalDailyPrice) / float64(apartmentCount)

	return gin.H{
		"total_apartments": apartmentCount,
		"by_status":        statusStats,
		"by_room_count":    roomCountStats,
		"avg_area":         avgArea,
		"avg_price":        avgPrice,
		"avg_daily_price":  avgDailyPrice,
	}
}

func (h *ApartmentHandler) calculateOwnerBookingStatistics(apartments []*domain.Apartment, dateFrom, dateTo time.Time) (gin.H, error) {
	if len(apartments) == 0 {
		return gin.H{
			"total_bookings":     0,
			"bookings":           []*domain.Booking{},
			"by_status":          map[string]int{},
			"by_apartment":       map[int]int{},
			"popular_apartments": []gin.H{},
		}, nil
	}

	var allBookings []*domain.Booking
	apartmentIDs := make([]int, len(apartments))
	for i, apt := range apartments {
		apartmentIDs[i] = apt.ID
	}

	filters := map[string]interface{}{
		"date_from": dateFrom.Format("2006-01-02"),
		"date_to":   dateTo.Format("2006-01-02"),
	}

	statusStats := make(map[string]int)
	apartmentStats := make(map[int]int)

	for _, apartmentID := range apartmentIDs {
		apartmentFilters := make(map[string]interface{})
		for k, v := range filters {
			apartmentFilters[k] = v
		}
		apartmentFilters["apartment_id"] = apartmentID

		bookings, _, err := h.bookingUseCase.AdminGetAllBookings(apartmentFilters, 1, 1000)
		if err != nil {
			continue
		}

		for _, booking := range bookings {
			allBookings = append(allBookings, booking)
			statusStats[string(booking.Status)]++
			apartmentStats[booking.ApartmentID]++
		}
	}

	type apartmentRating struct {
		ApartmentID  int    `json:"apartment_id"`
		Street       string `json:"street"`
		Building     string `json:"building"`
		ApartmentNum int    `json:"apartment_number"`
		BookingCount int    `json:"booking_count"`
	}

	var popularApartments []apartmentRating
	for _, apt := range apartments {
		count := apartmentStats[apt.ID]
		if count > 0 {
			popularApartments = append(popularApartments, apartmentRating{
				ApartmentID:  apt.ID,
				Street:       apt.Street,
				Building:     apt.Building,
				ApartmentNum: apt.ApartmentNumber,
				BookingCount: count,
			})
		}
	}

	for i := 0; i < len(popularApartments)-1; i++ {
		for j := i + 1; j < len(popularApartments); j++ {
			if popularApartments[i].BookingCount < popularApartments[j].BookingCount {
				popularApartments[i], popularApartments[j] = popularApartments[j], popularApartments[i]
			}
		}
	}

	if len(popularApartments) > 5 {
		popularApartments = popularApartments[:5]
	}

	return gin.H{
		"total_bookings":     len(allBookings),
		"bookings":           allBookings,
		"by_status":          statusStats,
		"by_apartment":       apartmentStats,
		"popular_apartments": popularApartments,
	}, nil
}

func (h *ApartmentHandler) calculateOwnerFinancialStatistics(bookings []*domain.Booking) gin.H {
	if len(bookings) == 0 {
		return gin.H{
			"total_revenue":       0,
			"avg_booking_value":   0,
			"revenue_by_month":    map[string]int{},
			"revenue_by_duration": map[string]int{},
		}
	}

	totalRevenue := 0
	monthlyRevenue := make(map[string]int)
	durationRevenue := make(map[string]int)

	for _, booking := range bookings {
		if booking.Status == domain.BookingStatusCompleted {
			totalRevenue += booking.FinalPrice

			month := booking.CreatedAt.Format("2006-01")
			monthlyRevenue[month] += booking.FinalPrice

			var durationCategory string
			if booking.Duration <= 3 {
				durationCategory = "short"
			} else if booking.Duration <= 12 {
				durationCategory = "medium"
			} else if booking.Duration < 24 {
				durationCategory = "long"
			} else {
				durationCategory = "daily"
			}
			durationRevenue[durationCategory] += booking.FinalPrice
		}
	}

	avgBookingValue := 0.0
	completedBookings := 0
	for _, booking := range bookings {
		if booking.Status == domain.BookingStatusCompleted {
			completedBookings++
		}
	}

	if completedBookings > 0 {
		avgBookingValue = float64(totalRevenue) / float64(completedBookings)
	}

	return gin.H{
		"total_revenue":       totalRevenue,
		"avg_booking_value":   avgBookingValue,
		"revenue_by_month":    monthlyRevenue,
		"revenue_by_duration": durationRevenue,
		"completed_bookings":  completedBookings,
	}
}

func (h *ApartmentHandler) calculateApartmentEfficiencyStatistics(apartments []*domain.Apartment, bookings []*domain.Booking) gin.H {
	if len(apartments) == 0 {
		return gin.H{
			"apartment_performance": []gin.H{},
			"top_performers":        []gin.H{},
			"low_performers":        []gin.H{},
		}
	}

	apartmentBookings := make(map[int][]*domain.Booking)
	for _, booking := range bookings {
		apartmentBookings[booking.ApartmentID] = append(apartmentBookings[booking.ApartmentID], booking)
	}

	type apartmentPerformance struct {
		ApartmentID       int     `json:"apartment_id"`
		Street            string  `json:"street"`
		Building          string  `json:"building"`
		ApartmentNumber   int     `json:"apartment_number"`
		Status            string  `json:"status"`
		TotalBookings     int     `json:"total_bookings"`
		CompletedBookings int     `json:"completed_bookings"`
		TotalRevenue      int     `json:"total_revenue"`
		AvgRevenue        float64 `json:"avg_revenue"`
		OccupancyRate     float64 `json:"occupancy_rate"`
	}

	var performances []apartmentPerformance

	for _, apt := range apartments {
		bookingList := apartmentBookings[apt.ID]
		totalBookings := len(bookingList)
		completedBookings := 0
		totalRevenue := 0

		for _, booking := range bookingList {
			if booking.Status == domain.BookingStatusCompleted {
				completedBookings++
				totalRevenue += booking.FinalPrice
			}
		}

		avgRevenue := 0.0
		if completedBookings > 0 {
			avgRevenue = float64(totalRevenue) / float64(completedBookings)
		}

		occupancyRate := 0.0
		if totalBookings > 0 && apt.Status == domain.AptStatusApproved {
			occupancyRate = float64(completedBookings) / 30.0 * 100
			if occupancyRate > 100 {
				occupancyRate = 100
			}
		}

		performances = append(performances, apartmentPerformance{
			ApartmentID:       apt.ID,
			Street:            apt.Street,
			Building:          apt.Building,
			ApartmentNumber:   apt.ApartmentNumber,
			Status:            string(apt.Status),
			TotalBookings:     totalBookings,
			CompletedBookings: completedBookings,
			TotalRevenue:      totalRevenue,
			AvgRevenue:        avgRevenue,
			OccupancyRate:     occupancyRate,
		})
	}

	for i := 0; i < len(performances)-1; i++ {
		for j := i + 1; j < len(performances); j++ {
			if performances[i].TotalRevenue < performances[j].TotalRevenue {
				performances[i], performances[j] = performances[j], performances[i]
			}
		}
	}

	var topPerformers, lowPerformers []apartmentPerformance

	for i, perf := range performances {
		if i < 3 && perf.TotalRevenue > 0 {
			topPerformers = append(topPerformers, perf)
		}
		if i >= len(performances)-3 && perf.TotalRevenue == 0 {
			lowPerformers = append(lowPerformers, perf)
		}
	}

	return gin.H{
		"apartment_performance": performances,
		"top_performers":        topPerformers,
		"low_performers":        lowPerformers,
	}
}

// @Summary История бронирований квартиры (админ)
// @Description Возвращает детальную историю всех бронирований конкретной квартиры с аналитикой
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
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
// @Router /admin/apartments/{id}/bookings-history [get]
func (h *ApartmentHandler) AdminGetApartmentBookingsHistory(c *gin.Context) {
	apartmentID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(apartmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
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

	filters := make(map[string]interface{})
	filters["apartment_id"] = apartmentID

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

	analytics := h.calculateApartmentAnalytics(bookings, apartment)

	bookingResponses := make([]gin.H, len(bookings))
	for i, booking := range bookings {
		bookingResponses[i] = gin.H{
			"id":             booking.ID,
			"booking_number": booking.BookingNumber,
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
		"apartment": gin.H{
			"id":               apartment.ID,
			"description":      apartment.Description,
			"street":           apartment.Street,
			"building":         apartment.Building,
			"apartment_number": apartment.ApartmentNumber,
			"price":            apartment.Price,
			"daily_price":      apartment.DailyPrice,
			"status":           apartment.Status,
		},
		"bookings": bookingResponses,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"pages":     (total + pageSize - 1) / pageSize,
		},
		"analytics": analytics,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("история бронирований квартиры получена", response))
}

func (h *ApartmentHandler) calculateApartmentAnalytics(bookings []*domain.Booking, apartment *domain.Apartment) gin.H {
	var totalRevenue, totalDuration int
	var completedBookings, canceledBookings, activeBookings int
	statusCounts := make(map[string]int)
	monthlyStats := make(map[string]int)

	for _, booking := range bookings {
		statusCounts[string(booking.Status)]++

		switch booking.Status {
		case domain.BookingStatusCompleted:
			completedBookings++
			totalRevenue += booking.FinalPrice
		case domain.BookingStatusCanceled:
			canceledBookings++
		case domain.BookingStatusActive:
			activeBookings++
		}

		totalDuration += booking.Duration

		month := booking.StartDate.Format("2006-01")
		monthlyStats[month]++
	}

	var avgPrice, avgDuration float64
	occupancyRate := 0.0

	if len(bookings) > 0 {
		avgDuration = float64(totalDuration) / float64(len(bookings))
		if completedBookings > 0 {
			avgPrice = float64(totalRevenue) / float64(completedBookings)
		}

		// Упрощенный расчет заполняемости (можно улучшить)
		if len(bookings) > 0 {
			occupancyRate = float64(completedBookings+activeBookings) / float64(len(bookings)) * 100
		}
	}

	return gin.H{
		"total_bookings":     len(bookings),
		"completed_bookings": completedBookings,
		"canceled_bookings":  canceledBookings,
		"active_bookings":    activeBookings,
		"total_revenue":      totalRevenue,
		"avg_price":          avgPrice,
		"avg_duration":       avgDuration,
		"occupancy_rate":     occupancyRate,
		"status_breakdown":   statusCounts,
		"monthly_stats":      monthlyStats,
		"revenue_per_hour":   float64(totalRevenue) / math.Max(float64(totalDuration), 1),
	}
}

// @Summary Обновление типа квартиры (админ)
// @Description Обновляет тип квартиры (только для админов и модераторов)
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Param request body domain.UpdateApartmentTypeIDRequest true "Данные для обновления типа квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartments/{id}/apartment-type [put]
func (h *ApartmentHandler) UpdateApartmentType(c *gin.Context) {
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

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав для обновления типа квартиры"))
		return
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	var req domain.UpdateApartmentTypeIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректные данные запроса"))
		return
	}

	var apartmentTypeID int
	if req.ApartmentTypeID != nil {
		apartmentTypeID = *req.ApartmentTypeID
	} else {
		apartmentTypeID = 0
	}

	err = h.apartmentUseCase.UpdateApartmentType(id, apartmentTypeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении типа квартиры: "+err.Error()))
		return
	}

	updatedApartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusOK, domain.NewSuccessResponse("тип квартиры обновлен", nil))
		return
	}

	enrichedApartment := h.enrichApartmentWithLocationData(updatedApartment)
	c.JSON(http.StatusOK, domain.NewSuccessResponse("тип квартиры обновлен", enrichedApartment))
}

// @Summary Обновление счетчиков квартиры (админ)
// @Description Позволяет админам изменять счетчики просмотров и бронирований
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID квартиры"
// @Param request body domain.AdminUpdateCountersRequest true "Данные для обновления счетчиков"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartments/{id}/counters [put]
func (h *ApartmentHandler) AdminUpdateCounters(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	var req domain.AdminUpdateCountersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректные данные запроса"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	if req.ViewCount != nil {
		if err := h.apartmentUseCase.AdminUpdateViewCount(id, *req.ViewCount); err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении счетчика просмотров"))
			return
		}
	}

	if req.BookingCount != nil {
		if err := h.apartmentUseCase.AdminUpdateBookingCount(id, *req.BookingCount); err != nil {
			c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении счетчика бронирований"))
			return
		}
	}

	updatedApartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении обновленных данных"))
		return
	}

	enrichedApartment := h.enrichApartmentWithLocationData(updatedApartment)
	c.JSON(http.StatusOK, domain.NewSuccessResponse("счетчики обновлены", enrichedApartment))
}

// @Summary Сброс счетчиков квартиры (админ)
// @Description Сбрасывает счетчики просмотров и бронирований в ноль
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID квартиры"
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/apartments/{id}/counters/reset [post]
func (h *ApartmentHandler) AdminResetCounters(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("некорректный ID квартиры"))
		return
	}

	apartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных квартиры"))
		return
	}
	if apartment == nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse("квартира не найдена"))
		return
	}

	if err := h.apartmentUseCase.AdminResetCounters(id); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при сбросе счетчиков"))
		return
	}

	updatedApartment, err := h.apartmentUseCase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении обновленных данных"))
		return
	}

	enrichedApartment := h.enrichApartmentWithLocationData(updatedApartment)
	c.JSON(http.StatusOK, domain.NewSuccessResponse("счетчики сброшены", enrichedApartment))
}
