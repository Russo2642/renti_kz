package usecase

import (
	"fmt"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
)

type chatUseCase struct {
	chatRoomRepo        domain.ChatRoomRepository
	chatMessageRepo     domain.ChatMessageRepository
	chatParticipantRepo domain.ChatParticipantRepository
	bookingRepo         domain.BookingRepository
	conciergeRepo       domain.ConciergeRepository
	renterRepo          domain.RenterRepository
	wsService           domain.ChatWebSocketService
	notificationRepo    domain.NotificationRepository
}

func NewChatUseCase(
	chatRoomRepo domain.ChatRoomRepository,
	chatMessageRepo domain.ChatMessageRepository,
	chatParticipantRepo domain.ChatParticipantRepository,
	bookingRepo domain.BookingRepository,
	conciergeRepo domain.ConciergeRepository,
	renterRepo domain.RenterRepository,
	wsService domain.ChatWebSocketService,
	notificationRepo domain.NotificationRepository,
) domain.ChatUseCase {
	return &chatUseCase{
		chatRoomRepo:        chatRoomRepo,
		chatMessageRepo:     chatMessageRepo,
		chatParticipantRepo: chatParticipantRepo,
		bookingRepo:         bookingRepo,
		conciergeRepo:       conciergeRepo,
		renterRepo:          renterRepo,
		wsService:           wsService,
		notificationRepo:    notificationRepo,
	}
}

func (uc *chatUseCase) CreateChatRoom(request *domain.CreateChatRoomRequest, userID int) (*domain.ChatRoom, error) {
	booking, err := uc.bookingRepo.GetByID(request.BookingID)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	if err := uc.validateUserBookingAccess(booking, userID); err != nil {
		return nil, err
	}

	existingRoom, err := uc.chatRoomRepo.GetByBookingID(request.BookingID)
	if err == nil && existingRoom != nil {
		return existingRoom, nil
	}

	concierges, err := uc.conciergeRepo.GetByApartmentIDActive(booking.ApartmentID)
	if err != nil {
		return nil, fmt.Errorf("no active concierge for apartment: %w", err)
	}
	if len(concierges) == 0 {
		return nil, fmt.Errorf("no active concierge found for apartment %d", booking.ApartmentID)
	}
	concierge := concierges[0]

	renter, err := uc.renterRepo.GetByID(booking.RenterID)
	if err != nil {
		return nil, fmt.Errorf("renter not found: %w", err)
	}

	room := &domain.ChatRoom{
		BookingID:   booking.ID,
		ConciergeID: concierge.ID,
		RenterID:    renter.ID,
		ApartmentID: booking.ApartmentID,
		Status:      domain.ChatRoomStatusPending,
	}

	err = uc.chatRoomRepo.Create(room)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}

	return room, nil
}

func (uc *chatUseCase) GetChatRoom(roomID, userID int) (*domain.ChatRoom, error) {
	canAccess, err := uc.CanUserAccessRoom(roomID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, fmt.Errorf("access denied to chat room")
	}

	room, err := uc.chatRoomRepo.GetByID(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat room: %w", err)
	}

	return room, nil
}

func (uc *chatUseCase) GetChatRoomByBookingID(bookingID int) (*domain.ChatRoom, error) {
	room, err := uc.chatRoomRepo.GetByBookingID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat room by booking ID: %w", err)
	}

	return room, nil
}

func (uc *chatUseCase) GetUserChatRooms(userID int, status []domain.ChatRoomStatus, page, pageSize int) ([]*domain.ChatRoom, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	rooms, total, err := uc.chatRoomRepo.GetByUserID(userID, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user chat rooms: %w", err)
	}

	return rooms, total, nil
}

func (uc *chatUseCase) GetConciergeChatRooms(conciergeID int, status []domain.ChatRoomStatus, page, pageSize int) ([]*domain.ChatRoom, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	rooms, total, err := uc.chatRoomRepo.GetByConciergeID(conciergeID, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get concierge chat rooms: %w", err)
	}

	return rooms, total, nil
}

