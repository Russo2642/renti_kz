package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/pkg/auth"
)

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         domain.User `json:"user"`
}

type AuthHandler struct {
	authUseCase       domain.AuthUseCase
	userUseCase       domain.UserUseCase
	otpUseCase        domain.OTPUseCase
	tokenManager      auth.TokenManager
	renterRepo        domain.RenterRepository
	propertyOwnerRepo domain.PropertyOwnerRepository
	renterUseCase     domain.RenterUseCase
}

func NewAuthHandler(
	authUseCase domain.AuthUseCase,
	userUseCase domain.UserUseCase,
	otpUseCase domain.OTPUseCase,
	tokenManager auth.TokenManager,
	renterRepo domain.RenterRepository,
	propertyOwnerRepo domain.PropertyOwnerRepository,
	renterUseCase domain.RenterUseCase,
) *AuthHandler {
	return &AuthHandler{
		authUseCase:       authUseCase,
		userUseCase:       userUseCase,
		otpUseCase:        otpUseCase,
		tokenManager:      tokenManager,
		renterRepo:        renterRepo,
		propertyOwnerRepo: propertyOwnerRepo,
		renterUseCase:     renterUseCase,
	}
}

func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/pre-register", h.PreRegister)
		auth.POST("/complete-register", h.CompleteRegistration)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
		auth.GET("/check-phone/:phone", h.CheckPhoneExists)

		auth.POST("/otp/request", h.RequestOTP)
		auth.POST("/otp/verify", h.VerifyOTP)
		auth.GET("/otp/status/:id", h.CheckOTPStatus)
	}
}

// @Summary Проверка существования телефона
// @Description Проверяет, зарегистрирован ли пользователь с данным номером телефона
// @Tags auth
// @Accept json
// @Produce json
// @Param phone path string true "Номер телефона"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/check-phone/{phone} [get]
func (h *AuthHandler) CheckPhoneExists(c *gin.Context) {
	phone := c.Param("phone")

	if phone == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("телефон обязателен"))
		return
	}

	if len(phone) != 11 || !isNumeric(phone) {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("телефон должен содержать 11 цифр без символов"))
		return
	}

	user, err := h.userUseCase.GetByPhone(phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке телефона"))
		return
	}

	exists := user != nil
	c.JSON(http.StatusOK, domain.NewSuccessResponse("проверка выполнена", gin.H{"exists": exists}))
}

// @Summary Предварительная регистрация
// @Description Первый этап регистрации - отправка данных пользователя и запрос OTP
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.PreRegistrationRequest true "Данные для предварительной регистрации"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/pre-register [post]
func (h *AuthHandler) PreRegister(c *gin.Context) {
	var req domain.PreRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	// Проверяем, что пользователь не существует
	existingUser, err := h.userUseCase.GetByPhone(req.Phone)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusConflict, domain.NewErrorResponse("пользователь с таким номером телефона уже существует"))
		return
	}

	// Проверяем email
	existingUser, err = h.userUseCase.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusConflict, domain.NewErrorResponse("пользователь с таким email уже существует"))
		return
	}

	// Отправляем OTP
	otpResponse, err := h.otpUseCase.RequestOTP(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при отправке OTP"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("OTP отправлен на номер телефона", gin.H{"otp_id": otpResponse.ID}))
}

// @Summary Завершение регистрации
// @Description Второй этап регистрации - проверка OTP и создание пользователя с автоматической авторизацией
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.CompleteRegistrationRequest true "Данные для завершения регистрации"
// @Success 201 {object} LoginResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/complete-register [post]
func (h *AuthHandler) CompleteRegistration(c *gin.Context) {
	var req domain.CompleteRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	isValid, err := h.otpUseCase.VerifyOTP(req.Phone, req.OTPCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке OTP"))
		return
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный OTP код"))
		return
	}

	user := &domain.User{
		Phone:     req.Phone,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		CityID:    req.CityID,
		IIN:       req.IIN,
		Role:      domain.RoleUser,
	}

	err = h.userUseCase.RegisterWithoutPassword(user)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, domain.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при регистрации пользователя"))
		return
	}

	registeredUser, err := h.userUseCase.GetByPhone(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("пользователь зарегистрирован, но не удалось получить данные"))
		return
	}

	accessToken, refreshToken, err := h.tokenManager.GenerateTokenPair(registeredUser.ID, string(registeredUser.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("пользователь зарегистрирован, но не удалось создать токены"))
		return
	}

	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *registeredUser,
	}

	c.JSON(http.StatusCreated, domain.NewSuccessResponse("пользователь успешно зарегистрирован и авторизован", response))
}

