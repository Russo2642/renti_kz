package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/russo2642/renti_kz/internal/domain"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	ID           string
	UserID       int
	RoomID       int
	Conn         *websocket.Conn
	Send         chan *domain.WSMessage
	Hub          *Hub
	LastActivity time.Time
}

type Hub struct {
	clients map[*Client]bool

	rooms map[int]map[*Client]bool

	users map[int]*Client

	register chan *Client

	unregister chan *Client

	broadcast chan *BroadcastMessage

	mutex sync.RWMutex

	chatUseCase domain.ChatUseCase
	userUseCase domain.UserUseCase
}

type BroadcastMessage struct {
	RoomID  int
	Message *domain.WSMessage
	Exclude *Client
}

type ChatWebSocketService struct {
	hub         *Hub
	chatUseCase domain.ChatUseCase
	userUseCase domain.UserUseCase
}

func NewChatWebSocketService(chatUseCase domain.ChatUseCase, userUseCase domain.UserUseCase) *ChatWebSocketService {
	hub := &Hub{
		clients:     make(map[*Client]bool),
		rooms:       make(map[int]map[*Client]bool),
		users:       make(map[int]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage),
		chatUseCase: chatUseCase,
		userUseCase: userUseCase,
	}

	go hub.run()

	return &ChatWebSocketService{
		hub:         hub,
		chatUseCase: chatUseCase,
		userUseCase: userUseCase,
	}
}

func (s *ChatWebSocketService) SetChatUseCase(chatUseCase domain.ChatUseCase) {
	s.chatUseCase = chatUseCase
	s.hub.chatUseCase = chatUseCase
}

func (s *ChatWebSocketService) HandleWebSocket(c interface{}) {
	ginCtx, ok := c.(*gin.Context)
	if !ok {
		return
	}
	roomIDStr := ginCtx.Param("roomId")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	userIDInterface, exists := ginCtx.Get("user_id")
	if !exists {
		ginCtx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(int)
	if !ok {
		ginCtx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	if s.chatUseCase != nil {
		canAccess, err := s.chatUseCase.CanUserAccessRoom(roomID, userID)
		if err != nil || !canAccess {
			ginCtx.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	conn, err := upgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:           fmt.Sprintf("%d_%d_%d", userID, roomID, time.Now().Unix()),
		UserID:       userID,
		RoomID:       roomID,
		Conn:         conn,
		Send:         make(chan *domain.WSMessage, 256),
		Hub:          s.hub,
		LastActivity: time.Now(),
	}

	s.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (s *ChatWebSocketService) HandleConnection(userID int, roomID int) error {
	return nil
}

func (s *ChatWebSocketService) BroadcastToRoom(roomID int, message *domain.WSMessage) error {
	s.hub.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: message,
	}
	return nil
}

func (s *ChatWebSocketService) SendToUser(userID int, message *domain.WSMessage) error {
	s.hub.mutex.RLock()
	client, exists := s.hub.users[userID]
	s.hub.mutex.RUnlock()

	if exists && client != nil {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(s.hub.users, userID)
		}
	}
	return nil
}

func (s *ChatWebSocketService) UserJoinRoom(userID, roomID int) error {
	message := &domain.WSMessage{
		Type:      domain.WSEventUserJoined,
		Data:      map[string]interface{}{"user_id": userID},
		Timestamp: time.Now(),
		UserID:    userID,
		RoomID:    roomID,
	}
	return s.BroadcastToRoom(roomID, message)
}

func (s *ChatWebSocketService) UserLeaveRoom(userID, roomID int) error {
	message := &domain.WSMessage{
		Type:      domain.WSEventUserLeft,
		Data:      map[string]interface{}{"user_id": userID},
		Timestamp: time.Now(),
		UserID:    userID,
		RoomID:    roomID,
	}
	return s.BroadcastToRoom(roomID, message)
}

func (s *ChatWebSocketService) GetOnlineUsers(roomID int) ([]int, error) {
	s.hub.mutex.RLock()
	defer s.hub.mutex.RUnlock()

	var userIDs []int
	if room, exists := s.hub.rooms[roomID]; exists {
		for client := range room {
			userIDs = append(userIDs, client.UserID)
		}
	}
	return userIDs, nil
}

func (s *ChatWebSocketService) IsUserOnline(userID int) bool {
	s.hub.mutex.RLock()
	defer s.hub.mutex.RUnlock()

	_, exists := s.hub.users[userID]
	return exists
}

func (h *Hub) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToRoom(message)

		case <-ticker.C:
			h.cleanupInactiveConnections()
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client] = true

	if h.rooms[client.RoomID] == nil {
		h.rooms[client.RoomID] = make(map[*Client]bool)
	}
	h.rooms[client.RoomID][client] = true

	if existingClient, exists := h.users[client.UserID]; exists {
		h.unregisterClientUnsafe(existingClient)
	}

	h.users[client.UserID] = client

	log.Printf("Client registered: User %d in room %d", client.UserID, client.RoomID)

	welcomeMsg := &domain.WSMessage{
		Type:      domain.WSEventUserJoined,
		Data:      map[string]interface{}{"message": "Connected to chat"},
		Timestamp: time.Now(),
		UserID:    client.UserID,
		RoomID:    client.RoomID,
	}

	select {
	case client.Send <- welcomeMsg:
	default:
		h.unregisterClientUnsafe(client)
	}
}

func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.unregisterClientUnsafe(client)
}

func (h *Hub) unregisterClientUnsafe(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)

		if room, exists := h.rooms[client.RoomID]; exists {
			delete(room, client)
			if len(room) == 0 {
				delete(h.rooms, client.RoomID)
			}
		}

		if h.users[client.UserID] == client {
			delete(h.users, client.UserID)
		}

		close(client.Send)
		client.Conn.Close()

		log.Printf("Client unregistered: User %d from room %d", client.UserID, client.RoomID)
	}
}

func (h *Hub) broadcastToRoom(message *BroadcastMessage) {
	h.mutex.RLock()
	room, exists := h.rooms[message.RoomID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	for client := range room {
		if message.Exclude != nil && client == message.Exclude {
			continue
		}

		select {
		case client.Send <- message.Message:
		default:
			h.unregister <- client
		}
	}
}

func (h *Hub) cleanupInactiveConnections() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	now := time.Now()
	timeout := 5 * time.Minute

	for client := range h.clients {
		if now.Sub(client.LastActivity) > timeout {
			h.unregisterClientUnsafe(client)
		}
	}
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		c.LastActivity = time.Now()
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.LastActivity = time.Now()

		var wsMsg domain.WSMessage
		if err := json.Unmarshal(messageBytes, &wsMsg); err != nil {
			log.Printf("Error parsing WebSocket message: %v", err)
			continue
		}

		c.handleIncomingMessage(&wsMsg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				w.Close()
				continue
			}

			w.Write(messageBytes)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				additionalMsg := <-c.Send
				additionalBytes, err := json.Marshal(additionalMsg)
				if err != nil {
					log.Printf("Error marshaling additional message: %v", err)
					continue
				}
				w.Write(additionalBytes)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleIncomingMessage(msg *domain.WSMessage) {
	switch msg.Type {
	case domain.WSEventHeartbeat:
		c.LastActivity = time.Now()

	case domain.WSEventUserTyping:
		broadcastMsg := &BroadcastMessage{
			RoomID:  c.RoomID,
			Message: msg,
			Exclude: c,
		}
		c.Hub.broadcast <- broadcastMsg

	default:
		log.Printf("Unknown WebSocket message type: %s", msg.Type)
	}
}
