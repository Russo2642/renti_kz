package postgres

import (
	"database/sql"
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type LocationRepository struct {
	db *sql.DB
}

func NewLocationRepository(db *sql.DB) *LocationRepository {
	return &LocationRepository{
		db: db,
	}
}

func (r *LocationRepository) GetAllRegions() ([]*domain.Region, error) {
	regions := make([]*domain.Region, 0)

	query := `
		SELECT id, name, created_at, updated_at
		FROM regions
		ORDER BY name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, utils.HandleSQLError(err, "regions", "get")
	}
	defer utils.CloseRows(rows)

	for rows.Next() {
		region := &domain.Region{}
		err := rows.Scan(
			&region.ID, &region.Name, &region.CreatedAt, &region.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "region", "scan")
		}
		regions = append(regions, region)
	}

	if err = utils.CheckRowsError(rows, "regions iteration"); err != nil {
		return nil, err
	}

	return regions, nil
}

func (r *LocationRepository) GetRegionByID(id int) (*domain.Region, error) {
	region := &domain.Region{}

	query := `
		SELECT id, name, created_at, updated_at
		FROM regions
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&region.ID, &region.Name, &region.CreatedAt, &region.UpdatedAt,
	)

	if err != nil {
		return nil, utils.HandleSQLErrorWithID(err, "region", "get", id)
	}

	return region, nil
}

func (r *LocationRepository) GetAllCities() ([]*domain.City, error) {
	cities := make([]*domain.City, 0)

	query := `
		SELECT c.id, c.name, c.region_id, c.created_at, c.updated_at,
		       r.id, r.name, r.created_at, r.updated_at,
		       cc.id, cc.latitude, cc.longitude, cc.created_at, cc.updated_at
		FROM cities c
		JOIN regions r ON c.region_id = r.id
		LEFT JOIN city_coordinates cc ON c.id = cc.city_id
		ORDER BY c.name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, utils.HandleSQLError(err, "cities", "get")
	}
	defer utils.CloseRows(rows)

	for rows.Next() {
		city := &domain.City{}
		region := &domain.Region{}
		var coordinatesID sql.NullInt64
		var latitude sql.NullFloat64
		var longitude sql.NullFloat64
		var coordCreatedAt sql.NullTime
		var coordUpdatedAt sql.NullTime

		err := rows.Scan(
			&city.ID, &city.Name, &city.RegionID, &city.CreatedAt, &city.UpdatedAt,
			&region.ID, &region.Name, &region.CreatedAt, &region.UpdatedAt,
			&coordinatesID, &latitude, &longitude, &coordCreatedAt, &coordUpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "city", "scan")
		}

		city.Region = region

		if coordinatesID.Valid && latitude.Valid && longitude.Valid {
			city.Coordinates = &domain.Coordinates{
				ID:        int(coordinatesID.Int64),
				CityID:    city.ID,
				Latitude:  latitude.Float64,
				Longitude: longitude.Float64,
				CreatedAt: coordCreatedAt.Time,
				UpdatedAt: coordUpdatedAt.Time,
			}
		}

		cities = append(cities, city)
	}

	if err = utils.CheckRowsError(rows, "cities iteration"); err != nil {
		return nil, err
	}

	return cities, nil
}

func (r *LocationRepository) GetCitiesByRegionID(regionID int) ([]*domain.City, error) {
	cities := make([]*domain.City, 0)

	query := `
		SELECT c.id, c.name, c.region_id, c.created_at, c.updated_at,
		       cc.id, cc.latitude, cc.longitude, cc.created_at, cc.updated_at
		FROM cities c
		LEFT JOIN city_coordinates cc ON c.id = cc.city_id
		WHERE c.region_id = $1
		ORDER BY c.name
	`

	rows, err := r.db.Query(query, regionID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "cities by region", "get")
	}
	defer utils.CloseRows(rows)

	for rows.Next() {
		city := &domain.City{RegionID: regionID}
		var coordinatesID sql.NullInt64
		var latitude sql.NullFloat64
		var longitude sql.NullFloat64
		var coordCreatedAt sql.NullTime
		var coordUpdatedAt sql.NullTime

		err := rows.Scan(
			&city.ID, &city.Name, &city.RegionID, &city.CreatedAt, &city.UpdatedAt,
			&coordinatesID, &latitude, &longitude, &coordCreatedAt, &coordUpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "city", "scan")
		}

		if coordinatesID.Valid && latitude.Valid && longitude.Valid {
			city.Coordinates = &domain.Coordinates{
				ID:        int(coordinatesID.Int64),
				CityID:    city.ID,
				Latitude:  latitude.Float64,
				Longitude: longitude.Float64,
				CreatedAt: coordCreatedAt.Time,
				UpdatedAt: coordUpdatedAt.Time,
			}
		}

		cities = append(cities, city)
	}

	if err = utils.CheckRowsError(rows, "cities by region iteration"); err != nil {
		return nil, err
	}

	return cities, nil
}

func (r *LocationRepository) GetCityByID(id int) (*domain.City, error) {
	city := &domain.City{}
	region := &domain.Region{}
	var coordinatesID sql.NullInt64
	var latitude sql.NullFloat64
	var longitude sql.NullFloat64
	var coordCreatedAt sql.NullTime
	var coordUpdatedAt sql.NullTime

	query := `
		SELECT c.id, c.name, c.region_id, c.created_at, c.updated_at,
		       r.id, r.name, r.created_at, r.updated_at,
		       cc.id, cc.latitude, cc.longitude, cc.created_at, cc.updated_at
		FROM cities c
		JOIN regions r ON c.region_id = r.id
		LEFT JOIN city_coordinates cc ON c.id = cc.city_id
		WHERE c.id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&city.ID, &city.Name, &city.RegionID, &city.CreatedAt, &city.UpdatedAt,
		&region.ID, &region.Name, &region.CreatedAt, &region.UpdatedAt,
		&coordinatesID, &latitude, &longitude, &coordCreatedAt, &coordUpdatedAt,
	)

	if err != nil {
		return nil, utils.HandleSQLErrorWithID(err, "city", "get", id)
	}

	city.Region = region

	if coordinatesID.Valid && latitude.Valid && longitude.Valid {
		city.Coordinates = &domain.Coordinates{
			ID:        int(coordinatesID.Int64),
			CityID:    city.ID,
			Latitude:  latitude.Float64,
			Longitude: longitude.Float64,
			CreatedAt: coordCreatedAt.Time,
			UpdatedAt: coordUpdatedAt.Time,
		}
	}

	return city, nil
}

