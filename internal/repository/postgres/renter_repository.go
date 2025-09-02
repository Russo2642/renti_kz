package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"github.com/russo2642/renti_kz/internal/utils"
)

type RenterRepository struct {
	db *sql.DB
}

func NewRenterRepository(db *sql.DB) *RenterRepository {
	return &RenterRepository{
		db: db,
	}
}

func (r *RenterRepository) marshalDocumentURL(documentURL map[string]string) ([]byte, error) {
	return json.Marshal(documentURL)
}

func (r *RenterRepository) unmarshalDocumentURL(documentURLJSON []byte) (map[string]string, error) {
	if len(documentURLJSON) == 0 {
		return make(map[string]string), nil
	}

	var docURLMap map[string]string
	if err := json.Unmarshal(documentURLJSON, &docURLMap); err != nil {
		return nil, utils.HandleSQLError(err, "document URL JSON", "unmarshal")
	}
	return docURLMap, nil
}

func (r *RenterRepository) scanRenterWithDocuments(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Renter, error) {
	renter := &domain.Renter{}
	var documentURLJSON []byte

	err := scanner.Scan(
		&renter.ID, &renter.UserID, &renter.DocumentType, &documentURLJSON, &renter.PhotoWithDocURL,
		&renter.VerificationStatus, &renter.CreatedAt, &renter.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	documentURL, err := r.unmarshalDocumentURL(documentURLJSON)
	if err != nil {
		return nil, err
	}
	renter.DocumentURL = documentURL

	return renter, nil
}

func (r *RenterRepository) Create(renter *domain.Renter) error {
	now := time.Now()
	renter.CreatedAt = now
	renter.UpdatedAt = now

	documentURLJSON, err := r.marshalDocumentURL(renter.DocumentURL)
	if err != nil {
		return utils.HandleSQLError(err, "document URL", "marshal to JSON")
	}

	query := `
		INSERT INTO renters (
			user_id, document_type, document_url, photo_with_doc_url,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id`

	err = r.db.QueryRow(
		query,
		renter.UserID, renter.DocumentType, documentURLJSON, renter.PhotoWithDocURL,
		renter.CreatedAt, renter.UpdatedAt,
	).Scan(&renter.ID)

	if err != nil {
		return utils.HandleSQLError(err, "renter", "create")
	}

	return nil
}

func (r *RenterRepository) GetByID(id int) (*domain.Renter, error) {
	query := `
		SELECT 
			id, user_id, document_type, document_url, photo_with_doc_url, verification_status,
			created_at, updated_at
		FROM renters
		WHERE id = $1`

	renter, err := r.scanRenterWithDocuments(r.db.QueryRow(query, id))
	if err != nil {
		return nil, utils.HandleSQLErrorWithID(err, "renter", "get", id)
	}

	return renter, nil
}

func (r *RenterRepository) GetByUserID(userID int) (*domain.Renter, error) {
	query := `
		SELECT 
			id, user_id, document_type, document_url, photo_with_doc_url, verification_status,
			created_at, updated_at
		FROM renters
		WHERE user_id = $1`

	renter, err := r.scanRenterWithDocuments(r.db.QueryRow(query, userID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "renter by user", "get")
	}

	return renter, nil
}

func (r *RenterRepository) GetByUserIDWithUser(userID int) (*domain.Renter, error) {
	renter := &domain.Renter{}
	renter.User = &domain.User{}
	var documentURLJSON []byte

	query := `
		SELECT 
			rt.id, rt.user_id, rt.document_type, rt.document_url, rt.photo_with_doc_url, rt.verification_status,
			rt.created_at, rt.updated_at,
			u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, 
			u.role_id, u.password_hash, u.created_at, u.updated_at, ur.name
		FROM renters rt
		JOIN users u ON rt.user_id = u.id
		JOIN user_roles ur ON u.role_id = ur.id
		WHERE rt.user_id = $1`

	err := r.db.QueryRow(query, userID).Scan(
		&renter.ID, &renter.UserID, &renter.DocumentType, &documentURLJSON, &renter.PhotoWithDocURL,
		&renter.VerificationStatus, &renter.CreatedAt, &renter.UpdatedAt,
		&renter.User.ID, &renter.User.Phone, &renter.User.FirstName, &renter.User.LastName, &renter.User.Email,
		&renter.User.CityID, &renter.User.IIN, &renter.User.RoleID, &renter.User.PasswordHash,
		&renter.User.CreatedAt, &renter.User.UpdatedAt, &renter.User.Role,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, utils.HandleSQLError(err, "renter with user", "get")
	}

	documentURL, err := r.unmarshalDocumentURL(documentURLJSON)
	if err != nil {
		return nil, err
	}
	renter.DocumentURL = documentURL

	return renter, nil
}

func (r *RenterRepository) GetByIDWithUser(id int) (*domain.Renter, error) {
	renter := &domain.Renter{}
	renter.User = &domain.User{}
	var documentURLJSON []byte

	query := `
		SELECT 
			rt.id, rt.user_id, rt.document_type, rt.document_url, rt.photo_with_doc_url, rt.verification_status,
			rt.created_at, rt.updated_at,
			u.id, u.phone, u.first_name, u.last_name, u.email, u.city_id, u.iin, 
			u.role_id, u.password_hash, u.created_at, u.updated_at, ur.name
		FROM renters rt
		JOIN users u ON rt.user_id = u.id
		JOIN user_roles ur ON u.role_id = ur.id
		WHERE rt.id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&renter.ID, &renter.UserID, &renter.DocumentType, &documentURLJSON, &renter.PhotoWithDocURL,
		&renter.VerificationStatus, &renter.CreatedAt, &renter.UpdatedAt,
		&renter.User.ID, &renter.User.Phone, &renter.User.FirstName, &renter.User.LastName, &renter.User.Email,
		&renter.User.CityID, &renter.User.IIN, &renter.User.RoleID, &renter.User.PasswordHash,
		&renter.User.CreatedAt, &renter.User.UpdatedAt, &renter.User.Role,
	)

	if err != nil {
		return nil, utils.HandleSQLErrorWithID(err, "renter with user", "get", id)
	}

	documentURL, err := r.unmarshalDocumentURL(documentURLJSON)
	if err != nil {
		return nil, err
	}
	renter.DocumentURL = documentURL

	return renter, nil
}

func (r *RenterRepository) Update(renter *domain.Renter) error {
	renter.UpdatedAt = time.Now()

	documentURLJSON, err := r.marshalDocumentURL(renter.DocumentURL)
	if err != nil {
		return utils.HandleSQLError(err, "document URL", "marshal to JSON")
	}

	query := `
		UPDATE renters
		SET 
			document_type = $2, document_url = $3, photo_with_doc_url = $4,
			verification_status = $5, updated_at = $6
		WHERE id = $1`

	_, err = r.db.Exec(
		query,
		renter.ID, renter.DocumentType, documentURLJSON, renter.PhotoWithDocURL,
		renter.VerificationStatus, renter.UpdatedAt,
	)

	if err != nil {
		return utils.HandleSQLErrorWithID(err, "renter", "update", renter.ID)
	}

	return nil
}

func (r *RenterRepository) Delete(id int) error {
	query := `DELETE FROM renters WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return utils.HandleSQLErrorWithID(err, "renter", "delete", id)
	}

	return nil
}
