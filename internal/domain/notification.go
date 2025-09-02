package domain

import (
	"time"
)

type NotificationType string

const (
	NotificationBookingApproved     NotificationType = "booking_approved"
	NotificationBookingRejected     NotificationType = "booking_rejected"
	NotificationBookingCanceled     NotificationType = "booking_canceled"
	NotificationBookingCompleted    NotificationType = "booking_completed"
	NotificationPasswordReady       NotificationType = "password_ready"
	NotificationExtensionRequest    NotificationType = "extension_request"
	NotificationExtensionApproved   NotificationType = "extension_approved"
	NotificationExtensionRejected   NotificationType = "extension_rejected"
	NotificationCheckoutReminder    NotificationType = "checkout_reminder"
	NotificationLockIssue           NotificationType = "lock_issue"
	NotificationNewBooking          NotificationType = "new_booking"
	NotificationSessionFinished     NotificationType = "session_finished"
	NotificationBookingStartingSoon NotificationType = "booking_starting_soon"
	NotificationBookingEnding       NotificationType = "booking_ending"
	NotificationPaymentRequired     NotificationType = "payment_required"

	NotificationApartmentCreated       NotificationType = "apartment_created"
	NotificationApartmentApproved      NotificationType = "apartment_approved"
	NotificationApartmentRejected      NotificationType = "apartment_rejected"
	NotificationApartmentUpdated       NotificationType = "apartment_updated"
	NotificationApartmentStatusChanged NotificationType = "apartment_status_changed"
)

type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

type DeviceType string

const (
	DeviceTypeIOS     DeviceType = "ios"
	DeviceTypeAndroid DeviceType = "android"
	DeviceTypeWeb     DeviceType = "web"
)

type Notification struct {
	ID          int                    `json:"id"`
	UserID      int                    `json:"user_id"`
	Type        NotificationType       `json:"type"`
	Priority    NotificationPriority   `json:"priority"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	IsRead      bool                   `json:"is_read"`
	IsPushed    bool                   `json:"is_pushed"`
	BookingID   *int                   `json:"booking_id,omitempty"`
	ApartmentID *int                   `json:"apartment_id,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
}

