package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/attendance/backend/internal/service"
	"github.com/attendance/backend/internal/utils"
	jwtPkg "github.com/attendance/backend/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Register godoc
// @Summary Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "Register request"
// @Success 201 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /api/v1/auth/register [post]
func (ctrl *AuthController) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := ctrl.authService.Register(&req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			utils.ErrorResponse(c, http.StatusConflict, "Email already exists", err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to register user", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", response)
}

// Login godoc
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "Login request"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /api/v1/auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := ctrl.authService.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials", err.Error())
			return
		}
		if errors.Is(err, service.ErrUserInactive) {
			utils.ErrorResponse(c, http.StatusForbidden, "User account is inactive", err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to login", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Refresh token"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/auth/refresh-token [post]
func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	// Get refresh token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header required", nil)
		return
	}

	// Extract token
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format", nil)
		return
	}

	refreshToken := tokenParts[1]

	// Generate new tokens
	tokens, err := ctrl.authService.RefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, jwtPkg.ErrInvalidToken) || errors.Is(err, jwtPkg.ErrExpiredToken) {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token", err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to refresh token", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Token refreshed successfully", tokens)
}

// GetMe godoc
// @Summary Get current user info
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/auth/me [get]
func (ctrl *AuthController) GetMe(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	user, err := ctrl.authService.GetUserByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User info retrieved", user.ToResponse())
}

// Logout godoc
// @Summary Logout user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/auth/logout [post]
func (ctrl *AuthController) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is handled client-side
	// by removing the token. For server-side logout, implement
	// token blacklisting with Redis
	utils.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}
