package usecase

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type LockAutoUpdateService interface {
	ProcessWebhookEvent(event *domain.TuyaWebhookEvent) error
	UpdateLock(lock *domain.Lock) error
	UpdateAllLocks()
}

type lockUseCase struct {
	lockRepo            domain.LockRepository
	apartmentRepo       domain.ApartmentRepository
	bookingRepo         domain.BookingRepository
	propertyOwnerRepo   domain.PropertyOwnerRepository
	renterRepo          domain.RenterRepository
	userUseCase         domain.UserUseCase
	tuyaService         domain.TuyaLockService
	autoUpdateService   LockAutoUpdateService
	notificationUseCase domain.NotificationUseCase
}

func NewLockUseCase(
	lockRepo domain.LockRepository,
	apartmentRepo domain.ApartmentRepository,
	bookingRepo domain.BookingRepository,
	propertyOwnerRepo domain.PropertyOwnerRepository,
	renterRepo domain.RenterRepository,
	userUseCase domain.UserUseCase,
	tuyaService domain.TuyaLockService,
	autoUpdateService LockAutoUpdateService,
) domain.LockUseCase {
	useCase := &lockUseCase{
		lockRepo:          lockRepo,
		apartmentRepo:     apartmentRepo,
		bookingRepo:       bookingRepo,
		propertyOwnerRepo: propertyOwnerRepo,
		renterRepo:        renterRepo,
		userUseCase:       userUseCase,
		tuyaService:       tuyaService,
		autoUpdateService: autoUpdateService,
	}

	return useCase
}

func (u *lockUseCase) SetNotificationUseCase(notificationUseCase domain.NotificationUseCase) {
	u.notificationUseCase = notificationUseCase
}

func (u *lockUseCase) generateNumericPassword() (string, error) {
	min := int64(1000000)
	max := int64(9999999)

	n, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return "", fmt.Errorf("ошибка генерации пароля: %w", err)
	}

	return fmt.Sprintf("%d", n.Int64()+min), nil
}

func (u *lockUseCase) CreateLock(request *domain.CreateLockRequest) (*domain.Lock, error) {

	_, err := u.lockRepo.GetByUniqueID(request.UniqueID)
	if err == nil {
		return nil, fmt.Errorf("замок с таким ID уже существует")
	}

	if request.ApartmentID != nil {
		_, err := u.apartmentRepo.GetByID(*request.ApartmentID)
		if err != nil {
			return nil, fmt.Errorf("квартира не найдена: %w", err)
		}
	}

	lock := &domain.Lock{
		UniqueID:        request.UniqueID,
		ApartmentID:     request.ApartmentID,
		Name:            request.Name,
		Description:     request.Description,
		CurrentStatus:   domain.LockStatusClosed,
		FirmwareVersion: request.FirmwareVersion,
		IsOnline:        false,
		TuyaDeviceID:    request.TuyaDeviceID,
		OwnerPassword:   request.OwnerPassword,
	}

	err = u.lockRepo.Create(lock)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать замок: %w", err)
	}

	return lock, nil
}

func (u *lockUseCase) GetLockByID(id int) (*domain.Lock, error) {
	return u.lockRepo.GetByID(id)
}

func (u *lockUseCase) GetLockByUniqueID(uniqueID string) (*domain.Lock, error) {
	return u.lockRepo.GetByUniqueID(uniqueID)
}

func (u *lockUseCase) GetLockByApartmentID(apartmentID int) (*domain.Lock, error) {
	return u.lockRepo.GetByApartmentID(apartmentID)
}

func (u *lockUseCase) GetAllLocks() ([]*domain.Lock, error) {
	return u.lockRepo.GetAll()
}

func (u *lockUseCase) GetAllWithFilters(filters map[string]interface{}, page, pageSize int) ([]*domain.Lock, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	locks, total, err := u.lockRepo.GetAllWithFilters(filters, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get locks with filters: %w", err)
	}

	return locks, total, nil
}

func (u *lockUseCase) UpdateLock(id int, request *domain.UpdateLockRequest) error {
	lock, err := u.lockRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	if request.Name != "" {
		lock.Name = request.Name
	}
	if request.Description != "" {
		lock.Description = request.Description
	}
	if request.FirmwareVersion != "" {
		lock.FirmwareVersion = request.FirmwareVersion
	}

	return u.lockRepo.Update(lock)
}

