package domain

import (
	"time"
)

type Favorite struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	ApartmentID int        `json:"apartment_id"`
	User        *User      `json:"user,omitempty"`
	Apartment   *Apartment `json:"apartment,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type FavoriteRepository interface {
	AddToFavorites(userID, apartmentID int) error

	RemoveFromFavorites(userID, apartmentID int) error

	GetUserFavorites(userID int, page, pageSize int) ([]*Favorite, int, error)

	IsFavorite(userID, apartmentID int) (bool, error)

	GetFavoriteCount(apartmentID int) (int, error)

	GetByID(id int) (*Favorite, error)

	GetByUserAndApartment(userID, apartmentID int) (*Favorite, error)
}

type FavoriteUseCase interface {
	AddToFavorites(userID, apartmentID int) error

	RemoveFromFavorites(userID, apartmentID int) error

	GetUserFavorites(userID int, page, pageSize int) ([]*Favorite, int, error)

	IsFavorite(userID, apartmentID int) (bool, error)

	GetFavoriteCount(userID, apartmentID int) (int, error)

	ToggleFavorite(userID, apartmentID int) (bool, error)
}
