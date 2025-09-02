package usecase

import (
	"fmt"
	"time"

	"github.com/russo2642/renti_kz/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCase struct {
	userRepo     domain.UserRepository
	roleRepo     domain.RoleRepository
	passwordSalt string
}

func NewUserUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	passwordSalt string,
) *UserUseCase {
	return &UserUseCase{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		passwordSalt: passwordSalt,
	}
}

func (uc *UserUseCase) Register(user *domain.User, password string) error {
	existingUser, err := uc.userRepo.GetByPhone(user.Phone)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if existingUser != nil {
		return fmt.Errorf("user with phone %s already exists", user.Phone)
	}

	existingUser, err = uc.userRepo.GetByEmail(user.Email)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if existingUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	hashedPassword, err := uc.hashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = hashedPassword

	if user.RoleID == 0 && user.Role != "" {
		role, err := uc.roleRepo.GetByName(string(user.Role))
		if err != nil {
			return fmt.Errorf("failed to get role: %w", err)
		}
		user.RoleID = role.ID
	}

	if err := uc.userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (uc *UserUseCase) RegisterWithoutPassword(user *domain.User) error {
	existingUser, err := uc.userRepo.GetByPhone(user.Phone)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if existingUser != nil {
		return fmt.Errorf("user with phone %s already exists", user.Phone)
	}

	existingUser, err = uc.userRepo.GetByEmail(user.Email)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if existingUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	user.PasswordHash = ""

	if user.RoleID == 0 && user.Role != "" {
		role, err := uc.roleRepo.GetByName(string(user.Role))
		if err != nil {
			return fmt.Errorf("failed to get role: %w", err)
		}
		user.RoleID = role.ID
	}

	if err := uc.userRepo.Create(user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (uc *UserUseCase) GetByID(id int) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

func (uc *UserUseCase) GetByPhone(phone string) (*domain.User, error) {
	user, err := uc.userRepo.GetByPhone(phone)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}
	return user, nil
}

func (uc *UserUseCase) GetByEmail(email string) (*domain.User, error) {
	user, err := uc.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (uc *UserUseCase) UpdateProfile(user *domain.User) error {
	existingUser, err := uc.userRepo.GetByID(user.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return fmt.Errorf("user with id %d not found", user.ID)
	}

	existingUser.FirstName = user.FirstName
	existingUser.LastName = user.LastName
	existingUser.Email = user.Email
	existingUser.CityID = user.CityID
	existingUser.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(existingUser); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (uc *UserUseCase) ChangePassword(userID int, oldPassword, newPassword string) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user with id %d not found", userID)
	}

	if !uc.checkPassword(oldPassword, user.PasswordHash) {
		return fmt.Errorf("incorrect old password")
	}

	hashedPassword, err := uc.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (uc *UserUseCase) DeleteUser(userID int, adminID int) error {
	admin, err := uc.userRepo.GetByID(adminID)
	if err != nil {
		return fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil {
		return fmt.Errorf("admin with id %d not found", adminID)
	}
	if admin.Role != domain.RoleAdmin {
		return fmt.Errorf("only admins can delete users")
	}

	if err := uc.userRepo.Delete(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (uc *UserUseCase) DeleteOwnAccount(userID int) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user with id %d not found", userID)
	}

	if err := uc.userRepo.Delete(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (uc *UserUseCase) GetAllUsers(filters map[string]interface{}, page, pageSize int) ([]*domain.User, int, error) {
	users, total, err := uc.userRepo.GetAll(filters, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}
	return users, total, nil
}

func (uc *UserUseCase) UpdateUserRole(userID int, role domain.UserRole, adminID int) error {
	admin, err := uc.userRepo.GetByID(adminID)
	if err != nil {
		return fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil {
		return fmt.Errorf("admin with id %d not found", adminID)
	}
	if admin.Role != domain.RoleAdmin {
		return fmt.Errorf("only admins can change user roles")
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user with id %d not found", userID)
	}

	roleEntity, err := uc.roleRepo.GetByName(string(role))
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if roleEntity == nil {
		return fmt.Errorf("role %s not found", role)
	}

	if err := uc.userRepo.UpdateRole(userID, role); err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	return nil
}

func (uc *UserUseCase) UpdateUserStatus(userID int, isActive bool, reason string, adminID int) error {
	admin, err := uc.userRepo.GetByID(adminID)
	if err != nil {
		return fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil {
		return fmt.Errorf("admin with id %d not found", adminID)
	}
	if admin.Role != domain.RoleAdmin {
		return fmt.Errorf("only admins can change user status")
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user with id %d not found", userID)
	}

	if userID == adminID {
		return fmt.Errorf("admin cannot change their own status")
	}

	if err := uc.userRepo.UpdateStatus(userID, isActive); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

func (uc *UserUseCase) hashPassword(password string) (string, error) {
	saltedPassword := password + uc.passwordSalt

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

func (uc *UserUseCase) checkPassword(password, hash string) bool {
	saltedPassword := password + uc.passwordSalt

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword))
	return err == nil
}

func (uc *UserUseCase) GetRoleStatistics() (map[string]int, error) {
	return uc.userRepo.GetRoleStatistics()
}

func (uc *UserUseCase) GetStatusStatistics() (map[string]int, error) {
	return uc.userRepo.GetStatusStatistics()
}