func (u *lockUseCase) DeleteLock(id int) error {
	_, err := u.lockRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	return u.lockRepo.Delete(id)
}

func (u *lockUseCase) UpdateLockStatus(request *domain.LockStatusUpdateRequest) error {
	lock, err := u.lockRepo.GetByUniqueID(request.UniqueID)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	oldStatus := lock.CurrentStatus

	timestamp := request.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	err = u.lockRepo.UpdateStatus(request.UniqueID, request.Status, &timestamp)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус: %w", err)
	}

	if request.BatteryLevel != nil || request.SignalStrength != nil {
		err = u.lockRepo.UpdateHeartbeat(request.UniqueID, timestamp, request.BatteryLevel, request.SignalStrength)
		if err != nil {
			return fmt.Errorf("не удалось обновить heartbeat: %w", err)
		}
	}

	if oldStatus != request.Status {
		statusLog := &domain.LockStatusLog{
			LockID:       lock.ID,
			OldStatus:    &oldStatus,
			NewStatus:    request.Status,
			ChangeSource: domain.LockChangeSourceManual,
			Notes:        "Статус обновлен устройством",
		}
		u.lockRepo.CreateStatusLog(statusLog)
	}

	return nil
}

func (u *lockUseCase) ProcessHeartbeat(request *domain.LockHeartbeatRequest) error {
	lock, err := u.lockRepo.GetByUniqueID(request.UniqueID)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	timestamp := request.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	err = u.lockRepo.UpdateHeartbeat(request.UniqueID, timestamp, request.BatteryLevel, request.SignalStrength)
	if err != nil {
		return fmt.Errorf("не удалось обновить heartbeat: %w", err)
	}

	if lock.CurrentStatus != request.Status {
		err = u.lockRepo.UpdateStatus(request.UniqueID, request.Status, &timestamp)
		if err != nil {
			return fmt.Errorf("не удалось обновить статус: %w", err)
		}

		statusLog := &domain.LockStatusLog{
			LockID:       lock.ID,
			OldStatus:    &lock.CurrentStatus,
			NewStatus:    request.Status,
			ChangeSource: domain.LockChangeSourceManual,
			Notes:        "Статус обновлен через heartbeat",
		}
		u.lockRepo.CreateStatusLog(statusLog)
	}

	return nil
}

func (u *lockUseCase) GeneratePasswordForBooking(uniqueID string, userID int, bookingID int) (string, error) {
	lock, err := u.lockRepo.GetByUniqueID(uniqueID)
	if err != nil {
		return "", fmt.Errorf("замок не найден: %w", err)
	}

	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return "", fmt.Errorf("бронирование не найдено: %w", err)
	}

	if lock.ApartmentID == nil || *lock.ApartmentID != booking.ApartmentID {
		return "", fmt.Errorf("бронирование не для этой квартиры")
	}

	canManage, err := u.CanUserManageLockViaBooking(booking, userID)
	if err != nil {
		return "", fmt.Errorf("ошибка проверки прав: %w", err)
	}
	if !canManage {
		return "", fmt.Errorf("у пользователя нет прав на управление замком через это бронирование")
	}

	existingPasswords, err := u.lockRepo.GetTempPasswordsByBookingID(bookingID)
	if err == nil && len(existingPasswords) > 0 {
		for _, p := range existingPasswords {
			if p.IsActive && p.ValidUntil.After(time.Now()) {
				return p.Password, nil
			}
		}
	}

	password, err := u.generateNumericPassword()
	if err != nil {
		return "", fmt.Errorf("ошибка генерации пароля: %w", err)
	}

	now := utils.GetCurrentTimeUTC()
	validFrom := now.Add(-5 * time.Minute)
	validUntil := booking.EndDate.Add(30 * time.Minute)

	validDays := int(validUntil.Sub(validFrom).Hours()/24) + 1
	if validDays < 1 {
		validDays = 1
	}

	passwordName := fmt.Sprintf("Booking_%d", bookingID)
	tuyaPassword, tuyaPasswordID, err := u.tuyaService.GenerateTemporaryPasswordWithTimes(
		lock.TuyaDeviceID,
		passwordName,
		password,
		validFrom,
		validUntil,
	)
	if err != nil {
		return "", fmt.Errorf("ошибка создания пароля в Tuya: %w", err)
	}

	tempPassword := &domain.LockTempPassword{
		LockID:         lock.ID,
		BookingID:      &bookingID,
		UserID:         &userID,
		Password:       tuyaPassword,
		TuyaPasswordID: tuyaPasswordID,
		Name:           passwordName,
		ValidFrom:      validFrom,
		ValidUntil:     validUntil,
		IsActive:       true,
	}

	err = u.lockRepo.CreateTempPassword(tempPassword)
	if err != nil {
		u.tuyaService.DeleteTempPassword(lock.TuyaDeviceID, tuyaPasswordID)
		return "", fmt.Errorf("ошибка сохранения пароля в БД: %w", err)
	}

	statusLog := &domain.LockStatusLog{
		LockID:       lock.ID,
		OldStatus:    &lock.CurrentStatus,
		NewStatus:    lock.CurrentStatus,
		ChangeSource: domain.LockChangeSourceAPI,
		UserID:       &userID,
		BookingID:    &bookingID,
		Notes:        fmt.Sprintf("Создан временный пароль для бронирования: %s", password),
	}
	u.lockRepo.CreateStatusLog(statusLog)

	if u.notificationUseCase != nil {
		apartment, _ := u.apartmentRepo.GetByID(booking.ApartmentID)
		apartmentTitle := "квартира"
		if apartment != nil {
			apartmentTitle = apartment.Description
		}
		go u.notificationUseCase.NotifyPasswordReady(userID, bookingID, apartmentTitle)
	}

	return tuyaPassword, nil
}

