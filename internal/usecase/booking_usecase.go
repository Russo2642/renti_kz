package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/services"
	"github.com/russo2642/renti_kz/internal/utils"
	"github.com/russo2642/renti_kz/pkg/logger"
)

type bookingUseCase struct {
	bookingRepo         domain.BookingRepository
	apartmentRepo       domain.ApartmentRepository
	renterRepo          domain.RenterRepository
	propertyOwnerRepo   domain.PropertyOwnerRepository
	lockUseCase         domain.LockUseCase
	userUseCase         domain.UserUseCase
	notificationUseCase domain.NotificationUseCase
	schedulerService    SchedulerServiceInterface
	chatUseCase         domain.ChatUseCase
	chatRoomRepo        domain.ChatRoomRepository
	conciergeRepo       domain.ConciergeRepository
	contractUseCase     domain.ContractUseCase
	settingsUseCase     domain.PlatformSettingsUseCase
	paymentUseCase      domain.PaymentUseCase
	paymentRepo         domain.PaymentRepository
	paymentLogRepo      domain.PaymentLogRepository
	availabilityService domain.ApartmentAvailabilityService
}

type SchedulerServiceInterface interface {
	RemoveScheduledTasksForBooking(bookingID int) error
	RescheduleCompletionTask(bookingID int, newEndDate time.Time) error
	GetSchedulerStats(ctx context.Context) map[string]interface{}
	GetMetrics() *SchedulerMetrics
}

type SchedulerMetrics = services.SchedulerMetrics

func NewBookingUseCase(
	bookingRepo domain.BookingRepository,
	apartmentRepo domain.ApartmentRepository,
	renterRepo domain.RenterRepository,
	propertyOwnerRepo domain.PropertyOwnerRepository,
	lockUseCase domain.LockUseCase,
	userUseCase domain.UserUseCase,
	notificationUseCase domain.NotificationUseCase,
	schedulerService SchedulerServiceInterface,
	chatUseCase domain.ChatUseCase,
	chatRoomRepo domain.ChatRoomRepository,
	conciergeRepo domain.ConciergeRepository,
	contractUseCase domain.ContractUseCase,
	settingsUseCase domain.PlatformSettingsUseCase,
	paymentUseCase domain.PaymentUseCase,
	paymentRepo domain.PaymentRepository,
	paymentLogRepo domain.PaymentLogRepository,
	availabilityService domain.ApartmentAvailabilityService,
) domain.BookingUseCase {
	return &bookingUseCase{
		bookingRepo:         bookingRepo,
		apartmentRepo:       apartmentRepo,
		renterRepo:          renterRepo,
		propertyOwnerRepo:   propertyOwnerRepo,
		lockUseCase:         lockUseCase,
		userUseCase:         userUseCase,
		notificationUseCase: notificationUseCase,
		schedulerService:    schedulerService,
		chatUseCase:         chatUseCase,
		chatRoomRepo:        chatRoomRepo,
		conciergeRepo:       conciergeRepo,
		contractUseCase:     contractUseCase,
		settingsUseCase:     settingsUseCase,
		paymentUseCase:      paymentUseCase,
		paymentRepo:         paymentRepo,
		paymentLogRepo:      paymentLogRepo,
		availabilityService: availabilityService,
	}
}

func validateRenterVerification(renter *domain.Renter) error {
	hasDocuments := false
	if len(renter.DocumentURL) > 0 {
		for _, url := range renter.DocumentURL {
			if url != "" {
				hasDocuments = true
				break
			}
		}
	}

	hasPhotoWithDoc := renter.PhotoWithDocURL != ""

	if !hasDocuments || !hasPhotoWithDoc {
		var missing []string
		if !hasDocuments {
			missing = append(missing, "фото документа")
		}
		if !hasPhotoWithDoc {
			missing = append(missing, "фото с документом")
		}

		if len(missing) == 1 {
			return fmt.Errorf("для создания бронирования необходимо завершить верификацию личности")
		}
		return fmt.Errorf("для создания бронирования необходимо завершить верификацию личности")
	}

	switch renter.VerificationStatus {
	case domain.VerificationPending:
		return fmt.Errorf("ваша верификация находится на рассмотрении. Дождитесь подтверждения для создания бронирования")
	case domain.VerificationRejected:
		return fmt.Errorf("ваша верификация была отклонена. Обратитесь в поддержку или повторите процедуру верификации")
	case domain.VerificationApproved:
		return nil
	default:
		return fmt.Errorf("статус верификации неизвестен. Обратитесь в поддержку")
	}
}

func (u *bookingUseCase) CreateBooking(userID int, request *domain.CreateBookingRequest) (*domain.Booking, error) {
	renter, err := u.renterRepo.GetByUserID(userID)
	if err != nil || renter == nil {
		user, getUserErr := u.userUseCase.GetByID(userID)
		if getUserErr != nil {
			return nil, fmt.Errorf("ошибка при получении данных пользователя: %w", getUserErr)
		}

		if user.Role == domain.RoleOwner {
			return nil, fmt.Errorf("для создания бронирования владелец недвижимости должен пройти верификацию как арендатор")
		} else {
			return nil, fmt.Errorf("для создания бронирования необходимо завершить верификацию личности")
		}
	}

	user, getUserErr := u.userUseCase.GetByID(userID)
	if getUserErr != nil {
		return nil, fmt.Errorf("ошибка при получении данных пользователя: %w", getUserErr)
	}

	if user.Role == domain.RoleOwner {
		if renter.VerificationStatus != domain.VerificationApproved {
			return nil, fmt.Errorf("для создания бронирования владелец недвижимости должен пройти верификацию как арендатор")
		}
	} else {
		if err := validateRenterVerification(renter); err != nil {
			return nil, err
		}
	}

	apartment, err := u.apartmentRepo.GetByID(request.ApartmentID)
	if err != nil {
		return nil, fmt.Errorf("квартира не найдена: %w", err)
	}

	if apartment == nil {
		return nil, fmt.Errorf("квартира с ID %d не найдена", request.ApartmentID)
	}

	if apartment.Status != domain.AptStatusApproved {
		return nil, fmt.Errorf("квартира недоступна для бронирования")
	}

	minDuration, err := u.settingsUseCase.GetMinBookingDurationHours()
	if err != nil {
		minDuration = 1
	}
	maxDuration, err := u.settingsUseCase.GetMaxBookingDurationHours()
	if err != nil {
		maxDuration = 720
	}

	if err := utils.ValidateRange(request.Duration, minDuration, maxDuration, "продолжительность аренды"); err != nil {
		return nil, err
	}

	startDate, err := utils.ParseUserInput(request.StartDate)
	if err != nil {
		return nil, fmt.Errorf("неверный формат даты: %s (ожидается 2006-01-02T15:04:05)", request.StartDate)
	}

	if err := utils.ValidateFutureDate(startDate); err != nil {
		return nil, fmt.Errorf("время бронирования не может быть в прошлом")
	}

	maxAdvanceDays, err := u.settingsUseCase.GetMaxAdvanceBookingDays()
	if err != nil {
		maxAdvanceDays = 90
	}

	if err := utils.ValidateDateNotTooFar(startDate, maxAdvanceDays); err != nil {
		return nil, fmt.Errorf("бронирование нельзя создать более чем на %d дней вперед", maxAdvanceDays)
	}

	if request.Duration == 24 {
		if !apartment.RentalTypeDaily {
			return nil, fmt.Errorf("данная квартира не поддерживает посуточную аренду")
		}
		if apartment.DailyPrice <= 0 {
			return nil, fmt.Errorf("для данной квартиры не установлена цена за сутки")
		}
	} else {
		if !apartment.RentalTypeHourly {
			return nil, fmt.Errorf("данная квартира не поддерживает почасовую аренду")
		}
		if apartment.Price <= 0 {
			return nil, fmt.Errorf("для данной квартиры не установлена почасовая цена")
		}
	}

	if !utils.ValidateRentalTime(startDate, request.Duration, apartment.RentalTypeHourly, apartment.RentalTypeDaily) {
		timeInfo := utils.GetRentalTimeInfo(startDate)
		if timeInfo["is_daytime"].(bool) {
			return nil, fmt.Errorf("в дневное время (10:00-22:00) можно бронировать на %d, %d, %d или %d часа", utils.RentalDuration3Hours, utils.RentalDuration6Hours, utils.RentalDuration12Hours, utils.RentalDuration24Hours)
		} else {
			return nil, fmt.Errorf("в ночное время (22:00-10:00) доступна только посуточная аренда (24 часа)")
		}
	}

	endDate := startDate.Add(time.Duration(request.Duration) * time.Hour)

	if request.Duration < 24 && apartment.RentalTypeHourly {
		endTimeLocal := utils.ConvertOutputFromUTC(endDate)
		if endTimeLocal.Hour() < 10 || endTimeLocal.Hour() > 22 || (endTimeLocal.Hour() == 22 && endTimeLocal.Minute() > 0) {
			maxStartHour := 22 - request.Duration
			if maxStartHour < 10 {
				return nil, fmt.Errorf("продолжительность %d часов слишком велика для почасовой аренды. Максимальная продолжительность: 12 часов (10:00-22:00)", request.Duration)
			}
			return nil, fmt.Errorf("почасовая аренда должна заканчиваться до 22:00. Для продолжительности %d ч. максимальное время начала: %02d:00",
				request.Duration, maxStartHour)
		}
	}

	isAvailable, err := u.bookingRepo.CheckApartmentAvailability(
		request.ApartmentID,
		startDate,
		endDate,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки доступности: %w", err)
	}

	if !isAvailable {
		return nil, fmt.Errorf("квартира недоступна в указанный период")
	}

	var totalPrice int
	if request.Duration == 24 && apartment.RentalTypeDaily {
		totalPrice = apartment.DailyPrice
	} else {
		totalPrice = utils.CalculateHourlyPrice(apartment.Price, request.Duration)
	}

	serviceFee := u.calculateServiceFee(totalPrice, request.Duration)
	finalPrice := totalPrice + serviceFee

	booking := &domain.Booking{
		RenterID:           renter.ID,
		ApartmentID:        request.ApartmentID,
		StartDate:          startDate,
		EndDate:            endDate,
		Duration:           request.Duration,
		CleaningDuration:   u.getDefaultCleaningDuration(),
		Status:             domain.BookingStatusCreated,
		TotalPrice:         totalPrice,
		ServiceFee:         serviceFee,
		FinalPrice:         finalPrice,
		IsContractAccepted: false,
		DoorStatus:         domain.DoorStatusClosed,
		CanExtend:          true,
		ExtensionRequested: false,
	}

	err = u.bookingRepo.Create(booking)
	if err != nil {
		if strings.Contains(err.Error(), "Apartment is not available for the selected period") {
			return nil, fmt.Errorf("квартира недоступна в указанный период - найдено пересекающееся бронирование")
		}
		return nil, fmt.Errorf("ошибка создания бронирования: %w", err)
	}

	booking.Renter = renter
	booking.Apartment = apartment

	if u.contractUseCase != nil {
		_, err = u.contractUseCase.CreateRentalContract(booking.ID)
		if err != nil {
			logger.Warn("failed to create rental contract",
				slog.Int("booking_id", booking.ID),
				slog.String("error", err.Error()))
		} else {
			u.loadContractID(booking)
		}
	}

	if u.availabilityService != nil {
		if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			logger.Warn("failed to recalculate apartment availability after booking creation",
				slog.Int("apartment_id", booking.ApartmentID),
				slog.Int("booking_id", booking.ID),
				slog.String("error", err.Error()))
		}
	}

	go func() {
		if err := u.apartmentRepo.IncrementBookingCount(booking.ApartmentID); err != nil {
			logger.Warn("failed to increment booking count",
				slog.Int("apartment_id", booking.ApartmentID),
				slog.Int("booking_id", booking.ID),
				slog.String("error", err.Error()))
		}
	}()

	return booking, nil
}

