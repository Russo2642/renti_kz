package usecase

import (
	"fmt"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
)

type contractUseCase struct {
	contractRepo      domain.ContractRepository
	contractService   domain.ContractService
	bookingRepo       domain.BookingRepository
	apartmentRepo     domain.ApartmentRepository
	userRepo          domain.UserRepository
	renterRepo        domain.RenterRepository
	propertyOwnerRepo domain.PropertyOwnerRepository
}

func NewContractUseCase(
	contractRepo domain.ContractRepository,
	contractService domain.ContractService,
	bookingRepo domain.BookingRepository,
	apartmentRepo domain.ApartmentRepository,
	userRepo domain.UserRepository,
	renterRepo domain.RenterRepository,
	propertyOwnerRepo domain.PropertyOwnerRepository,
) domain.ContractUseCase {
	return &contractUseCase{
		contractRepo:      contractRepo,
		contractService:   contractService,
		bookingRepo:       bookingRepo,
		apartmentRepo:     apartmentRepo,
		userRepo:          userRepo,
		renterRepo:        renterRepo,
		propertyOwnerRepo: propertyOwnerRepo,
	}
}

func (u *contractUseCase) CreateRentalContract(bookingID int) (*domain.Contract, error) {
	exists, err := u.contractRepo.ExistsForBooking(bookingID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существования договора: %w", err)
	}
	if exists {
		return u.contractRepo.GetByBookingID(bookingID)
	}

	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("бронирование не найдено: %w", err)
	}

	snapshot, err := u.createRentalSnapshot(booking)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания снапшота: %w", err)
	}

	contract := &domain.Contract{
		Type:            domain.ContractTypeRental,
		ApartmentID:     booking.ApartmentID,
		BookingID:       &bookingID,
		TemplateVersion: 1,
		Status:          domain.ContractStatusDraft,
		IsActive:        true,
		ExpiresAt:       nil,
	}

	err = contract.SetSnapshotData(snapshot)
	if err != nil {
		return nil, fmt.Errorf("ошибка установки снапшота: %w", err)
	}

	err = u.contractRepo.Create(contract)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания договора: %w", err)
	}

	go func() {
		_ = u.contractService.WarmupContractCache(contract.ID)
	}()

	return contract, nil
}

func (u *contractUseCase) CreateApartmentContract(apartmentID int) (*domain.Contract, error) {
	existing, err := u.contractRepo.GetByApartmentID(apartmentID, domain.ContractTypeApartment)
	if err == nil && existing != nil {
		return existing, nil
	}

	apartment, err := u.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("квартира не найдена: %w", err)
	}

	owner, err := u.propertyOwnerRepo.GetByID(apartment.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("владелец не найден: %w", err)
	}

	user, err := u.userRepo.GetByID(owner.UserID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}

	snapshot, err := u.createApartmentSnapshot(apartment, user)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания снапшота: %w", err)
	}

	contract := &domain.Contract{
		Type:            domain.ContractTypeApartment,
		ApartmentID:     apartmentID,
		BookingID:       nil,
		TemplateVersion: 1,
		Status:          domain.ContractStatusDraft,
		IsActive:        true,
		ExpiresAt:       nil,
	}

	err = contract.SetSnapshotData(snapshot)
	if err != nil {
		return nil, fmt.Errorf("ошибка установки снапшота: %w", err)
	}

	err = u.contractRepo.Create(contract)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания договора: %w", err)
	}

	go func() {
		_ = u.contractService.WarmupContractCache(contract.ID)
	}()

	return contract, nil
}

func (u *contractUseCase) GetContractByID(id int) (*domain.Contract, error) {
	contract, err := u.contractRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("договор не найден: %w", err)
	}

	err = u.loadContractRelations(contract)
	if err != nil {
		return nil, err
	}

	return contract, nil
}

func (u *contractUseCase) GetContractByBookingID(bookingID int) (*domain.Contract, error) {
	contract, err := u.contractRepo.GetByBookingID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("договор для бронирования не найден: %w", err)
	}

	err = u.loadContractRelations(contract)
	if err != nil {
		return nil, err
	}

	return contract, nil
}

func (u *contractUseCase) GetApartmentContract(apartmentID int) (*domain.Contract, error) {
	contract, err := u.contractRepo.GetByApartmentID(apartmentID, domain.ContractTypeApartment)
	if err != nil {
		return nil, fmt.Errorf("договор квартиры не найден: %w", err)
	}

	err = u.loadContractRelations(contract)
	if err != nil {
		return nil, err
	}

	return contract, nil
}

func (u *contractUseCase) GetContractHTML(contractID int) (string, error) {
	_, err := u.contractRepo.GetByID(contractID)
	if err != nil {
		return "", fmt.Errorf("договор не найден: %w", err)
	}

	html, err := u.contractService.GetOrGenerateContractHTML(contractID)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации HTML: %w", err)
	}

	return html, nil
}

func (u *contractUseCase) UpdateContractStatus(contractID int, status domain.ContractStatus) error {
	contract, err := u.contractRepo.GetByID(contractID)
	if err != nil {
		return fmt.Errorf("договор не найден: %w", err)
	}

	contract.Status = status
	err = u.contractRepo.Update(contract)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса: %w", err)
	}

	_ = u.contractService.InvalidateContractCache(contractID)

	return nil
}

func (u *contractUseCase) ConfirmContract(contractID int) error {
	return u.UpdateContractStatus(contractID, domain.ContractStatusConfirmed)
}

