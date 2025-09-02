package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
)

type DictionaryHandler struct {
	apartmentUseCase domain.ApartmentUseCase
}

func NewDictionaryHandler(
	apartmentUseCase domain.ApartmentUseCase,
) *DictionaryHandler {
	return &DictionaryHandler{
		apartmentUseCase: apartmentUseCase,
	}
}

func (h *DictionaryHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/conditions", h.GetAllConditions)
	router.GET("/house-rules", h.GetAllHouseRules)
	router.GET("/amenities", h.GetAllPopularAmenities)
}

// @Summary Получение состояний квартир
// @Description Получает список всех доступных состояний квартир
// @Tags dictionaries
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /dictionaries/conditions [get]
func (h *DictionaryHandler) GetAllConditions(c *gin.Context) {
	conditions, err := h.apartmentUseCase.GetAllConditions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении списка состояний квартир"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("список состояний квартир успешно получен", conditions))
}

// @Summary Получение правил дома
// @Description Получает список всех доступных правил дома
// @Tags dictionaries
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /dictionaries/house-rules [get]
func (h *DictionaryHandler) GetAllHouseRules(c *gin.Context) {
	rules, err := h.apartmentUseCase.GetAllHouseRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении списка правил дома"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("список правил дома успешно получен", rules))
}

// @Summary Получение популярных удобств
// @Description Получает список всех популярных удобств
// @Tags dictionaries
// @Accept json
// @Produce json
// @Success 200 {object} domain.SuccessResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /dictionaries/amenities [get]
func (h *DictionaryHandler) GetAllPopularAmenities(c *gin.Context) {
	amenities, err := h.apartmentUseCase.GetAllPopularAmenities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.NewErrorResponse("ошибка при получении списка популярных удобств"))
		return
	}

	c.JSON(http.StatusOK, domain.NewSuccessResponse("список популярных удобств успешно получен", amenities))
}
