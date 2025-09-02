package domain

import (
	"time"
)

type Region struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Cities    []*City   `json:"cities,omitempty"`
}

type City struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	RegionID    int          `json:"region_id"`
	Region      *Region      `json:"region,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Districts   []*District  `json:"districts,omitempty"`
	Coordinates *Coordinates `json:"coordinates,omitempty"`
}

type District struct {
	ID             int              `json:"id"`
	Name           string           `json:"name"`
	CityID         int              `json:"city_id"`
	City           *City            `json:"city,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	Microdistricts []*Microdistrict `json:"microdistricts,omitempty"`
}

type Microdistrict struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	DistrictID int       `json:"district_id"`
	District   *District `json:"district,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Coordinates struct {
	ID        int       `json:"id"`
	CityID    int       `json:"city_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LocationRepository interface {
	GetAllRegions() ([]*Region, error)
	GetRegionByID(id int) (*Region, error)

	GetAllCities() ([]*City, error)
	GetCitiesByRegionID(regionID int) ([]*City, error)
	GetCityByID(id int) (*City, error)

	GetAllDistricts() ([]*District, error)
	GetDistrictsByCityID(cityID int) ([]*District, error)
	GetDistrictByID(id int) (*District, error)

	GetAllMicrodistricts() ([]*Microdistrict, error)
	GetMicrodistrictsByDistrictID(districtID int) ([]*Microdistrict, error)
	GetMicrodistrictByID(id int) (*Microdistrict, error)
}

type LocationUseCase interface {
	GetAllRegions() ([]*Region, error)
	GetRegionByID(id int) (*Region, error)

	GetAllCities() ([]*City, error)
	GetCitiesByRegionID(regionID int) ([]*City, error)
	GetCityByID(id int) (*City, error)

	GetAllDistricts() ([]*District, error)
	GetDistrictsByCityID(cityID int) ([]*District, error)
	GetDistrictByID(id int) (*District, error)

	GetAllMicrodistricts() ([]*Microdistrict, error)
	GetMicrodistrictsByDistrictID(districtID int) ([]*Microdistrict, error)
	GetMicrodistrictByID(id int) (*Microdistrict, error)
}
