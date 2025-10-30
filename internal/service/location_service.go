package service

import (
	"errors"

	"github.com/attendance/backend/internal/model"
	"github.com/attendance/backend/internal/utils"
	"gorm.io/gorm"
)

type LocationService struct {
	db *gorm.DB
}

func NewLocationService(db *gorm.DB) *LocationService {
	return &LocationService{db: db}
}

// CreateLocationRequest represents create location request
type CreateLocationRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude" binding:"required"`
	Longitude   float64 `json:"longitude" binding:"required"`
	Radius      int     `json:"radius" binding:"required,min=1"`
}

// UpdateLocationRequest represents update location request
type UpdateLocationRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Radius      int     `json:"radius" binding:"min=1"`
	IsActive    *bool   `json:"is_active"`
}

// GetNearbyLocationsRequest represents nearby locations request
type GetNearbyLocationsRequest struct {
	Latitude  float64 `form:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
	RadiusKm  float64 `form:"radius_km" binding:"required,min=0.1,max=50"` // max 50km
}

// CreateLocation creates a new attendance location
func (s *LocationService) CreateLocation(req *CreateLocationRequest, createdBy uint) (*model.AttendanceLocation, error) {
	location := model.AttendanceLocation{
		Name:        req.Name,
		Description: req.Description,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Radius:      req.Radius,
		IsActive:    true,
		CreatedBy:   &createdBy,
	}

	if err := s.db.Create(&location).Error; err != nil {
		return nil, err
	}

	// Load creator info
	s.db.Preload("Creator").First(&location, location.ID)

	return &location, nil
}

// GetLocationByID retrieves location by ID
func (s *LocationService) GetLocationByID(id uint) (*model.AttendanceLocation, error) {
	var location model.AttendanceLocation
	if err := s.db.Preload("Creator").First(&location, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("location not found")
		}
		return nil, err
	}
	return &location, nil
}

// GetAllLocations retrieves all locations with optional filters
func (s *LocationService) GetAllLocations(isActive *bool) ([]model.AttendanceLocation, error) {
	var locations []model.AttendanceLocation
	query := s.db.Preload("Creator")

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Find(&locations).Error; err != nil {
		return nil, err
	}

	return locations, nil
}

// GetNearbyLocations retrieves locations near user's current position
func (s *LocationService) GetNearbyLocations(req *GetNearbyLocationsRequest) ([]model.AttendanceLocation, error) {
	var allLocations []model.AttendanceLocation

	// Get all active locations
	if err := s.db.Where("is_active = ?", true).Find(&allLocations).Error; err != nil {
		return nil, err
	}

	// Filter locations within radius
	var nearbyLocations []model.AttendanceLocation
	for _, loc := range allLocations {
		if utils.IsWithinRadius(req.Latitude, req.Longitude, loc.Latitude, loc.Longitude, req.RadiusKm) {
			nearbyLocations = append(nearbyLocations, loc)
		}
	}

	return nearbyLocations, nil
}

// UpdateLocation updates location information
func (s *LocationService) UpdateLocation(id uint, req *UpdateLocationRequest) (*model.AttendanceLocation, error) {
	location, err := s.GetLocationByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		location.Name = req.Name
	}
	if req.Description != "" {
		location.Description = req.Description
	}
	if req.Latitude != 0 {
		location.Latitude = req.Latitude
	}
	if req.Longitude != 0 {
		location.Longitude = req.Longitude
	}
	if req.Radius > 0 {
		location.Radius = req.Radius
	}
	if req.IsActive != nil {
		location.IsActive = *req.IsActive
	}

	if err := s.db.Save(&location).Error; err != nil {
		return nil, err
	}

	return location, nil
}

// DeleteLocation deletes a location
func (s *LocationService) DeleteLocation(id uint) error {
	// Check if location exists
	if _, err := s.GetLocationByID(id); err != nil {
		return err
	}

	// Soft delete
	if err := s.db.Delete(&model.AttendanceLocation{}, id).Error; err != nil {
		return err
	}

	return nil
}

// ValidateLocationForAttendance validates if user can check-in at location
func (s *LocationService) ValidateLocationForAttendance(locationID uint, userLat, userLon float64) (bool, float64, error) {
	location, err := s.GetLocationByID(locationID)
	if err != nil {
		return false, 0, err
	}

	if !location.IsActive {
		return false, 0, errors.New("location is not active")
	}

	isValid, distance := utils.ValidateLocation(
		userLat, userLon,
		location.Latitude, location.Longitude,
		float64(location.Radius),
	)

	return isValid, distance, nil
}
