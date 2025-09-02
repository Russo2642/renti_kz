package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
)

type FavoriteHandler struct {
	favoriteUseCase domain.FavoriteUseCase
}

func NewFavoriteHandler(favoriteUseCase domain.FavoriteUseCase) *FavoriteHandler {
	return &FavoriteHandler{
		favoriteUseCase: favoriteUseCase,
	}
}

func (h *FavoriteHandler) RegisterRoutes(router *gin.RouterGroup) {

	favorites := router.Group("/favorites")
	{
		favorites.GET("", h.GetUserFavorites)
		favorites.POST("/:apartment_id", h.AddToFavorites)
		favorites.DELETE("/:apartment_id", h.RemoveFromFavorites)
		favorites.POST("/:apartment_id/toggle", h.ToggleFavorite)
		favorites.GET("/:apartment_id/check", h.IsFavorite)
		favorites.GET("/count/:apartment_id", h.GetFavoriteCount)
	}
}

// @Summary Добавление в избранное
// @Description Добавляет квартиру в список избранных пользователя
// @Tags favorites
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param apartment_id path int true "ID квартиры"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /favorites/{apartment_id} [post]
func (h *FavoriteHandler) AddToFavorites(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	apartmentIDStr := c.Param("apartment_id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID квартиры"})
		return
	}

	err = h.favoriteUseCase.AddToFavorites(userID.(int), apartmentID)
	if err != nil {
		switch {
		case err.Error() == "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		case err.Error() == "apartment not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Квартира не найдена"})
		case err.Error() == "apartment is not available for favorites":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Квартира недоступна для добавления в избранное"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить в избранное"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Квартира добавлена в избранное",
	})
}

// @Summary Удаление из избранного
// @Description Удаляет указанную квартиру из избранного пользователя
// @Tags favorites
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param apartment_id path int true "ID квартиры"
// @Success 200 {object} domain.SuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /favorites/{apartment_id} [delete]
func (h *FavoriteHandler) RemoveFromFavorites(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	apartmentIDStr := c.Param("apartment_id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID квартиры"})
		return
	}

	err = h.favoriteUseCase.RemoveFromFavorites(userID.(int), apartmentID)
	if err != nil {
		switch {
		case err.Error() == "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		case err.Error() == "apartment not found in favorites":
			c.JSON(http.StatusNotFound, gin.H{"error": "Квартира не найдена в избранном"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить из избранного"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Квартира удалена из избранного",
	})
}

// @Summary Получение избранных квартир
// @Description Получает список избранных квартир пользователя с пагинацией
// @Tags favorites
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /favorites [get]
func (h *FavoriteHandler) GetUserFavorites(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	favorites, total, err := h.favoriteUseCase.GetUserFavorites(userID.(int), page, pageSize)
	if err != nil {
		switch {
		case err.Error() == "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить избранные квартиры"})
		}
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"data": favorites,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (h *FavoriteHandler) ToggleFavorite(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	apartmentIDStr := c.Param("apartment_id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID квартиры"})
		return
	}

	isAdded, err := h.favoriteUseCase.ToggleFavorite(userID.(int), apartmentID)
	if err != nil {
		switch {
		case err.Error() == "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		case err.Error() == "apartment not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Квартира не найдена"})
		case err.Error() == "apartment is not available for favorites":
			c.JSON(http.StatusBadRequest, gin.H{"error": "Квартира недоступна для добавления в избранное"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось изменить статус избранного"})
		}
		return
	}

	message := "Квартира удалена из избранного"
	if isAdded {
		message = "Квартира добавлена в избранное"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     message,
		"is_favorite": isAdded,
	})
}

func (h *FavoriteHandler) IsFavorite(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	apartmentIDStr := c.Param("apartment_id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID квартиры"})
		return
	}

	isFavorite, err := h.favoriteUseCase.IsFavorite(userID.(int), apartmentID)
	if err != nil {
		switch {
		case err.Error() == "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось проверить статус избранного"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_favorite": isFavorite,
	})
}

func (h *FavoriteHandler) GetFavoriteCount(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	apartmentIDStr := c.Param("apartment_id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID квартиры"})
		return
	}

	count, err := h.favoriteUseCase.GetFavoriteCount(userID.(int), apartmentID)
	if err != nil {
		switch {
		case err.Error() == "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		case err.Error() == "apartment not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Квартира не найдена"})
		case err.Error() == "user is not a property owner":
			c.JSON(http.StatusForbidden, gin.H{"error": "Пользователь не является владельцем недвижимости"})
		case err.Error() == "access denied: you can only view favorite count for your own apartments":
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен: вы можете просматривать количество избранных только для своих квартир"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить количество избранных"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"favorite_count": count,
	})
}
