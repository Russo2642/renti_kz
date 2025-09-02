package postgres

import (
	"database/sql"
	"fmt"

	"github.com/russo2642/renti_kz/internal/domain"
)

type favoriteRepository struct {
	db *sql.DB
}

func NewFavoriteRepository(db *sql.DB) domain.FavoriteRepository {
	return &favoriteRepository{
		db: db,
	}
}

func (r *favoriteRepository) AddToFavorites(userID, apartmentID int) error {
	query := `
		INSERT INTO favorites (user_id, apartment_id) 
		VALUES ($1, $2)
		ON CONFLICT (user_id, apartment_id) DO NOTHING`

	_, err := r.db.Exec(query, userID, apartmentID)
	if err != nil {
		return fmt.Errorf("failed to add apartment to favorites: %w", err)
	}

	return nil
}

func (r *favoriteRepository) RemoveFromFavorites(userID, apartmentID int) error {
	query := `DELETE FROM favorites WHERE user_id = $1 AND apartment_id = $2`

	result, err := r.db.Exec(query, userID, apartmentID)
	if err != nil {
		return fmt.Errorf("failed to remove apartment from favorites: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("apartment not found in favorites")
	}

	return nil
}

func (r *favoriteRepository) GetUserFavorites(userID int, page, pageSize int) ([]*domain.Favorite, int, error) {
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`
	var totalCount int
	err := r.db.QueryRow(countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get favorites count: %w", err)
	}

	query := `
		SELECT 
			f.id, f.user_id, f.apartment_id, f.created_at, f.updated_at,
			a.id, a.owner_id, a.city_id, a.district_id, a.microdistrict_id,
			a.street, a.building, a.apartment_number, a.room_count, a.total_area,
			a.kitchen_area, a.floor, a.total_floors, a.condition_id, a.price,
			a.is_free, a.status, a.moderator_comment, a.description,
			a.created_at, a.updated_at,
			c.id, c.name, c.region_id,
			d.id, d.name, d.city_id,
			ac.id, ac.name, ac.description
		FROM favorites f
		JOIN apartments a ON f.apartment_id = a.id
		JOIN cities c ON a.city_id = c.id
		JOIN districts d ON a.district_id = d.id
		JOIN apartment_conditions ac ON a.condition_id = ac.id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user favorites: %w", err)
	}
	defer rows.Close()

	var favorites []*domain.Favorite
	for rows.Next() {
		favorite := &domain.Favorite{
			Apartment: &domain.Apartment{
				City:      &domain.City{},
				District:  &domain.District{},
				Condition: &domain.ApartmentCondition{},
			},
		}

		var microdistrictID sql.NullInt64
		var moderatorComment sql.NullString
		var description sql.NullString

		err := rows.Scan(
			&favorite.ID, &favorite.UserID, &favorite.ApartmentID,
			&favorite.CreatedAt, &favorite.UpdatedAt,
			&favorite.Apartment.ID, &favorite.Apartment.OwnerID,
			&favorite.Apartment.CityID, &favorite.Apartment.DistrictID,
			&microdistrictID, &favorite.Apartment.Street,
			&favorite.Apartment.Building, &favorite.Apartment.ApartmentNumber,
			&favorite.Apartment.RoomCount, &favorite.Apartment.TotalArea,
			&favorite.Apartment.KitchenArea, &favorite.Apartment.Floor,
			&favorite.Apartment.TotalFloors, &favorite.Apartment.ConditionID,
			&favorite.Apartment.Price, &favorite.Apartment.IsFree,
			&favorite.Apartment.Status, &moderatorComment, &description,
			&favorite.Apartment.CreatedAt, &favorite.Apartment.UpdatedAt,
			&favorite.Apartment.City.ID, &favorite.Apartment.City.Name,
			&favorite.Apartment.City.RegionID,
			&favorite.Apartment.District.ID, &favorite.Apartment.District.Name,
			&favorite.Apartment.District.CityID,
			&favorite.Apartment.Condition.ID, &favorite.Apartment.Condition.Name,
			&favorite.Apartment.Condition.Description,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan favorite: %w", err)
		}

		if microdistrictID.Valid {
			favorite.Apartment.MicrodistrictID = &[]int{int(microdistrictID.Int64)}[0]
		}

		if moderatorComment.Valid {
			favorite.Apartment.ModeratorComment = moderatorComment.String
		}

		if description.Valid {
			favorite.Apartment.Description = description.String
		}

		favorite.Apartment.IsFavorite = true

		favorites = append(favorites, favorite)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating over favorites: %w", err)
	}

	return favorites, totalCount, nil
}

func (r *favoriteRepository) IsFavorite(userID, apartmentID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND apartment_id = $2)`

	var exists bool
	err := r.db.QueryRow(query, userID, apartmentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if apartment is favorite: %w", err)
	}

	return exists, nil
}

func (r *favoriteRepository) GetFavoriteCount(apartmentID int) (int, error) {
	query := `SELECT COUNT(*) FROM favorites WHERE apartment_id = $1`

	var count int
	err := r.db.QueryRow(query, apartmentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get favorite count: %w", err)
	}

	return count, nil
}

func (r *favoriteRepository) GetByID(id int) (*domain.Favorite, error) {
	query := `
		SELECT id, user_id, apartment_id, created_at, updated_at 
		FROM favorites 
		WHERE id = $1`

	favorite := &domain.Favorite{}
	err := r.db.QueryRow(query, id).Scan(
		&favorite.ID, &favorite.UserID, &favorite.ApartmentID,
		&favorite.CreatedAt, &favorite.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("favorite not found")
		}
		return nil, fmt.Errorf("failed to get favorite by ID: %w", err)
	}

	return favorite, nil
}

func (r *favoriteRepository) GetByUserAndApartment(userID, apartmentID int) (*domain.Favorite, error) {
	query := `
		SELECT id, user_id, apartment_id, created_at, updated_at 
		FROM favorites 
		WHERE user_id = $1 AND apartment_id = $2`

	favorite := &domain.Favorite{}
	err := r.db.QueryRow(query, userID, apartmentID).Scan(
		&favorite.ID, &favorite.UserID, &favorite.ApartmentID,
		&favorite.CreatedAt, &favorite.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("favorite not found")
		}
		return nil, fmt.Errorf("failed to get favorite by user and apartment: %w", err)
	}

	return favorite, nil
}