func (uc *chatUseCase) CanUserAccessRoom(roomID, userID int) (bool, error) {
	canAccess, err := uc.chatRoomRepo.CanUserAccessRoom(roomID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check room access: %w", err)
	}

	return canAccess, nil
}

func (uc *chatUseCase) SendMessage(roomID int, request *domain.SendMessageRequest, userID int) (*domain.ChatMessage, error) {
	canAccess, err := uc.CanUserAccessRoom(roomID, userID)
	if err != nil || !canAccess {
		return nil, fmt.Errorf("access denied to chat room")
	}

	room, err := uc.chatRoomRepo.GetByID(roomID)
	if err != nil {
		return nil, fmt.Errorf("chat room not found: %w", err)
	}

	if room.Status != domain.ChatRoomStatusActive {
		return nil, fmt.Errorf("chat room is not active")
	}

	message := &domain.ChatMessage{
		ChatRoomID: roomID,
		SenderID:   userID,
		Type:       request.Type,
		Content:    request.Content,
		Status:     domain.MessageStatusSent,
		ReplyToID:  request.ReplyToID,
	}

	err = uc.chatMessageRepo.Create(message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	wsMessage := &domain.WSMessage{
		Type:      domain.WSEventNewMessage,
		Data:      message,
		Timestamp: time.Now(),
		UserID:    userID,
		RoomID:    roomID,
	}

	uc.wsService.BroadcastToRoom(roomID, wsMessage)
	go uc.NotifyNewMessage(message)

	return message, nil
}

func (uc *chatUseCase) GetMessages(roomID, userID int, page, pageSize int) ([]*domain.ChatMessage, int, error) {
	canAccess, err := uc.CanUserAccessRoom(roomID, userID)
	if err != nil || !canAccess {
		return nil, 0, fmt.Errorf("access denied to chat room")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}

	messages, total, err := uc.chatMessageRepo.GetByRoomID(roomID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, total, nil
}

func (uc *chatUseCase) UpdateMessage(messageID int, request *domain.UpdateMessageRequest, userID int) error {
	message, err := uc.chatMessageRepo.GetByID(messageID)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	if message.SenderID != userID {
		return fmt.Errorf("access denied: not message sender")
	}

	if time.Since(message.CreatedAt) > 15*time.Minute {
		return fmt.Errorf("message too old to edit")
	}

	message.Content = request.Content
	err = uc.chatMessageRepo.Update(message)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

func (uc *chatUseCase) DeleteMessage(messageID, userID int) error {
	message, err := uc.chatMessageRepo.GetByID(messageID)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	if message.SenderID != userID {
		return fmt.Errorf("access denied: not message sender")
	}

	err = uc.chatMessageRepo.Delete(messageID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

func (uc *chatUseCase) MarkMessagesAsRead(roomID int, request *domain.MarkMessagesReadRequest, userID int) error {
	canAccess, err := uc.CanUserAccessRoom(roomID, userID)
	if err != nil || !canAccess {
		return fmt.Errorf("access denied to chat room")
	}

	err = uc.chatMessageRepo.MarkAsRead(request.MessageIDs, userID)
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	return nil
}

func (uc *chatUseCase) GetUnreadCount(roomID, userID int) (int, error) {
	canAccess, err := uc.CanUserAccessRoom(roomID, userID)
	if err != nil || !canAccess {
		return 0, fmt.Errorf("access denied to chat room")
	}

	count, err := uc.chatMessageRepo.GetUnreadCount(roomID, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

func (uc *chatUseCase) OpenScheduledChats() error {
	rooms, err := uc.chatRoomRepo.GetRoomsPendingOpen()
	if err != nil {
		return fmt.Errorf("failed to get rooms pending open: %w", err)
	}

	for _, room := range rooms {
		room.Status = domain.ChatRoomStatusActive
		now := time.Now()
		room.OpenedAt = &now

		err = uc.chatRoomRepo.Update(room)
		if err != nil {
			continue
		}

		uc.CreateWelcomeMessage(room.ID)
	}

	return nil
}

func (uc *chatUseCase) CloseExpiredChats() error {
	rooms, err := uc.chatRoomRepo.GetRoomsPendingClose()
	if err != nil {
		return fmt.Errorf("failed to get rooms pending close: %w", err)
	}

	for _, room := range rooms {
		room.Status = domain.ChatRoomStatusClosed
		now := time.Now()
		room.ClosedAt = &now

		uc.chatRoomRepo.Update(room)
	}

	return nil
}

func (uc *chatUseCase) CreateWelcomeMessage(roomID int) error {
	message := &domain.ChatMessage{
		ChatRoomID: roomID,
		SenderID:   0,
		Type:       domain.MessageTypeWelcome,
		Content:    "Добро пожаловать в чат с консьержем!",
		Status:     domain.MessageStatusSent,
	}

	return uc.chatMessageRepo.Create(message)
}

func (uc *chatUseCase) NotifyNewMessage(message *domain.ChatMessage) error {
	room, err := uc.chatRoomRepo.GetByID(message.ChatRoomID)
	if err != nil {
		return err
	}

	var recipientUserID int
	var senderTitle string

	if room.Concierge != nil && room.Concierge.UserID == message.SenderID {
		if room.Renter != nil {
			recipientUserID = room.Renter.UserID
			senderTitle = "Консьерж"
		}
	} else if room.Renter != nil && room.Renter.UserID == message.SenderID {
		if room.Concierge != nil {
			recipientUserID = room.Concierge.UserID
			senderTitle = "Арендатор"
		}
	}

	if recipientUserID > 0 {
		notification := &domain.Notification{
			UserID:  recipientUserID,
			Type:    "chat_message",
			Title:   fmt.Sprintf("Новое сообщение от: %s", senderTitle),
			Message: message.Content,
		}

		return uc.notificationRepo.CreateNotification(notification)
	}

	return nil
}

func (uc *chatUseCase) validateUserBookingAccess(booking *domain.Booking, userID int) error {
	renter, err := uc.renterRepo.GetByID(booking.RenterID)
	if err == nil && renter.UserID == userID {
		return nil
	}

	concierge, err := uc.conciergeRepo.GetByUserID(userID)
	if err == nil && concierge != nil {
		for _, apartment := range concierge.Apartments {
			if apartment.ID == booking.ApartmentID {
				return nil
			}
		}
	}

	return fmt.Errorf("access denied: user is not the renter or concierge of this booking")
}

func (uc *chatUseCase) ActivateChat(roomID, userID int) error {
	canAccess, err := uc.CanUserAccessRoom(roomID, userID)
	if err != nil || !canAccess {
		return fmt.Errorf("access denied to chat room")
	}

	room, err := uc.chatRoomRepo.GetByID(roomID)
	if err != nil {
		return fmt.Errorf("chat room not found: %w", err)
	}

	if room.Status != domain.ChatRoomStatusPending {
		return fmt.Errorf("can only activate pending chat rooms")
	}

	now := time.Now()
	room.Status = domain.ChatRoomStatusActive
	room.OpenedAt = &now

	err = uc.chatRoomRepo.Update(room)
	if err != nil {
		return fmt.Errorf("failed to activate chat room: %w", err)
	}

	uc.CreateWelcomeMessage(roomID)

	return nil
}

func (uc *chatUseCase) CloseChat(roomID, userID int) error {
	canAccess, err := uc.CanUserAccessRoom(roomID, userID)
	if err != nil || !canAccess {
		return fmt.Errorf("access denied to chat room")
	}

	room, err := uc.chatRoomRepo.GetByID(roomID)
	if err != nil {
		return fmt.Errorf("chat room not found: %w", err)
	}

	if room.Status == domain.ChatRoomStatusClosed {
		return fmt.Errorf("chat room is already closed")
	}

	now := time.Now()
	room.Status = domain.ChatRoomStatusClosed
	room.ClosedAt = &now

	err = uc.chatRoomRepo.Update(room)
	if err != nil {
		return fmt.Errorf("failed to close chat room: %w", err)
	}

	return nil
}