func (r *LocationRepository) GetAllDistricts() ([]*domain.District, error) {
	districts := make([]*domain.District, 0)

	query := `
		SELECT d.id, d.name, d.city_id, d.created_at, d.updated_at,
		       c.id, c.name, c.region_id, c.created_at, c.updated_at
		FROM districts d
		JOIN cities c ON d.city_id = c.id
		ORDER BY d.name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, utils.HandleSQLError(err, "districts", "get")
	}
	defer utils.CloseRows(rows)

	for rows.Next() {
		district := &domain.District{}
		city := &domain.City{}

		err := rows.Scan(
			&district.ID, &district.Name, &district.CityID, &district.CreatedAt, &district.UpdatedAt,
			&city.ID, &city.Name, &city.RegionID, &city.CreatedAt, &city.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "district", "scan")
		}

		district.City = city
		districts = append(districts, district)
	}

	if err = utils.CheckRowsError(rows, "districts iteration"); err != nil {
		return nil, err
	}

	return districts, nil
}

func (r *LocationRepository) GetDistrictsByCityID(cityID int) ([]*domain.District, error) {
	districts := make([]*domain.District, 0)

	query := `
		SELECT id, name, city_id, created_at, updated_at
		FROM districts
		WHERE city_id = $1
		ORDER BY name
	`

	rows, err := r.db.Query(query, cityID)
	if err != nil {
		return nil, utils.HandleSQLError(err, "districts by city", "get")
	}
	defer utils.CloseRows(rows)

	for rows.Next() {
		district := &domain.District{CityID: cityID}
		err := rows.Scan(
			&district.ID, &district.Name, &district.CityID, &district.CreatedAt, &district.UpdatedAt,
		)
		if err != nil {
			return nil, utils.HandleSQLError(err, "district", "scan")
		}
		districts = append(districts, district)
	}

	if err = utils.CheckRowsError(rows, "districts by city iteration"); err != nil {
		return nil, err
	}

	return districts, nil
}

func (r *LocationRepository) GetDistrictByID(id int) (*domain.District, error) {
	district := &domain.District{}
	city := &domain.City{}

	query := `
		SELECT d.id, d.name, d.city_id, d.created_at, d.updated_at,
		       c.id, c.name, c.region_id, c.created_at, c.updated_at
		FROM districts d
		JOIN cities c ON d.city_id = c.id
		WHERE d.id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&district.ID, &district.Name, &district.CityID, &district.CreatedAt, &district.UpdatedAt,
		&city.ID, &city.Name, &city.RegionID, &city.CreatedAt, &city.UpdatedAt,
	)

	if err != nil {
		return nil, utils.HandleSQLErrorWithID(err, "district", "get", id)
	}

	district.City = city
	return district, nil
}

