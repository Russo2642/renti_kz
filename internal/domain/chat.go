package domain

import (
	"time"
)

type ChatRoomStatus string

const (
	ChatRoomStatusPending  ChatRoomStatus = "pending"
	ChatRoomStatusActive   ChatRoomStatus = "active"
	ChatRoomStatusClosed   ChatRoomStatus = "closed"
	ChatRoomStatusArchived ChatRoomStatus = "archived"
)

type MessageType string

const (
	MessageTypeText    MessageType = "text"
	MessageTypeImage   MessageType = "image"
	MessageTypeFile    MessageType = "file"
	MessageTypeSystem  MessageType = "system"
	MessageTypeWelcome MessageType = "welcome"
	MessageTypeStatus  MessageType = "status"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

type ChatRoom struct {
	ID            int            `json:"id"`
	BookingID     int            `json:"booking_id"`
	Booking       *Booking       `json:"booking,omitempty"`
	ConciergeID   int            `json:"concierge_id"`
	Concierge     *Concierge     `json:"concierge,omitempty"`
	RenterID      int            `json:"renter_id"`
	Renter        *Renter        `json:"renter,omitempty"`
	ApartmentID   int            `json:"apartment_id"`
	Apartment     *Apartment     `json:"apartment,omitempty"`
	Status        ChatRoomStatus `json:"status"`
	OpenedAt      *time.Time     `json:"opened_at"`
	ClosedAt      *time.Time     `json:"closed_at"`
	LastMessageAt *time.Time     `json:"last_message_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type ChatMessage struct {
	ID         int           `json:"id"`
	ChatRoomID int           `json:"chat_room_id"`
	ChatRoom   *ChatRoom     `json:"chat_room,omitempty"`
	SenderID   int           `json:"sender_id"`
	Sender     *User         `json:"sender,omitempty"`
	Type       MessageType   `json:"type"`
	Content    string        `json:"content"`
	FileURL    *string       `json:"file_url,omitempty"`
	FileName   *string       `json:"file_name,omitempty"`
	FileSize   *int64        `json:"file_size,omitempty"`
	Status     MessageStatus `json:"status"`
	ReadAt     *time.Time    `json:"read_at"`
	ReplyToID  *int          `json:"reply_to_id,omitempty"`
	ReplyTo    *ChatMessage  `json:"reply_to,omitempty"`
	EditedAt   *time.Time    `json:"edited_at"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type ChatParticipant struct {
	ID         int                 `json:"id"`
	ChatRoomID int                 `json:"chat_room_id"`
	UserID     int                 `json:"user_id"`
	User       *User               `json:"user,omitempty"`
	Role       ChatParticipantRole `json:"role"`
	JoinedAt   time.Time           `json:"joined_at"`
	LeftAt     *time.Time          `json:"left_at"`
	LastReadAt *time.Time          `json:"last_read_at"`
	IsOnline   bool                `json:"is_online"`
	LastSeenAt *time.Time          `json:"last_seen_at"`
}

type ChatParticipantRole string

const (
	ChatRoleRenter    ChatParticipantRole = "renter"
	ChatRoleConcierge ChatParticipantRole = "concierge"
	ChatRoleModerator ChatParticipantRole = "moderator"
	ChatRoleAdmin     ChatParticipantRole = "admin"
)

type WSEventType string

const (
	WSEventNewMessage  WSEventType = "new_message"
	WSEventMessageRead WSEventType = "message_read"
	WSEventUserJoined  WSEventType = "user_joined"
	WSEventUserLeft    WSEventType = "user_left"
	WSEventUserTyping  WSEventType = "user_typing"
	WSEventChatOpened  WSEventType = "chat_opened"
	WSEventChatClosed  WSEventType = "chat_closed"
	WSEventError       WSEventType = "error"
	WSEventHeartbeat   WSEventType = "heartbeat"
)

type WSMessage struct {
	Type      WSEventType `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	UserID    int         `json:"user_id,omitempty"`
	RoomID    int         `json:"room_id,omitempty"`
}

type CreateChatRoomRequest struct {
	BookingID int `json:"booking_id" validate:"required"`
}

type SendMessageRequest struct {
	Type      MessageType `json:"type" validate:"required"`
	Content   string      `json:"content" validate:"required"`
	ReplyToID *int        `json:"reply_to_id,omitempty"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" validate:"required"`
}

type MarkMessagesReadRequest struct {
	MessageIDs []int `json:"message_ids" validate:"required"`
}

type ChatRoomRepository interface {
	Create(room *ChatRoom) error
	GetByID(id int) (*ChatRoom, error)
	GetByBookingID(bookingID int) (*ChatRoom, error)
	GetByUserID(userID int, status []ChatRoomStatus, page, pageSize int) ([]*ChatRoom, int, error)
	GetByConciergeID(conciergeID int, status []ChatRoomStatus, page, pageSize int) ([]*ChatRoom, int, error)
	GetActiveRooms() ([]*ChatRoom, error)
	GetRoomsPendingOpen() ([]*ChatRoom, error)
	GetRoomsPendingClose() ([]*ChatRoom, error)
	Update(room *ChatRoom) error
	Delete(id int) error
	CanUserAccessRoom(roomID, userID int) (bool, error)
}

type ChatMessageRepository interface {
	Create(message *ChatMessage) error
	GetByID(id int) (*ChatMessage, error)
	GetByRoomID(roomID int, page, pageSize int) ([]*ChatMessage, int, error)
	GetUnreadCount(roomID, userID int) (int, error)
	GetUnreadMessages(roomID, userID int) ([]*ChatMessage, error)
	Update(message *ChatMessage) error
	Delete(id int) error
	MarkAsRead(messageIDs []int, userID int) error
	MarkRoomAsRead(roomID, userID int) error
	GetLastMessage(roomID int) (*ChatMessage, error)
}

type ChatParticipantRepository interface {
	Create(participant *ChatParticipant) error
	GetByRoomID(roomID int) ([]*ChatParticipant, error)
	GetByUserAndRoom(userID, roomID int) (*ChatParticipant, error)
	UpdateLastSeen(userID, roomID int, lastSeen time.Time) error
	UpdateOnlineStatus(userID, roomID int, isOnline bool) error
	UpdateLastRead(userID, roomID int, lastRead time.Time) error
}

type ChatUseCase interface {
	CreateChatRoom(request *CreateChatRoomRequest, userID int) (*ChatRoom, error)
	GetChatRoom(roomID, userID int) (*ChatRoom, error)
	GetChatRoomByBookingID(bookingID int) (*ChatRoom, error)
	GetUserChatRooms(userID int, status []ChatRoomStatus, page, pageSize int) ([]*ChatRoom, int, error)
	GetConciergeChatRooms(conciergeID int, status []ChatRoomStatus, page, pageSize int) ([]*ChatRoom, int, error)
	CanUserAccessRoom(roomID, userID int) (bool, error)

	SendMessage(roomID int, request *SendMessageRequest, userID int) (*ChatMessage, error)
	GetMessages(roomID, userID int, page, pageSize int) ([]*ChatMessage, int, error)
	UpdateMessage(messageID int, request *UpdateMessageRequest, userID int) error
	DeleteMessage(messageID, userID int) error
	MarkMessagesAsRead(roomID int, request *MarkMessagesReadRequest, userID int) error
	GetUnreadCount(roomID, userID int) (int, error)

	OpenScheduledChats() error
	CloseExpiredChats() error
	CreateWelcomeMessage(roomID int) error
	NotifyNewMessage(message *ChatMessage) error

	ActivateChat(roomID, userID int) error
	CloseChat(roomID, userID int) error
}

type ChatWebSocketService interface {
	HandleWebSocket(c interface{})
	HandleConnection(userID int, roomID int) error
	BroadcastToRoom(roomID int, message *WSMessage) error
	SendToUser(userID int, message *WSMessage) error
	UserJoinRoom(userID, roomID int) error
	UserLeaveRoom(userID, roomID int) error
	GetOnlineUsers(roomID int) ([]int, error)
	IsUserOnline(userID int) bool
}

type ChatRoomResponse struct {
	ID            int               `json:"id"`
	BookingID     int               `json:"booking_id"`
	Booking       *Booking          `json:"booking,omitempty"`
	ConciergeID   int               `json:"concierge_id"`
	Concierge     *Concierge        `json:"concierge,omitempty"`
	RenterID      int               `json:"renter_id"`
	Renter        *Renter           `json:"renter,omitempty"`
	ApartmentID   int               `json:"apartment_id"`
	Apartment     *Apartment        `json:"apartment,omitempty"`
	Status        ChatRoomStatus    `json:"status"`
	UnreadCount   int               `json:"unread_count"`
	LastMessage   *ChatMessage      `json:"last_message,omitempty"`
	Participants  []ChatParticipant `json:"participants,omitempty"`
	OpenedAt      *string           `json:"opened_at"`
	ClosedAt      *string           `json:"closed_at"`
	LastMessageAt *string           `json:"last_message_at"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
}

type ChatMessageResponse struct {
	ID         int           `json:"id"`
	ChatRoomID int           `json:"chat_room_id"`
	SenderID   int           `json:"sender_id"`
	Sender     *User         `json:"sender,omitempty"`
	Type       MessageType   `json:"type"`
	Content    string        `json:"content"`
	FileURL    *string       `json:"file_url,omitempty"`
	FileName   *string       `json:"file_name,omitempty"`
	FileSize   *int64        `json:"file_size,omitempty"`
	Status     MessageStatus `json:"status"`
	ReadAt     *string       `json:"read_at"`
	ReplyToID  *int          `json:"reply_to_id,omitempty"`
	ReplyTo    *ChatMessage  `json:"reply_to,omitempty"`
	EditedAt   *string       `json:"edited_at"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
}
