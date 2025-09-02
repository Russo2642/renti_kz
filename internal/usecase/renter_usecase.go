package usecase

import (
	"fmt"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/pkg/storage/s3"
)

type RenterUseCase struct {
	renterRepo   domain.RenterRepository
	userRepo     domain.UserRepository
	roleRepo     domain.RoleRepository
	s3Storage    *s3.Storage
	passwordSalt string
}

func NewRenterUseCase(
	renterRepo domain.RenterRepository,
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	s3Storage *s3.Storage,
	passwordSalt string,
) *RenterUseCase {
	return &RenterUseCase{
		renterRepo:   renterRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		s3Storage:    s3Storage,
		passwordSalt: passwordSalt,
	}
}

func (uc *RenterUseCase) Register(renter *domain.Renter, user *domain.User, password string) error {
	userUseCase := NewUserUseCase(uc.userRepo, uc.roleRepo, uc.passwordSalt)
	user.Role = domain.RoleUser

	if err := userUseCase.Register(user, password); err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	renter.UserID = user.ID

	if err := uc.renterRepo.Create(renter); err != nil {
		uc.userRepo.Delete(user.ID)
		return fmt.Errorf("failed to create renter: %w", err)
	}

	return nil
}

func (uc *RenterUseCase) GetByID(id int) (*domain.Renter, error) {
	renter, err := uc.renterRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get renter by id: %w", err)
	}
	return renter, nil
}

func (uc *RenterUseCase) GetByUserID(userID int) (*domain.Renter, error) {
	renter, err := uc.renterRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get renter by user id: %w", err)
	}
	return renter, nil
}

func (uc *RenterUseCase) UpdateProfile(renter *domain.Renter) error {
	existingRenter, err := uc.renterRepo.GetByID(renter.ID)
	if err != nil {
		return fmt.Errorf("failed to get renter: %w", err)
	}
	if existingRenter == nil {
		return fmt.Errorf("renter with id %d not found", renter.ID)
	}

	if renter.DocumentType != "" {
		existingRenter.DocumentType = renter.DocumentType
	}
	if renter.PhotoWithDocURL != "" {
		existingRenter.PhotoWithDocURL = renter.PhotoWithDocURL
	}
	if len(renter.DocumentURL) > 0 {
		existingRenter.DocumentURL = renter.DocumentURL
	}
	existingRenter.UpdatedAt = time.Now()

	if err := uc.renterRepo.Update(existingRenter); err != nil {
		return fmt.Errorf("failed to update renter: %w", err)
	}

	return nil
}

func (uc *RenterUseCase) UploadDocument(renterID int, documentType domain.DocumentType, fileData []byte) (string, error) {
	renter, err := uc.renterRepo.GetByID(renterID)
	if err != nil {
		return "", fmt.Errorf("failed to get renter: %w", err)
	}
	if renter == nil {
		return "", fmt.Errorf("renter with id %d not found", renterID)
	}

	user, err := uc.userRepo.GetByID(renter.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to get user for renter: %w", err)
	}

	url, err := uc.s3Storage.UploadUserDocument(user.Phone, string(documentType), fileData)
	if err != nil {
		return "", fmt.Errorf("failed to upload document: %w", err)
	}

	if renter.DocumentURL == nil {
		renter.DocumentURL = make(map[string]string)
	}
	renter.DocumentURL[string(documentType)] = url
	renter.DocumentType = documentType

	if err := uc.renterRepo.Update(renter); err != nil {
		return "", fmt.Errorf("failed to update renter document info: %w", err)
	}

	return url, nil
}

func (uc *RenterUseCase) UploadDocumentsParallel(renterID int, documentType domain.DocumentType, documentsData [][]byte) (map[string]string, error) {
	renter, err := uc.renterRepo.GetByID(renterID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении арендатора: %w", err)
	}

	if renter == nil {
		return nil, fmt.Errorf("арендатор с ID %d не найден", renterID)
	}

	user, err := uc.userRepo.GetByID(renter.UserID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	urls, err := uc.s3Storage.UploadUserDocumentsParallel(user.Phone, string(documentType), documentsData)
	if err != nil {
		return nil, fmt.Errorf("ошибка при загрузке документов в S3: %w", err)
	}

	documentURLs := make(map[string]string, len(urls))
	for i, url := range urls {
		page := fmt.Sprintf("page%d", i+1)
		documentURLs[page] = url
	}

	return documentURLs, nil
}

func (uc *RenterUseCase) UploadPhotoWithDoc(renterID int, fileData []byte) (string, error) {
	renter, err := uc.renterRepo.GetByID(renterID)
	if err != nil {
		return "", fmt.Errorf("failed to get renter: %w", err)
	}
	if renter == nil {
		return "", fmt.Errorf("renter with id %d not found", renterID)
	}

	user, err := uc.userRepo.GetByID(renter.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to get user for renter: %w", err)
	}

	url, err := uc.s3Storage.UploadUserPhotoWithDoc(user.Phone, fileData)
	if err != nil {
		return "", fmt.Errorf("failed to upload photo: %w", err)
	}

	renter.PhotoWithDocURL = url

	if err := uc.renterRepo.Update(renter); err != nil {
		return "", fmt.Errorf("failed to update renter photo info: %w", err)
	}

	return url, nil
}

func (uc *RenterUseCase) UpdateVerificationStatus(renterID int, status domain.VerificationStatus) error {
	renter, err := uc.renterRepo.GetByID(renterID)
	if err != nil {
		return fmt.Errorf("failed to get renter: %w", err)
	}
	if renter == nil {
		return fmt.Errorf("renter with id %d not found", renterID)
	}

	renter.VerificationStatus = status

	if err := uc.renterRepo.Update(renter); err != nil {
		return fmt.Errorf("failed to update verification status: %w", err)
	}

	return nil
}