func (u *lockUseCase) GetOwnerPassword(uniqueID string, userID int) (string, error) {
	lock, err := u.lockRepo.GetByUniqueID(uniqueID)
	if err != nil {
		return "", fmt.Errorf("замок не найден: %w", err)
	}

	if lock.ApartmentID == nil {
		return "", fmt.Errorf("замок не привязан к квартире")
	}

	apartment, err := u.apartmentRepo.GetByID(*lock.ApartmentID)
	if err != nil {
		return "", fmt.Errorf("квартира не найдена: %w", err)
	}

	user, err := u.userUseCase.GetByID(userID)
	if err != nil {
		return "", fmt.Errorf("пользователь не найден: %w", err)
	}

	if user.Role == domain.RoleOwner {
		propertyOwner, err := u.propertyOwnerRepo.GetByUserID(userID)
		if err == nil && propertyOwner.ID == apartment.OwnerID {
			return lock.OwnerPassword, nil
		}
	}

	return "", fmt.Errorf("только владелец квартиры может получить постоянный пароль")
}

func (u *lockUseCase) DeactivatePasswordForBooking(bookingID int) error {
	passwords, err := u.lockRepo.GetTempPasswordsByBookingID(bookingID)
	if err != nil {
		return fmt.Errorf("ошибка получения паролей: %w", err)
	}

	for _, password := range passwords {
		if password.IsActive {
			err = u.lockRepo.DeactivateTempPassword(password.ID)
			if err != nil {
				log.Printf("❌ Ошибка деактивации пароля %d в БД: %v", password.ID, err)
				continue
			}
			log.Printf("✅ Пароль %d деактивирован в БД", password.ID)

			lock, err := u.lockRepo.GetByID(password.LockID)
			if err != nil {
				log.Printf("❌ Ошибка получения замка %d: %v", password.LockID, err)
				continue
			}
			if err := u.tuyaService.DeleteTempPassword(lock.TuyaDeviceID, password.TuyaPasswordID); err != nil {
				log.Printf("❌ Ошибка удаления пароля из Tuya: %v", err)
			}
		}
	}

	return nil
}

