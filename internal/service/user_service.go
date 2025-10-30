package service

import (
	"errors"
	"fmt"

	"github.com/attendance/backend/internal/model"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Phone    string `json:"phone"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Role     string `json:"role" binding:"omitempty,oneof=admin user"`
	IsActive *bool  `json:"is_active"`
}

// ChangePasswordRequest represents the request to change user password
type ChangePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdateMyProfileRequest represents the request to update own profile
type UpdateMyProfileRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

// UpdateMyPasswordRequest represents the request to update own password
type UpdateMyPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers() ([]model.User, error) {
	var users []model.User

	result := s.db.Order("created_at DESC").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(userID uint) (*model.User, error) {
	var user model.User

	result := s.db.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	var user model.User

	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(req *CreateUserRequest) (*model.User, error) {
	// Check if email already exists
	var existingUser model.User
	result := s.db.Where("email = ?", req.Email).First(&existingUser)
	if result.Error == nil {
		return nil, errors.New("email already exists")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// Create new user
	user := &model.User{
		Email:    req.Email,
		FullName: req.FullName,
		Phone:    req.Phone,
		Role:     req.Role,
		IsActive: true,
	}

	// Hash password
	if err := user.HashPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Save to database
	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(userID uint, req *UpdateUserRequest) (*model.User, error) {
	// Get user
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Check if email is being changed and already exists
	if req.Email != "" && req.Email != user.Email {
		var existingUser model.User
		result := s.db.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser)
		if result.Error == nil {
			return nil, errors.New("email already exists")
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		user.Email = req.Email
	}

	// Update fields
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Save changes
	if err := s.db.Save(user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(userID uint) error {
	// Get user to ensure it exists
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Prevent deleting the last admin
	if user.Role == "admin" {
		var adminCount int64
		s.db.Model(&model.User{}).Where("role = ?", "admin").Count(&adminCount)
		if adminCount <= 1 {
			return errors.New("cannot delete the last admin user")
		}
	}

	// Delete user
	if err := s.db.Delete(user).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ChangeUserPassword changes a user's password
func (s *UserService) ChangeUserPassword(userID uint, req *ChangePasswordRequest) error {
	// Get user
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Hash new password
	if err := user.HashPassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Save changes
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	return nil
}

// GetUserStats returns user statistics
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	var totalUsers int64
	var activeUsers int64
	var adminUsers int64
	var regularUsers int64

	s.db.Model(&model.User{}).Count(&totalUsers)
	s.db.Model(&model.User{}).Where("is_active = ?", true).Count(&activeUsers)
	s.db.Model(&model.User{}).Where("role = ?", "admin").Count(&adminUsers)
	s.db.Model(&model.User{}).Where("role = ?", "user").Count(&regularUsers)

	stats := map[string]interface{}{
		"total_users":   totalUsers,
		"active_users":  activeUsers,
		"admin_users":   adminUsers,
		"regular_users": regularUsers,
		"inactive_users": totalUsers - activeUsers,
	}

	return stats, nil
}

// UpdateMyProfile updates the authenticated user's profile
func (s *UserService) UpdateMyProfile(userID uint, req *UpdateMyProfileRequest) (*model.User, error) {
	// Get user
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Check if email is being changed and already exists
	if req.Email != "" && req.Email != user.Email {
		var existingUser model.User
		result := s.db.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser)
		if result.Error == nil {
			return nil, errors.New("email already exists")
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		user.Email = req.Email
	}

	// Update fields
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	// Save changes
	if err := s.db.Save(user).Error; err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return user, nil
}

// UpdateMyPassword updates the authenticated user's password
func (s *UserService) UpdateMyPassword(userID uint, req *UpdateMyPasswordRequest) error {
	// Get user
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Verify old password
	if !user.CheckPassword(req.OldPassword) {
		return errors.New("old password is incorrect")
	}

	// Hash new password
	if err := user.HashPassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Save changes
	if err := s.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