func (u *bookingUseCase) GetBookingByID(id int) (*domain.Booking, error) {
	booking, err := u.bookingRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	utils.LoadBookingRelatedData(booking, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
	u.loadContractID(booking)

	return booking, nil
}

func (u *bookingUseCase) GetBookingByNumber(bookingNumber string) (*domain.Booking, error) {
	booking, err := u.bookingRepo.GetByBookingNumber(bookingNumber)
	if err != nil {
		return nil, err
	}

	utils.LoadBookingRelatedData(booking, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
	u.loadContractID(booking)

	return booking, nil
}

func (u *bookingUseCase) GetRenterBookings(userID int, status []domain.BookingStatus, dateFrom, dateTo *time.Time, page, pageSize int) ([]*domain.Booking, int, error) {
	renter, err := u.renterRepo.GetByUserID(userID)
	if err != nil || renter == nil {
		return []*domain.Booking{}, 0, nil
	}

	if renter.VerificationStatus == domain.VerificationRejected {
		return nil, 0, fmt.Errorf("ваша верификация была отклонена. Для создания новых бронирований обратитесь в поддержку или повторите процедуру верификации")
	}

	bookings, total, err := u.bookingRepo.GetByRenterID(renter.ID, status, dateFrom, dateTo, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	utils.LoadBookingsRelatedData(bookings, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
	u.loadContractIDs(bookings)

	return bookings, total, nil
}

func (u *bookingUseCase) GetOwnerBookings(userID int, status []domain.BookingStatus, dateFrom, dateTo *time.Time, page, pageSize int) ([]*domain.Booking, int, error) {
	owner, err := u.propertyOwnerRepo.GetByUserID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("владелец не найден: %w", err)
	}

	if owner == nil {
		return nil, 0, fmt.Errorf("пользователь не является владельцем недвижимости")
	}

	filteredStatuses := make([]domain.BookingStatus, 0, len(status))
	for _, s := range status {
		if s != domain.BookingStatusCreated {
			filteredStatuses = append(filteredStatuses, s)
		}
	}

	if len(status) == 0 {
		filteredStatuses = []domain.BookingStatus{
			domain.BookingStatusPending,
			domain.BookingStatusApproved,
			domain.BookingStatusRejected,
			domain.BookingStatusActive,
			domain.BookingStatusCompleted,
			domain.BookingStatusCanceled,
		}
	}

	bookings, total, err := u.bookingRepo.GetByOwnerID(owner.ID, filteredStatuses, dateFrom, dateTo, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	utils.LoadBookingsRelatedData(bookings, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
	u.loadContractIDs(bookings)

	return bookings, total, nil
}

func (u *bookingUseCase) ApproveBooking(bookingID, userID int) error {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("бронирование не найдено: %w", err)
	}

	apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
	if err != nil {
		return fmt.Errorf("квартира не найдена: %w", err)
	}

	if apartment == nil {
		return fmt.Errorf("квартира с ID %d не найдена", booking.ApartmentID)
	}

	propertyOwner, err := u.propertyOwnerRepo.GetByUserID(userID)
	if err != nil || propertyOwner == nil {
		return fmt.Errorf("пользователь не является владельцем недвижимости")
	}

	if apartment.OwnerID != propertyOwner.ID {
		return fmt.Errorf("нет прав для подтверждения этого бронирования")
	}

	if booking.Status != domain.BookingStatusPending {
		return fmt.Errorf("можно подтвердить только ожидающие бронирования")
	}

	booking.Status = domain.BookingStatusApproved

	now := utils.GetCurrentTimeUTC()

	if booking.StartDate.Before(now) || booking.StartDate.Equal(now) {
		booking.Status = domain.BookingStatusActive

		if u.availabilityService != nil {
			if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
				return fmt.Errorf("ошибка пересчета статуса квартиры: %w", err)
			}
		}
	}

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return err
	}

	if u.chatUseCase != nil {
		fmt.Printf("🔄 Начинаем создание комнаты чата для бронирования %d\n", booking.ID)
		err = u.createChatRoomForBooking(booking)
		if err != nil {
			fmt.Printf("❌ Ошибка создания комнаты чата для бронирования %d: %v\n", booking.ID, err)
		} else {
			logger.Info("chat room created successfully for booking", slog.Int("booking_id", booking.ID))
		}
	} else {
		logger.Warn("ChatUseCase not initialized")
	}

	if u.notificationUseCase != nil {
		apartmentTitle := fmt.Sprintf("%s, кв. %d", apartment.Street, apartment.ApartmentNumber)

		renter, renterErr := u.renterRepo.GetByID(booking.RenterID)
		if renterErr == nil && renter != nil {
			notifyErr := u.notificationUseCase.NotifyBookingApproved(renter.UserID, booking.ID, apartmentTitle)
			if notifyErr != nil {
				logger.Warn("failed to send booking approval notification",
					slog.String("error", notifyErr.Error()))
			}
		} else {
			logger.Warn("failed to get renter for approval notification",
				slog.String("error", renterErr.Error()))
		}
	}

	if u.availabilityService != nil {
		if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			logger.Warn("failed to recalculate apartment availability after booking approval",
				slog.Int("apartment_id", booking.ApartmentID),
				slog.Int("booking_id", booking.ID),
				slog.String("error", err.Error()))
		}
	}

	return nil
}

func (u *bookingUseCase) RejectBooking(bookingID, userID int, comment string) error {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("бронирование не найдено: %w", err)
	}

	owner, err := u.propertyOwnerRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("владелец не найден: %w", err)
	}

	if owner == nil {
		return fmt.Errorf("пользователь не является владельцем недвижимости")
	}

	apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
	if err != nil {
		return fmt.Errorf("квартира не найдена: %w", err)
	}

	if apartment == nil {
		return fmt.Errorf("квартира не найдена")
	}

	if apartment.OwnerID != owner.ID {
		return fmt.Errorf("нет прав для отклонения этого бронирования")
	}

	if booking.Status != domain.BookingStatusPending {
		return fmt.Errorf("можно отклонить только ожидающие бронирования")
	}

	booking.Status = domain.BookingStatusRejected
	booking.OwnerComment = &comment

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return err
	}

	if u.notificationUseCase != nil {
		apartmentTitle := fmt.Sprintf("%s, кв. %d", apartment.Street, apartment.ApartmentNumber)

		renter, renterErr := u.renterRepo.GetByID(booking.RenterID)
		if renterErr == nil && renter != nil {
			notifyErr := u.notificationUseCase.NotifyBookingRejected(renter.UserID, booking.ID, apartmentTitle, comment)
			if notifyErr != nil {
				logger.Warn("failed to send booking rejection notification",
					slog.String("error", notifyErr.Error()))
			}
		} else {
			logger.Warn("failed to get renter for rejection notification",
				slog.String("error", renterErr.Error()))
		}
	}

	if u.availabilityService != nil {
		if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			logger.Warn("failed to recalculate apartment availability after booking rejection",
				slog.Int("apartment_id", booking.ApartmentID),
				slog.Int("booking_id", booking.ID),
				slog.String("error", err.Error()))
		}
	}

	return nil
}

func (u *bookingUseCase) CancelBooking(bookingID, userID int, reason string) error {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("бронирование не найдено: %w", err)
	}

	renter, err := utils.GetRenterByUserID(u.renterRepo, userID)
	if err != nil {
		return err
	}

	if booking.RenterID != renter.ID {
		return fmt.Errorf("нет прав для отмены этого бронирования")
	}

	if booking.Status == domain.BookingStatusCompleted ||
		booking.Status == domain.BookingStatusCanceled ||
		booking.Status == domain.BookingStatusActive {
		return fmt.Errorf("нельзя отменить завершенное, уже отмененное или активное бронирование")
	}

	now := time.Now()
	hoursUntilStart := booking.StartDate.Sub(now).Hours()

	shouldRefund := true
	if (booking.Status == domain.BookingStatusApproved || booking.Status == domain.BookingStatusPending) && hoursUntilStart < 6 {
		shouldRefund = false
	}

	wasActive := booking.Status == domain.BookingStatusActive

	if shouldRefund && booking.PaymentID != nil {
		paymentRecord, err := u.paymentRepo.GetByID(*booking.PaymentID)
		if err == nil && paymentRecord != nil && u.paymentUseCase != nil {
			logger.Info("attempting to refund payment for cancelled booking",
				slog.Int("booking_id", bookingID),
				slog.String("payment_id", paymentRecord.PaymentID))

			refundResponse, refundErr := u.paymentUseCase.RefundPayment(paymentRecord.PaymentID, nil)
			if refundErr != nil {
				logger.Error("failed to refund payment for cancelled booking",
					slog.Int("booking_id", bookingID),
					slog.String("payment_id", paymentRecord.PaymentID),
					slog.String("error", refundErr.Error()))
			} else if refundResponse.Success {
				logger.Info("payment refunded successfully for cancelled booking",
					slog.Int("booking_id", bookingID),
					slog.String("payment_id", paymentRecord.PaymentID))
			}
		}
	} else if !shouldRefund && booking.PaymentID != nil {
		logger.Info("booking cancelled without refund due to late cancellation policy",
			slog.Int("booking_id", bookingID),
			slog.String("status", string(booking.Status)),
			slog.Float64("hours_until_start", hoursUntilStart))
	}

	booking.Status = domain.BookingStatusCanceled
	booking.CancellationReason = &reason

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return err
	}

	if wasActive {
		if u.availabilityService != nil {
			if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
				return fmt.Errorf("ошибка пересчета доступности квартиры: %w", err)
			}
		}
	}

	if u.schedulerService != nil {
		err = u.schedulerService.RemoveScheduledTasksForBooking(bookingID)
		if err != nil {
			logger.Warn("failed to remove scheduler tasks for cancelled booking",
				slog.Int("booking_id", bookingID),
				slog.String("error", err.Error()))
		}
	}

	if u.notificationUseCase != nil {
		apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
		if err == nil && apartment != nil {
			apartmentTitle := fmt.Sprintf("%s, кв. %d", apartment.Street, apartment.ApartmentNumber)

			err = u.notificationUseCase.NotifyBookingCanceled(renter.UserID, bookingID, apartmentTitle, reason)
			if err != nil {
				logger.Warn("failed to send cancellation notification to renter",
					slog.String("error", err.Error()))
			}
		}
	}

	if u.availabilityService != nil {
		if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			logger.Warn("failed to recalculate apartment availability after booking cancellation",
				slog.Int("apartment_id", booking.ApartmentID),
				slog.Int("booking_id", booking.ID),
				slog.String("error", err.Error()))
		}
	}

	return nil
}

