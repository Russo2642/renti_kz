package usecase

import (
	"fmt"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
)

type conciergeUseCase struct {
	conciergeRepo domain.ConciergeRepository
	userRepo      domain.UserRepository
	apartmentRepo domain.ApartmentRepository
	roleRepo      domain.RoleRepository
	bookingRepo   domain.BookingRepository
	chatRoomRepo  domain.ChatRoomRepository
}

func NewConciergeUseCase(
	conciergeRepo domain.ConciergeRepository,
	userRepo domain.UserRepository,
	apartmentRepo domain.ApartmentRepository,
	roleRepo domain.RoleRepository,
	bookingRepo domain.BookingRepository,
	chatRoomRepo domain.ChatRoomRepository,
) domain.ConciergeUseCase {
	return &conciergeUseCase{
		conciergeRepo: conciergeRepo,
		userRepo:      userRepo,
		apartmentRepo: apartmentRepo,
		roleRepo:      roleRepo,
		bookingRepo:   bookingRepo,
		chatRoomRepo:  chatRoomRepo,
	}
}

func (uc *conciergeUseCase) CreateConcierge(request *domain.CreateConciergeRequest, adminID int) (*domain.Concierge, error) {
	if err := uc.validateAdminAccess(adminID); err != nil {
		return nil, err
	}

	user, err := uc.userRepo.GetByID(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var apartments []*domain.Apartment
	for _, apartmentID := range request.ApartmentIDs {
		apartment, err := uc.apartmentRepo.GetByID(apartmentID)
		if err != nil {
			return nil, fmt.Errorf("apartment with ID %d not found: %w", apartmentID, err)
		}
		apartments = append(apartments, apartment)
	}

	isAlreadyConcierge, err := uc.conciergeRepo.IsUserConcierge(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user is concierge: %w", err)
	}
	if isAlreadyConcierge {
		return nil, fmt.Errorf("user is already a concierge")
	}

	concierge := &domain.Concierge{
		UserID:   request.UserID,
		IsActive: true,
		Schedule: request.Schedule,
	}

	err = uc.conciergeRepo.Create(concierge)
	if err != nil {
		return nil, fmt.Errorf("failed to create concierge: %w", err)
	}

	for _, apartmentID := range request.ApartmentIDs {
		err = uc.conciergeRepo.AssignToApartment(concierge.ID, apartmentID)
		if err != nil {
			uc.conciergeRepo.Delete(concierge.ID)
			return nil, fmt.Errorf("failed to assign concierge to apartment %d: %w", apartmentID, err)
		}
	}

	err = uc.userRepo.UpdateRole(request.UserID, domain.RoleConcierge)
	if err != nil {
		uc.conciergeRepo.Delete(concierge.ID)
		return nil, fmt.Errorf("failed to update user role to concierge: %w", err)
	}

	user.Role = domain.RoleConcierge

	concierge.User = user
	concierge.Apartments = apartments

	return concierge, nil
}

func (uc *conciergeUseCase) GetConciergeByID(id int) (*domain.Concierge, error) {
	concierge, err := uc.conciergeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get concierge: %w", err)
	}

	return concierge, nil
}

func (uc *conciergeUseCase) GetConciergesByApartment(apartmentID int) ([]*domain.Concierge, error) {
	concierges, err := uc.conciergeRepo.GetByApartmentIDActive(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get concierges by apartment: %w", err)
	}

	return concierges, nil
}

func (uc *conciergeUseCase) GetAllConcierges(filters map[string]interface{}, page, pageSize int) ([]*domain.Concierge, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	concierges, total, err := uc.conciergeRepo.GetAll(filters, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get concierges: %w", err)
	}

	return concierges, total, nil
}

func (uc *conciergeUseCase) UpdateConcierge(id int, request *domain.UpdateConciergeRequest, adminID int) error {
	if err := uc.validateAdminAccess(adminID); err != nil {
		return err
	}

	concierge, err := uc.conciergeRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("concierge not found: %w", err)
	}

	concierge.IsActive = request.IsActive
	if request.Schedule != nil {
		concierge.Schedule = request.Schedule
	}

	err = uc.conciergeRepo.Update(concierge)
	if err != nil {
		return fmt.Errorf("failed to update concierge: %w", err)
	}

	return nil
}

func (uc *conciergeUseCase) DeleteConcierge(id int, adminID int) error {
	if err := uc.validateAdminAccess(adminID); err != nil {
		return err
	}

	concierge, err := uc.conciergeRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("concierge not found: %w", err)
	}

	err = uc.conciergeRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete concierge: %w", err)
	}

	err = uc.userRepo.UpdateRole(concierge.UserID, domain.RoleUser)
	if err != nil {
		fmt.Printf("Warning: failed to update user role back to user for user %d: %v\n", concierge.UserID, err)
	}

	return nil
}

func (uc *conciergeUseCase) GetConciergesByOwner(ownerID int) ([]*domain.Concierge, error) {
	concierges, err := uc.conciergeRepo.GetConciergesByOwner(ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get concierges by owner: %w", err)
	}

	return concierges, nil
}

func (uc *conciergeUseCase) IsUserConcierge(userID int) (bool, error) {
	isActive, err := uc.conciergeRepo.IsUserConcierge(userID)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is concierge: %w", err)
	}

	return isActive, nil
}

