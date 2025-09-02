package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type ChatHandler struct {
	chatUseCase domain.ChatUseCase
	wsService   domain.ChatWebSocketService
}

func NewChatHandler(
	chatUseCase domain.ChatUseCase,
	wsService domain.ChatWebSocketService,
) *ChatHandler {
	return &ChatHandler{
		chatUseCase: chatUseCase,
		wsService:   wsService,
	}
}

func (h *ChatHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/ws/chat/:roomId", h.HandleWebSocket)

	chat := router.Group("/chat")
	{
		chat.POST("/rooms", h.CreateChatRoom)
		chat.GET("/rooms", h.GetUserChatRooms)
		chat.GET("/rooms/:roomId", h.GetChatRoom)

		chat.POST("/rooms/:roomId/messages", h.SendMessage)
		chat.GET("/rooms/:roomId/messages", h.GetMessages)
		chat.PUT("/messages/:messageId", h.UpdateMessage)
		chat.DELETE("/messages/:messageId", h.DeleteMessage)

		chat.POST("/rooms/:roomId/read", h.MarkMessagesAsRead)
		chat.GET("/rooms/:roomId/unread-count", h.GetUnreadCount)

		chat.POST("/rooms/:roomId/activate", h.ActivateChat)
		chat.POST("/rooms/:roomId/close", h.CloseChat)
	}
}

func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	roomIDStr := c.Param("roomId")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse("Invalid room ID"))
		return
	}

	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(401, domain.NewErrorResponse("Unauthorized"))
		return
	}

	canAccess, err := h.chatUseCase.CanUserAccessRoom(roomID, userID)
	if err != nil || !canAccess {
		c.JSON(403, domain.NewErrorResponse("Access denied"))
		return
	}

	h.wsService.HandleWebSocket(c)
}

// @Summary Создание комнаты чата
// @Description Создает новую комнату чата для бронирования с консьержем
// @Tags chat
// @Accept json
// @Produce json
// @Param request body domain.CreateChatRoomRequest true "Данные для создания комнаты"
// @Success 201 {object} domain.SuccessResponse{data=domain.ChatRoom} "Комната чата создана"
// @Failure 400 {object} domain.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Security ApiKeyAuth
// @Router /chat/rooms [post]
func (h *ChatHandler) CreateChatRoom(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	var request domain.CreateChatRoomRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, domain.NewErrorResponse("Invalid request body"))
		return
	}

	room, err := h.chatUseCase.CreateChatRoom(&request, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(201, domain.NewSuccessResponse("Chat room created successfully", room))
}