type UserDevice struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	DeviceToken string     `json:"device_token"`
	DeviceType  DeviceType `json:"device_type"`
	IsActive    bool       `json:"is_active"`
	AppVersion  string     `json:"app_version,omitempty"`
	OSVersion   string     `json:"os_version,omitempty"`
	LastSeenAt  time.Time  `json:"last_seen_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PushMessage struct {
	UserID           int                    `json:"user_id"`
	NotificationType NotificationType       `json:"notification_type"`
	Priority         NotificationPriority   `json:"priority"`
	Title            string                 `json:"title"`
	Body             string                 `json:"body"`
	Data             map[string]interface{} `json:"data,omitempty"`
	BookingID        *int                   `json:"booking_id,omitempty"`
	ApartmentID      *int                   `json:"apartment_id,omitempty"`
}

type CreateNotificationRequest struct {
	UserID      int                  `json:"user_id" validate:"required"`
	Type        NotificationType     `json:"type" validate:"required"`
	Priority    NotificationPriority `json:"priority"`
	Title       string               `json:"title" validate:"required"`
	Body        string               `json:"body" validate:"required"`
	Data        map[string]string    `json:"data,omitempty"`
	BookingID   *int                 `json:"booking_id,omitempty"`
	ApartmentID *int                 `json:"apartment_id,omitempty"`
	SendPush    bool                 `json:"send_push"`
}

type RegisterDeviceRequest struct {
	DeviceToken string     `json:"device_token" validate:"required"`
	DeviceType  DeviceType `json:"device_type" validate:"required"`
	AppVersion  string     `json:"app_version"`
	OSVersion   string     `json:"os_version"`
}

type NotificationRepository interface {
	CreateNotification(notification *Notification) error
	GetNotificationByID(id int) (*Notification, error)
	GetUserNotifications(userID int, limit, offset int) ([]*Notification, error)
	GetUnreadNotifications(userID int) ([]*Notification, error)
	MarkAsRead(notificationID, userID int) error
	MarkMultipleAsRead(notificationIDs []int, userID int) error
	MarkAllAsRead(userID int) error
	GetUnreadCount(userID int) (int, error)
	DeleteNotification(id int) error
	DeleteNotificationsByUser(notificationID, userID int) error
	DeleteReadNotifications(userID int) (int, error)
	DeleteAllNotifications(userID int) (int, error)

	RegisterDevice(device *UserDevice) error
	GetDevicesByUserID(userID int) ([]*UserDevice, error)
	GetDeviceByToken(token string) (*UserDevice, error)
	UpdateDevice(device *UserDevice) error
	DeactivateDevice(token string) error
	UpdateDeviceHeartbeat(token string) error
}

type NotificationUseCase interface {
	CreateNotification(notification *Notification) error
	CreateDelayedNotification(notification *Notification, delay time.Duration) error
	GetUserNotifications(userID int, limit, offset int) ([]*Notification, error)
	GetUnreadNotifications(userID int) ([]*Notification, error)
	GetUnreadCount(userID int) (int, error)
	MarkAsRead(notificationID, userID int) error
	MarkMultipleAsRead(notificationIDs []int, userID int) error
	MarkAllAsRead(userID int) error
	DeleteNotification(notificationID, userID int) error
	DeleteReadNotifications(userID int) (int, error)
	DeleteAllNotifications(userID int) (int, error)

	RegisterDevice(device *UserDevice) error
	UpdateDeviceHeartbeat(deviceToken string) error
	DeactivateDevice(deviceToken string) error
	GetDevicesByUserID(userID int) ([]*UserDevice, error)

	NotifyBookingApproved(userID int, bookingID int, apartmentTitle string) error
	NotifyBookingRejected(userID int, bookingID int, apartmentTitle, reason string) error
	NotifyBookingCanceled(userID int, bookingID int, apartmentTitle string, reason string) error
	NotifyBookingCompleted(userID int, bookingID int, apartmentTitle string) error
	NotifyPasswordReady(userID int, bookingID int, apartmentTitle string) error
	NotifyBookingStartingSoon(userID int, bookingID int, apartmentTitle string, startsIn time.Duration) error
	NotifyBookingEnding(userID int, bookingID int, apartmentTitle string) error
	NotifyPaymentRequired(userID int, bookingID int, apartmentTitle string, amount float64) error
	NotifySessionFinished(ownerUserID int, bookingID int, apartmentTitle string, renterName string) error
	NotifyExtensionRequested(ownerUserID int, bookingID int, apartmentTitle string, renterName string, duration int) error
	NotifyExtensionApproved(renterUserID int, bookingID int, apartmentTitle string, duration int) error
	NotifyExtensionRejected(renterUserID int, bookingID int, apartmentTitle string, duration int) error
	NotifyExtensionTimeoutRefund(renterUserID int, bookingID int, apartmentTitle string, duration int) error
	NotifyNewBookingRequest(ownerUserID int, bookingID int, apartmentTitle string, renterName string) error
	NotifyBookingStarted(ownerUserID int, bookingID int, apartmentTitle string, renterName string) error
	NotifyRenterBookingStarted(renterUserID int, bookingID int, apartmentTitle string) error

	NotifyApartmentCreated(ownerUserID int, apartmentID int, apartmentTitle string) error
	NotifyApartmentApproved(ownerUserID int, apartmentID int, apartmentTitle string) error
	NotifyApartmentRejected(ownerUserID int, apartmentID int, apartmentTitle string, reason string) error
	NotifyApartmentUpdated(ownerUserID int, apartmentID int, apartmentTitle string) error
	NotifyApartmentStatusChanged(ownerUserID int, apartmentID int, apartmentTitle string, oldStatus, newStatus string) error

	StartNotificationConsumer()
}

type PushNotificationService interface {
	SendPush(userID int, message *PushMessage) error
	SendPushToDevice(deviceToken string, deviceType DeviceType, message *PushMessage) error
	SendMulticast(userIDs []int, message *PushMessage) error
}

type MessageQueueService interface {
	PublishNotification(message *PushMessage) error
	PublishDelayedNotification(message *PushMessage, delay time.Duration) error
	ConsumeNotifications(handler func(*PushMessage) error) error
}

type NotificationResponse struct {
	ID          int                    `json:"id"`
	Type        NotificationType       `json:"type"`
	Priority    NotificationPriority   `json:"priority"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Data        map[string]interface{} `json:"data,omitempty"`
	IsRead      bool                   `json:"is_read"`
	BookingID   *int                   `json:"booking_id,omitempty"`
	ApartmentID *int                   `json:"apartment_id,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	ReadAt      *string                `json:"read_at,omitempty"`
}
