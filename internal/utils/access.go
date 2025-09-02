package utils

import (
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
)

func GetUserWithRoleCheck(userUseCase domain.UserUseCase, userID int, allowedRoles ...domain.UserRole) (*domain.User, error) {
	user, err := userUseCase.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении данных пользователя: %w", err)
	}

	if len(allowedRoles) > 0 {
		hasRole := false
		for _, role := range allowedRoles {
			if user.Role == role {
				hasRole = true
				break
			}
		}
		if !hasRole {
			return nil, fmt.Errorf("недостаточно прав")
		}
	}

	return user, nil
}

func GetRenterByUserID(renterRepo domain.RenterRepository, userID int) (*domain.Renter, error) {
	renter, err := renterRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("арендатор не найден: %w", err)
	}
	if renter == nil {
		return nil, fmt.Errorf("пользователь не является арендатором")
	}
	return renter, nil
}

func GetPropertyOwnerByUserID(ownerRepo domain.PropertyOwnerRepository, userID int) (*domain.PropertyOwner, error) {
	owner, err := ownerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("владелец не найден: %w", err)
	}
	if owner == nil {
		return nil, fmt.Errorf("пользователь не является владельцем недвижимости")
	}
	return owner, nil
}

func CheckOwnershipAccess(ownerRepo domain.PropertyOwnerRepository, userID, apartmentOwnerID int) (bool, error) {
	owner, err := ownerRepo.GetByUserID(userID)
	if err != nil {
		return false, err
	}
	if owner == nil {
		return false, nil
	}
	return owner.ID == apartmentOwnerID, nil
}

func IsAdmin(user *domain.User) bool {
	return user.Role == domain.RoleAdmin
}

func IsModerator(user *domain.User) bool {
	return user.Role == domain.RoleModerator
}

func IsAdminOrModerator(user *domain.User) bool {
	return IsAdmin(user) || IsModerator(user)
}
