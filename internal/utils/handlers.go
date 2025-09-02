package utils

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
)

func GetOwnerIDForUser(ownerUseCase domain.PropertyOwnerUseCase, userID int) (int, error) {
	owner, err := ownerUseCase.GetByUserID(userID)
	if err != nil {
		return 0, err
	}
	if owner == nil {
		return 0, fmt.Errorf("профиль владельца не найден")
	}
	return owner.ID, nil
}

func RequireOwnerRole(c *gin.Context, userUseCase domain.UserUseCase, ownerUseCase domain.PropertyOwnerUseCase) (int, int, bool) {
	userID, ok := RequireAuth(c)
	if !ok {
		return 0, 0, false
	}

	user, err := userUseCase.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении данных пользователя"))
		return 0, 0, false
	}

	if user.Role != domain.RoleOwner && user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator {
		c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав"))
		return 0, 0, false
	}

	if user.Role == domain.RoleOwner {
		ownerID, err := GetOwnerIDForUser(ownerUseCase, userID)
		if err != nil {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("профиль владельца не найден"))
			return 0, 0, false
		}
		return userID, ownerID, true
	}

	return userID, 0, true
}

func CheckOwnerPermission(c *gin.Context, userID int, user *domain.User, apartmentOwnerID int, ownerUseCase domain.PropertyOwnerUseCase) bool {
	if IsAdminOrModerator(user) {
		return true
	}

	if user.Role == domain.RoleOwner {
		owner, err := ownerUseCase.GetByUserID(userID)
		if err != nil || owner == nil {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("профиль владельца не найден"))
			return false
		}

		if owner.ID != apartmentOwnerID {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse("нет прав для изменения этой квартиры"))
			return false
		}
		return true
	}

	c.JSON(http.StatusForbidden, domain.NewErrorResponse("недостаточно прав"))
	return false
}

func ParseOptionalQueryInt(c *gin.Context, paramName string) *int {
	if param := c.Query(paramName); param != "" {
		if value, err := strconv.Atoi(param); err == nil {
			return &value
		}
	}
	return nil
}

func ParseOptionalQueryFloat(c *gin.Context, paramName string) *float64 {
	if param := c.Query(paramName); param != "" {
		if value, err := strconv.ParseFloat(param, 64); err == nil {
			return &value
		}
	}
	return nil
}

func ParseStatusFilter(c *gin.Context) []domain.BookingStatus {
	statusParam := c.Query("status")
	var statuses []domain.BookingStatus
	if statusParam != "" {
		statusStrings := strings.Split(statusParam, ",")
		for _, s := range statusStrings {
			statuses = append(statuses, domain.BookingStatus(strings.TrimSpace(s)))
		}
	}
	return statuses
}

func ConvertBookingToResponse(booking *domain.Booking) *domain.BookingResponse {


	return &domain.BookingResponse{
		ID:                 booking.ID,
		RenterID:           booking.RenterID,
		Renter:             booking.Renter,
		ApartmentID:        booking.ApartmentID,
		Apartment:          booking.Apartment,
		ContractID:         booking.ContractID,

		StartDate:          FormatForUser(booking.StartDate),
		EndDate:            FormatForUser(booking.EndDate),
		Duration:           booking.Duration,
		CleaningDuration:   booking.CleaningDuration,
		Status:             booking.Status,
		TotalPrice:         booking.TotalPrice,
		ServiceFee:         booking.ServiceFee,
		FinalPrice:         booking.FinalPrice,
		IsContractAccepted: booking.IsContractAccepted,
		CancellationReason: booking.CancellationReason,
		OwnerComment:       booking.OwnerComment,
		BookingNumber:      booking.BookingNumber,
		DoorStatus:         booking.DoorStatus,
		LastDoorAction:     FormatForUserPtr(booking.LastDoorAction),
		CanExtend:          booking.CanExtend,
		ExtensionRequested: booking.ExtensionRequested,
		ExtensionEndDate:   FormatForUserPtr(booking.ExtensionEndDate),
		ExtensionDuration:  booking.ExtensionDuration,
		ExtensionPrice:     booking.ExtensionPrice,
		CreatedAt:          FormatForUser(booking.CreatedAt),
		UpdatedAt:          FormatForUser(booking.UpdatedAt),
	}
}

func ConvertBookingsToResponse(bookings []*domain.Booking) []*domain.BookingResponse {
	responses := make([]*domain.BookingResponse, len(bookings))
	for i, booking := range bookings {
		responses[i] = ConvertBookingToResponse(booking)
	}
	return responses
}

func ValidateCreateBookingRequest(request *domain.CreateBookingRequest) error {
	if request.ApartmentID <= 0 {
		return fmt.Errorf("некорректный ID квартиры")
	}

	if request.Duration <= 0 {
		return fmt.Errorf("продолжительность должна быть больше 0")
	}

	if request.StartDate == "" {
		return fmt.Errorf("необходимо указать дату начала")
	}

	return nil
}

func HandleFileUpload(c *gin.Context, fieldName string) ([]byte, string, error) {
	file, header, err := c.Request.FormFile(fieldName)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка получения файла: %w", err)
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка чтения файла: %w", err)
	}

	return fileData, header.Filename, nil
}

func ValidateImageUpload(filename string) error {
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	filename = strings.ToLower(filename)
	for _, ext := range validExtensions {
		if strings.HasSuffix(filename, ext) {
			return nil
		}
	}

	return fmt.Errorf("поддерживаются только изображения (jpg, jpeg, png, gif, webp)")
}

func ShouldBindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorResponse("неверный формат данных: "+err.Error()))
		return false
	}
	return true
}

func ValidateStruct(c *gin.Context, obj interface{}) bool {
	// Здесь можно добавить дополнительную валидацию если нужно
	// Пока что просто возвращаем true, так как gin уже провалидировал JSON
	return true
}