// @Summary Список комнат чата
// @Description Получает список комнат чата пользователя с пагинацией
// @Tags chat
// @Produce json
// @Param status query string false "Фильтр по статусу комнаты (pending, active, closed)"
// @Param page query integer false "Номер страницы" default(1)
// @Param page_size query integer false "Размер страницы" default(20)
// @Success 200 {object} object{success=boolean,data=[]domain.ChatRoom,pagination=object} "Список комнат чата"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 500 {object} domain.ErrorResponse "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /chat/rooms [get]
func (h *ChatHandler) GetUserChatRooms(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	statusParam := c.Query("status")
	var statusFilter []domain.ChatRoomStatus
	if statusParam != "" {
		statusFilter = append(statusFilter, domain.ChatRoomStatus(statusParam))
	}

	page, pageSize := utils.ParsePagination(c)

	rooms, total, err := h.chatUseCase.GetUserChatRooms(userID, statusFilter, page, pageSize)
	if err != nil {
		c.JSON(500, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    rooms,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// @Summary Получение комнаты чата
// @Description Получает информацию о комнате чата по ID
// @Tags chat
// @Produce json
// @Param roomId path integer true "ID комнаты чата"
// @Success 200 {object} domain.SuccessResponse{data=domain.ChatRoom} "Комната чата"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Доступ запрещен"
// @Security ApiKeyAuth
// @Router /chat/rooms/{roomId} [get]
func (h *ChatHandler) GetChatRoom(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	roomID, ok := utils.ParseIDParam(c, "roomId")
	if !ok {
		return
	}

	room, err := h.chatUseCase.GetChatRoom(roomID, userID)
	if err != nil {
		c.JSON(403, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Chat room retrieved successfully", room))
}

// @Summary Отправка сообщения
// @Description Отправляет сообщение в указанную комнату чата
// @Tags chat
// @Accept json
// @Produce json
// @Param roomId path integer true "ID комнаты чата"
// @Param request body domain.SendMessageRequest true "Данные сообщения"
// @Success 201 {object} domain.SuccessResponse{data=domain.ChatMessage} "Сообщение отправлено"
// @Failure 400 {object} domain.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Доступ запрещен"
// @Security ApiKeyAuth
// @Router /chat/rooms/{roomId}/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	roomID, ok := utils.ParseIDParam(c, "roomId")
	if !ok {
		return
	}

	var request domain.SendMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, domain.NewErrorResponse("Invalid request body"))
		return
	}

	message, err := h.chatUseCase.SendMessage(roomID, &request, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(201, domain.NewSuccessResponse("Message sent successfully", message))
}

// @Summary Получение сообщений
// @Description Получает список сообщений из указанной комнаты чата с пагинацией
// @Tags chat
// @Produce json
// @Param roomId path integer true "ID комнаты чата"
// @Param page query integer false "Номер страницы" default(1)
// @Param page_size query integer false "Размер страницы" default(50)
// @Success 200 {object} object{success=boolean,data=[]domain.ChatMessage,pagination=object} "Список сообщений"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Доступ запрещен"
// @Security ApiKeyAuth
// @Router /chat/rooms/{roomId}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	roomID, ok := utils.ParseIDParam(c, "roomId")
	if !ok {
		return
	}

	page, pageSize := utils.ParsePagination(c)

	messages, total, err := h.chatUseCase.GetMessages(roomID, userID, page, pageSize)
	if err != nil {
		c.JSON(403, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    messages,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// @Summary Редактирование сообщения
// @Description Редактирует содержимое сообщения (только автор может редактировать свои сообщения)
// @Tags chat
// @Accept json
// @Produce json
// @Param messageId path integer true "ID сообщения"
// @Param request body domain.UpdateMessageRequest true "Новый текст сообщения"
// @Success 200 {object} domain.SuccessResponse "Сообщение обновлено"
// @Failure 400 {object} domain.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Доступ запрещен"
// @Security ApiKeyAuth
// @Router /chat/messages/{messageId} [put]
func (h *ChatHandler) UpdateMessage(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	messageID, ok := utils.ParseIDParam(c, "messageId")
	if !ok {
		return
	}

	var request domain.UpdateMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, domain.NewErrorResponse("Invalid request body"))
		return
	}

	err := h.chatUseCase.UpdateMessage(messageID, &request, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Message updated successfully", nil))
}

// @Summary Удаление сообщения
// @Description Удаляет сообщение (только автор может удалить свое сообщение)
// @Tags chat
// @Produce json
// @Param messageId path integer true "ID сообщения"
// @Success 200 {object} domain.SuccessResponse "Сообщение удалено"
// @Failure 400 {object} domain.ErrorResponse "Ошибка при удалении"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Доступ запрещен"
// @Security ApiKeyAuth
// @Router /chat/messages/{messageId} [delete]
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	messageID, ok := utils.ParseIDParam(c, "messageId")
	if !ok {
		return
	}

	err := h.chatUseCase.DeleteMessage(messageID, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Message deleted successfully", nil))
}

// @Summary Отметить сообщения как прочитанные
// @Description Отмечает указанные сообщения в комнате как прочитанные
// @Tags chat
// @Accept json
// @Produce json
// @Param roomId path integer true "ID комнаты чата"
// @Param request body domain.MarkMessagesReadRequest true "Список ID сообщений"
// @Success 200 {object} domain.SuccessResponse "Сообщения отмечены как прочитанные"
// @Failure 400 {object} domain.ErrorResponse "Неверные данные запроса"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Доступ запрещен"
// @Security ApiKeyAuth
// @Router /chat/rooms/{roomId}/read [post]
func (h *ChatHandler) MarkMessagesAsRead(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	roomID, ok := utils.ParseIDParam(c, "roomId")
	if !ok {
		return
	}

	var request domain.MarkMessagesReadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, domain.NewErrorResponse("Invalid request body"))
		return
	}

	err := h.chatUseCase.MarkMessagesAsRead(roomID, &request, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Messages marked as read", nil))
}

// @Summary Получить количество непрочитанных сообщений
// @Description Получает количество непрочитанных сообщений в комнате чата
// @Tags chat
// @Produce json
// @Param roomId path integer true "ID комнаты чата"
// @Success 200 {object} domain.SuccessResponse{data=object{unread_count=integer}} "Количество непрочитанных"
// @Failure 401 {object} domain.ErrorResponse "Не авторизован"
// @Failure 403 {object} domain.ErrorResponse "Доступ запрещен"
// @Security ApiKeyAuth
// @Router /chat/rooms/{roomId}/unread-count [get]
func (h *ChatHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	roomID, ok := utils.ParseIDParam(c, "roomId")
	if !ok {
		return
	}

	count, err := h.chatUseCase.GetUnreadCount(roomID, userID)
	if err != nil {
		c.JSON(403, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Unread count retrieved successfully", gin.H{
		"unread_count": count,
	}))
}

// @Summary Активировать чат
// @Description Активирует комнату чата (переводит из pending в active)
// @Tags chat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param roomId path int true "ID комнаты"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /chat/rooms/{roomId}/activate [post]
func (h *ChatHandler) ActivateChat(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	roomID, ok := utils.ParseIDParam(c, "roomId")
	if !ok {
		return
	}

	err := h.chatUseCase.ActivateChat(roomID, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Chat activated successfully", nil))
}

// @Summary Закрыть чат
// @Description Закрывает комнату чата (переводит в closed)
// @Tags chat
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param roomId path int true "ID комнаты"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /chat/rooms/{roomId}/close [post]
func (h *ChatHandler) CloseChat(c *gin.Context) {
	userID, ok := utils.RequireAuth(c)
	if !ok {
		return
	}

	roomID, ok := utils.ParseIDParam(c, "roomId")
	if !ok {
		return
	}

	err := h.chatUseCase.CloseChat(roomID, userID)
	if err != nil {
		c.JSON(400, domain.NewErrorResponse(err.Error()))
		return
	}

	c.JSON(200, domain.NewSuccessResponse("Chat closed successfully", nil))
}