func (u *contractUseCase) CanUserAccessContract(contractID, userID int) (bool, error) {
	contract, err := u.contractRepo.GetByID(contractID)
	if err != nil {
		return false, err
	}

	user, err := u.userRepo.GetByID(userID)
	if err == nil && (user.Role == domain.RoleAdmin || user.Role == domain.RoleModerator) {
		return true, nil
	}

	if contract.IsRentalContract() && contract.BookingID != nil {
		booking, err := u.bookingRepo.GetByID(*contract.BookingID)
		if err == nil {
			renter, err := u.renterRepo.GetByID(booking.RenterID)
			if err == nil && renter.UserID == userID {
				return true, nil
			}
		}
	}

	apartment, err := u.apartmentRepo.GetByID(contract.ApartmentID)
	if err == nil {
		owner, err := u.propertyOwnerRepo.GetByID(apartment.OwnerID)
		if err == nil && owner.UserID == userID {
			return true, nil
		}
	}

	return false, nil
}

func (u *contractUseCase) GetUserRentalContracts(userID int) ([]*domain.Contract, error) {
	renter, err := u.renterRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не является арендатором: %w", err)
	}

	bookings, _, err := u.bookingRepo.GetByRenterID(renter.ID, nil, nil, nil, 1, 1000)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения бронирований: %w", err)
	}

	var contracts []*domain.Contract
	for _, booking := range bookings {
		contract, err := u.contractRepo.GetByBookingID(booking.ID)
		if err == nil {
			err = u.loadContractRelations(contract)
			if err == nil {
				contracts = append(contracts, contract)
			}
		}
	}

	return contracts, nil
}

func (u *contractUseCase) GetOwnerApartmentContracts(ownerID int) ([]*domain.Contract, error) {
	apartments, err := u.apartmentRepo.GetByOwnerID(ownerID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения квартир: %w", err)
	}

	var contracts []*domain.Contract
	for _, apartment := range apartments {
		apartmentContracts, err := u.contractRepo.GetActiveByApartmentID(apartment.ID)
		if err == nil {
			for _, contract := range apartmentContracts {
				err = u.loadContractRelations(contract)
				if err == nil {
					contracts = append(contracts, contract)
				}
			}
		}
	}

	return contracts, nil
}

func (u *contractUseCase) GetAllContracts(limit, offset int) ([]*domain.Contract, error) {
	contracts, err := u.contractRepo.GetAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения договоров: %w", err)
	}

	for _, contract := range contracts {
		_ = u.loadContractRelations(contract)
	}

	return contracts, nil
}

func (u *contractUseCase) RefreshContractData(contractID int) error {
	return u.contractService.InvalidateContractCache(contractID)
}

func (u *contractUseCase) createRentalSnapshot(booking *domain.Booking) (*domain.RentalContractSnapshot, error) {
	apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
	if err != nil {
		return nil, fmt.Errorf("квартира не найдена: %w", err)
	}

	renter, err := u.renterRepo.GetByID(booking.RenterID)
	if err != nil {
		return nil, fmt.Errorf("арендатор не найден: %w", err)
	}

	renterUser, err := u.userRepo.GetByID(renter.UserID)
	if err != nil {
		return nil, fmt.Errorf("пользователь арендатора не найден: %w", err)
	}

	owner, err := u.propertyOwnerRepo.GetByID(apartment.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("арендодатель не найден: %w", err)
	}

	ownerUser, err := u.userRepo.GetByID(owner.UserID)
	if err != nil {
		return nil, fmt.Errorf("пользователь арендодателя не найден: %w", err)
	}

	apartmentAddress := fmt.Sprintf("г. Алматы, ул. %s, д. %s, кв. %d",
		apartment.Street, apartment.Building, apartment.ApartmentNumber)
	apartmentTitle := fmt.Sprintf("%d-комнатная квартира, %.1f м²",
		apartment.RoomCount, apartment.TotalArea)

	return &domain.RentalContractSnapshot{
		BookingNumber:    booking.BookingNumber,
		StartDate:        booking.StartDate,
		EndDate:          booking.EndDate,
		Duration:         booking.Duration,
		TotalPrice:       booking.TotalPrice,
		ServiceFee:       booking.ServiceFee,
		FinalPrice:       booking.FinalPrice,
		RenterName:       fmt.Sprintf("%s %s", renterUser.FirstName, renterUser.LastName),
		RenterIIN:        renterUser.IIN,
		OwnerName:        fmt.Sprintf("%s %s", ownerUser.FirstName, ownerUser.LastName),
		OwnerIIN:         ownerUser.IIN,
		ApartmentAddress: apartmentAddress,
		ApartmentTitle:   apartmentTitle,
		ContractDate:     time.Now().Format("02.01.2006"),
		CreatedByUserID:  renterUser.ID,
	}, nil
}

func (u *contractUseCase) createApartmentSnapshot(apartment *domain.Apartment, user *domain.User) (*domain.ApartmentContractSnapshot, error) {
	snapshot := &domain.ApartmentContractSnapshot{
		ContractDate:    time.Now().Format("02.01.2006"),
		CreatedByUserID: user.ID,
	}

	return snapshot, nil
}

func (u *contractUseCase) loadContractRelations(contract *domain.Contract) error {
	if apartment, err := u.apartmentRepo.GetByID(contract.ApartmentID); err == nil {
		contract.Apartment = apartment
	}

	if contract.BookingID != nil {
		if booking, err := u.bookingRepo.GetByID(*contract.BookingID); err == nil {
			contract.Booking = booking
		}
	}

	return nil
}
