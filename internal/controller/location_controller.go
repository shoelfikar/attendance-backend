package controller

import (
	"net/http"
	"strconv"

	"github.com/attendance/backend/internal/service"
	"github.com/attendance/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type LocationController struct {
	locationService *service.LocationService
}

func NewLocationController(locationService *service.LocationService) *LocationController {
	return &LocationController{
		locationService: locationService,
	}
}

// GetNearbyLocations godoc
// @Summary Get nearby attendance locations
// @Tags locations
// @Produce json
// @Security BearerAuth
// @Param latitude query float64 true "User latitude"
// @Param longitude query float64 true "User longitude"
// @Param radius_km query float64 true "Search radius in km"
// @Success 200 {object} utils.Response
// @Router /api/v1/attendance/locations [get]
func (ctrl *LocationController) GetNearbyLocations(c *gin.Context) {
	var req service.GetNearbyLocationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	locations, err := ctrl.locationService.GetNearbyLocations(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get locations", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Nearby locations retrieved", locations)
}

// ValidateLocation godoc
// @Summary Validate if user is within location radius
// @Tags locations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/attendance/validate-location [post]
func (ctrl *LocationController) ValidateLocation(c *gin.Context) {
	var req struct {
		LocationID uint    `json:"location_id" binding:"required"`
		Latitude   float64 `json:"latitude" binding:"required"`
		Longitude  float64 `json:"longitude" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	isValid, distance, err := ctrl.locationService.ValidateLocationForAttendance(
		req.LocationID,
		req.Latitude,
		req.Longitude,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Location validated", gin.H{
		"is_valid": isValid,
		"distance": distance,
	})
}

// CreateLocation godoc
// @Summary Create new attendance location (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateLocationRequest true "Create location request"
// @Success 201 {object} utils.Response
// @Router /api/v1/admin/locations [post]
func (ctrl *LocationController) CreateLocation(c *gin.Context) {
	var req service.CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	// Get userID from context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID type", nil)
		return
	}

	location, err := ctrl.locationService.CreateLocation(&req, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create location", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Location created successfully", location.ToResponse())
}

// GetAllLocations godoc
// @Summary Get all locations (Admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/locations [get]
func (ctrl *LocationController) GetAllLocations(c *gin.Context) {
	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		activeBool, _ := strconv.ParseBool(activeStr)
		isActive = &activeBool
	}

	locations, err := ctrl.locationService.GetAllLocations(isActive)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get locations", err.Error())
		return
	}

	// Convert to responses
	responses := make([]interface{}, len(locations))
	for i, loc := range locations {
		responses[i] = loc.ToResponse()
	}

	utils.SuccessResponse(c, http.StatusOK, "Locations retrieved", responses)
}

// GetLocationByID godoc
// @Summary Get location by ID (Admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Location ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/locations/:id [get]
func (ctrl *LocationController) GetLocationByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid location ID", err.Error())
		return
	}

	location, err := ctrl.locationService.GetLocationByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Location not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Location retrieved", location.ToResponse())
}

// UpdateLocation godoc
// @Summary Update location (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Location ID"
// @Param request body service.UpdateLocationRequest true "Update location request"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/locations/:id [put]
func (ctrl *LocationController) UpdateLocation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid location ID", err.Error())
		return
	}

	var req service.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	location, err := ctrl.locationService.UpdateLocation(uint(id), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update location", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Location updated successfully", location.ToResponse())
}

// DeleteLocation godoc
// @Summary Delete location (Admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Location ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/locations/:id [delete]
func (ctrl *LocationController) DeleteLocation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid location ID", err.Error())
		return
	}

	if err := ctrl.locationService.DeleteLocation(uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete location", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Location deleted successfully", nil)
}
