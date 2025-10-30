package service

import (
	"errors"

	"github.com/attendance/backend/internal/config"
	"github.com/attendance/backend/internal/model"
	"github.com/attendance/backend/pkg/jwt"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
)

type AuthService struct {
	db     *gorm.DB
	config *config.Config
}

func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:     db,
		config: cfg,
	}
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Phone    string `json:"phone"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         model.UserResponse `json:"user"`
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
}

// Register creates a new user account
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Check if email already exists
	var existingUser model.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, ErrEmailAlreadyExists
	}

	// Create new user
	user := model.User{
		Email:    req.Email,
		FullName: req.FullName,
		Phone:    req.Phone,
		Role:     "user",
		IsActive: true,
	}

	// Hash password
	if err := user.HashPassword(req.Password); err != nil {
		return nil, err
	}

	// Save to database
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// Generate tokens
	tokens, err := jwt.GenerateTokenPair(
		user.ID,
		user.Email,
		user.Role,
		s.config.JWT.Secret,
		s.config.JWT.Expiration,
		s.config.JWT.RefreshExpiration,
	)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Find user by email
	var user model.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokens, err := jwt.GenerateTokenPair(
		user.ID,
		user.Email,
		user.Role,
		s.config.JWT.Secret,
		s.config.JWT.Expiration,
		s.config.JWT.RefreshExpiration,
	)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

// GetUserByID retrieves user by ID
func (s *AuthService) GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// RefreshToken generates new access token from refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*jwt.TokenPair, error) {
	// Validate refresh token
	claims, err := jwt.ValidateToken(refreshToken, s.config.JWT.Secret)
	if err != nil {
		return nil, err
	}

	// Get user to ensure still active
	user, err := s.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Generate new token pair
	return jwt.GenerateTokenPair(
		user.ID,
		user.Email,
		user.Role,
		s.config.JWT.Secret,
		s.config.JWT.Expiration,
		s.config.JWT.RefreshExpiration,
	)
}
