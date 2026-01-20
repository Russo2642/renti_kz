package services

import (
	"fmt"
	"log/slog"

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/pkg/logger"
)

type PushNotificationService struct {
	client             *expo.PushClient
	notificationRepo   domain.NotificationRepository
	enablePushMessages bool
}

func NewPushNotificationService(
	notificationRepo domain.NotificationRepository,
) domain.PushNotificationService {

	service := &PushNotificationService{
		notificationRepo:   notificationRepo,
		enablePushMessages: true,
		client:             expo.NewPushClient(nil),
	}

	logger.Info("Expo Push Notification Service initialized successfully")

	return service
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
				slog.String("device_token", maskToken(device.DeviceToken)),
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

	pushToken, err := expo.NewExponentPushToken(deviceToken)
	if err != nil {
		logger.Error("invalid expo push token",
			slog.String("token", maskToken(deviceToken)),
			slog.String("error", err.Error()))

		s.notificationRepo.DeactivateDevice(deviceToken)
		return fmt.Errorf("невалидный expo push token: %w", err)
	}

	expoPushMessage := s.buildExpoPushMessage(pushToken, deviceType, message)

	response, err := s.client.Publish(expoPushMessage)
	if err != nil {
		return fmt.Errorf("ошибка отправки push через Expo: %w", err)
	}

	if err := response.ValidateResponse(); err != nil {
		logger.Error("expo push failed",
			slog.String("token", maskToken(deviceToken)),
			slog.String("error", err.Error()))

		if isTokenError(err) {
			logger.Info("deactivating invalid expo token", slog.String("token", maskToken(deviceToken)))
			s.notificationRepo.DeactivateDevice(deviceToken)
		}

		return fmt.Errorf("expo push validation failed: %w", err)
	}

	logger.Debug("push notification sent successfully via Expo",
		slog.String("token", maskToken(deviceToken)))
	return nil
}

func (s *PushNotificationService) SendMulticast(userIDs []int, message *domain.PushMessage) error {
	if !s.enablePushMessages {
		return nil
	}

	if len(userIDs) == 0 {
		return nil
	}

	var allTokens []expo.ExponentPushToken
	var tokenToDevice = make(map[string]*domain.UserDevice)

	for _, userID := range userIDs {
		devices, err := s.notificationRepo.GetDevicesByUserID(userID)
		if err != nil {
			logger.Error("failed to get user devices", slog.Int("user_id", userID), slog.String("error", err.Error()))
			continue
		}

		for _, device := range devices {
			pushToken, err := expo.NewExponentPushToken(device.DeviceToken)
			if err != nil {
				logger.Warn("skipping invalid token",
					slog.String("token", maskToken(device.DeviceToken)),
					slog.String("error", err.Error()))
				continue
			}

			allTokens = append(allTokens, pushToken)
			tokenToDevice[device.DeviceToken] = device
		}
	}

	if len(allTokens) == 0 {
		logger.Debug("no active devices for multicast push")
		return nil
	}

	batchSize := 100
	for i := 0; i < len(allTokens); i += batchSize {
		end := i + batchSize
		if end > len(allTokens) {
			end = len(allTokens)
		}

		batch := allTokens[i:end]
		err := s.sendMulticastBatch(batch, message)
		if err != nil {
			logger.Error("failed to send push batch", slog.Int("start", i), slog.Int("end", end), slog.String("error", err.Error()))
		}
	}

	return nil
}

func (s *PushNotificationService) sendMulticastBatch(tokens []expo.ExponentPushToken, message *domain.PushMessage) error {
	if len(tokens) == 0 {
		return nil
	}

	expoPushMessage := &expo.PushMessage{
		To:       tokens,
		Title:    message.Title,
		Body:     message.Body,
		Data:     s.buildDataPayload(message),
		Sound:    "default",
		Priority: s.getExpoPriority(message.Priority),
	}

	if badge := s.getBadgeCount(message.UserID); badge != nil {
		expoPushMessage.Badge = *badge
	}

	if channelID := s.getChannelID(message.NotificationType); channelID != "" {
		expoPushMessage.ChannelID = channelID
	}

	response, err := s.client.Publish(expoPushMessage)
	if err != nil {
		return fmt.Errorf("ошибка multicast отправки через Expo: %w", err)
	}

	if err := response.ValidateResponse(); err != nil {
		logger.Error("multicast push validation failed", slog.String("error", err.Error()))
		return fmt.Errorf("multicast validation failed: %w", err)
	}

	logger.Info("multicast push sent successfully",
		slog.Int("token_count", len(tokens)))

	return nil
}

func (s *PushNotificationService) buildExpoPushMessage(token expo.ExponentPushToken, deviceType domain.DeviceType, message *domain.PushMessage) *expo.PushMessage {
	expoPushMessage := &expo.PushMessage{
		To:       []expo.ExponentPushToken{token},
		Title:    message.Title,
		Body:     message.Body,
		Data:     s.buildDataPayload(message),
		Sound:    "default",
		Priority: s.getExpoPriority(message.Priority),
	}

	if badge := s.getBadgeCount(message.UserID); badge != nil {
		expoPushMessage.Badge = *badge
	}

	if deviceType == domain.DeviceTypeAndroid {
		if channelID := s.getChannelID(message.NotificationType); channelID != "" {
			expoPushMessage.ChannelID = channelID
		}
	}

	return expoPushMessage
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

func (s *PushNotificationService) getExpoPriority(priority domain.NotificationPriority) string {
	switch priority {
	case domain.NotificationPriorityUrgent, domain.NotificationPriorityHigh:
		return "high"
	case domain.NotificationPriorityNormal:
		return "default"
	case domain.NotificationPriorityLow:
		return "default"
	default:
		return "default"
	}
}


func maskToken(token string) string {
	if len(token) <= 20 {
		return "***"
	}
	return token[:10] + "..." + token[len(token)-5:]
}

func isTokenError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()
	return contains(errorStr, "DeviceNotRegistered") ||
		contains(errorStr, "InvalidCredentials") ||
		contains(errorStr, "MessageTooBig") ||
		contains(errorStr, "MessageRateExceeded")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)*2 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