func (u *lockUseCase) ExtendPasswordForBooking(bookingID int, newEndDate time.Time) error {
	passwords, err := u.lockRepo.GetTempPasswordsByBookingID(bookingID)
	if err != nil {
		return fmt.Errorf("ошибка получения паролей: %w", err)
	}

	newValidUntil := newEndDate.Add(30 * time.Minute)
	extendedCount := 0

	for _, password := range passwords {
		if password.IsActive {
			lock, err := u.lockRepo.GetByID(password.LockID)
			if err != nil {
				log.Printf("❌ Ошибка получения замка %d для продления пароля: %v", password.LockID, err)
				continue
			}

			err = u.tuyaService.DeleteTempPassword(lock.TuyaDeviceID, password.TuyaPasswordID)
			if err != nil {
				log.Printf("❌ Ошибка удаления старого пароля %d из Tuya API: %v", password.ID, err)
				continue
			}

			_, newTuyaPasswordID, err := u.tuyaService.GenerateTemporaryPasswordWithTimes(
				lock.TuyaDeviceID,
				password.Name,
				password.Password,
				password.ValidFrom,
				newValidUntil,
			)
			if err != nil {
				log.Printf("❌ Ошибка создания нового пароля %d в Tuya API: %v", password.ID, err)
				continue
			}

			password.ValidUntil = newValidUntil
			password.TuyaPasswordID = newTuyaPasswordID
			err = u.lockRepo.UpdateTempPassword(password)
			if err != nil {
				log.Printf("❌ Ошибка обновления пароля %d в БД: %v", password.ID, err)
				u.tuyaService.DeleteTempPassword(lock.TuyaDeviceID, newTuyaPasswordID)
				continue
			}

			log.Printf("✅ Пароль %d продлен до %s", password.ID, newValidUntil.Format("2006-01-02 15:04:05"))
			extendedCount++
		}
	}

	if extendedCount == 0 {
		return fmt.Errorf("не найдено активных паролей для продления")
	}

	log.Printf("🔑 Продлено %d паролей для бронирования %d до %s", extendedCount, bookingID, newValidUntil.Format("2006-01-02 15:04:05"))
	return nil
}

func (u *lockUseCase) GetLockStatus(uniqueID string) (*domain.Lock, error) {
	return u.lockRepo.GetByUniqueID(uniqueID)
}

func (u *lockUseCase) GetLockHistory(uniqueID string, limit int) ([]*domain.LockStatusLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return u.lockRepo.GetStatusLogsByUniqueID(uniqueID, limit)
}

