package usecase

import (
	"fmt"
	"log"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
)

type notificationUseCase struct {
	notificationRepo domain.NotificationRepository
	pushService      domain.PushNotificationService
	queueService     domain.MessageQueueService
}

func NewNotificationUseCase(
	notificationRepo domain.NotificationRepository,
	pushService domain.PushNotificationService,
	queueService domain.MessageQueueService,
) domain.NotificationUseCase {
	return &notificationUseCase{
		notificationRepo: notificationRepo,
		pushService:      pushService,
		queueService:     queueService,
	}
}

func (uc *notificationUseCase) CreateNotification(notification *domain.Notification) error {
	err := uc.notificationRepo.CreateNotification(notification)
	if err != nil {
		return fmt.Errorf("ошибка создания уведомления: %w", err)
	}

	pushMessage := &domain.PushMessage{
		UserID:           notification.UserID,
		Title:            notification.Title,
		Body:             notification.Message,
		NotificationType: notification.Type,
		Priority:         notification.Priority,
		Data: map[string]interface{}{
			"notification_id": notification.ID,
			"type":            string(notification.Type),
		},
	}

	err = uc.queueService.PublishNotification(pushMessage)
	if err != nil {
		log.Printf("❌ Ошибка добавления в очередь: %v", err)
		err = uc.pushService.SendPush(notification.UserID, pushMessage)
		if err != nil {
			log.Printf("❌ Ошибка прямой отправки push: %v", err)
		}
	}

	return nil
}

func (uc *notificationUseCase) CreateDelayedNotification(notification *domain.Notification, delay time.Duration) error {
	err := uc.notificationRepo.CreateNotification(notification)
	if err != nil {
		return fmt.Errorf("ошибка создания отложенного уведомления: %w", err)
	}

	pushMessage := &domain.PushMessage{
		UserID:           notification.UserID,
		Title:            notification.Title,
		Body:             notification.Message,
		NotificationType: notification.Type,
		Priority:         notification.Priority,
		Data: map[string]interface{}{
			"notification_id": notification.ID,
			"type":            string(notification.Type),
		},
	}

	return uc.queueService.PublishDelayedNotification(pushMessage, delay)
}

func (uc *notificationUseCase) GetUserNotifications(userID int, limit, offset int) ([]*domain.Notification, error) {
	return uc.notificationRepo.GetUserNotifications(userID, limit, offset)
}

func (uc *notificationUseCase) GetUnreadNotifications(userID int) ([]*domain.Notification, error) {
	return uc.notificationRepo.GetUnreadNotifications(userID)
}

func (uc *notificationUseCase) GetUnreadCount(userID int) (int, error) {
	return uc.notificationRepo.GetUnreadCount(userID)
}

func (uc *notificationUseCase) MarkAsRead(notificationID, userID int) error {
	return uc.notificationRepo.MarkAsRead(notificationID, userID)
}

func (uc *notificationUseCase) MarkMultipleAsRead(notificationIDs []int, userID int) error {
	return uc.notificationRepo.MarkMultipleAsRead(notificationIDs, userID)
}

func (uc *notificationUseCase) MarkAllAsRead(userID int) error {
	return uc.notificationRepo.MarkAllAsRead(userID)
}

func (uc *notificationUseCase) RegisterDevice(device *domain.UserDevice) error {
	return uc.notificationRepo.RegisterDevice(device)
}

func (uc *notificationUseCase) UpdateDeviceHeartbeat(deviceToken string) error {
	return uc.notificationRepo.UpdateDeviceHeartbeat(deviceToken)
}

func (uc *notificationUseCase) DeactivateDevice(deviceToken string) error {
	return uc.notificationRepo.DeactivateDevice(deviceToken)
}

func (uc *notificationUseCase) GetDevicesByUserID(userID int) ([]*domain.UserDevice, error) {
	return uc.notificationRepo.GetDevicesByUserID(userID)
}

