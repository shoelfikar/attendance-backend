package controller

import (
	"net/http"
	"strconv"

	"github.com/attendance/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Get all users (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/users [get]
func (ctrl *UserController) GetAllUsers(c *gin.Context) {
	users, err := ctrl.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to retrieve users",
			"error":   err.Error(),
		})
		return
	}

	// Convert to response format (without password hash)
	var userResponses []interface{}
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Users retrieved successfully",
		"data":    userResponses,
	})
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get a specific user by ID (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id} [get]
func (ctrl *UserController) GetUserByID(c *gin.Context) {
	// Parse user ID
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}

	user, err := ctrl.userService.GetUserByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User retrieved successfully",
		"data":    user.ToResponse(),
	})
}

// CreateUser godoc
// @Summary Create new user
// @Description Create a new user (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body service.CreateUserRequest true "User data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/users [post]
func (ctrl *UserController) CreateUser(c *gin.Context) {
	var req service.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	user, err := ctrl.userService.CreateUser(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "email already exists" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "User created successfully",
		"data":    user.ToResponse(),
	})
}

// UpdateUser godoc
// @Summary Update user
// @Description Update an existing user (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param user body service.UpdateUserRequest true "User data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id} [put]
func (ctrl *UserController) UpdateUser(c *gin.Context) {
	// Parse user ID
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	user, err := ctrl.userService.UpdateUser(uint(userID), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "email already exists" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User updated successfully",
		"data":    user.ToResponse(),
	})
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id} [delete]
func (ctrl *UserController) DeleteUser(c *gin.Context) {
	// Parse user ID
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}

	// Prevent admin from deleting themselves
	currentUserID, exists := c.Get("userID")
	if exists && currentUserID.(uint) == uint(userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Cannot delete your own account",
		})
		return
	}

	err = ctrl.userService.DeleteUser(uint(userID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot delete the last admin user" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User deleted successfully",
	})
}

// ChangeUserPassword godoc
// @Summary Change user password
// @Description Change a user's password (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param password body service.ChangePasswordRequest true "New password"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id}/password [put]
func (ctrl *UserController) ChangeUserPassword(c *gin.Context) {
	// Parse user ID
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	err = ctrl.userService.ChangeUserPassword(uint(userID), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Password changed successfully",
	})
}

// GetUserStats godoc
// @Summary Get user statistics
// @Description Get statistics about users (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /admin/users/stats [get]
func (ctrl *UserController) GetUserStats(c *gin.Context) {
	stats, err := ctrl.userService.GetUserStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to retrieve user statistics",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User statistics retrieved successfully",
		"data":    stats,
	})
}

// GetMyProfile godoc
// @Summary Get my profile
// @Description Get authenticated user's profile
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /admin/profile [get]
func (ctrl *UserController) GetMyProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	user, err := ctrl.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile retrieved successfully",
		"data":    user.ToResponse(),
	})
}

// UpdateMyProfile godoc
// @Summary Update my profile
// @Description Update authenticated user's profile
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body service.UpdateMyProfileRequest true "Profile data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /admin/profile [put]
func (ctrl *UserController) UpdateMyProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	var req service.UpdateMyProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	user, err := ctrl.userService.UpdateMyProfile(userID.(uint), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "email already exists" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile updated successfully",
		"data":    user.ToResponse(),
	})
}

// UpdateMyPassword godoc
// @Summary Update my password
// @Description Update authenticated user's password
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body service.UpdateMyPasswordRequest true "Password data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /admin/profile/password [put]
func (ctrl *UserController) UpdateMyPassword(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	var req service.UpdateMyPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	err := ctrl.userService.UpdateMyPassword(userID.(uint), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "old password is incorrect" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Password updated successfully",
	})
}