func (u *lockUseCase) CanUserControlLock(uniqueID string, userID int) (bool, error) {
	lock, err := u.lockRepo.GetByUniqueID(uniqueID)
	if err != nil {
		return false, fmt.Errorf("замок не найден: %w", err)
	}

	if lock.ApartmentID == nil {
		return false, fmt.Errorf("замок не привязан к квартире")
	}

	apartment, err := u.apartmentRepo.GetByID(*lock.ApartmentID)
	if err != nil {
		return false, fmt.Errorf("квартира не найдена: %w", err)
	}

	user, err := u.userUseCase.GetByID(userID)
	if err != nil {
		return false, fmt.Errorf("пользователь не найден: %w", err)
	}

	if user.Role == domain.RoleOwner {
		propertyOwner, err := u.propertyOwnerRepo.GetByUserID(userID)
		if err == nil && propertyOwner.ID == apartment.OwnerID {
			return true, nil
		}
	}

	if user.Role == domain.RoleUser {
		renter, err := u.renterRepo.GetByUserID(userID)
		if err != nil {
			return false, nil
		}

		now := time.Now()
		bookings, err := u.bookingRepo.GetByApartmentID(apartment.ID, []domain.BookingStatus{domain.BookingStatusApproved, domain.BookingStatusActive})
		if err != nil {
			return false, nil
		}

		for _, booking := range bookings {
			if booking.RenterID == renter.ID &&
				(booking.Status == domain.BookingStatusApproved || booking.Status == domain.BookingStatusActive) &&
				!booking.StartDate.After(now) &&
				booking.EndDate.After(now) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (u *lockUseCase) CheckOfflineLocks() ([]*domain.Lock, error) {
	allLocks, err := u.lockRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список замков: %w", err)
	}

	var offlineLocks []*domain.Lock
	offlineThreshold := 5 * time.Minute

	for _, lock := range allLocks {
		if lock.LastHeartbeat == nil {
			offlineLocks = append(offlineLocks, lock)
			continue
		}

		if time.Since(*lock.LastHeartbeat) > offlineThreshold {
			offlineLocks = append(offlineLocks, lock)

			if lock.IsOnline {
				u.lockRepo.UpdateOnlineStatus(lock.UniqueID, false)
			}
		}
	}

	return offlineLocks, nil
}

func (u *lockUseCase) GetTempPasswordsByLockID(lockID int) ([]*domain.LockTempPassword, error) {
	return u.lockRepo.GetTempPasswordsByLockID(lockID)
}

func (u *lockUseCase) GetTempPasswordsByBookingID(bookingID int) ([]*domain.LockTempPassword, error) {
	return u.lockRepo.GetTempPasswordsByBookingID(bookingID)
}

func (u *lockUseCase) CanUserManageLockViaBooking(booking *domain.Booking, userID int) (bool, error) {
	if booking == nil {
		return false, fmt.Errorf("бронирование не указано")
	}

	renter, err := u.renterRepo.GetByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("у пользователя нет профиля арендатора: %w", err)
	}

	if booking.RenterID != renter.ID {
		return false, fmt.Errorf("бронирование не принадлежит пользователю")
	}

	if booking.Status != domain.BookingStatusApproved && booking.Status != domain.BookingStatusActive {
		return false, fmt.Errorf("бронирование должно быть одобренным или активным (текущий статус: %s)", booking.Status)
	}

	now := utils.GetCurrentTimeUTC()

	if booking.EndDate.Before(now) {
		return false, fmt.Errorf("время выезда уже прошло (%s)", booking.EndDate.Format("2006-01-02 15:04"))
	}

	if booking.Status == domain.BookingStatusActive {
		return true, nil
	}

	if booking.Status == domain.BookingStatusApproved {
		timeUntilStart := booking.StartDate.Sub(now)

		minBookingDuration := 3 * time.Hour
		cleaningBuffer := 1 * time.Hour
		earlyAccessThreshold := minBookingDuration + cleaningBuffer

		if timeUntilStart <= earlyAccessThreshold {
			return true, nil
		}

		standardAccessTime := booking.StartDate.Add(-15 * time.Minute)
		if now.Before(standardAccessTime) {
			availableAtKZ := standardAccessTime.In(utils.KazakhstanTZ)
			return false, fmt.Errorf("пароль можно получить не ранее чем за 15 минут до заезда (доступен с %s)", availableAtKZ.Format("2006-01-02 15:04"))
		}
	}

	return true, nil
}

func (u *lockUseCase) BindLockToApartment(lockID, apartmentID int) error {
	lock, err := u.lockRepo.GetByID(lockID)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	_, err = u.apartmentRepo.GetByID(apartmentID)
	if err != nil {
		return fmt.Errorf("квартира не найдена: %w", err)
	}

	existingLock, err := u.lockRepo.GetByApartmentID(apartmentID)
	if err == nil && existingLock.ID != lockID {
		existingLock.ApartmentID = nil
		u.lockRepo.Update(existingLock)
	}

	lock.ApartmentID = &apartmentID
	err = u.lockRepo.Update(lock)
	if err != nil {
		return fmt.Errorf("не удалось обновить замок: %w", err)
	}

	return nil
}

func (u *lockUseCase) UnbindLockFromApartment(lockID int) error {
	lock, err := u.lockRepo.GetByID(lockID)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	if lock.ApartmentID == nil {
		return fmt.Errorf("замок не привязан к квартире")
	}

	lock.ApartmentID = nil
	err = u.lockRepo.Update(lock)
	if err != nil {
		return fmt.Errorf("не удалось обновить замок: %w", err)
	}

	return nil
}

func (u *lockUseCase) EmergencyResetLock(lockID int) error {
	lock, err := u.lockRepo.GetByID(lockID)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	tempPasswords, err := u.lockRepo.GetTempPasswordsByLockID(lockID)
	if err == nil {
		for _, pwd := range tempPasswords {
			if pwd.IsActive {
				err = u.lockRepo.DeactivateTempPassword(pwd.ID)
				if err != nil {
					log.Printf("❌ Ошибка деактивации пароля %d: %v", pwd.ID, err)
					continue
				}

				if err := u.tuyaService.DeleteTempPassword(lock.TuyaDeviceID, pwd.TuyaPasswordID); err != nil {
					log.Printf("❌ Ошибка удаления пароля из Tuya: %v", err)
				}
			}
		}
	}

	statusLog := &domain.LockStatusLog{
		LockID:       lock.ID,
		OldStatus:    &lock.CurrentStatus,
		NewStatus:    domain.LockStatusClosed,
		ChangeSource: domain.LockChangeSourceSystem,
		Notes:        "Экстренный сброс замка администратором",
	}
	u.lockRepo.CreateStatusLog(statusLog)

	now := utils.GetCurrentTimeUTC()
	err = u.lockRepo.UpdateStatus(lock.UniqueID, domain.LockStatusClosed, &now)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус замка: %w", err)
	}

	return nil
}

