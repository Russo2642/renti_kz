package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type ContractHandler struct {
	contractUseCase domain.ContractUseCase
	userUseCase     domain.UserUseCase
}

func NewContractHandler(contractUseCase domain.ContractUseCase, userUseCase domain.UserUseCase) *ContractHandler {
	return &ContractHandler{
		contractUseCase: contractUseCase,
		userUseCase:     userUseCase,
	}
}

func (h *ContractHandler) RegisterRoutes(api *gin.RouterGroup) {
	contracts := api.Group("/contracts")
	{
		contracts.GET("/:id", h.GetContractByID)
		contracts.GET("/:id/html", h.GetContractHTML)
		contracts.GET("/booking/:booking_id", h.GetContractByBookingID)
		contracts.GET("/apartment/:apartment_id", h.GetApartmentContract)
		contracts.GET("/my/rental", h.GetUserRentalContracts)
		contracts.GET("/my/apartment", h.GetOwnerApartmentContracts)
		contracts.PUT("/:id/status", h.UpdateContractStatus)
		contracts.PUT("/:id/confirm", h.ConfirmContract)
		contracts.GET("", h.GetAllContracts)
	}

	bookings := api.Group("/bookings")
	{
		bookings.POST("/:id/contract", h.CreateRentalContract)
	}

	apartments := api.Group("/apartments")
	{
		apartments.POST("/:id/contract", h.CreateApartmentContract)
		apartments.GET("/:id/contract", h.GetApartmentContract)
	}
}

// @Summary Получение договора по ID
// @Description Получает договор аренды по его идентификатору с мета-данными
// @Tags contracts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID договора"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /contracts/{id} [get]
func (h *ContractHandler) GetContractByID(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	contractID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	canAccess, err := h.contractUseCase.CanUserAccessContract(contractID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("Ошибка проверки прав доступа"))
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("Недостаточно прав для просмотра договора"))
		return
	}

	contract, err := h.contractUseCase.GetContractByID(contractID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
		return
	}

	response := h.toContractResponse(contract)
	c.JSON(http.StatusOK, domain.NewSuccessResponse("Договор получен", response))
}

// @Summary Получение HTML договора
// @Description Получает HTML содержимое договора для просмотра/печати
// @Tags contracts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID договора"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /contracts/{id}/html [get]
func (h *ContractHandler) GetContractHTML(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	contractID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	canAccess, err := h.contractUseCase.CanUserAccessContract(contractID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("Ошибка проверки прав доступа"))
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("Недостаточно прав для просмотра договора"))
		return
	}

	contract, err := h.contractUseCase.GetContractByID(contractID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
		return
	}

	html, err := h.contractUseCase.GetContractHTML(contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	response := &domain.ContractHTMLResponse{
		ContractID: contractID,
		HTML:       html,
		Status:     contract.Status,
		CachedAt:   "", // TODO: добавить время кэширования из Redis если нужно
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("HTML договора получен", response))
}

// @Summary Получение договора по ID бронирования
// @Description Получает договор аренды по идентификатору бронирования
// @Tags contracts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param booking_id path int true "ID бронирования"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /contracts/booking/{booking_id} [get]
func (h *ContractHandler) GetContractByBookingID(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "booking_id")
	if !ok {
		return
	}

	contract, err := h.contractUseCase.GetContractByBookingID(bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
		return
	}

	canAccess, err := h.contractUseCase.CanUserAccessContract(contract.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("Ошибка проверки прав доступа"))
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("Недостаточно прав для просмотра договора"))
		return
	}

	response := h.toContractResponse(contract)
	c.JSON(http.StatusOK, domain.NewSuccessResponse("Договор получен", response))
}

// @Summary Получение договора квартиры
// @Description Получает договор между компанией и арендодателем для квартиры
// @Tags contracts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param apartment_id path int true "ID квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /contracts/apartment/{apartment_id} [get]
func (h *ContractHandler) GetApartmentContract(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	apartmentID, ok := utils.ParseIDParam(c, "apartment_id")
	if !ok {
		return
	}

	contract, err := h.contractUseCase.GetApartmentContract(apartmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.NewErrorResponse(err.Error()))
		return
	}

	canAccess, err := h.contractUseCase.CanUserAccessContract(contract.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("Ошибка проверки прав доступа"))
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("Недостаточно прав для просмотра договора"))
		return
	}

	response := h.toContractResponse(contract)
	c.JSON(http.StatusOK, domain.NewSuccessResponse("Договор квартиры получен", response))
}

// @Summary Получение rental договоров пользователя
// @Description Получает список договоров аренды текущего пользователя как арендатора
// @Tags contracts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /contracts/my/rental [get]
func (h *ContractHandler) GetUserRentalContracts(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	contracts, err := h.contractUseCase.GetUserRentalContracts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	responses := make([]*domain.ContractResponse, len(contracts))
	for i, contract := range contracts {
		responses[i] = h.toContractResponse(contract)
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("Договоры аренды получены", responses))
}

