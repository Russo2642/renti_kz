package utils

import (
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
)

func LoadBookingRelatedData(booking *domain.Booking, apartmentRepo domain.ApartmentRepository, renterRepo domain.RenterRepository, ownerRepo domain.PropertyOwnerRepository) {
	if apartment, err := apartmentRepo.GetByID(booking.ApartmentID); err == nil {
		booking.Apartment = apartment

		if photos, photoErr := apartmentRepo.GetPhotosByApartmentID(apartment.ID); photoErr == nil {
			booking.Apartment.Photos = photos
		}

		if location, locationErr := apartmentRepo.GetLocationByApartmentID(apartment.ID); locationErr == nil {
			booking.Apartment.Location = location
		}

		if owner, err := ownerRepo.GetByIDWithUser(apartment.OwnerID); err == nil {
			booking.Apartment.Owner = owner
		}
	}

	if renter, err := renterRepo.GetByIDWithUser(booking.RenterID); err == nil {
		booking.Renter = renter
	}
}

func LoadBookingsRelatedData(bookings []*domain.Booking, apartmentRepo domain.ApartmentRepository, renterRepo domain.RenterRepository, ownerRepo domain.PropertyOwnerRepository) {
	for _, booking := range bookings {
		LoadBookingRelatedData(booking, apartmentRepo, renterRepo, ownerRepo)
	}
}

func ValidateOwnerAccess(userID, apartmentID int, apartmentRepo domain.ApartmentRepository, ownerRepo domain.PropertyOwnerRepository) error {
	apartment, err := apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return fmt.Errorf("квартира не найдена: %w", err)
	}

	owner, err := ownerRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("владелец не найден: %w", err)
	}

	if owner == nil {
		return fmt.Errorf("пользователь не является владельцем недвижимости")
	}

	if apartment.OwnerID != owner.ID {
		return fmt.Errorf("нет прав для управления этой квартирой")
	}

	return nil
}

func CheckUserExists(userID int, userUseCase domain.UserUseCase) (*domain.User, error) {
	user, err := userUseCase.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("пользователь не найден")
	}
	return user, nil
}

func ValidateBookingOwnership(booking *domain.Booking, userID int, userUseCase domain.UserUseCase, renterRepo domain.RenterRepository, ownerRepo domain.PropertyOwnerRepository, apartmentRepo domain.ApartmentRepository) (bool, error) {
	user, err := CheckUserExists(userID, userUseCase)
	if err != nil {
		return false, err
	}

	if IsAdminOrModerator(user) {
		return true, nil
	}

	renter, err := renterRepo.GetByUserID(userID)
	if err == nil && renter != nil && renter.ID == booking.RenterID {
		return true, nil
	}

	apartment, err := apartmentRepo.GetByID(booking.ApartmentID)
	if err != nil {
		return false, err
	}

	owner, err := ownerRepo.GetByUserID(userID)
	if err == nil && owner != nil && owner.ID == apartment.OwnerID {
		return true, nil
	}

	return false, nil
}

func CreateBookingNumber(prefix string, id int) string {
	return fmt.Sprintf("%s%d", prefix, id)
}