func (u *lockUseCase) ProcessTuyaWebhookEvent(event *domain.TuyaWebhookEvent) error {
	log.Printf("🔄 Обработка Tuya webhook события: %s для устройства %s", event.BizCode, event.DevID)

	if u.autoUpdateService == nil {
		return fmt.Errorf("сервис автообновления не инициализирован")
	}

	return u.autoUpdateService.ProcessWebhookEvent(event)
}

func (u *lockUseCase) SyncAllLocksWithTuya() error {
	log.Println("🔄 Начинаем синхронизацию всех замков с Tuya API...")

	if u.autoUpdateService == nil {
		return fmt.Errorf("сервис автообновления не инициализирован")
	}

	u.autoUpdateService.UpdateAllLocks()
	log.Println("✅ Синхронизация всех замков завершена")

	return nil
}

func (u *lockUseCase) SyncLockWithTuya(uniqueID string) error {
	log.Printf("🔄 Синхронизация замка %s с Tuya API...", uniqueID)

	lock, err := u.lockRepo.GetByUniqueID(uniqueID)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	if u.autoUpdateService == nil {
		return fmt.Errorf("сервис автообновления не инициализирован")
	}

	if err := u.autoUpdateService.UpdateLock(lock); err != nil {
		return fmt.Errorf("ошибка синхронизации замка: %w", err)
	}

	log.Printf("✅ Замок %s успешно синхронизирован", uniqueID)
	return nil
}

func (u *lockUseCase) EnableAutoUpdate(uniqueID string) error {
	log.Printf("🔄 Включение автообновления для замка %s...", uniqueID)

	_, err := u.lockRepo.GetByUniqueID(uniqueID)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	if err := u.lockRepo.EnableAutoUpdate(uniqueID, true); err != nil {
		return fmt.Errorf("ошибка включения автообновления: %w", err)
	}

	log.Printf("✅ Автообновление включено для замка %s", uniqueID)
	return nil
}

func (u *lockUseCase) DisableAutoUpdate(uniqueID string) error {
	log.Printf("🔄 Отключение автообновления для замка %s...", uniqueID)

	_, err := u.lockRepo.GetByUniqueID(uniqueID)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	if err := u.lockRepo.EnableAutoUpdate(uniqueID, false); err != nil {
		return fmt.Errorf("ошибка отключения автообновления: %w", err)
	}

	log.Printf("✅ Автообновление отключено для замка %s", uniqueID)
	return nil
}

func (u *lockUseCase) ConfigureTuyaWebhooks(uniqueID string) error {
	log.Printf("🔄 Настройка Tuya webhooks для замка %s...", uniqueID)

	lock, err := u.lockRepo.GetByUniqueID(uniqueID)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	if lock.TuyaDeviceID == "" {
		return fmt.Errorf("у замка отсутствует TuyaDeviceID")
	}

	// Здесь можно добавить реальную настройку webhook'ов через Tuya API
	// Пока просто помечаем как настроенные
	if err := u.lockRepo.ConfigureWebhook(uniqueID, true); err != nil {
		return fmt.Errorf("ошибка настройки webhook: %w", err)
	}

	log.Printf("✅ Tuya webhooks настроены для замка %s", uniqueID)
	return nil
}

func (u *lockUseCase) UpdateOnlineStatus(uniqueID string, isOnline bool) error {
	log.Printf("🔄 Обновление онлайн статуса замка %s: %t", uniqueID, isOnline)

	if err := u.lockRepo.UpdateOnlineStatus(uniqueID, isOnline); err != nil {
		return fmt.Errorf("ошибка обновления онлайн статуса: %w", err)
	}

	log.Printf("✅ Онлайн статус замка %s обновлен: %t", uniqueID, isOnline)
	return nil
}

func (u *lockUseCase) UpdateTuyaSync(uniqueID string, syncTime time.Time) error {
	log.Printf("🔄 Обновление времени синхронизации Tuya для замка %s", uniqueID)

	if err := u.lockRepo.UpdateTuyaSync(uniqueID, syncTime); err != nil {
		return fmt.Errorf("ошибка обновления времени синхронизации: %w", err)
	}

	log.Printf("✅ Время синхронизации Tuya для замка %s обновлено", uniqueID)
	return nil
}