func (u *bookingUseCase) CompleteBooking(bookingID int) error {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("бронирование не найдено: %w", err)
	}

	if booking.Status != domain.BookingStatusActive {
		return fmt.Errorf("можно завершить только активные бронирования")
	}

	booking.Status = domain.BookingStatusCompleted
	booking.DoorStatus = domain.DoorStatusClosed

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return err
	}

	if u.availabilityService != nil {
		if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			return fmt.Errorf("ошибка пересчета доступности квартиры: %w", err)
		}
	}

	if u.notificationUseCase != nil {
		apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
		if err == nil && apartment != nil {
			apartmentTitle := fmt.Sprintf("%s, кв. %d", apartment.Street, apartment.ApartmentNumber)

			renter, err := u.renterRepo.GetByID(booking.RenterID)
			if err == nil && renter != nil {
				err = u.notificationUseCase.NotifyBookingCompleted(renter.UserID, bookingID, apartmentTitle)
				if err != nil {
					logger.Warn("failed to send booking completion notification",
						slog.String("error", err.Error()))
				}
			}
		}
	}

	return nil
}

func (u *bookingUseCase) FinishSession(bookingID, userID int) error {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("бронирование не найдено: %w", err)
	}

	renter, err := utils.GetRenterByUserID(u.renterRepo, userID)
	if err != nil {
		return fmt.Errorf("пользователь не является арендатором: %w", err)
	}

	if booking.RenterID != renter.ID {
		return fmt.Errorf("нет прав для завершения этого бронирования")
	}

	if booking.Status != domain.BookingStatusActive {
		return fmt.Errorf("завершить можно только активные бронирования")
	}

	apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
	if err != nil {
		return fmt.Errorf("квартира не найдена: %w", err)
	}

	if apartment == nil {
		return fmt.Errorf("квартира с ID %d не найдена", booking.ApartmentID)
	}

	propertyOwner, err := u.propertyOwnerRepo.GetByID(apartment.OwnerID)
	if err != nil {
		return fmt.Errorf("владелец не найден: %w", err)
	}

	booking.Status = domain.BookingStatusCompleted
	booking.DoorStatus = domain.DoorStatusClosed

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса бронирования: %w", err)
	}

	if u.availabilityService != nil {
		if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			return fmt.Errorf("ошибка пересчета доступности квартиры: %w", err)
		}
	}

	err = u.lockUseCase.DeactivatePasswordForBooking(bookingID)
	if err != nil {
		return fmt.Errorf("ошибка деактивации паролей замка: %w", err)
	}

	if u.schedulerService != nil {
		err = u.schedulerService.RemoveScheduledTasksForBooking(bookingID)
		if err != nil {
			logger.Warn("failed to remove scheduler tasks for booking",
				slog.Int("booking_id", bookingID),
				slog.String("error", err.Error()))
		}
	}

	if u.notificationUseCase != nil && propertyOwner != nil {
		apartmentTitle := fmt.Sprintf("%s, кв. %d", apartment.Street, apartment.ApartmentNumber)

		renterUser, err := u.userUseCase.GetByID(renter.UserID)
		renterName := "арендатор"
		if err == nil && renterUser != nil {
			renterName = fmt.Sprintf("%s %s", renterUser.FirstName, renterUser.LastName)
		}

		err = u.notificationUseCase.NotifySessionFinished(
			propertyOwner.UserID,
			bookingID,
			apartmentTitle,
			renterName,
		)
		if err != nil {
			logger.Warn("failed to send completion notification to owner",
				slog.String("error", err.Error()))
		}

		err = u.notificationUseCase.NotifyBookingCompleted(renter.UserID, bookingID, apartmentTitle)
		if err != nil {
			logger.Warn("failed to send completion notification to renter",
				slog.String("error", err.Error()))
		}
	}

	if u.availabilityService != nil {
		if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			logger.Warn("failed to recalculate apartment availability after session finish",
				slog.Int("apartment_id", booking.ApartmentID),
				slog.Int("booking_id", booking.ID),
				slog.String("error", err.Error()))
		}
	}

	return nil
}

func (u *bookingUseCase) RequestExtension(bookingID, userID int, request *domain.ExtendBookingRequest) error {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("бронирование не найдено: %w", err)
	}

	renter, err := utils.GetRenterByUserID(u.renterRepo, userID)
	if err != nil {
		return err
	}

	if booking.RenterID != renter.ID {
		return fmt.Errorf("нет прав для продления этого бронирования")
	}

	if !booking.CanExtend {
		return fmt.Errorf("продление данного бронирования недоступно")
	}

	if booking.Status != domain.BookingStatusActive {
		return fmt.Errorf("можно продлить только активные бронирования")
	}

	newEndDate := booking.EndDate.Add(time.Duration(request.Duration) * time.Hour)

	isAvailable, err := u.bookingRepo.CheckApartmentAvailability(
		booking.ApartmentID,
		booking.EndDate,
		newEndDate,
		&booking.ID,
	)
	if err != nil {
		return fmt.Errorf("ошибка проверки доступности: %w", err)
	}

	if !isAvailable {
		return fmt.Errorf("квартира недоступна для продления в указанный период")
	}

	apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
	if err != nil {
		return fmt.Errorf("квартира не найдена: %w", err)
	}

	if apartment == nil {
		return fmt.Errorf("квартира с ID %d не найдена", booking.ApartmentID)
	}

	extensionPrice := utils.CalculateHourlyPrice(apartment.Price, request.Duration)

	extension := &domain.BookingExtension{
		BookingID: bookingID,
		Duration:  request.Duration,
		Price:     extensionPrice,
		Status:    domain.BookingStatusAwaitingPayment,
	}

	err = u.bookingRepo.CreateExtension(extension)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса на продление: %w", err)
	}

	return nil
}

