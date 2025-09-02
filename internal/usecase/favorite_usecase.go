package usecase

import (
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
)

type favoriteUseCase struct {
	favoriteRepo      domain.FavoriteRepository
	apartmentRepo     domain.ApartmentRepository
	userRepo          domain.UserRepository
	propertyOwnerRepo domain.PropertyOwnerRepository
}

func NewFavoriteUseCase(
	favoriteRepo domain.FavoriteRepository,
	apartmentRepo domain.ApartmentRepository,
	userRepo domain.UserRepository,
	propertyOwnerRepo domain.PropertyOwnerRepository,
) domain.FavoriteUseCase {
	return &favoriteUseCase{
		favoriteRepo:      favoriteRepo,
		apartmentRepo:     apartmentRepo,
		userRepo:          userRepo,
		propertyOwnerRepo: propertyOwnerRepo,
	}
}

func (uc *favoriteUseCase) AddToFavorites(userID, apartmentID int) error {
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	apartment, err := uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return fmt.Errorf("apartment not found: %w", err)
	}
	if apartment == nil {
		return fmt.Errorf("apartment not found")
	}

	if apartment.Status != domain.AptStatusApproved {
		return fmt.Errorf("apartment is not available for favorites")
	}

	err = uc.favoriteRepo.AddToFavorites(userID, apartmentID)
	if err != nil {
		return fmt.Errorf("failed to add to favorites: %w", err)
	}

	return nil
}

func (uc *favoriteUseCase) RemoveFromFavorites(userID, apartmentID int) error {
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	err = uc.favoriteRepo.RemoveFromFavorites(userID, apartmentID)
	if err != nil {
		return fmt.Errorf("failed to remove from favorites: %w", err)
	}

	return nil
}

func (uc *favoriteUseCase) GetUserFavorites(userID int, page, pageSize int) ([]*domain.Favorite, int, error) {
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("user not found: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	favorites, total, err := uc.favoriteRepo.GetUserFavorites(userID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user favorites: %w", err)
	}

	for _, favorite := range favorites {
		photos, err := uc.apartmentRepo.GetPhotosByApartmentID(favorite.ApartmentID)
		if err == nil {
			favorite.Apartment.Photos = photos
		}

		location, err := uc.apartmentRepo.GetLocationByApartmentID(favorite.ApartmentID)
		if err == nil {
			favorite.Apartment.Location = location
		}

		amenities, err := uc.apartmentRepo.GetAmenitiesByApartmentID(favorite.ApartmentID)
		if err == nil {
			favorite.Apartment.Amenities = amenities
		}

		houseRules, err := uc.apartmentRepo.GetHouseRulesByApartmentID(favorite.ApartmentID)
		if err == nil {
			favorite.Apartment.HouseRules = houseRules
		}
	}

	return favorites, total, nil
}

func (uc *favoriteUseCase) IsFavorite(userID, apartmentID int) (bool, error) {
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	isFavorite, err := uc.favoriteRepo.IsFavorite(userID, apartmentID)
	if err != nil {
		return false, fmt.Errorf("failed to check favorite status: %w", err)
	}

	return isFavorite, nil
}

func (uc *favoriteUseCase) GetFavoriteCount(userID, apartmentID int) (int, error) {
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return 0, fmt.Errorf("user not found: %w", err)
	}

	apartment, err := uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return 0, fmt.Errorf("apartment not found: %w", err)
	}
	if apartment == nil {
		return 0, fmt.Errorf("apartment not found")
	}

	owner, err := uc.propertyOwnerRepo.GetByUserID(userID)
	if err != nil {
		return 0, fmt.Errorf("user is not a property owner")
	}
	if owner == nil {
		return 0, fmt.Errorf("user is not a property owner")
	}

	if apartment.OwnerID != owner.ID {
		return 0, fmt.Errorf("access denied: you can only view favorite count for your own apartments")
	}

	count, err := uc.favoriteRepo.GetFavoriteCount(apartmentID)
	if err != nil {
		return 0, fmt.Errorf("failed to get favorite count: %w", err)
	}

	return count, nil
}

func (uc *favoriteUseCase) ToggleFavorite(userID, apartmentID int) (bool, error) {
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	apartment, err := uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return false, fmt.Errorf("apartment not found: %w", err)
	}
	if apartment == nil {
		return false, fmt.Errorf("apartment not found")
	}

	if apartment.Status != domain.AptStatusApproved {
		return false, fmt.Errorf("apartment is not available for favorites")
	}

	isFavorite, err := uc.favoriteRepo.IsFavorite(userID, apartmentID)
	if err != nil {
		return false, fmt.Errorf("failed to check favorite status: %w", err)
	}

	if isFavorite {
		err = uc.favoriteRepo.RemoveFromFavorites(userID, apartmentID)
		if err != nil {
			return false, fmt.Errorf("failed to remove from favorites: %w", err)
		}
		return false, nil
	} else {
		err = uc.favoriteRepo.AddToFavorites(userID, apartmentID)
		if err != nil {
			return false, fmt.Errorf("failed to add to favorites: %w", err)
		}
		return true, nil
	}
}
