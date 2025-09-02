package usecase

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/pkg/logger"
)

type apartmentTypeUseCase struct {
	apartmentTypeRepo domain.ApartmentTypeRepository
	userUseCase       domain.UserUseCase
}

func NewApartmentTypeUseCase(
	apartmentTypeRepo domain.ApartmentTypeRepository,
	userUseCase domain.UserUseCase,
) domain.ApartmentTypeUseCase {
	return &apartmentTypeUseCase{
		apartmentTypeRepo: apartmentTypeRepo,
		userUseCase:       userUseCase,
	}
}

func (u *apartmentTypeUseCase) Create(request *domain.CreateApartmentTypeRequest, adminID int) (*domain.ApartmentType, error) {
	user, err := u.userUseCase.GetByID(adminID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator {
		return nil, fmt.Errorf("недостаточно прав для создания типов квартир")
	}

	apartmentType := &domain.ApartmentType{
		Name:        strings.TrimSpace(request.Name),
		Description: strings.TrimSpace(request.Description),
	}

	err = u.apartmentTypeRepo.Create(apartmentType)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания типа квартиры: %w", err)
	}

	logger.Info("apartment type created",
		slog.Int("apartment_type_id", apartmentType.ID),
		slog.String("name", apartmentType.Name),
		slog.Int("admin_id", adminID))

	return apartmentType, nil
}

func (u *apartmentTypeUseCase) GetByID(id int) (*domain.ApartmentType, error) {
	apartmentType, err := u.apartmentTypeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("тип квартиры не найден: %w", err)
	}

	return apartmentType, nil
}

func (u *apartmentTypeUseCase) GetAll() ([]*domain.ApartmentType, error) {
	apartmentTypes, err := u.apartmentTypeRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения типов квартир: %w", err)
	}

	return apartmentTypes, nil
}

func (u *apartmentTypeUseCase) Update(id int, request *domain.UpdateApartmentTypeRequest, adminID int) (*domain.ApartmentType, error) {
	user, err := u.userUseCase.GetByID(adminID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}

	if user.Role != domain.RoleAdmin && user.Role != domain.RoleModerator {
		return nil, fmt.Errorf("недостаточно прав для обновления типов квартир")
	}

	apartmentType, err := u.apartmentTypeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("тип квартиры не найден: %w", err)
	}

	apartmentType.Name = strings.TrimSpace(request.Name)
	apartmentType.Description = strings.TrimSpace(request.Description)

	err = u.apartmentTypeRepo.Update(apartmentType)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления типа квартиры: %w", err)
	}

	logger.Info("apartment type updated",
		slog.Int("apartment_type_id", apartmentType.ID),
		slog.String("name", apartmentType.Name),
		slog.Int("admin_id", adminID))

	return apartmentType, nil
}

func (u *apartmentTypeUseCase) Delete(id int, adminID int) error {
	user, err := u.userUseCase.GetByID(adminID)
	if err != nil {
		return fmt.Errorf("пользователь не найден: %w", err)
	}

	if user.Role != domain.RoleAdmin {
		return fmt.Errorf("недостаточно прав для удаления типов квартир")
	}

	apartmentType, err := u.apartmentTypeRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("тип квартиры не найден: %w", err)
	}

	err = u.apartmentTypeRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("ошибка удаления типа квартиры: %w", err)
	}

	logger.Info("apartment type deleted",
		slog.Int("apartment_type_id", id),
		slog.String("name", apartmentType.Name),
		slog.Int("admin_id", adminID))

	return nil
}