// @Summary Регистрация пользователя
// @Description Регистрирует нового пользователя в системе
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	user := &domain.User{
		Phone:     req.Phone,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		CityID:    req.CityID,
		IIN:       req.IIN,
		Role:      domain.RoleUser,
	}

	err := h.userUseCase.RegisterWithoutPassword(user)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, domain.NewErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при регистрации пользователя"))
		return
	}

	registeredUser, err := h.userUseCase.GetByPhone(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("пользователь зарегистрирован, но не удалось получить данные"))
		return
	}

	c.JSON(http.StatusCreated, domain.NewSuccessResponse("пользователь успешно зарегистрирован", registeredUser))
}

// @Summary Авторизация пользователя
// @Description Авторизует пользователя по номеру телефона и паролю
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Данные для входа"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	tokens, user, err := h.authUseCase.SignIn(req.Phone, req.Password)
	if err != nil {

		if strings.Contains(err.Error(), "invalid credentials") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("неверные учетные данные"))
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при входе в систему"))
		return
	}

	response := LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         *user,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("вход выполнен успешно", response))
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// @Summary Обновление токенов
// @Description Обновляет access и refresh токены
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh токен"
// @Success 200 {object} RefreshTokenResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	tokens, err := h.authUseCase.RefreshTokens(req.RefreshToken)
	if err != nil {

		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("недействительный или истекший токен"))
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при обновлении токена"))
		return
	}

	response := RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("токены успешно обновлены", response))
}

type LogoutRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
}

// @Summary Выход из системы
// @Description Аннулирует токен пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Access токен"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	_, err := h.tokenManager.ParseAccessToken(req.AccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.NewErrorResponse("недействительный токен"))
		return
	}

	if err := h.authUseCase.SignOut(req.AccessToken); err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при выходе из системы"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("выход выполнен успешно", nil))
}

func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

// @Summary Запрос OTP кода
// @Description Отправляет OTP код на указанный номер телефона
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.OTPRequest true "Номер телефона для OTP"
// @Success 200 {object} domain.OTPRequestResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/otp/request [post]
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var req domain.OTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	if len(req.Phone) != 11 || !isNumeric(req.Phone) {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("телефон должен содержать 11 цифр без символов"))
		return
	}

	response, err := h.otpUseCase.RequestOTP(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при отправке OTP кода"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("OTP код отправлен", response))
}

// @Summary Проверка OTP кода и аутентификация
// @Description Проверяет OTP код и выполняет аутентификацию пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.OTPVerifyRequest true "Данные для проверки OTP"
// @Success 200 {object} domain.OTPAuthResponse
// @Failure 400 {object} domain.ErrorResponse "Неверный OTP код или неверные данные запроса"
// @Failure 404 {object} domain.ErrorResponse "Сессия OTP не найдена"
// @Failure 409 {object} domain.ErrorResponse "OTP код уже был использован"
// @Failure 410 {object} domain.ErrorResponse "Время действия OTP кода истекло"
// @Failure 500 {object} domain.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/otp/verify [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req domain.OTPVerifyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse(err.Error()))
		return
	}

	if len(req.Phone) != 11 || !isNumeric(req.Phone) {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("телефон должен содержать 11 цифр без символов"))
		return
	}

	if len(req.Code) != 4 || !isNumeric(req.Code) {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("код должен содержать 4 цифры"))
		return
	}

	response, err := h.otpUseCase.VerifyOTPAndAuthenticate(req.ID, req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке OTP кода"))
		return
	}

	if response.RequiresRegistration {
		c.JSON(http.StatusOK, domain.NewSuccessResponse(response.Message, gin.H{
			"requires_registration": true,
			"phone":                 req.Phone,
		}))
		return
	}

	if response.AccessToken == nil {
		var statusCode int
		switch response.ErrorType {
		case domain.OTPErrorNotFound:
			statusCode = http.StatusNotFound
		case domain.OTPErrorPhoneMismatch:
			statusCode = http.StatusBadRequest
		case domain.OTPErrorExpired:
			statusCode = http.StatusGone
		case domain.OTPErrorAlreadyUsed:
			statusCode = http.StatusConflict
		case domain.OTPErrorInvalidCode:
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, domain.NewErrorResponse(response.Message))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse(response.Message, gin.H{
		"requires_registration": false,
		"access_token":          *response.AccessToken,
		"refresh_token":         *response.RefreshToken,
		"user":                  response.User,
	}))
}

// @Summary Проверка статуса OTP
// @Description Проверяет статус OTP сессии
// @Tags auth
// @Accept json
// @Produce json
// @Param id path string true "ID OTP сессии"
// @Success 200 {object} domain.OTPStatusResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /auth/otp/status/{id} [get]
func (h *AuthHandler) CheckOTPStatus(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("ID сессии обязателен"))
		return
	}

	response, err := h.otpUseCase.CheckStatus(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при проверке статуса OTP"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("статус получен", response))
}
