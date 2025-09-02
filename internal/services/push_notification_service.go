package services

import (
	"context"
	"fmt"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"github.com/russo2642/renti_kz/internal/config"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/pkg/logger"
)

type PushNotificationService struct {
	client             *messaging.Client
	notificationRepo   domain.NotificationRepository
	enablePushMessages bool
}

func NewPushNotificationService(
	firebaseConfig config.FirebaseConfig,
	notificationRepo domain.NotificationRepository,
) (domain.PushNotificationService, error) {

	service := &PushNotificationService{
		notificationRepo:   notificationRepo,
		enablePushMessages: true,
	}

	if firebaseConfig.CredentialsFile == "" {
		logger.Warn("Firebase credentials not provided, push notifications disabled")
		service.enablePushMessages = false
		return service, nil
	}

	opt := option.WithCredentialsFile(firebaseConfig.CredentialsFile)

	firebaseAppConfig := &firebase.Config{
		ProjectID: firebaseConfig.ProjectID,
	}

	app, err := firebase.NewApp(context.Background(), firebaseAppConfig, opt)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации Firebase: %w", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ошибка создания messaging client: %w", err)
	}

	service.client = client
	logger.Info("Firebase Cloud Messaging initialized successfully")

	return service, nil
}

func (s *PushNotificationService) SendPush(userID int, message *domain.PushMessage) error {
	if !s.enablePushMessages {
		logger.Debug("push notifications disabled for user", slog.Int("user_id", userID))
		return nil
	}

	devices, err := s.notificationRepo.GetDevicesByUserID(userID)
	if err != nil {
		return fmt.Errorf("ошибка получения устройств пользователя %d: %w", userID, err)
	}

	if len(devices) == 0 {
		logger.Debug("no active devices for user", slog.Int("user_id", userID))
		return nil
	}

	var successCount, errorCount int

	for _, device := range devices {
		err := s.SendPushToDevice(device.DeviceToken, device.DeviceType, message)
		if err != nil {
			logger.Error("failed to send push notification",
				slog.String("device_token", device.DeviceToken[:10]+"..."),
				slog.String("error", err.Error()))
			errorCount++
		} else {
			successCount++
		}
	}

	logger.Info("push notifications sent",
		slog.Int("user_id", userID),
		slog.Int("success_count", successCount),
		slog.Int("error_count", errorCount))
	return nil
}

func (s *PushNotificationService) SendPushToDevice(deviceToken string, deviceType domain.DeviceType, message *domain.PushMessage) error {
	if !s.enablePushMessages {
		return nil
	}

	fcmMessage := s.buildFCMMessage(deviceToken, deviceType, message)

	response, err := s.client.Send(context.Background(), fcmMessage)
	if err != nil {
		if messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err) {
			logger.Info("deactivating invalid device token", slog.String("token", deviceToken[:10]+"..."))
			s.notificationRepo.DeactivateDevice(deviceToken)
		}
		return fmt.Errorf("ошибка отправки FCM: %w", err)
	}

	logger.Debug("push notification sent successfully", slog.String("response", response))
	return nil
}

func (s *PushNotificationService) SendMulticast(userIDs []int, message *domain.PushMessage) error {
	if !s.enablePushMessages {
		return nil
	}

	if len(userIDs) == 0 {
		return nil
	}

	var allTokens []string
	var tokenToDevice = make(map[string]*domain.UserDevice)

	for _, userID := range userIDs {
		devices, err := s.notificationRepo.GetDevicesByUserID(userID)
		if err != nil {
			logger.Error("failed to get user devices", slog.Int("user_id", userID), slog.String("error", err.Error()))
			continue
		}

		for _, device := range devices {
			allTokens = append(allTokens, device.DeviceToken)
			tokenToDevice[device.DeviceToken] = device
		}
	}

	if len(allTokens) == 0 {
		logger.Debug("no active devices for multicast push")
		return nil
	}

	batchSize := 500
	for i := 0; i < len(allTokens); i += batchSize {
		end := i + batchSize
		if end > len(allTokens) {
			end = len(allTokens)
		}

		batch := allTokens[i:end]
		err := s.sendMulticastBatch(batch, tokenToDevice, message)
		if err != nil {
			logger.Error("failed to send push batch", slog.Int("start", i), slog.Int("end", end), slog.String("error", err.Error()))
		}
	}

	return nil
}

func (s *PushNotificationService) sendMulticastBatch(tokens []string, tokenToDevice map[string]*domain.UserDevice, message *domain.PushMessage) error {
	if len(tokens) == 0 {
		return nil
	}

	deviceGroups := s.groupTokensByDeviceType(tokens, tokenToDevice)

	for deviceType, deviceTokens := range deviceGroups {
		logger.Debug("grouped devices for multicast", slog.String("device_type", string(deviceType)), slog.Int("token_count", len(deviceTokens)))
	}

	for deviceType, deviceTokens := range deviceGroups {
		if len(deviceTokens) == 0 {
			continue
		}

		err := s.sendMulticastByDeviceType(deviceTokens, deviceType, message)
		if err != nil {
			logger.Error("failed to send multicast", slog.String("device_type", string(deviceType)), slog.String("error", err.Error()))
		}
	}

	return nil
}

func (s *PushNotificationService) groupTokensByDeviceType(tokens []string, tokenToDevice map[string]*domain.UserDevice) map[domain.DeviceType][]string {
	groups := make(map[domain.DeviceType][]string)
	skippedTokens := 0

	for _, token := range tokens {
		device, exists := tokenToDevice[token]
		if !exists {
			logger.Warn("device not found for token", slog.String("token", token[:10]+"..."))
			skippedTokens++
			continue
		}

		if !device.IsActive {
			logger.Debug("skipping inactive device", slog.String("token", token[:10]+"..."))
			skippedTokens++
			continue
		}

		groups[device.DeviceType] = append(groups[device.DeviceType], token)
	}

	if skippedTokens > 0 {
		logger.Debug("skipped tokens", slog.Int("skipped", skippedTokens), slog.Int("total", len(tokens)))
	}

	return groups
}