func (uc *notificationUseCase) NotifyBookingApproved(userID int, bookingID int, apartmentTitle string) error {
	notification := &domain.Notification{
		UserID:    userID,
		Type:      domain.NotificationBookingApproved,
		Title:     "Бронирование одобрено!",
		Message:   fmt.Sprintf("Ваше бронирование квартиры '%s' было одобрено владельцем", apartmentTitle),
		Priority:  domain.NotificationPriorityHigh,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyBookingRejected(userID int, bookingID int, apartmentTitle, reason string) error {
	message := fmt.Sprintf("Ваше бронирование квартиры '%s' было отклонено", apartmentTitle)
	if reason != "" {
		message += fmt.Sprintf(". Причина: %s", reason)
	}

	notification := &domain.Notification{
		UserID:    userID,
		Type:      domain.NotificationBookingRejected,
		Title:     "Бронирование отклонено",
		Message:   message,
		Priority:  domain.NotificationPriorityNormal,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyPasswordReady(userID int, bookingID int, apartmentTitle string) error {
	notification := &domain.Notification{
		UserID:    userID,
		Type:      domain.NotificationPasswordReady,
		Title:     "Пароль для замка готов!",
		Message:   fmt.Sprintf("Временный пароль для квартиры '%s' создан и готов к использованию", apartmentTitle),
		Priority:  domain.NotificationPriorityHigh,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyBookingStartingSoon(userID int, bookingID int, apartmentTitle string, startsIn time.Duration) error {
	timeStr := "скоро"
	if startsIn.Hours() >= 24 {
		timeStr = fmt.Sprintf("через %d дней", int(startsIn.Hours()/24))
	} else if startsIn.Hours() >= 1 {
		timeStr = fmt.Sprintf("через %d часов", int(startsIn.Hours()))
	} else {
		timeStr = fmt.Sprintf("через %d минут", int(startsIn.Minutes()))
	}

	notification := &domain.Notification{
		UserID:    userID,
		Type:      domain.NotificationBookingStartingSoon,
		Title:     "Бронирование начинается скоро!",
		Message:   fmt.Sprintf("Ваше бронирование квартиры '%s' начинается %s", apartmentTitle, timeStr),
		Priority:  domain.NotificationPriorityNormal,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
		},
	}

	delay := startsIn - time.Hour
	if delay <= 0 {
		return uc.CreateNotification(notification)
	}

	return uc.CreateDelayedNotification(notification, delay)
}

func (uc *notificationUseCase) NotifyBookingEnding(userID int, bookingID int, apartmentTitle string) error {
	notification := &domain.Notification{
		UserID:    userID,
		Type:      domain.NotificationBookingEnding,
		Title:     "Бронирование заканчивается",
		Message:   fmt.Sprintf("Ваше бронирование квартиры '%s' заканчивается через час", apartmentTitle),
		Priority:  domain.NotificationPriorityNormal,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyPaymentRequired(userID int, bookingID int, apartmentTitle string, amount float64) error {
	notification := &domain.Notification{
		UserID:    userID,
		Type:      domain.NotificationPaymentRequired,
		Title:     "Требуется оплата",
		Message:   fmt.Sprintf("Для завершения бронирования квартиры '%s' необходимо оплатить %.0f тенге", apartmentTitle, amount),
		Priority:  domain.NotificationPriorityHigh,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
			"amount":     amount,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifySessionFinished(ownerUserID int, bookingID int, apartmentTitle string, renterName string) error {
	notification := &domain.Notification{
		UserID:    ownerUserID,
		Type:      domain.NotificationSessionFinished,
		Title:     "Сеанс завершен досрочно",
		Message:   fmt.Sprintf("Арендатор %s завершил сеанс в квартире '%s' раньше запланированного времени", renterName, apartmentTitle),
		Priority:  domain.NotificationPriorityNormal,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyExtensionRequested(ownerUserID int, bookingID int, apartmentTitle string, renterName string, duration int) error {
	notification := &domain.Notification{
		UserID:    ownerUserID,
		Type:      domain.NotificationExtensionRequest,
		Title:     "Запрос на продление",
		Message:   fmt.Sprintf("Арендатор %s запросил продление бронирования квартиры '%s' на %d часов", renterName, apartmentTitle, duration),
		Priority:  domain.NotificationPriorityNormal,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
			"duration":   duration,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyExtensionApproved(renterUserID int, bookingID int, apartmentTitle string, duration int) error {
	notification := &domain.Notification{
		UserID:    renterUserID,
		Type:      domain.NotificationExtensionApproved,
		Title:     "Продление одобрено",
		Message:   fmt.Sprintf("Ваш запрос на продление бронирования квартиры '%s' на %d часов был одобрен владельцем", apartmentTitle, duration),
		Priority:  domain.NotificationPriorityHigh,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
			"duration":   duration,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyExtensionRejected(renterUserID int, bookingID int, apartmentTitle string, duration int) error {
	notification := &domain.Notification{
		UserID:    renterUserID,
		Type:      domain.NotificationExtensionRejected,
		Title:     "Продление отклонено",
		Message:   fmt.Sprintf("Ваш запрос на продление бронирования квартиры '%s' на %d часов был отклонен владельцем", apartmentTitle, duration),
		Priority:  domain.NotificationPriorityNormal,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
			"duration":   duration,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyExtensionTimeoutRefund(renterUserID int, bookingID int, apartmentTitle string, duration int) error {
	notification := &domain.Notification{
		UserID:    renterUserID,
		Type:      domain.NotificationExtensionRejected,
		Title:     "Возврат за продление",
		Message:   fmt.Sprintf("Владелец не ответил на запрос продления квартиры '%s' на %d часов. Средства возвращены на ваш счет", apartmentTitle, duration),
		Priority:  domain.NotificationPriorityHigh,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
			"duration":   duration,
			"refund":     true,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyNewBookingRequest(ownerUserID int, bookingID int, apartmentTitle string, renterName string) error {
	notification := &domain.Notification{
		UserID:    ownerUserID,
		Type:      domain.NotificationNewBooking,
		Title:     "Новый запрос на бронирование",
		Message:   fmt.Sprintf("Пользователь %s хочет забронировать вашу квартиру '%s'", renterName, apartmentTitle),
		Priority:  domain.NotificationPriorityHigh,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyBookingStarted(ownerUserID int, bookingID int, apartmentTitle string, renterName string) error {
	notification := &domain.Notification{
		UserID:    ownerUserID,
		Type:      domain.NotificationBookingStartingSoon,
		Title:     "Аренда началась",
		Message:   fmt.Sprintf("Началась аренда вашей квартиры '%s' пользователем %s", apartmentTitle, renterName),
		Priority:  domain.NotificationPriorityNormal,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyRenterBookingStarted(renterUserID int, bookingID int, apartmentTitle string) error {
	notification := &domain.Notification{
		UserID:    renterUserID,
		Type:      domain.NotificationBookingStartingSoon,
		Title:     "Аренда началась",
		Message:   fmt.Sprintf("Ваша аренда квартиры '%s' началась. Добро пожаловать!", apartmentTitle),
		Priority:  domain.NotificationPriorityHigh,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) DeleteNotification(notificationID, userID int) error {
	return uc.notificationRepo.DeleteNotificationsByUser(notificationID, userID)
}

func (uc *notificationUseCase) DeleteReadNotifications(userID int) (int, error) {
	return uc.notificationRepo.DeleteReadNotifications(userID)
}

func (uc *notificationUseCase) DeleteAllNotifications(userID int) (int, error) {
	return uc.notificationRepo.DeleteAllNotifications(userID)
}

func (uc *notificationUseCase) StartNotificationConsumer() {
	log.Println("🚀 Запуск notification consumer...")

	handler := func(message *domain.PushMessage) error {
		return uc.pushService.SendPush(message.UserID, message)
	}

	go func() {
		err := uc.queueService.ConsumeNotifications(handler)
		if err != nil {
			log.Printf("❌ Ошибка notification consumer: %v", err)
		}
	}()
}

func (uc *notificationUseCase) NotifyBookingCanceled(userID int, bookingID int, apartmentTitle string, reason string) error {
	notification := &domain.Notification{
		UserID:    userID,
		Type:      domain.NotificationBookingCanceled,
		Title:     "Бронирование отменено",
		Message:   fmt.Sprintf("Ваше бронирование квартиры '%s' было отменено. Причина: %s", apartmentTitle, reason),
		Priority:  domain.NotificationPriorityNormal,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
			"reason":     reason,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyBookingCompleted(userID int, bookingID int, apartmentTitle string) error {
	notification := &domain.Notification{
		UserID:    userID,
		Type:      domain.NotificationBookingCompleted,
		Title:     "Бронирование завершено",
		Message:   fmt.Sprintf("Ваше бронирование квартиры '%s' успешно завершено. Спасибо за использование нашего сервиса!", apartmentTitle),
		Priority:  domain.NotificationPriorityNormal,
		IsRead:    false,
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"booking_id": bookingID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyApartmentCreated(ownerUserID int, apartmentID int, apartmentTitle string) error {
	notification := &domain.Notification{
		UserID:      ownerUserID,
		Type:        domain.NotificationApartmentCreated,
		Title:       "Квартира добавлена",
		Message:     fmt.Sprintf("Ваша квартира '%s' успешно добавлена и отправлена на модерацию", apartmentTitle),
		Priority:    domain.NotificationPriorityNormal,
		IsRead:      false,
		CreatedAt:   time.Now(),
		ApartmentID: &apartmentID,
		Data: map[string]interface{}{
			"apartment_id": apartmentID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyApartmentApproved(ownerUserID int, apartmentID int, apartmentTitle string) error {
	notification := &domain.Notification{
		UserID:      ownerUserID,
		Type:        domain.NotificationApartmentApproved,
		Title:       "Квартира одобрена!",
		Message:     fmt.Sprintf("Ваша квартира '%s' прошла модерацию и опубликована на платформе", apartmentTitle),
		Priority:    domain.NotificationPriorityHigh,
		IsRead:      false,
		CreatedAt:   time.Now(),
		ApartmentID: &apartmentID,
		Data: map[string]interface{}{
			"apartment_id": apartmentID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyApartmentRejected(ownerUserID int, apartmentID int, apartmentTitle string, reason string) error {
	message := fmt.Sprintf("Ваша квартира '%s' была отклонена модератором", apartmentTitle)
	if reason != "" {
		message += fmt.Sprintf(". Причина: %s", reason)
	}

	notification := &domain.Notification{
		UserID:      ownerUserID,
		Type:        domain.NotificationApartmentRejected,
		Title:       "Квартира отклонена",
		Message:     message,
		Priority:    domain.NotificationPriorityNormal,
		IsRead:      false,
		CreatedAt:   time.Now(),
		ApartmentID: &apartmentID,
		Data: map[string]interface{}{
			"apartment_id": apartmentID,
			"reason":       reason,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyApartmentUpdated(ownerUserID int, apartmentID int, apartmentTitle string) error {
	notification := &domain.Notification{
		UserID:      ownerUserID,
		Type:        domain.NotificationApartmentUpdated,
		Title:       "Квартира обновлена",
		Message:     fmt.Sprintf("Информация о квартире '%s' обновлена и отправлена на повторную модерацию", apartmentTitle),
		Priority:    domain.NotificationPriorityNormal,
		IsRead:      false,
		CreatedAt:   time.Now(),
		ApartmentID: &apartmentID,
		Data: map[string]interface{}{
			"apartment_id": apartmentID,
		},
	}

	return uc.CreateNotification(notification)
}

func (uc *notificationUseCase) NotifyApartmentStatusChanged(ownerUserID int, apartmentID int, apartmentTitle string, oldStatus, newStatus string) error {
	var title string
	var message string
	var priority domain.NotificationPriority

	switch newStatus {
	case "approved":
		title = "Квартира одобрена!"
		message = fmt.Sprintf("Ваша квартира '%s' одобрена и опубликована", apartmentTitle)
		priority = domain.NotificationPriorityHigh
	case "rejected":
		title = "Квартира отклонена"
		message = fmt.Sprintf("Ваша квартира '%s' была отклонена модератором", apartmentTitle)
		priority = domain.NotificationPriorityNormal
	case "blocked":
		title = "Квартира заблокирована"
		message = fmt.Sprintf("Ваша квартира '%s' была заблокирована администратором", apartmentTitle)
		priority = domain.NotificationPriorityHigh
	case "inactive":
		title = "Квартира деактивирована"
		message = fmt.Sprintf("Ваша квартира '%s' была деактивирована", apartmentTitle)
		priority = domain.NotificationPriorityNormal
	default:
		title = "Статус квартиры изменен"
		message = fmt.Sprintf("Статус вашей квартиры '%s' изменен с '%s' на '%s'", apartmentTitle, oldStatus, newStatus)
		priority = domain.NotificationPriorityNormal
	}

	notification := &domain.Notification{
		UserID:      ownerUserID,
		Type:        domain.NotificationApartmentStatusChanged,
		Title:       title,
		Message:     message,
		Priority:    priority,
		IsRead:      false,
		CreatedAt:   time.Now(),
		ApartmentID: &apartmentID,
		Data: map[string]interface{}{
			"apartment_id": apartmentID,
			"old_status":   oldStatus,
			"new_status":   newStatus,
		},
	}

	return uc.CreateNotification(notification)
}
