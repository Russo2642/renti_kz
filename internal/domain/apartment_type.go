package domain

import (
	"time"
)

type ApartmentType struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateApartmentTypeRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateApartmentTypeRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
}

type ApartmentTypeRepository interface {
	Create(apartmentType *ApartmentType) error
	GetByID(id int) (*ApartmentType, error)
	GetAll() ([]*ApartmentType, error)
	Update(apartmentType *ApartmentType) error
	Delete(id int) error
}

type ApartmentTypeUseCase interface {
	Create(request *CreateApartmentTypeRequest, adminID int) (*ApartmentType, error)
	GetByID(id int) (*ApartmentType, error)
	GetAll() ([]*ApartmentType, error)
	Update(id int, request *UpdateApartmentTypeRequest, adminID int) (*ApartmentType, error)
	Delete(id int, adminID int) error
}