// @Summary Получение договоров квартир владельца
// @Description Получает список всех договоров квартир текущего пользователя как арендодателя
// @Tags contracts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /contracts/my/apartment [get]
func (h *ContractHandler) GetOwnerApartmentContracts(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	// Находим ID арендодателя по userID
	// TODO: Добавить метод в PropertyOwnerUseCase для получения по userID
	// Пока используем простую заглушку

	contracts, err := h.contractUseCase.GetOwnerApartmentContracts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	responses := make([]*domain.ContractResponse, len(contracts))
	for i, contract := range contracts {
		responses[i] = h.toContractResponse(contract)
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("Договоры квартир получены", responses))
}

type UpdateContractStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft confirmed signed"`
}

// @Summary Обновление статуса договора
// @Description Обновляет статус договора (только для администраторов)
// @Tags contracts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID договора"
// @Param request body UpdateContractStatusRequest true "Новый статус"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /contracts/{id}/status [put]
func (h *ContractHandler) UpdateContractStatus(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	user, err := h.userUseCase.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("Ошибка получения данных пользователя"))
		return
	}

	if user.Role != domain.RoleAdmin {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("Недостаточно прав доступа"))
		return
	}

	contractID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var req UpdateContractStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("Неверный формат данных: "+err.Error()))
		return
	}

	status := domain.ContractStatus(req.Status)
	err = h.contractUseCase.UpdateContractStatus(contractID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("Статус договора обновлен", nil))
}

// @Summary Подтверждение договора
// @Description Подтверждает договор (меняет статус на confirmed)
// @Tags contracts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID договора"
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /contracts/{id}/confirm [put]
func (h *ContractHandler) ConfirmContract(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	contractID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	canAccess, err := h.contractUseCase.CanUserAccessContract(contractID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("Ошибка проверки прав доступа"))
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("Недостаточно прав для подтверждения договора"))
		return
	}

	err = h.contractUseCase.ConfirmContract(contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("Договор подтвержден", nil))
}

// @Summary Получение всех договоров
// @Description Получает список всех договоров с пагинацией (только для администраторов)
// @Tags contracts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(50)
// @Success 200 {object} domain.SuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /contracts [get]
func (h *ContractHandler) GetAllContracts(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	user, err := h.userUseCase.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("Ошибка получения данных пользователя"))
		return
	}

	if user.Role != domain.RoleAdmin {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("Недостаточно прав доступа"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	contracts, err := h.contractUseCase.GetAllContracts(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	responses := make([]*domain.ContractResponse, len(contracts))
	for i, contract := range contracts {
		responses[i] = h.toContractResponse(contract)
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("Договоры получены", responses))
}

// @Summary Создание rental договора для бронирования
// @Description Создает новый договор аренды для указанного бронирования
// @Tags contracts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID бронирования"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /bookings/{id}/contract [post]
func (h *ContractHandler) CreateRentalContract(c *gin.Context) {
	_, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	bookingID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	contract, err := h.contractUseCase.CreateRentalContract(bookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	response := h.toContractResponse(contract)
	c.JSON(http.StatusCreated, domain.NewSuccessResponse("Договор аренды создан", response))
}

// @Summary Создание apartment договора
// @Description Создает новый договор между компанией и арендодателем для квартиры
// @Tags contracts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID квартиры"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /apartments/{id}/contract [post]
func (h *ContractHandler) CreateApartmentContract(c *gin.Context) {
	_, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	apartmentID, ok := utils.ParseIDParam(c, "id")
	if !ok {
		return
	}

	contract, err := h.contractUseCase.CreateApartmentContract(apartmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse(err.Error()))
		return
	}

	response := h.toContractResponse(contract)
	c.JSON(http.StatusCreated, domain.NewSuccessResponse("Договор квартиры создан", response))
}

func (h *ContractHandler) toContractResponse(contract *domain.Contract) *domain.ContractResponse {
	response := &domain.ContractResponse{
		ID:              contract.ID,
		Type:            contract.Type,
		ApartmentID:     contract.ApartmentID,
		BookingID:       contract.BookingID,
		TemplateVersion: contract.TemplateVersion,
		Status:          contract.Status,
		IsActive:        contract.IsActive,
		CreatedAt:       contract.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       contract.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if contract.ExpiresAt != nil {
		expiresAt := contract.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		response.ExpiresAt = &expiresAt
	}

	if contract.Apartment != nil {
		response.Apartment = contract.Apartment
	}

	if contract.Booking != nil {
		response.Booking = contract.Booking
	}

	return response
}