func (u *bookingUseCase) ProcessExtensionPayment(extensionID int, paymentID string) (*domain.BookingExtension, error) {
	extension, err := u.bookingRepo.GetExtensionByID(extensionID)
	if err != nil {
		return nil, fmt.Errorf("продление не найдено: %w", err)
	}

	if extension.Status != domain.BookingStatusAwaitingPayment {
		return nil, fmt.Errorf("можно оплатить только продления со статусом 'awaiting_payment'")
	}

	existingPayment, err := u.paymentRepo.GetByPaymentID(paymentID)
	if err == nil && existingPayment != nil {
		return nil, fmt.Errorf("платеж %s уже использован", paymentID)
	}

	paymentStatus, err := u.paymentUseCase.CheckPaymentStatus(paymentID)
	if err != nil {
		logger.Error("failed to check extension payment status",
			slog.String("payment_id", paymentID),
			slog.Int("extension_id", extensionID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("ошибка проверки платежа: %w", err)
	}

	if paymentStatus.Status != "success" {
		logger.Warn("extension payment not successful",
			slog.String("payment_id", paymentID),
			slog.Int("extension_id", extensionID),
			slog.String("status", paymentStatus.Status))
		return nil, fmt.Errorf("платеж не завершен: %s", paymentStatus.Status)
	}

	payment := &domain.Payment{
		BookingID:      int64(extension.BookingID),
		PaymentID:      paymentID,
		Amount:         extension.Price,
		Currency:       "KZT",
		Status:         domain.PaymentStatusSuccess,
		PaymentMethod:  &paymentStatus.PaymentMethod,
		ProviderStatus: &paymentStatus.Status,
		ProcessedAt:    &[]time.Time{time.Now()}[0],
	}

	if err := u.paymentRepo.Create(payment); err != nil {
		logger.Error("failed to save extension payment record",
			slog.String("payment_id", paymentID),
			slog.Int("extension_id", extensionID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("ошибка сохранения платежа в БД: %w", err)
	}

	extension.Status = domain.BookingStatusPending
	extension.PaymentID = &payment.ID

	err = u.bookingRepo.UpdateExtension(extension)
	if err != nil {
		logger.Error("failed to update extension after payment",
			slog.String("payment_id", paymentID),
			slog.Int("extension_id", extensionID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("ошибка обновления продления: %w", err)
	}

	booking, err := u.bookingRepo.GetByID(extension.BookingID)
	if err != nil {
		return nil, fmt.Errorf("бронирование не найдено: %w", err)
	}

	newEndDate := booking.EndDate.Add(time.Duration(extension.Duration) * time.Hour)
	booking.ExtensionRequested = true
	booking.ExtensionEndDate = &newEndDate
	booking.ExtensionDuration = extension.Duration
	booking.ExtensionPrice = extension.Price

	err = u.bookingRepo.Update(booking)
	if err != nil {
		logger.Error("failed to update booking after extension payment",
			slog.String("payment_id", paymentID),
			slog.Int("extension_id", extensionID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("ошибка обновления бронирования: %w", err)
	}

	if u.notificationUseCase != nil {
		apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
		if err == nil && apartment != nil {
			propertyOwner, ownerErr := u.propertyOwnerRepo.GetByID(apartment.OwnerID)
			if ownerErr == nil && propertyOwner != nil {
				apartmentTitle := fmt.Sprintf("%s, кв. %d", apartment.Street, apartment.ApartmentNumber)

				renter, renterErr := u.renterRepo.GetByID(booking.RenterID)
				renterName := "арендатор"
				if renterErr == nil && renter != nil {
					renterUser, userErr := u.userUseCase.GetByID(renter.UserID)
					if userErr == nil && renterUser != nil {
						renterName = fmt.Sprintf("%s %s", renterUser.FirstName, renterUser.LastName)
					}
				}

				notifyErr := u.notificationUseCase.NotifyExtensionRequested(
					propertyOwner.UserID,
					booking.ID,
					apartmentTitle,
					renterName,
					extension.Duration,
				)
				if notifyErr != nil {
					logger.Warn("failed to send extension request notification",
						slog.String("error", notifyErr.Error()))
				}
			}
		}
	}

	logger.Info("extension payment processed successfully",
		slog.String("payment_id", paymentID),
		slog.Int("extension_id", extensionID),
		slog.Int("booking_id", extension.BookingID))

	return extension, nil
}

func (u *bookingUseCase) ProcessExtensionPaymentWithOrder(extensionID int, orderID string) (*domain.BookingExtension, error) {
	extension, err := u.bookingRepo.GetExtensionByID(extensionID)
	if err != nil {
		return nil, fmt.Errorf("продление не найдено: %w", err)
	}

	paymentStatus, err := u.paymentUseCase.CheckPaymentStatusByOrderID(orderID, int64(extension.BookingID))
	if err != nil {
		logger.Error("failed to get payment status by order ID for extension",
			slog.String("order_id", orderID),
			slog.Int("extension_id", extensionID),
			slog.Int("booking_id", extension.BookingID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("ошибка получения статуса платежа по order_id: %w", err)
	}

	if !paymentStatus.Exists {
		return nil, fmt.Errorf("платеж с order_id %s не найден: %s", orderID, paymentStatus.ErrorMessage)
	}

	logger.Info("converting order_id to payment_id for extension",
		slog.String("order_id", orderID),
		slog.String("payment_id", paymentStatus.PaymentID),
		slog.Int("extension_id", extensionID),
		slog.Int("booking_id", extension.BookingID))

	return u.ProcessExtensionPayment(extensionID, paymentStatus.PaymentID)
}

func (u *bookingUseCase) ApproveExtension(extensionID, userID int) error {
	extension, err := u.bookingRepo.GetExtensionByID(extensionID)
	if err != nil {
		return fmt.Errorf("продление не найдено: %w", err)
	}

	if extension.Status != domain.BookingStatusPending {
		return fmt.Errorf("можно подтвердить только оплаченные продления (статус: pending)")
	}

	booking, err := u.bookingRepo.GetByID(extension.BookingID)
	if err != nil {
		return fmt.Errorf("бронирование не найдено: %w", err)
	}

	owner, err := u.propertyOwnerRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("владелец не найден: %w", err)
	}

	if owner == nil {
		return fmt.Errorf("пользователь не является владельцем недвижимости")
	}

	apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
	if err != nil {
		return fmt.Errorf("квартира не найдена: %w", err)
	}

	if apartment == nil {
		return fmt.Errorf("квартира не найдена")
	}

	if apartment.OwnerID != owner.ID {
		return fmt.Errorf("нет прав для подтверждения продления")
	}

	now := time.Now()
	extension.Status = domain.BookingStatusApproved
	extension.ApprovedAt = &now

	err = u.bookingRepo.UpdateExtension(extension)
	if err != nil {
		return err
	}

	if booking.ExtensionEndDate == nil {
		return fmt.Errorf("данные о продлении повреждены: отсутствует дата окончания")
	}

	newEndDate := *booking.ExtensionEndDate
	extensionDuration := booking.ExtensionDuration

	booking.EndDate = newEndDate
	booking.Duration += booking.ExtensionDuration
	booking.TotalPrice += booking.ExtensionPrice
	booking.FinalPrice = booking.TotalPrice + booking.ServiceFee
	booking.ExtensionRequested = false

	booking.ExtensionEndDate = nil
	booking.ExtensionDuration = 0
	booking.ExtensionPrice = 0

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return err
	}

	if u.lockUseCase != nil {
		err = u.lockUseCase.ExtendPasswordForBooking(booking.ID, newEndDate)
		if err != nil {
			logger.Warn("failed to extend lock password lifetime",
				slog.Int("booking_id", booking.ID),
				slog.String("error", err.Error()))
		} else {
			logger.Info("lock password lifetime extended",
				slog.Int("booking_id", booking.ID),
				slog.String("new_end_date", newEndDate.Format("2006-01-02 15:04:05")))
		}
	}

	if u.schedulerService != nil {
		err = u.schedulerService.RescheduleCompletionTask(booking.ID, newEndDate)
		if err != nil {
			logger.Warn("failed to reschedule tasks for extended booking",
				slog.Int("booking_id", booking.ID),
				slog.String("error", err.Error()))
		}
	}

	if u.notificationUseCase != nil {
		apartmentTitle := fmt.Sprintf("%s, кв. %d", apartment.Street, apartment.ApartmentNumber)

		renter, renterErr := u.renterRepo.GetByID(booking.RenterID)
		if renterErr == nil && renter != nil {
			notifyErr := u.notificationUseCase.NotifyExtensionApproved(
				renter.UserID,
				booking.ID,
				apartmentTitle,
				extensionDuration,
			)
			if notifyErr != nil {
				logger.Warn("failed to send extension approval notification",
					slog.String("error", notifyErr.Error()))
			}
		}
	}

	return nil
}

func (u *bookingUseCase) RejectExtension(extensionID, userID int) error {
	extension, err := u.bookingRepo.GetExtensionByID(extensionID)
	if err != nil {
		return fmt.Errorf("продление не найдено: %w", err)
	}

	if extension.Status != domain.BookingStatusPending {
		return fmt.Errorf("можно отклонить только ожидающие продления")
	}

	booking, err := u.bookingRepo.GetByID(extension.BookingID)
	if err != nil {
		return fmt.Errorf("бронирование не найдено: %w", err)
	}

	owner, err := u.propertyOwnerRepo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("владелец не найден: %w", err)
	}

	if owner == nil {
		return fmt.Errorf("пользователь не является владельцем недвижимости")
	}

	apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
	if err != nil {
		return fmt.Errorf("квартира не найдена: %w", err)
	}

	if apartment == nil {
		return fmt.Errorf("квартира не найдена")
	}

	if apartment.OwnerID != owner.ID {
		return fmt.Errorf("нет прав для отклонения продления")
	}

	extensionDuration := booking.ExtensionDuration

	extension.Status = domain.BookingStatusRejected
	err = u.bookingRepo.UpdateExtension(extension)
	if err != nil {
		return err
	}

	if extension.PaymentID != nil {
		paymentRecord, paymentErr := u.paymentRepo.GetByID(*extension.PaymentID)
		if paymentErr != nil {
			logger.Error("failed to get payment record for refund",
				slog.Int("extension_id", extensionID),
				slog.Int64("payment_id", *extension.PaymentID),
				slog.String("error", paymentErr.Error()))
		} else {
			logger.Info("attempting to refund payment for rejected extension",
				slog.Int("extension_id", extensionID),
				slog.String("payment_id", paymentRecord.PaymentID))

			refundResponse, refundErr := u.paymentUseCase.RefundPayment(paymentRecord.PaymentID, nil)
			if refundErr != nil {
				logger.Error("failed to refund payment for rejected extension",
					slog.Int("extension_id", extensionID),
					slog.String("payment_id", paymentRecord.PaymentID),
					slog.String("error", refundErr.Error()))
			} else if refundResponse.Success {
				logger.Info("payment refunded successfully for rejected extension",
					slog.Int("extension_id", extensionID),
					slog.String("payment_id", paymentRecord.PaymentID))
			}
		}
	}

	booking.ExtensionRequested = false
	booking.ExtensionEndDate = nil
	booking.ExtensionDuration = 0
	booking.ExtensionPrice = 0

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return err
	}

	if u.notificationUseCase != nil {
		apartmentTitle := fmt.Sprintf("%s, кв. %d", apartment.Street, apartment.ApartmentNumber)

		renter, renterErr := u.renterRepo.GetByID(booking.RenterID)
		if renterErr == nil && renter != nil {
			notifyErr := u.notificationUseCase.NotifyExtensionRejected(
				renter.UserID,
				booking.ID,
				apartmentTitle,
				extensionDuration,
			)
			if notifyErr != nil {
				logger.Warn("failed to send extension rejection notification",
					slog.String("error", notifyErr.Error()))
			}
		}
	}

	return nil
}

func (u *bookingUseCase) CanUserAccessBooking(bookingID, userID int) (bool, error) {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return false, err
	}

	user, err := u.userUseCase.GetByID(userID)
	if err != nil {
		return false, err
	}

	if user.Role == domain.RoleAdmin || user.Role == domain.RoleModerator {
		return true, nil
	}

	renter, err := u.renterRepo.GetByUserID(userID)
	if err == nil && renter != nil && renter.ID == booking.RenterID {
		return true, nil
	}

	apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
	if err != nil {
		return false, err
	}

	if apartment == nil {
		return false, fmt.Errorf("квартира с ID %d не найдена", booking.ApartmentID)
	}

	owner, err := u.propertyOwnerRepo.GetByUserID(userID)
	if err == nil && owner != nil && owner.ID == apartment.OwnerID {
		return true, nil
	}

	return false, nil
}

func (u *bookingUseCase) CanUserManageDoor(bookingID, userID int) (bool, error) {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return false, fmt.Errorf("бронирование не найдено: %w", err)
	}

	return u.lockUseCase.CanUserManageLockViaBooking(booking, userID)
}

func (u *bookingUseCase) IsBookingActive(bookingID int) (bool, error) {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return false, err
	}

	if booking.Status != domain.BookingStatusActive {
		return false, nil
	}

	now := utils.GetCurrentTimeUTC()

	return utils.IsTimeInRange(now, booking.StartDate, booking.EndDate, 0), nil
}

func (u *bookingUseCase) CheckApartmentAvailability(apartmentID int, startDate, endDate time.Time) (bool, error) {
	return u.bookingRepo.CheckApartmentAvailability(apartmentID, startDate, endDate, nil)
}

func (u *bookingUseCase) calculateServiceFee(totalPrice, duration int) int {
	if duration == 24 {
		return 3000
	}

	serviceFeePercentage, err := u.settingsUseCase.GetServiceFeePercentage()
	if err != nil {
		serviceFeePercentage = 15
	}
	return totalPrice * serviceFeePercentage / 100
}

func (u *bookingUseCase) getDefaultCleaningDuration() int {
	cleaningDuration, err := u.settingsUseCase.GetDefaultCleaningDurationMinutes()
	if err != nil {
		return 60
	}
	return cleaningDuration
}

func (u *bookingUseCase) GetBookingExtensions(bookingID int) ([]*domain.BookingExtension, error) {
	return u.bookingRepo.GetExtensionsByBookingID(bookingID)
}

func (u *bookingUseCase) GetAvailableExtensions(bookingID, userID int) (*domain.AvailableExtensionsResponse, error) {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("бронирование не найдено: %w", err)
	}

	canAccess, err := u.CanUserAccessBooking(bookingID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, fmt.Errorf("нет прав для доступа к данному бронированию")
	}

	if !booking.CanExtend {
		return nil, fmt.Errorf("продление данного бронирования недоступно")
	}

	if booking.Status != domain.BookingStatusActive {
		return nil, fmt.Errorf("можно продлить только активные бронирования")
	}

	apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
	if err != nil {
		return nil, fmt.Errorf("квартира не найдена: %w", err)
	}

	cleaningDurationMinutes := u.getDefaultCleaningDuration()

	nextBooking, err := u.bookingRepo.GetNextBookingAfterDate(
		booking.ApartmentID,
		booking.EndDate,
		&booking.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске следующих бронирований: %w", err)
	}

	currentTime := utils.GetCurrentTimeUTC()
	timeInfo := utils.GetRentalTimeInfo(currentTime)

	var allDurations []int
	if apartment.RentalTypeDaily {
		allDurations = append(allDurations, 24)
	}
	if apartment.RentalTypeHourly {
		allDurations = append(allDurations, utils.RentalDuration3Hours, utils.RentalDuration6Hours, utils.RentalDuration12Hours)
	}

	var availableExtensions []int
	var maxPossibleExtension int
	limitations := make(map[string]string)

	for _, duration := range allDurations {
		proposedEndTime := booking.EndDate.Add(time.Duration(duration) * time.Hour)

		if duration < 24 && apartment.RentalTypeHourly {
			proposedEndTimeLocal := utils.ConvertOutputFromUTC(proposedEndTime)

			if proposedEndTimeLocal.Hour() > 22 || (proposedEndTimeLocal.Hour() == 22 && proposedEndTimeLocal.Minute() > 0) {
				limitations[fmt.Sprintf("%d_hours", duration)] = "Превышает дневные часы аренды (до 22:00)"
				continue
			}

			if proposedEndTimeLocal.Hour() < 10 {
				limitations[fmt.Sprintf("%d_hours", duration)] = "Окончание продления попадает на ночное время (22:00-10:00)"
				continue
			}
		}

		if nextBooking != nil {
			cleaningEndTime := proposedEndTime.Add(time.Duration(cleaningDurationMinutes) * time.Minute)
			if cleaningEndTime.After(nextBooking.StartDate) {
				nextBookingLocalTime := utils.ConvertOutputFromUTC(nextBooking.StartDate)
				limitations[fmt.Sprintf("%d_hours", duration)] = fmt.Sprintf("Конфликт со следующим бронированием в %s", nextBookingLocalTime.Format("15:04"))
				continue
			}
		}

		availableExtensions = append(availableExtensions, duration)
		if duration > maxPossibleExtension {
			maxPossibleExtension = duration
		}
	}

	response := &domain.AvailableExtensionsResponse{
		BookingID:               booking.ID,
		CurrentEndDate:          utils.FormatForUser(booking.EndDate),
		AvailableExtensions:     availableExtensions,
		MaxPossibleExtension:    maxPossibleExtension,
		CleaningDurationMinutes: cleaningDurationMinutes,
		Limitations:             limitations,
		TimeInfo:                timeInfo,
	}

	if nextBooking != nil {
		nextBookingTime := utils.FormatForUser(nextBooking.StartDate)
		response.NextBookingStartsAt = &nextBookingTime
	}

	return response, nil
}

func (u *bookingUseCase) DebugBookingAccess(bookingID, userID int) (map[string]interface{}, error) {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("бронирование не найдено: %w", err)
	}

	user, err := u.userUseCase.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}

	apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
	if err != nil {
		return nil, fmt.Errorf("квартира не найдена: %w", err)
	}

	if apartment == nil {
		return nil, fmt.Errorf("квартира с ID %d не найдена", booking.ApartmentID)
	}

	renter, err := u.renterRepo.GetByID(booking.RenterID)
	if err != nil {
		return nil, fmt.Errorf("арендатор не найден: %w", err)
	}

	userRenter, renterErr := u.renterRepo.GetByUserID(userID)
	var isUserRenter bool
	var userRenterID *int
	if renterErr == nil && userRenter != nil {
		isUserRenter = true
		userRenterID = &userRenter.ID
	}

	userOwner, ownerErr := u.propertyOwnerRepo.GetByUserID(userID)
	var isUserOwner bool
	var userOwnerID *int
	if ownerErr == nil && userOwner != nil {
		isUserOwner = true
		userOwnerID = &userOwner.ID
	}

	canAccess, accessErr := u.CanUserAccessBooking(bookingID, userID)
	canManageDoor, doorErr := u.CanUserManageDoor(bookingID, userID)

	now := time.Now()
	isActive := booking.Status == domain.BookingStatusActive
	isInTimeFrame := now.After(booking.StartDate) && now.Before(booking.EndDate)

	debugInfo := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   userID,
			"role": user.Role,
		},
		"booking": map[string]interface{}{
			"id":         booking.ID,
			"renter_id":  booking.RenterID,
			"status":     booking.Status,
			"start_date": booking.StartDate.Format("2006-01-02T15:04:05"),
			"end_date":   booking.EndDate.Format("2006-01-02T15:04:05"),
		},
		"apartment": map[string]interface{}{
			"id":       apartment.ID,
			"owner_id": apartment.OwnerID,
		},
		"renter": map[string]interface{}{
			"id":      renter.ID,
			"user_id": renter.UserID,
		},
		"user_roles": map[string]interface{}{
			"is_renter":      isUserRenter,
			"user_renter_id": userRenterID,
			"is_owner":       isUserOwner,
			"user_owner_id":  userOwnerID,
		},
		"permissions": map[string]interface{}{
			"can_access":      canAccess,
			"can_manage_door": canManageDoor,
			"access_error":    getErrorString(accessErr),
			"door_error":      getErrorString(doorErr),
		},
		"time_checks": map[string]interface{}{
			"now":             now.Format("2006-01-02T15:04:05"),
			"is_active":       isActive,
			"is_in_timeframe": isInTimeFrame,
		},
	}

	return debugInfo, nil
}

func getErrorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func (u *bookingUseCase) createChatRoomForBooking(booking *domain.Booking) error {
	logger.Debug("checking existing chat room for booking", slog.Int("booking_id", booking.ID))
	existingRoom, err := u.chatUseCase.GetChatRoomByBookingID(booking.ID)
	if err == nil && existingRoom != nil {
		fmt.Printf("ℹ️ Комната чата уже существует для бронирования %d (ID: %d)\n", booking.ID, existingRoom.ID)
		return nil
	}

	fmt.Printf("🔍 Ищем консьержа для квартиры %d\n", booking.ApartmentID)
	concierge, err := u.getConciergeForApartment(booking.ApartmentID)
	if err != nil {
		fmt.Printf("❌ Не найден консьерж для квартиры %d: %v\n", booking.ApartmentID, err)
		return fmt.Errorf("failed to find concierge for apartment %d: %w", booking.ApartmentID, err)
	}

	logger.Debug("concierge found for apartment",
		slog.Int("concierge_id", concierge.ID),
		slog.Int("apartment_id", booking.ApartmentID))

	room := &domain.ChatRoom{
		BookingID:   booking.ID,
		ConciergeID: concierge.ID,
		RenterID:    booking.RenterID,
		ApartmentID: booking.ApartmentID,
		Status:      domain.ChatRoomStatusPending,
	}

	fmt.Printf("💾 Создаем комнату чата в БД\n")
	err = u.createChatRoomDirect(room)
	if err != nil {
		fmt.Printf("❌ Ошибка создания комнаты в БД: %v\n", err)
		return fmt.Errorf("failed to create chat room: %w", err)
	}

	fmt.Printf("🎉 Комната чата успешно создана\n")
	return nil
}

func (u *bookingUseCase) getConciergeForApartment(apartmentID int) (*domain.Concierge, error) {
	concierge, err := u.getConciergeByApartmentID(apartmentID)
	if err == nil && concierge != nil {
		return concierge, nil
	}

	return u.getOrCreateDefaultConcierge()
}

func (u *bookingUseCase) getConciergeByApartmentID(apartmentID int) (*domain.Concierge, error) {
	concierges, err := u.conciergeRepo.GetByApartmentIDActive(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("concierge not found for apartment %d: %w", apartmentID, err)
	}
	if len(concierges) == 0 {
		return nil, fmt.Errorf("no active concierge found for apartment %d", apartmentID)
	}
	return concierges[0], nil
}

func (u *bookingUseCase) getOrCreateDefaultConcierge() (*domain.Concierge, error) {
	filters := map[string]interface{}{
		"is_active": true,
	}
	concierges, _, err := u.conciergeRepo.GetAll(filters, 1, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get concierges: %w", err)
	}

	if len(concierges) == 0 {
		return nil, fmt.Errorf("no active concierge available - please create a concierge first")
	}

	return concierges[0], nil
}

func (u *bookingUseCase) createChatRoomDirect(room *domain.ChatRoom) error {
	err := u.chatRoomRepo.Create(room)
	if err != nil {
		return fmt.Errorf("failed to create chat room: %w", err)
	}
	return nil
}

func (u *bookingUseCase) ConfirmBooking(bookingID, userID int, request *domain.ConfirmBookingRequest) (*domain.Booking, error) {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("бронирование не найдено: %w", err)
	}

	renter, err := utils.GetRenterByUserID(u.renterRepo, userID)
	if err != nil {
		return nil, err
	}

	if booking.RenterID != renter.ID {
		return nil, fmt.Errorf("нет прав для подтверждения этого бронирования")
	}

	if booking.Status != domain.BookingStatusCreated {
		return nil, fmt.Errorf("можно подтвердить только бронирования со статусом 'created'")
	}

	if !request.IsContractAccepted {
		return nil, fmt.Errorf("необходимо принять условия договора аренды")
	}

	booking.Status = domain.BookingStatusAwaitingPayment
	booking.IsContractAccepted = request.IsContractAccepted

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления бронирования: %w", err)
	}

	if u.contractUseCase != nil {
		contract, contractErr := u.contractUseCase.CreateRentalContract(booking.ID)
		if contractErr != nil {
			logger.Warn("failed to create contract for booking",
				slog.Int("booking_id", booking.ID),
				slog.String("error", contractErr.Error()))
		} else {
			if contract.Status == domain.ContractStatusDraft {
				updateErr := u.contractUseCase.UpdateContractStatus(contract.ID, domain.ContractStatusConfirmed)
				if updateErr != nil {
					logger.Warn("failed to update contract status",
						slog.Int("contract_id", contract.ID),
						slog.String("error", updateErr.Error()))
				} else {
					logger.Info("contract status updated to confirmed", slog.Int("contract_id", contract.ID))
				}
			}
			logger.Info("contract processed successfully for booking", slog.Int("booking_id", booking.ID))
		}
	}

	utils.LoadBookingRelatedData(booking, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
	u.loadContractID(booking)

	if u.notificationUseCase != nil && booking.Apartment != nil {
		propertyOwner, ownerErr := u.propertyOwnerRepo.GetByID(booking.Apartment.OwnerID)
		if ownerErr == nil && propertyOwner != nil {
			apartmentTitle := fmt.Sprintf("%s, кв. %d", booking.Apartment.Street, booking.Apartment.ApartmentNumber)

			renterUser, userErr := u.userUseCase.GetByID(renter.UserID)
			renterName := "арендатор"
			if userErr == nil && renterUser != nil {
				renterName = fmt.Sprintf("%s %s", renterUser.FirstName, renterUser.LastName)
			}

			if booking.Status == domain.BookingStatusPending {
				notifyErr := u.notificationUseCase.NotifyNewBookingRequest(
					propertyOwner.UserID,
					booking.ID,
					apartmentTitle,
					renterName,
				)
				if notifyErr != nil {
					logger.Warn("failed to send new booking request notification",
						slog.String("error", notifyErr.Error()))
				}
			} else if booking.Status == domain.BookingStatusApproved {
				notifyErr := u.notificationUseCase.NotifyBookingApproved(renter.UserID, booking.ID, apartmentTitle)
				if notifyErr != nil {
					logger.Warn("failed to send auto-confirmation notification",
						slog.String("error", notifyErr.Error()))
				}

				notifyErr = u.notificationUseCase.NotifyBookingStarted(propertyOwner.UserID, booking.ID, apartmentTitle, renterName)
				if notifyErr != nil {
					logger.Warn("failed to send confirmed booking notification to owner",
						slog.String("error", notifyErr.Error()))
				}
			} else if booking.Status == domain.BookingStatusActive {
				notifyErr := u.notificationUseCase.NotifyRenterBookingStarted(renter.UserID, booking.ID, apartmentTitle)
				if notifyErr != nil {
					logger.Warn("failed to send booking start notification to renter",
						slog.String("error", notifyErr.Error()))
				}

				notifyErr = u.notificationUseCase.NotifyBookingStarted(propertyOwner.UserID, booking.ID, apartmentTitle, renterName)
				if notifyErr != nil {
					logger.Warn("failed to send booking start notification to owner",
						slog.String("error", notifyErr.Error()))
				}
			}
		}
	}

	if u.availabilityService != nil {
		if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			logger.Warn("failed to recalculate apartment availability after booking confirmation",
				slog.Int("apartment_id", booking.ApartmentID),
				slog.Int("booking_id", booking.ID),
				slog.String("error", err.Error()))
		}
	}

	return booking, nil
}