func (s *PushNotificationService) sendMulticastByDeviceType(tokens []string, deviceType domain.DeviceType, message *domain.PushMessage) error {
	multicastMessage := s.buildMulticastMessage(tokens, deviceType, message)

	response, err := s.client.SendEachForMulticast(context.Background(), multicastMessage)
	if err != nil {
		return fmt.Errorf("ошибка multicast FCM для %s: %w", deviceType, err)
	}

	logger.Info("multicast sent",
		slog.String("device_type", string(deviceType)),
		slog.Int("success_count", int(response.SuccessCount)),
		slog.Int("failure_count", int(response.FailureCount)))

	for i, resp := range response.Responses {
		if !resp.Success && i < len(tokens) {
			token := tokens[i]
			if messaging.IsUnregistered(resp.Error) || messaging.IsInvalidArgument(resp.Error) {
				logger.Info("deactivating invalid token",
					slog.String("device_type", string(deviceType)),
					slog.String("token", token[:10]+"..."))
				s.notificationRepo.DeactivateDevice(token)
			} else {
				logger.Error("failed to send to device",
					slog.String("device_type", string(deviceType)),
					slog.String("error", resp.Error.Error()))
			}
		}
	}

	return nil
}

func (s *PushNotificationService) buildFCMMessage(token string, deviceType domain.DeviceType, message *domain.PushMessage) *messaging.Message {
	fcmMessage := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: message.Title,
			Body:  message.Body,
		},
		Data: s.buildDataPayload(message),
	}

	switch deviceType {
	case domain.DeviceTypeIOS:
		fcmMessage.APNS = &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: message.Title,
						Body:  message.Body,
					},
					Badge:            s.getBadgeCount(message.UserID),
					Sound:            "default",
					ContentAvailable: true,
				},
			},
		}
	case domain.DeviceTypeAndroid:
		fcmMessage.Android = &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Title:     message.Title,
				Body:      message.Body,
				Icon:      "ic_notification",
				Color:     "#2196F3",
				Sound:     "default",
				ChannelID: s.getChannelID(message.NotificationType),
				Priority:  s.getAndroidPriority(message.Priority),
			},
			Priority: s.getAndroidMessagePriority(message.Priority),
		}
	case domain.DeviceTypeWeb:
	}

	return fcmMessage
}

func (s *PushNotificationService) buildMulticastMessage(tokens []string, deviceType domain.DeviceType, message *domain.PushMessage) *messaging.MulticastMessage {
	multicastMessage := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: message.Title,
			Body:  message.Body,
		},
		Data: s.buildDataPayload(message),
	}

	switch deviceType {
	case domain.DeviceTypeIOS:
		multicastMessage.APNS = &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: message.Title,
						Body:  message.Body,
					},
					Sound:            "default",
					ContentAvailable: true,
				},
			},
		}
	case domain.DeviceTypeAndroid:
		multicastMessage.Android = &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Title:     message.Title,
				Body:      message.Body,
				Icon:      "ic_notification",
				Color:     "#2196F3",
				Sound:     "default",
				ChannelID: s.getChannelID(message.NotificationType),
				Priority:  s.getAndroidPriority(message.Priority),
			},
			Priority: s.getAndroidMessagePriority(message.Priority),
		}
	case domain.DeviceTypeWeb:
	}

	return multicastMessage
}

func (s *PushNotificationService) buildDataPayload(message *domain.PushMessage) map[string]string {
	data := map[string]string{
		"type":     string(message.NotificationType),
		"priority": string(message.Priority),
		"user_id":  fmt.Sprintf("%d", message.UserID),
	}

	if message.BookingID != nil {
		data["booking_id"] = fmt.Sprintf("%d", *message.BookingID)
	}

	if message.ApartmentID != nil {
		data["apartment_id"] = fmt.Sprintf("%d", *message.ApartmentID)
	}

	if message.Data != nil {
		for k, v := range message.Data {
			if s, ok := v.(string); ok {
				data[k] = s
			} else {
				data[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	return data
}

func (s *PushNotificationService) getBadgeCount(userID int) *int {
	count, err := s.notificationRepo.GetUnreadCount(userID)
	if err != nil {
		return nil
	}
	return &count
}

func (s *PushNotificationService) getChannelID(notificationType domain.NotificationType) string {
	switch notificationType {
	case domain.NotificationBookingApproved, domain.NotificationBookingRejected:
		return "booking_updates"
	case domain.NotificationPasswordReady:
		return "access_codes"
	case domain.NotificationCheckoutReminder:
		return "reminders"
	case domain.NotificationLockIssue:
		return "urgent"
	default:
		return "general"
	}
}

func (s *PushNotificationService) getAndroidPriority(priority domain.NotificationPriority) messaging.AndroidNotificationPriority {
	switch priority {
	case domain.NotificationPriorityUrgent:
		return messaging.PriorityMax
	case domain.NotificationPriorityHigh:
		return messaging.PriorityHigh
	case domain.NotificationPriorityNormal:
		return messaging.PriorityDefault
	case domain.NotificationPriorityLow:
		return messaging.PriorityMin
	default:
		return messaging.PriorityDefault
	}
}

func (s *PushNotificationService) getAndroidMessagePriority(priority domain.NotificationPriority) string {
	switch priority {
	case domain.NotificationPriorityUrgent, domain.NotificationPriorityHigh:
		return "high"
	default:
		return "normal"
	}
}
