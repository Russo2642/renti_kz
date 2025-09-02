package usecase

import (
	"github.com/russo2642/renti_kz/internal/domain"
)

type LocationUseCase struct {
	locationRepo domain.LocationRepository
}

func NewLocationUseCase(locationRepo domain.LocationRepository) *LocationUseCase {
	return &LocationUseCase{
		locationRepo: locationRepo,
	}
}

func (uc *LocationUseCase) GetAllRegions() ([]*domain.Region, error) {
	return uc.locationRepo.GetAllRegions()
}

func (uc *LocationUseCase) GetRegionByID(id int) (*domain.Region, error) {
	return uc.locationRepo.GetRegionByID(id)
}

func (uc *LocationUseCase) GetAllCities() ([]*domain.City, error) {
	return uc.locationRepo.GetAllCities()
}

func (uc *LocationUseCase) GetCitiesByRegionID(regionID int) ([]*domain.City, error) {
	return uc.locationRepo.GetCitiesByRegionID(regionID)
}

func (uc *LocationUseCase) GetCityByID(id int) (*domain.City, error) {
	return uc.locationRepo.GetCityByID(id)
}

func (uc *LocationUseCase) GetAllDistricts() ([]*domain.District, error) {
	return uc.locationRepo.GetAllDistricts()
}

func (uc *LocationUseCase) GetDistrictsByCityID(cityID int) ([]*domain.District, error) {
	return uc.locationRepo.GetDistrictsByCityID(cityID)
}

func (uc *LocationUseCase) GetDistrictByID(id int) (*domain.District, error) {
	return uc.locationRepo.GetDistrictByID(id)
}

func (uc *LocationUseCase) GetAllMicrodistricts() ([]*domain.Microdistrict, error) {
	return uc.locationRepo.GetAllMicrodistricts()
}

func (uc *LocationUseCase) GetMicrodistrictsByDistrictID(districtID int) ([]*domain.Microdistrict, error) {
	return uc.locationRepo.GetMicrodistrictsByDistrictID(districtID)
}

func (uc *LocationUseCase) GetMicrodistrictByID(id int) (*domain.Microdistrict, error) {
	return uc.locationRepo.GetMicrodistrictByID(id)
}
