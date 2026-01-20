package domain

import (
	"time"
)

type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleOwner     UserRole = "owner"
	RoleModerator UserRole = "moderator"
	RoleAdmin     UserRole = "admin"
	RoleConcierge UserRole = "concierge"
	RoleCleaner   UserRole = "cleaner"
)

type Role struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type DocumentType string

const (
	DocTypeID       DocumentType = "udv"
	DocTypePassport DocumentType = "passport"
)

type VerificationStatus string

const (
	VerificationPending  VerificationStatus = "pending"
	VerificationApproved VerificationStatus = "approved"
	VerificationRejected VerificationStatus = "rejected"
)

type User struct {
	ID           int       `json:"id"`
	Phone        string    `json:"phone"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	CityID       int       `json:"city_id"`
	City         *City     `json:"city,omitempty"`
	IIN          string    `json:"iin"`
	RoleID       int       `json:"role_id"`
	Role         UserRole  `json:"role"`
	IsActive     bool      `json:"is_active"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PropertyOwner struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      *User     `json:"user,omitempty"`
}

type Renter struct {
	ID                 int                `json:"id"`
	UserID             int                `json:"user_id"`
	DocumentType       DocumentType       `json:"document_type"`
	DocumentURL        map[string]string  `json:"document_url"`
	PhotoWithDocURL    string             `json:"photo_with_doc_url"`
	VerificationStatus VerificationStatus `json:"verification_status"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	User               *User              `json:"user,omitempty"`
}

type RoleRepository interface {
	GetAll() ([]*Role, error)
	GetByID(id int) (*Role, error)
	GetByName(name string) (*Role, error)
}

type UserRepository interface {
	Create(user *User) error
	GetByID(id int) (*User, error)
	GetByPhone(phone string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id int) error

	GetAll(filters map[string]interface{}, page, pageSize int) ([]*User, int, error)
	UpdateRole(userID int, role UserRole) error
	UpdateStatus(userID int, isActive bool) error
	GetRoleStatistics() (map[string]int, error)
	GetStatusStatistics() (map[string]int, error)
}

type PropertyOwnerRepository interface {
	Create(owner *PropertyOwner) error
	GetByID(id int) (*PropertyOwner, error)
	GetByIDWithUser(id int) (*PropertyOwner, error)
	GetByUserID(userID int) (*PropertyOwner, error)
	GetByUserIDWithUser(userID int) (*PropertyOwner, error)
	Update(owner *PropertyOwner) error
	Delete(id int) error
}

type RenterRepository interface {
	Create(renter *Renter) error
	GetByID(id int) (*Renter, error)
	GetByIDWithUser(id int) (*Renter, error)
	GetByUserID(userID int) (*Renter, error)
	GetByUserIDWithUser(userID int) (*Renter, error)
	Update(renter *Renter) error
	Delete(id int) error
}

type UserUseCase interface {
	Register(user *User, password string) error
	RegisterWithoutPassword(user *User) error
	GetByID(id int) (*User, error)
	GetByPhone(phone string) (*User, error)
	GetByEmail(email string) (*User, error)
	UpdateProfile(user *User) error
	ChangePassword(userID int, oldPassword, newPassword string) error
	DeleteUser(userID int, adminID int) error
	DeleteOwnAccount(userID int) error

	GetAllUsers(filters map[string]interface{}, page, pageSize int) ([]*User, int, error)
	UpdateUserRole(userID int, role UserRole, adminID int) error
	UpdateUserStatus(userID int, isActive bool, reason string, adminID int) error
	AdminSetPassword(userID int, newPassword string, adminID int) error
	GetRoleStatistics() (map[string]int, error)
	GetStatusStatistics() (map[string]int, error)
}

type PropertyOwnerUseCase interface {
	Register(owner *PropertyOwner, user *User, password string) error
	GetByID(id int) (*PropertyOwner, error)
	GetByUserID(userID int) (*PropertyOwner, error)
	UpdateProfile(owner *PropertyOwner) error
}

type RenterUseCase interface {
	Register(renter *Renter, user *User, password string) error
	GetByID(id int) (*Renter, error)
	GetByUserID(userID int) (*Renter, error)
	UpdateProfile(renter *Renter) error
	UploadDocument(renterID int, documentType DocumentType, fileData []byte) (string, error)
	UploadDocumentsParallel(renterID int, documentType DocumentType, documentsData [][]byte) (map[string]string, error)
	UploadPhotoWithDoc(renterID int, fileData []byte) (string, error)
	UpdateVerificationStatus(renterID int, status VerificationStatus) error
}
