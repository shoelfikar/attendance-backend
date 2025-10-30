package service

import (
	"errors"
	"time"

	"github.com/attendance/backend/internal/model"
	"gorm.io/gorm"
)

type AttendanceService struct {
	db              *gorm.DB
	locationService *LocationService
}

func NewAttendanceService(db *gorm.DB, locationService *LocationService) *AttendanceService {
	return &AttendanceService{
		db:              db,
		locationService: locationService,
	}
}

// CheckInRequest represents check-in request
type CheckInRequest struct {
	LocationID uint    `json:"location_id" binding:"required"`
	Latitude   float64 `json:"latitude" binding:"required"`
	Longitude  float64 `json:"longitude" binding:"required"`
	PhotoURL   string  `json:"photo_url"`
	Notes      string  `json:"notes"`
}

// CheckOutRequest represents check-out request
type CheckOutRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Notes     string  `json:"notes"`
}

// CheckIn creates a new attendance record
func (s *AttendanceService) CheckIn(userID uint, req *CheckInRequest) (*model.Attendance, error) {
	// Check if already checked in today
	hasCheckedIn, err := s.HasCheckedInToday(userID)
	if err != nil {
		return nil, err
	}
	if hasCheckedIn {
		return nil, errors.New("already checked in today")
	}

	// Validate location
	isValid, distance, err := s.locationService.ValidateLocationForAttendance(
		req.LocationID,
		req.Latitude,
		req.Longitude,
	)
	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, errors.New("you are outside the allowed radius")
	}

	// Determine status based on time
	status := s.determineAttendanceStatus(time.Now())

	// Create attendance record
	attendance := model.Attendance{
		UserID:               userID,
		LocationID:           req.LocationID,
		CheckInTime:          time.Now(),
		CheckInLatitude:      req.Latitude,
		CheckInLongitude:     req.Longitude,
		DistanceFromLocation: distance,
		Status:               status,
		Notes:                req.Notes,
		PhotoURL:             req.PhotoURL,
	}

	if err := s.db.Create(&attendance).Error; err != nil {
		return nil, err
	}

	// Load relations
	s.db.Preload("User").Preload("Location").First(&attendance, attendance.ID)

	return &attendance, nil
}

// CheckOut updates attendance record with check-out time
func (s *AttendanceService) CheckOut(userID uint, req *CheckOutRequest) (*model.Attendance, error) {
	// Get today's attendance
	attendance, err := s.GetTodayAttendance(userID)
	if err != nil {
		return nil, err
	}

	if attendance.CheckOutTime != nil {
		return nil, errors.New("already checked out today")
	}

	// Validate location (should be near check-in location)
	isValid, _, err := s.locationService.ValidateLocationForAttendance(
		attendance.LocationID,
		req.Latitude,
		req.Longitude,
	)
	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, errors.New("you are outside the allowed radius for check-out")
	}

	// Update check-out info
	now := time.Now()
	attendance.CheckOutTime = &now
	attendance.CheckOutLatitude = &req.Latitude
	attendance.CheckOutLongitude = &req.Longitude

	if req.Notes != "" {
		if attendance.Notes != "" {
			attendance.Notes += " | " + req.Notes
		} else {
			attendance.Notes = req.Notes
		}
	}

	if err := s.db.Save(&attendance).Error; err != nil {
		return nil, err
	}

	// Reload with relations
	s.db.Preload("User").Preload("Location").First(&attendance, attendance.ID)

	return attendance, nil
}

// HasCheckedInToday checks if user has checked in today
func (s *AttendanceService) HasCheckedInToday(userID uint) (bool, error) {
	var count int64
	today := time.Now().Format("2006-01-02")

	err := s.db.Model(&model.Attendance{}).
		Where("user_id = ? AND DATE(check_in_time) = ?", userID, today).
		Count(&count).Error

	return count > 0, err
}

// GetTodayAttendance gets user's attendance for today
func (s *AttendanceService) GetTodayAttendance(userID uint) (*model.Attendance, error) {
	var attendance model.Attendance
	today := time.Now().Format("2006-01-02")

	err := s.db.Preload("User").Preload("Location").
		Where("user_id = ? AND DATE(check_in_time) = ?", userID, today).
		First(&attendance).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no attendance record found for today")
		}
		return nil, err
	}

	return &attendance, nil
}

// GetAttendanceStatus gets current attendance status
func (s *AttendanceService) GetAttendanceStatus(userID uint) (map[string]interface{}, error) {
	attendance, err := s.GetTodayAttendance(userID)
	if err != nil {
		// No check-in today
		return map[string]interface{}{
			"has_checked_in":  false,
			"has_checked_out": false,
			"message":         "You haven't checked in today",
		}, nil
	}

	return map[string]interface{}{
		"has_checked_in":  true,
		"has_checked_out": attendance.CheckOutTime != nil,
		"check_in_time":   attendance.CheckInTime,
		"check_out_time":  attendance.CheckOutTime,
		"location":        attendance.Location.Name,
		"status":          attendance.Status,
	}, nil
}

// GetUserAttendanceHistory gets attendance history for a user
func (s *AttendanceService) GetUserAttendanceHistory(userID uint, limit, offset int) ([]model.Attendance, int64, error) {
	var attendances []model.Attendance
	var total int64

	// Count total
	s.db.Model(&model.Attendance{}).Where("user_id = ?", userID).Count(&total)

	// Get paginated records
	err := s.db.Preload("Location").
		Where("user_id = ?", userID).
		Order("check_in_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&attendances).Error

	if err != nil {
		return nil, 0, err
	}

	return attendances, total, nil
}

// GetAllAttendances gets all attendances with filters (Admin)
func (s *AttendanceService) GetAllAttendances(filters map[string]interface{}, limit, offset int) ([]model.Attendance, int64, error) {
	var attendances []model.Attendance
	var total int64

	query := s.db.Model(&model.Attendance{})

	// Apply filters
	if userID, ok := filters["user_id"].(uint); ok && userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if locationID, ok := filters["location_id"].(uint); ok && locationID > 0 {
		query = query.Where("location_id = ?", locationID)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if dateFrom, ok := filters["date_from"].(string); ok && dateFrom != "" {
		query = query.Where("DATE(check_in_time) >= ?", dateFrom)
	}
	if dateTo, ok := filters["date_to"].(string); ok && dateTo != "" {
		query = query.Where("DATE(check_in_time) <= ?", dateTo)
	}

	// Count total
	query.Count(&total)

	// Get paginated records
	err := query.Preload("User").Preload("Location").
		Order("check_in_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&attendances).Error

	if err != nil {
		return nil, 0, err
	}

	return attendances, total, nil
}

// determineAttendanceStatus determines status based on check-in time
func (s *AttendanceService) determineAttendanceStatus(checkInTime time.Time) string {
	// For now, simple logic: late if after 9 AM
	hour := checkInTime.Hour()

	if hour < 9 {
		return "present"
	} else if hour == 9 {
		return "present"
	} else if hour < 12 {
		return "late"
	} else {
		return "half_day"
	}
}