func (u *lockUseCase) AdminGeneratePassword(uniqueID string, request *domain.AdminGeneratePasswordRequest) (*domain.LockTempPassword, error) {
	lock, err := u.lockRepo.GetByUniqueID(uniqueID)
	if err != nil {
		return nil, fmt.Errorf("замок не найден: %w", err)
	}

	validFrom, err := time.Parse(time.RFC3339, request.ValidFrom)
	if err != nil {
		return nil, fmt.Errorf("неверный формат даты начала: %w", err)
	}

	validUntil, err := time.Parse(time.RFC3339, request.ValidUntil)
	if err != nil {
		return nil, fmt.Errorf("неверный формат даты окончания: %w", err)
	}

	if validUntil.Before(validFrom) {
		return nil, fmt.Errorf("дата окончания не может быть раньше даты начала")
	}

	if validUntil.Before(time.Now()) {
		return nil, fmt.Errorf("дата окончания не может быть в прошлом")
	}

	password, err := u.generateNumericPassword()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации пароля: %w", err)
	}

	tuyaPassword, tuyaPasswordID, err := u.tuyaService.GenerateTemporaryPasswordWithTimes(
		lock.TuyaDeviceID,
		request.Name,
		password,
		validFrom,
		validUntil,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пароля в Tuya: %w", err)
	}

	tempPassword := &domain.LockTempPassword{
		LockID:         lock.ID,
		BookingID:      nil,
		UserID:         request.UserID,
		Password:       tuyaPassword,
		TuyaPasswordID: tuyaPasswordID,
		Name:           request.Name,
		ValidFrom:      validFrom,
		ValidUntil:     validUntil,
		IsActive:       true,
	}

	err = u.lockRepo.CreateTempPassword(tempPassword)
	if err != nil {
		u.tuyaService.DeleteTempPassword(lock.TuyaDeviceID, tuyaPasswordID)
		return nil, fmt.Errorf("ошибка сохранения пароля в БД: %w", err)
	}

	statusLog := &domain.LockStatusLog{
		LockID:       lock.ID,
		OldStatus:    &lock.CurrentStatus,
		NewStatus:    lock.CurrentStatus,
		ChangeSource: domain.LockChangeSourceAPI,
		UserID:       request.UserID,
		BookingID:    nil,
		Notes:        fmt.Sprintf("Администратор создал временный пароль: %s", request.Name),
	}
	u.lockRepo.CreateStatusLog(statusLog)

	return tempPassword, nil
}

func (u *lockUseCase) AdminGetAllLockPasswords(uniqueID string) ([]*domain.LockTempPassword, error) {
	lock, err := u.lockRepo.GetByUniqueID(uniqueID)
	if err != nil {
		return nil, fmt.Errorf("замок не найден: %w", err)
	}

	return u.lockRepo.GetTempPasswordsByLockID(lock.ID)
}

func (u *lockUseCase) AdminDeactivatePassword(passwordID int) error {
	password, err := u.lockRepo.GetTempPasswordByID(passwordID)
	if err != nil {
		return fmt.Errorf("пароль не найден: %w", err)
	}

	if !password.IsActive {
		return fmt.Errorf("пароль уже деактивирован")
	}

	err = u.lockRepo.DeactivateTempPassword(passwordID)
	if err != nil {
		return fmt.Errorf("ошибка деактивации пароля в БД: %w", err)
	}

	lock, err := u.lockRepo.GetByID(password.LockID)
	if err != nil {
		return fmt.Errorf("замок не найден: %w", err)
	}

	err = u.tuyaService.DeleteTempPassword(lock.TuyaDeviceID, password.TuyaPasswordID)
	if err != nil {
		return fmt.Errorf("ошибка удаления пароля из Tuya API: %w", err)
	}

	return u.lockRepo.DeactivateTempPassword(passwordID)
}

func (u *lockUseCase) GeneratePasswordForBookingByID(bookingID, userID int) (string, error) {
	booking, err := u.bookingRepo.GetByID(bookingID)
	if err != nil {
		return "", fmt.Errorf("бронирование не найдено: %w", err)
	}

	canManage, err := u.CanUserManageLockViaBooking(booking, userID)
	if err != nil {
		return "", fmt.Errorf("ошибка проверки прав: %w", err)
	}
	if !canManage {
		return "", fmt.Errorf("у пользователя нет прав на управление замком через это бронирование")
	}

	lock, err := u.GetLockByApartmentID(booking.ApartmentID)
	if err != nil {
		return "", fmt.Errorf("замок не найден для квартиры: %w", err)
	}

	return u.GeneratePasswordForBooking(lock.UniqueID, userID, bookingID)
}