func (r *LocationRepository) GetAllMicrodistricts() ([]*domain.Microdistrict, error) {
	microdistricts := make([]*domain.Microdistrict, 0)

	query := `
		SELECT m.id, m.name, m.district_id, m.created_at, m.updated_at,
		       d.id, d.name, d.city_id, d.created_at, d.updated_at
		FROM microdistricts m
		JOIN districts d ON m.district_id = d.id
		ORDER BY m.name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get microdistricts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		microdistrict := &domain.Microdistrict{}
		district := &domain.District{}

		err := rows.Scan(
			&microdistrict.ID, &microdistrict.Name, &microdistrict.DistrictID, &microdistrict.CreatedAt, &microdistrict.UpdatedAt,
			&district.ID, &district.Name, &district.CityID, &district.CreatedAt, &district.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan microdistrict: %w", err)
		}

		microdistrict.District = district
		microdistricts = append(microdistricts, microdistrict)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return microdistricts, nil
}

func (r *LocationRepository) GetMicrodistrictsByDistrictID(districtID int) ([]*domain.Microdistrict, error) {
	microdistricts := make([]*domain.Microdistrict, 0)

	query := `
		SELECT id, name, district_id, created_at, updated_at
		FROM microdistricts
		WHERE district_id = $1
		ORDER BY name
	`

	rows, err := r.db.Query(query, districtID)
	if err != nil {
		return nil, fmt.Errorf("failed to get microdistricts by district: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		microdistrict := &domain.Microdistrict{DistrictID: districtID}
		err := rows.Scan(
			&microdistrict.ID, &microdistrict.Name, &microdistrict.DistrictID, &microdistrict.CreatedAt, &microdistrict.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan microdistrict: %w", err)
		}
		microdistricts = append(microdistricts, microdistrict)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return microdistricts, nil
}

func (r *LocationRepository) GetMicrodistrictByID(id int) (*domain.Microdistrict, error) {
	microdistrict := &domain.Microdistrict{}
	district := &domain.District{}

	query := `
		SELECT m.id, m.name, m.district_id, m.created_at, m.updated_at,
		       d.id, d.name, d.city_id, d.created_at, d.updated_at
		FROM microdistricts m
		JOIN districts d ON m.district_id = d.id
		WHERE m.id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&microdistrict.ID, &microdistrict.Name, &microdistrict.DistrictID, &microdistrict.CreatedAt, &microdistrict.UpdatedAt,
		&district.ID, &district.Name, &district.CityID, &district.CreatedAt, &district.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("microdistrict with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get microdistrict: %w", err)
	}

	microdistrict.District = district
	return microdistrict, nil
}