func (u *bookingUseCase) GetAvailableTimeSlots(apartmentID int, date string, duration int) ([]string, error) {
	apartment, err := u.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("квартира не найдена: %w", err)
	}

	if apartment == nil {
		return nil, fmt.Errorf("квартира с ID %d не найдена", apartmentID)
	}

	if apartment.Status != domain.AptStatusApproved {
		return nil, fmt.Errorf("квартира недоступна для бронирования")
	}

	if duration == 24 {
		if !apartment.RentalTypeDaily {
			return nil, fmt.Errorf("данная квартира не поддерживает посуточную аренду")
		}
	} else {
		if !apartment.RentalTypeHourly {
			return nil, fmt.Errorf("данная квартира не поддерживает почасовую аренду")
		}
	}

	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("неверный формат даты: %s (ожидается 2006-01-02)", date)
	}

	nowUTC := utils.GetCurrentTimeUTC()
	nowLocal := utils.ConvertOutputFromUTC(nowUTC)
	todayLocal := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, utils.KazakhstanTZ)

	targetDateLocal := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, utils.KazakhstanTZ)
	if targetDateLocal.Before(todayLocal) {
		return nil, fmt.Errorf("дата не может быть в прошлом")
	}

	var availableSlots []string

	startHour := 0
	endHour := 24

	if duration < 24 && apartment.RentalTypeHourly {
		startHour = 10
		endHour = 22
	}

	for hour := startHour; hour < endHour; hour++ {
		startTimeLocal := time.Date(
			targetDate.Year(), targetDate.Month(), targetDate.Day(),
			hour, 0, 0, 0, utils.KazakhstanTZ,
		)

		if targetDateLocal.Equal(todayLocal) && startTimeLocal.Before(nowLocal) {
			continue
		}

		if duration < 24 && apartment.RentalTypeHourly {
			endHour := hour + duration
			if endHour > 22 {
				continue
			}
		}

		startTimeUTC := startTimeLocal.UTC()
		endTimeUTC := startTimeUTC.Add(time.Duration(duration) * time.Hour)

		isAvailable, err := u.bookingRepo.CheckApartmentAvailability(
			apartmentID,
			startTimeUTC,
			endTimeUTC,
			nil,
		)

		if err == nil && isAvailable {
			timeStr := fmt.Sprintf("%02d:00", hour)
			availableSlots = append(availableSlots, timeStr)
		}
	}

	return availableSlots, nil
}

