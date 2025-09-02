package usecase

import (
	"fmt"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/pkg/storage/s3"
)

type PropertyOwnerUseCase struct {
	propertyOwnerRepo domain.PropertyOwnerRepository
	userRepo          domain.UserRepository
	roleRepo          domain.RoleRepository
	s3Storage         *s3.Storage
	passwordSalt      string
}

func NewPropertyOwnerUseCase(
	propertyOwnerRepo domain.PropertyOwnerRepository,
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	s3Storage *s3.Storage,
	passwordSalt string,
) *PropertyOwnerUseCase {
	return &PropertyOwnerUseCase{
		propertyOwnerRepo: propertyOwnerRepo,
		userRepo:          userRepo,
		roleRepo:          roleRepo,
		s3Storage:         s3Storage,
		passwordSalt:      passwordSalt,
	}
}

func (uc *PropertyOwnerUseCase) Register(owner *domain.PropertyOwner, user *domain.User, password string) error {
	userUseCase := NewUserUseCase(uc.userRepo, uc.roleRepo, uc.passwordSalt)
	user.Role = domain.RoleOwner

	if err := userUseCase.Register(user, password); err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	owner.UserID = user.ID

	if err := uc.propertyOwnerRepo.Create(owner); err != nil {
		uc.userRepo.Delete(user.ID)
		return fmt.Errorf("failed to create property owner: %w", err)
	}

	return nil
}

func (uc *PropertyOwnerUseCase) GetByID(id int) (*domain.PropertyOwner, error) {
	owner, err := uc.propertyOwnerRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get property owner by id: %w", err)
	}

	user, err := uc.userRepo.GetByID(owner.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for property owner: %w", err)
	}

	owner.User = user
	return owner, nil
}

func (uc *PropertyOwnerUseCase) GetByUserID(userID int) (*domain.PropertyOwner, error) {
	owner, err := uc.propertyOwnerRepo.GetByUserIDWithUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property owner by user id: %w", err)
	}

	return owner, nil
}

func (uc *PropertyOwnerUseCase) UpdateProfile(owner *domain.PropertyOwner) error {
	existingOwner, err := uc.propertyOwnerRepo.GetByID(owner.ID)
	if err != nil {
		return fmt.Errorf("failed to get property owner: %w", err)
	}
	if existingOwner == nil {
		return fmt.Errorf("property owner with id %d not found", owner.ID)
	}

	existingOwner.UpdatedAt = time.Now()

	if err := uc.propertyOwnerRepo.Update(existingOwner); err != nil {
		return fmt.Errorf("failed to update property owner: %w", err)
	}

	return nil
}
