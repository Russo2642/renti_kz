package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/russo2642/renti_kz/internal/domain"
)

type LocationHandler struct {
	locationUseCase domain.LocationUseCase
}

func NewLocationHandler(g *gin.RouterGroup, locationUseCase domain.LocationUseCase) {
	handler := &LocationHandler{
		locationUseCase: locationUseCase,
	}

	// Регистрируем роуты напрямую без создания подгруппы
	g.GET("/regions", handler.GetAllRegions)
	g.GET("/regions/:id", handler.GetRegionByID)

	g.GET("/cities", handler.GetAllCities)
	g.GET("/regions/:id/cities", handler.GetCitiesByRegionID)
	g.GET("/cities/:id", handler.GetCityByID)

	g.GET("/districts", handler.GetAllDistricts)
	g.GET("/cities/:id/districts", handler.GetDistrictsByCityID)
	g.GET("/districts/:id", handler.GetDistrictByID)

	g.GET("/microdistricts", handler.GetAllMicrodistricts)
	g.GET("/districts/:id/microdistricts", handler.GetMicrodistrictsByDistrictID)
	g.GET("/microdistricts/:id", handler.GetMicrodistrictByID)
}

// @Summary Получение всех регионов
// @Description Получает список всех регионов
// @Tags locations
// @Accept json
// @Produce json
// @Success 200 {array} domain.Region
// @Failure 500 {object} map[string]string
// @Router /locations/regions [get]
func (h *LocationHandler) GetAllRegions(c *gin.Context) {
	regions, err := h.locationUseCase.GetAllRegions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, regions)
}

// @Summary Получение региона по ID
// @Description Получает информацию о регионе по его идентификатору
// @Tags locations
// @Accept json
// @Produce json
// @Param id path int true "ID региона"
// @Success 200 {object} domain.Region
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /locations/regions/{id} [get]
func (h *LocationHandler) GetRegionByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id parameter"})
		return
	}

	region, err := h.locationUseCase.GetRegionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, region)
}

// @Summary Получение всех городов
// @Description Получает список всех городов
// @Tags locations
// @Accept json
// @Produce json
// @Success 200 {array} domain.City
// @Failure 500 {object} map[string]string
// @Router /locations/cities [get]
func (h *LocationHandler) GetAllCities(c *gin.Context) {
	cities, err := h.locationUseCase.GetAllCities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cities)
}

// @Summary Получение городов региона
// @Description Получает список городов для указанного региона
// @Tags locations
// @Accept json
// @Produce json
// @Param id path int true "ID региона"
// @Success 200 {array} domain.City
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /locations/regions/{id}/cities [get]
func (h *LocationHandler) GetCitiesByRegionID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id parameter"})
		return
	}

	cities, err := h.locationUseCase.GetCitiesByRegionID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cities)
}

// @Summary Получение города по ID
// @Description Получает информацию о городе по его идентификатору
// @Tags locations
// @Accept json
// @Produce json
// @Param id path int true "ID города"
// @Success 200 {object} domain.City
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /locations/cities/{id} [get]
func (h *LocationHandler) GetCityByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id parameter"})
		return
	}

	city, err := h.locationUseCase.GetCityByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, city)
}

// @Summary Получение всех районов
// @Description Получает список всех районов
// @Tags locations
// @Accept json
// @Produce json
// @Success 200 {array} domain.District
// @Failure 500 {object} map[string]string
// @Router /locations/districts [get]
func (h *LocationHandler) GetAllDistricts(c *gin.Context) {
	districts, err := h.locationUseCase.GetAllDistricts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, districts)
}

// @Summary Получение районов города
// @Description Получает список районов для указанного города
// @Tags locations
// @Accept json
// @Produce json
// @Param id path int true "ID города"
// @Success 200 {array} domain.District
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /locations/cities/{id}/districts [get]
func (h *LocationHandler) GetDistrictsByCityID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id parameter"})
		return
	}

	districts, err := h.locationUseCase.GetDistrictsByCityID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, districts)
}

// @Summary Получение района по ID
// @Description Получает информацию о районе по его идентификатору
// @Tags locations
// @Accept json
// @Produce json
// @Param id path int true "ID района"
// @Success 200 {object} domain.District
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /locations/districts/{id} [get]
func (h *LocationHandler) GetDistrictByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id parameter"})
		return
	}

	district, err := h.locationUseCase.GetDistrictByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, district)
}

// @Summary Получение всех микрорайонов
// @Description Получает список всех микрорайонов
// @Tags locations
// @Accept json
// @Produce json
// @Success 200 {array} domain.Microdistrict
// @Failure 500 {object} map[string]string
// @Router /locations/microdistricts [get]
func (h *LocationHandler) GetAllMicrodistricts(c *gin.Context) {
	microdistricts, err := h.locationUseCase.GetAllMicrodistricts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, microdistricts)
}

// @Summary Получение микрорайонов района
// @Description Получает список микрорайонов для указанного района
// @Tags locations
// @Accept json
// @Produce json
// @Param id path int true "ID района"
// @Success 200 {array} domain.Microdistrict
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /locations/districts/{id}/microdistricts [get]
func (h *LocationHandler) GetMicrodistrictsByDistrictID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id parameter"})
		return
	}

	microdistricts, err := h.locationUseCase.GetMicrodistrictsByDistrictID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, microdistricts)
}

// @Summary Получение микрорайона по ID
// @Description Получает информацию о микрорайоне по его идентификатору
// @Tags locations
// @Accept json
// @Produce json
// @Param id path int true "ID микрорайона"
// @Success 200 {object} domain.Microdistrict
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /locations/microdistricts/{id} [get]
func (h *LocationHandler) GetMicrodistrictByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id parameter"})
		return
	}

	microdistrict, err := h.locationUseCase.GetMicrodistrictByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, microdistrict)
}