func (u *bookingUseCase) AdminGetAllBookings(filters map[string]interface{}, page, pageSize int) ([]*domain.Booking, int, error) {
	bookings, total, err := u.bookingRepo.GetAll(filters, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get bookings: %w", err)
	}

	for _, booking := range bookings {
		utils.LoadBookingRelatedData(booking, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
		u.loadContractID(booking)
	}

	return bookings, total, nil
}

func (u *bookingUseCase) AdminUpdateBookingStatus(bookingID int, status domain.BookingStatus, reason string, adminID int) error {
	admin, err := u.userUseCase.GetByID(adminID)
	if err != nil {
		return fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil {
		return fmt.Errorf("admin with id %d not found", adminID)
	}
	if admin.Role != domain.RoleAdmin {
		return fmt.Errorf("only admins can update booking status")
	}

	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("booking with id %d not found", bookingID)
	}

	oldStatus := booking.Status

	booking.Status = status
	if reason != "" {
		if status == domain.BookingStatusRejected {
			booking.OwnerComment = &reason
		} else if status == domain.BookingStatusCanceled {
			booking.CancellationReason = &reason
		}
	}

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	switch status {
	case domain.BookingStatusActive:
		if oldStatus == domain.BookingStatusApproved {
			utils.LoadBookingRelatedData(booking, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
			if u.notificationUseCase != nil && booking.Apartment != nil && booking.Renter != nil {
				apartmentTitle := fmt.Sprintf("%s, кв. %d", booking.Apartment.Street, booking.Apartment.ApartmentNumber)
				u.notificationUseCase.NotifyRenterBookingStarted(booking.Renter.UserID, bookingID, apartmentTitle)
			}
		}

	case domain.BookingStatusCompleted:
		if u.lockUseCase != nil {
			u.lockUseCase.DeactivatePasswordForBooking(bookingID)
		}
		utils.LoadBookingRelatedData(booking, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
		if u.notificationUseCase != nil && booking.Apartment != nil && booking.Renter != nil {
			apartmentTitle := fmt.Sprintf("%s, кв. %d", booking.Apartment.Street, booking.Apartment.ApartmentNumber)
			u.notificationUseCase.NotifyBookingCompleted(booking.Renter.UserID, bookingID, apartmentTitle)
		}

	case domain.BookingStatusCanceled:
		if u.lockUseCase != nil {
			u.lockUseCase.DeactivatePasswordForBooking(bookingID)
		}
		if u.schedulerService != nil {
			u.schedulerService.RemoveScheduledTasksForBooking(bookingID)
		}
		utils.LoadBookingRelatedData(booking, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
		if u.notificationUseCase != nil && booking.Apartment != nil && booking.Renter != nil {
			apartmentTitle := fmt.Sprintf("%s, кв. %d", booking.Apartment.Street, booking.Apartment.ApartmentNumber)
			u.notificationUseCase.NotifyBookingCanceled(booking.Renter.UserID, bookingID, apartmentTitle, reason)
		}

	case domain.BookingStatusRejected:
		utils.LoadBookingRelatedData(booking, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
		if u.notificationUseCase != nil && booking.Apartment != nil && booking.Renter != nil {
			apartmentTitle := fmt.Sprintf("%s, кв. %d", booking.Apartment.Street, booking.Apartment.ApartmentNumber)
			u.notificationUseCase.NotifyBookingRejected(booking.Renter.UserID, bookingID, apartmentTitle, reason)
		}
	}

	return nil
}

func (u *bookingUseCase) AdminCancelBooking(bookingID int, reason string, adminID int) error {
	admin, err := u.userUseCase.GetByID(adminID)
	if err != nil {
		return fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil {
		return fmt.Errorf("admin with id %d not found", adminID)
	}
	if admin.Role != domain.RoleAdmin {
		return fmt.Errorf("only admins can cancel bookings")
	}

	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("booking with id %d not found", bookingID)
	}

	if booking.Status == domain.BookingStatusCanceled {
		return fmt.Errorf("booking is already canceled")
	}
	if booking.Status == domain.BookingStatusCompleted {
		return fmt.Errorf("cannot cancel completed booking")
	}

	if booking.PaymentID != nil {
		paymentRecord, err := u.paymentRepo.GetByID(*booking.PaymentID)
		if err == nil && paymentRecord != nil && u.paymentUseCase != nil {
			logger.Info("attempting to refund payment for admin cancelled booking",
				slog.Int("booking_id", bookingID),
				slog.String("payment_id", paymentRecord.PaymentID),
				slog.Int("admin_id", adminID))

			refundResponse, refundErr := u.paymentUseCase.RefundPayment(paymentRecord.PaymentID, nil)
			if refundErr != nil {
				logger.Error("failed to refund payment for admin cancelled booking",
					slog.Int("booking_id", bookingID),
					slog.String("payment_id", paymentRecord.PaymentID),
					slog.Int("admin_id", adminID),
					slog.String("error", refundErr.Error()))
			} else if refundResponse.Success {
				logger.Info("payment refunded successfully for admin cancelled booking",
					slog.Int("booking_id", bookingID),
					slog.String("payment_id", paymentRecord.PaymentID),
					slog.Int("admin_id", adminID))
			}
		}
	}

	booking.Status = domain.BookingStatusCanceled
	if reason != "" {
		fullReason := fmt.Sprintf("Отменено администратором: %s", reason)
		booking.CancellationReason = &fullReason
	} else {
		defaultReason := "Отменено администратором"
		booking.CancellationReason = &defaultReason
	}

	err = u.bookingRepo.Update(booking)
	if err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	if u.lockUseCase != nil {
		err = u.lockUseCase.DeactivatePasswordForBooking(bookingID)
		if err != nil {
			logger.Warn("failed to deactivate lock passwords for booking",
				slog.Int("booking_id", bookingID),
				slog.String("error", err.Error()))
		}
	}

	if u.schedulerService != nil {
		err = u.schedulerService.RemoveScheduledTasksForBooking(bookingID)
		if err != nil {
			logger.Warn("failed to remove scheduled tasks for booking",
				slog.Int("booking_id", bookingID),
				slog.String("error", err.Error()))
		}
	}

	if u.notificationUseCase != nil {
		utils.LoadBookingRelatedData(booking, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
		if booking.Apartment != nil && booking.Renter != nil {
			apartmentTitle := fmt.Sprintf("%s, кв. %d", booking.Apartment.Street, booking.Apartment.ApartmentNumber)
			notifyReason := reason
			if notifyReason == "" {
				notifyReason = "Решение администрации"
			}
			err = u.notificationUseCase.NotifyBookingCanceled(booking.Renter.UserID, bookingID, apartmentTitle, notifyReason)
			if err != nil {
				logger.Warn("failed to send cancellation notification",
					slog.String("error", err.Error()))
			}
		}
	}

	return nil
}

func (u *bookingUseCase) GetStatusStatistics() (map[string]int, error) {
	return u.bookingRepo.GetStatusStatistics()
}

func (u *bookingUseCase) loadContractID(booking *domain.Booking) {
	if u.contractUseCase != nil {
		// Используем contractRepo через contractUseCase - нужно добавить метод
		// Пока используем простой способ через GetContractByBookingID
		contract, err := u.contractUseCase.GetContractByBookingID(booking.ID)
		if err == nil && contract != nil {
			booking.ContractID = &contract.ID
		}
	}
}

func (u *bookingUseCase) loadContractIDs(bookings []*domain.Booking) {
	for _, booking := range bookings {
		u.loadContractID(booking)
	}
}

func (u *bookingUseCase) GetMyBookingsLockAccess(userID int) (*domain.MyBookingsLockAccessResponse, error) {
	statuses := []domain.BookingStatus{
		domain.BookingStatusApproved,
		domain.BookingStatusActive,
	}

	bookings, _, err := u.GetRenterBookings(userID, statuses, nil, nil, 1, 1000)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения бронирований: %w", err)
	}

	var bookingLockAccesses []domain.BookingLockAccess

	for _, booking := range bookings {
		lockAccess, err := u.calculateLockAccessForBooking(booking, userID)
		if err != nil {
			continue
		}

		bookingLockAccesses = append(bookingLockAccesses, *lockAccess)
	}

	return &domain.MyBookingsLockAccessResponse{
		Bookings: bookingLockAccesses,
		Total:    len(bookingLockAccesses),
	}, nil
}

func (u *bookingUseCase) GetBookingLockAccess(bookingID, userID int) (*domain.BookingLockAccessResponse, error) {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("бронирование не найдено: %w", err)
	}

	renter, err := u.renterRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("профиль арендатора не найден: %w", err)
	}

	if booking.RenterID != renter.ID {
		return nil, fmt.Errorf("нет прав для доступа к этому бронированию")
	}

	lockAccess, err := u.calculateLockAccessForBooking(booking, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка расчета доступа к замку: %w", err)
	}

	return &domain.BookingLockAccessResponse{
		BookingID:  bookingID,
		LockAccess: lockAccess.LockAccess,
	}, nil
}

func (u *bookingUseCase) calculateLockAccessForBooking(booking *domain.Booking, userID int) (*domain.BookingLockAccess, error) {
	if booking.Apartment == nil {
		apartment, err := u.apartmentRepo.GetByID(booking.ApartmentID)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения квартиры: %w", err)
		}
		booking.Apartment = apartment
	}

	response := &domain.BookingLockAccess{
		BookingID: booking.ID,
		ApartmentInfo: domain.ApartmentInfo{
			ID:      booking.Apartment.ID,
			Title:   fmt.Sprintf("%s, кв. %d", booking.Apartment.Street, booking.Apartment.ApartmentNumber),
			Address: booking.Apartment.Street,
		},
		BookingPeriod: domain.BookingPeriod{
			StartDate: booking.StartDate,
			EndDate:   booking.EndDate,
			Status:    booking.Status,
		},
	}

	if u.lockUseCase != nil {
		lock, err := u.lockUseCase.GetLockByApartmentID(booking.ApartmentID)
		if err == nil && lock != nil {
			response.LockInfo = &domain.LockInfo{
				UniqueID: lock.UniqueID,
				Name:     lock.Name,
			}
		}
	}

	lockAccess := u.determineLockAccessStatus(booking, userID)
	response.LockAccess = lockAccess

	return response, nil
}

func (u *bookingUseCase) determineLockAccessStatus(booking *domain.Booking, userID int) domain.LockAccess {
	now := utils.GetCurrentTimeUTC()

	if booking.Status != domain.BookingStatusApproved && booking.Status != domain.BookingStatusActive {
		return domain.LockAccess{
			Status:         domain.LockAccessNotAvailable,
			CanGenerateNow: false,
			PasswordExists: false,
			Message:        "Бронирование не активно",
			DetailedReason: &[]string{fmt.Sprintf("Статус бронирования: %s", booking.Status)}[0],
		}
	}

	if booking.EndDate.Before(now) {
		return domain.LockAccess{
			Status:         domain.LockAccessNotAvailable,
			CanGenerateNow: false,
			PasswordExists: false,
			Message:        "Бронирование завершено",
			DetailedReason: &[]string{"Время выезда уже прошло"}[0],
		}
	}

	if u.lockUseCase != nil {
		existingPasswords, err := u.lockUseCase.GetTempPasswordsByBookingID(booking.ID)
		if err == nil {
			for _, pwd := range existingPasswords {
				if pwd.IsActive && pwd.ValidUntil.After(now) {
					usageInstructions := "Введите пароль на клавиатуре замка и нажмите #"
					return domain.LockAccess{
						Status:             domain.LockAccessPasswordExists,
						CanGenerateNow:     false,
						PasswordExists:     true,
						Password:           &pwd.Password,
						PasswordValidFrom:  &pwd.ValidFrom,
						PasswordValidUntil: &pwd.ValidUntil,
						Message:            fmt.Sprintf("Пароль активен до %s", pwd.ValidUntil.In(utils.KazakhstanTZ).Format("02.01 15:04")),
						UsageInstructions:  &usageInstructions,
					}
				}
			}
		}
	}

	timeUntilStart := booking.StartDate.Sub(now)

	minBookingDuration := 3 * time.Hour
	cleaningBuffer := 1 * time.Hour
	earlyAccessThreshold := minBookingDuration + cleaningBuffer
	standardAccessTime := 15 * time.Minute

	if timeUntilStart <= earlyAccessThreshold {
		return domain.LockAccess{
			Status:         domain.LockAccessAvailableNow,
			CanGenerateNow: true,
			PasswordExists: false,
			Message:        "Можно получить пароль прямо сейчас",
			DetailedReason: &[]string{"Ближайшая бронь - пароль будет активен с момента заезда"}[0],
		}
	}

	standardAvailableAt := booking.StartDate.Add(-standardAccessTime)

	if now.After(standardAvailableAt) || now.Equal(standardAvailableAt) {
		return domain.LockAccess{
			Status:         domain.LockAccessAvailableNow,
			CanGenerateNow: true,
			PasswordExists: false,
			Message:        "Можно получить пароль прямо сейчас",
		}
	}

	availableAtKZ := standardAvailableAt.In(utils.KazakhstanTZ)
	hoursUntil := timeUntilStart.Hours()
	detailedReason := fmt.Sprintf("До начала бронирования осталось %.1f часов", hoursUntil)

	return domain.LockAccess{
		Status:         domain.LockAccessAvailableSoon,
		CanGenerateNow: false,
		PasswordExists: false,
		AvailableAt:    &standardAvailableAt,
		Message:        fmt.Sprintf("Пароль будет доступен с %s (за 15 минут до заезда)", availableAtKZ.Format("02.01 15:04")),
		DetailedReason: &detailedReason,
	}
}

func (u *bookingUseCase) AdminGetBookingByID(bookingID int) (*domain.Booking, error) {
	return u.GetBookingByID(bookingID)
}

func (u *bookingUseCase) AdminGetBookingStatistics() (map[string]interface{}, error) {
	statusStats, err := u.GetStatusStatistics()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения статистики статусов: %w", err)
	}

	return map[string]interface{}{
		"status_statistics": statusStats,
	}, nil
}

func (u *bookingUseCase) ProcessPayment(bookingID int, paymentID string) (*domain.Booking, error) {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("бронирование не найдено: %w", err)
	}

	if booking.Status != domain.BookingStatusAwaitingPayment {
		return nil, fmt.Errorf("можно оплатить только бронирования со статусом 'awaiting_payment'")
	}

	existingPayment, err := u.paymentRepo.GetByPaymentID(paymentID)
	if err == nil && existingPayment != nil {
		if existingPayment.BookingID != int64(bookingID) {
			logger.Error("payment_id already used for another booking",
				slog.String("payment_id", paymentID),
				slog.Int("current_booking_id", bookingID),
				slog.Int64("existing_booking_id", existingPayment.BookingID),
				slog.Int64("existing_payment_id", existingPayment.ID))

			duplicateLog := &domain.PaymentLog{
				PaymentID:    &existingPayment.ID,
				BookingID:    int64(bookingID),
				FPPaymentID:  paymentID,
				Action:       domain.PaymentLogActionProcessPayment,
				Source:       domain.PaymentLogSourceAPI,
				Success:      false,
				ErrorMessage: &[]string{fmt.Sprintf("Попытка повторного использования payment_id для бронирования %d", bookingID)}[0],
			}
			u.paymentLogRepo.Create(duplicateLog)

			return nil, fmt.Errorf("платеж %s уже использован для бронирования #%d", paymentID, existingPayment.BookingID)
		}

		if booking.Status == domain.BookingStatusApproved || booking.Status == domain.BookingStatusActive {
			logger.Info("payment already processed for this booking",
				slog.String("payment_id", paymentID),
				slog.Int("booking_id", bookingID),
				slog.String("status", string(booking.Status)))
			return booking, nil
		}
	}

	paymentStatus, err := u.paymentUseCase.CheckPaymentStatusWithBooking(paymentID, int64(bookingID))
	if err != nil {
		logger.Error("failed to check payment status",
			slog.String("payment_id", paymentID),
			slog.Int("booking_id", bookingID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("ошибка проверки платежа: %w", err)
	}

	if paymentStatus.Status != "success" {
		logger.Warn("payment not successful",
			slog.String("payment_id", paymentID),
			slog.Int("booking_id", bookingID),
			slog.String("status", paymentStatus.Status),
			slog.String("error_message", paymentStatus.ErrorMessage))
		return nil, fmt.Errorf("платеж не завершен: %s", paymentStatus.Status)
	}

	payment := &domain.Payment{
		BookingID:      int64(bookingID),
		PaymentID:      paymentID,
		Amount:         booking.FinalPrice,
		Currency:       "KZT",
		Status:         domain.PaymentStatusSuccess,
		PaymentMethod:  &paymentStatus.PaymentMethod,
		ProviderStatus: &paymentStatus.Status,
		ProcessedAt:    &[]time.Time{time.Now()}[0],
	}

	if err := u.paymentRepo.Create(payment); err != nil {
		logger.Error("failed to save payment record",
			slog.String("payment_id", paymentID),
			slog.Int("booking_id", bookingID),
			slog.String("error", err.Error()))

		// КРИТИЧНО: если не можем сохранить платеж - останавливаем процесс
		return nil, fmt.Errorf("ошибка сохранения платежа в БД: %w", err)
	}

	logger.Info("payment record created successfully",
		slog.String("payment_id", paymentID),
		slog.Int("booking_id", bookingID),
		slog.Int64("payment_db_id", payment.ID))

	now := utils.GetCurrentTimeUTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	bookingDate := time.Date(booking.StartDate.Year(), booking.StartDate.Month(), booking.StartDate.Day(), 0, 0, 0, 0, booking.StartDate.Location())

	if bookingDate.Before(today) || bookingDate.Equal(today) && booking.StartDate.Before(now) {
		booking.Status = domain.BookingStatusActive

		if u.availabilityService != nil {
			if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
				logger.Error("failed to recalculate apartment availability after payment",
					slog.Int("apartment_id", booking.ApartmentID),
					slog.String("error", err.Error()))
				return nil, fmt.Errorf("ошибка пересчета статуса квартиры: %w", err)
			}
		}

		logger.Info("booking activated immediately after payment",
			slog.Int("booking_id", bookingID),
			slog.String("payment_id", paymentID))
	} else if bookingDate.Equal(today) {
		booking.Status = domain.BookingStatusApproved

		logger.Info("booking approved after payment, waiting for start time",
			slog.Int("booking_id", bookingID),
			slog.String("payment_id", paymentID))
	} else {
		booking.Status = domain.BookingStatusPending

		logger.Info("booking set to pending after payment, requires owner approval",
			slog.Int("booking_id", bookingID),
			slog.String("payment_id", paymentID))
	}

	booking.PaymentID = &payment.ID

	err = u.bookingRepo.Update(booking)
	if err != nil {
		logger.Error("failed to update booking after payment",
			slog.Int("booking_id", bookingID),
			slog.String("payment_id", paymentID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("ошибка обновления бронирования: %w", err)
	}

	finalBookingStatus := string(booking.Status)
	payment.FinalBookingStatus = &finalBookingStatus
	if updateErr := u.paymentRepo.Update(payment); updateErr != nil {
		logger.Warn("failed to update payment record with final status",
			slog.String("payment_id", paymentID),
			slog.Int("booking_id", bookingID),
			slog.String("error", updateErr.Error()))
	}

	oldStatus := "awaiting_payment"
	processLog := &domain.PaymentLog{
		PaymentID:   &payment.ID,
		BookingID:   int64(bookingID),
		FPPaymentID: paymentID,
		Action:      domain.PaymentLogActionProcessPayment,
		OldStatus:   &oldStatus,
		NewStatus:   &finalBookingStatus,
		Source:      domain.PaymentLogSourceAPI,
		Success:     true,
	}

	if logErr := u.paymentLogRepo.Create(processLog); logErr != nil {
		logger.Warn("failed to save payment process log",
			slog.String("payment_id", paymentID),
			slog.Int("booking_id", bookingID),
			slog.String("error", logErr.Error()))
	}

	utils.LoadBookingRelatedData(booking, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)
	u.loadContractID(booking)

	if u.notificationUseCase != nil && booking.Apartment != nil {
		apartmentTitle := fmt.Sprintf("%s, кв. %d", booking.Apartment.Street, booking.Apartment.ApartmentNumber)

		renter, renterErr := u.renterRepo.GetByID(booking.RenterID)
		if renterErr != nil {
			logger.Warn("failed to get renter for payment notification",
				slog.Int("renter_id", booking.RenterID),
				slog.String("error", renterErr.Error()))
			return booking, nil
		}

		if booking.Status == domain.BookingStatusApproved {
			err = u.notificationUseCase.NotifyBookingApproved(renter.UserID, booking.ID, apartmentTitle)
			if err != nil {
				logger.Warn("failed to send booking approved notification",
					slog.Int("booking_id", bookingID),
					slog.String("error", err.Error()))
			}
		} else if booking.Status == domain.BookingStatusActive {
			err = u.notificationUseCase.NotifyRenterBookingStarted(renter.UserID, booking.ID, apartmentTitle)
			if err != nil {
				logger.Warn("failed to send booking started notification",
					slog.Int("booking_id", bookingID),
					slog.String("error", err.Error()))
			}
		}
	}

	logger.Info("payment processed successfully",
		slog.Int("booking_id", bookingID),
		slog.String("payment_id", paymentID),
		slog.String("new_status", string(booking.Status)),
		slog.String("amount", paymentStatus.Amount),
		slog.String("currency", paymentStatus.Currency))

	if u.availabilityService != nil {
		if err := u.availabilityService.RecalculateApartmentAvailability(booking.ApartmentID); err != nil {
			logger.Warn("failed to recalculate apartment availability after payment processing",
				slog.Int("apartment_id", booking.ApartmentID),
				slog.Int("booking_id", booking.ID),
				slog.String("error", err.Error()))
		}
	}

	return booking, nil
}

func (u *bookingUseCase) ProcessPaymentWithOrder(bookingID int, orderID string) (*domain.Booking, error) {
	paymentStatus, err := u.paymentUseCase.CheckPaymentStatusByOrderID(orderID, int64(bookingID))
	if err != nil {
		logger.Error("failed to get payment status by order ID",
			slog.String("order_id", orderID),
			slog.Int("booking_id", bookingID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("ошибка получения статуса платежа по order_id: %w", err)
	}

	if !paymentStatus.Exists {
		return nil, fmt.Errorf("платеж с order_id %s не найден: %s", orderID, paymentStatus.ErrorMessage)
	}

	logger.Info("converting order_id to payment_id",
		slog.String("order_id", orderID),
		slog.String("payment_id", paymentStatus.PaymentID),
		slog.Int("booking_id", bookingID))

	return u.ProcessPayment(bookingID, paymentStatus.PaymentID)
}

func (u *bookingUseCase) GetPaymentReceipt(bookingID, userID int) (*domain.PaymentReceipt, error) {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("бронирование не найдено: %w", err)
	}

	user, userErr := u.userUseCase.GetByID(userID)
	if userErr != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", userErr)
	}

	if user.Role == domain.RoleAdmin {
	} else if user.Role == domain.RoleOwner {
		apartment, aptErr := u.apartmentRepo.GetByID(booking.ApartmentID)
		if aptErr != nil {
			return nil, fmt.Errorf("квартира не найдена для проверки прав владельца: %w", aptErr)
		}

		owner, ownerErr := u.propertyOwnerRepo.GetByUserID(userID)
		if ownerErr != nil {
			return nil, fmt.Errorf("данные владельца не найдены: %w", ownerErr)
		}
		if owner == nil {
			return nil, fmt.Errorf("пользователь не является зарегистрированным владельцем")
		}
		if apartment.OwnerID != owner.ID {
			return nil, fmt.Errorf("отказано в доступе: вы не являетесь владельцем данной квартиры")
		}
	} else {
		renter, err := utils.GetRenterByUserID(u.renterRepo, userID)
		if err != nil {
			return nil, fmt.Errorf("отказано в доступе: недостаточно прав для просмотра чека (роль: %s)", user.Role)
		}

		if booking.RenterID != renter.ID {
			return nil, fmt.Errorf("отказано в доступе: вы можете просматривать только свои чеки об оплате")
		}
	}

	var paymentRecord *domain.Payment
	var fpPaymentID string

	if booking.PaymentID != nil {
		paymentRecord, err = u.paymentRepo.GetByID(*booking.PaymentID)
		if err != nil {
			return nil, fmt.Errorf("платеж не найден: %w", err)
		}
		fpPaymentID = paymentRecord.PaymentID
	} else {
		paymentRecords, err := u.paymentRepo.GetByBookingID(int64(bookingID))
		if err != nil || len(paymentRecords) == 0 {
			return nil, fmt.Errorf("платеж для данного бронирования не найден")
		}
		for i := len(paymentRecords) - 1; i >= 0; i-- {
			if paymentRecords[i].Status == domain.PaymentStatusSuccess {
				paymentRecord = paymentRecords[i]
				break
			}
		}
		if paymentRecord == nil {
			return nil, fmt.Errorf("успешный платеж для данного бронирования не найден")
		}
		fpPaymentID = paymentRecord.PaymentID
	}

	paymentStatus, err := u.paymentUseCase.CheckPaymentStatus(fpPaymentID)
	if err != nil {
		logger.Warn("failed to get payment status for receipt",
			slog.Int("booking_id", bookingID),
			slog.String("payment_id", fpPaymentID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("ошибка получения данных платежа")
	}

	if !paymentStatus.Exists {
		return nil, fmt.Errorf("платеж не найден в системе")
	}

	var cardPan string = "****-****-****-****"
	if u.paymentUseCase != nil {
		// Здесь мы получаем полную информацию о платеже от FreedomPay
		// В будущем можно добавить метод для получения полного ответа FreedomPay
		// который включает замаскированный номер карты
		cardPan = "****-****-****-****" // Пока оставляем placeholder
	}

	utils.LoadBookingRelatedData(booking, u.apartmentRepo, u.renterRepo, u.propertyOwnerRepo)

	rentalType := "hourly"
	if booking.Duration == 24 {
		rentalType = "daily"
	}

	apartmentAddress := fmt.Sprintf("%s, кв. %d",
		booking.Apartment.Street, booking.Apartment.ApartmentNumber)
	if booking.Apartment.ResidentialComplex != nil && *booking.Apartment.ResidentialComplex != "" {
		apartmentAddress = fmt.Sprintf("%s, %s", *booking.Apartment.ResidentialComplex, apartmentAddress)
	}

	receiptID := fmt.Sprintf("RECEIPT-%s", fpPaymentID)

	receipt := &domain.PaymentReceipt{
		ReceiptID:     receiptID,
		BookingNumber: booking.BookingNumber,
		PaymentID:     fpPaymentID,
		OrderID:       fmt.Sprintf("booking_%d", booking.ID),
		Status:        "paid",
		PaymentDate:   paymentStatus.CreateDate,
		PaymentMethod: paymentStatus.PaymentMethod,
		CardPan:       cardPan,
		Amounts: domain.ReceiptAmounts{
			TotalPrice: booking.TotalPrice,
			ServiceFee: booking.ServiceFee,
			FinalPrice: booking.FinalPrice,
			Currency:   paymentStatus.Currency,
		},
		BookingDetails: domain.ReceiptBookingDetails{
			ApartmentAddress: apartmentAddress,
			StartDate:        booking.StartDate.Format(time.RFC3339),
			EndDate:          booking.EndDate.Format(time.RFC3339),
			DurationHours:    booking.Duration,
			RentalType:       rentalType,
		},
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	logger.Info("payment receipt generated",
		slog.Int("booking_id", bookingID),
		slog.String("payment_id", fpPaymentID),
		slog.String("receipt_id", receiptID))

	return receipt, nil
}