func (uc *conciergeUseCase) AssignConciergeToApartment(conciergeID, apartmentID, adminID int) error {
	if err := uc.validateAdminAccess(adminID); err != nil {
		return err
	}

	_, err := uc.conciergeRepo.GetByID(conciergeID)
	if err != nil {
		return fmt.Errorf("concierge not found: %w", err)
	}

	_, err = uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return fmt.Errorf("apartment not found: %w", err)
	}

	err = uc.conciergeRepo.AssignToApartment(conciergeID, apartmentID)
	if err != nil {
		return fmt.Errorf("failed to assign concierge to apartment: %w", err)
	}

	return nil
}

func (uc *conciergeUseCase) RemoveConciergeFromApartment(conciergeID, apartmentID, adminID int) error {
	if err := uc.validateAdminAccess(adminID); err != nil {
		return err
	}

	_, err := uc.conciergeRepo.GetByID(conciergeID)
	if err != nil {
		return fmt.Errorf("concierge not found: %w", err)
	}

	_, err = uc.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return fmt.Errorf("apartment not found: %w", err)
	}

	err = uc.conciergeRepo.RemoveFromApartment(conciergeID, apartmentID)
	if err != nil {
		return fmt.Errorf("failed to remove concierge from apartment: %w", err)
	}

	return nil
}

func (uc *conciergeUseCase) validateAdminAccess(userID int) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.Role != domain.RoleAdmin {
		return fmt.Errorf("access denied: admin role required")
	}

	return nil
}

func (uc *conciergeUseCase) GetConciergeByUserID(userID int) (*domain.Concierge, error) {
	concierge, err := uc.conciergeRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("concierge not found: %w", err)
	}

	return concierge, nil
}

func (uc *conciergeUseCase) GetConciergeApartments(userID int) ([]*domain.Apartment, error) {
	concierge, err := uc.conciergeRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("concierge not found: %w", err)
	}

	return concierge.Apartments, nil
}

func (uc *conciergeUseCase) GetConciergeBookings(userID int, filters map[string]interface{}, page, pageSize int) ([]*domain.Booking, int, error) {
	concierge, err := uc.conciergeRepo.GetByUserID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("concierge not found: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	if len(concierge.Apartments) == 0 {
		return []*domain.Booking{}, 0, nil
	}

	var apartmentIDs []int
	for _, apartment := range concierge.Apartments {
		apartmentIDs = append(apartmentIDs, apartment.ID)
	}

	filters["apartment_ids"] = apartmentIDs

	bookings, total, err := uc.bookingRepo.GetAll(filters, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get bookings: %w", err)
	}

	return bookings, total, nil
}

func (uc *conciergeUseCase) GetConciergeStats(userID int) (map[string]interface{}, error) {
	concierge, err := uc.conciergeRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("concierge not found: %w", err)
	}

	stats := make(map[string]interface{})

	if len(concierge.Apartments) == 0 {
		stats["active_bookings"] = 0
		stats["active_chats"] = 0
		stats["today_bookings"] = 0
		stats["completed_chats"] = 0
		return stats, nil
	}

	var apartmentIDs []int
	for _, apartment := range concierge.Apartments {
		apartmentIDs = append(apartmentIDs, apartment.ID)
	}

	activeBookingsFilters := map[string]interface{}{
		"apartment_ids": apartmentIDs,
		"status":        "confirmed",
		"active":        true,
	}
	activeBookings, _, err := uc.bookingRepo.GetAll(activeBookingsFilters, 1, 1000)
	if err == nil {
		stats["active_bookings"] = len(activeBookings)
	} else {
		stats["active_bookings"] = 0
	}

	activeChats, _, err := uc.chatRoomRepo.GetByConciergeID(concierge.ID, []domain.ChatRoomStatus{domain.ChatRoomStatusActive}, 1, 1000)
	if err == nil {
		stats["active_chats"] = len(activeChats)
	} else {
		stats["active_chats"] = 0
	}

	todayStart := time.Now().Truncate(24 * time.Hour)
	todayBookingsFilters := map[string]interface{}{
		"apartment_ids":   apartmentIDs,
		"start_date_from": todayStart.Format("2006-01-02"),
		"start_date_to":   todayStart.Add(24 * time.Hour).Format("2006-01-02"),
	}
	todayBookings, _, err := uc.bookingRepo.GetAll(todayBookingsFilters, 1, 1000)
	if err == nil {
		stats["today_bookings"] = len(todayBookings)
	} else {
		stats["today_bookings"] = 0
	}

	completedChats, _, err := uc.chatRoomRepo.GetByConciergeID(concierge.ID, []domain.ChatRoomStatus{domain.ChatRoomStatusClosed}, 1, 1000)
	if err == nil {
		stats["completed_chats"] = len(completedChats)
	} else {
		stats["completed_chats"] = 0
	}

	return stats, nil
}

func (uc *conciergeUseCase) UpdateConciergeSchedule(userID int, schedule *domain.ConciergeSchedule) error {
	concierge, err := uc.conciergeRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("concierge not found: %w", err)
	}

	concierge.Schedule = schedule

	err = uc.conciergeRepo.Update(concierge)
	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	return nil
}
