package model

import (
	"time"
)

type Attendance struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	UserID               uint       `gorm:"not null" json:"user_id"`
	LocationID           uint       `gorm:"not null" json:"location_id"`
	CheckInTime          time.Time  `gorm:"not null" json:"check_in_time"`
	CheckOutTime         *time.Time `json:"check_out_time"`
	CheckInLatitude      float64    `gorm:"not null;type:decimal(10,8)" json:"check_in_latitude"`
	CheckInLongitude     float64    `gorm:"not null;type:decimal(11,8)" json:"check_in_longitude"`
	CheckOutLatitude     *float64   `gorm:"type:decimal(10,8)" json:"check_out_latitude"`
	CheckOutLongitude    *float64   `gorm:"type:decimal(11,8)" json:"check_out_longitude"`
	DistanceFromLocation float64    `gorm:"type:decimal(10,2)" json:"distance_from_location"` // in meters
	Status               string     `gorm:"default:present" json:"status"`                     // 'present', 'late', 'half_day'
	Notes                string     `json:"notes"`
	PhotoURL             string     `json:"photo_url"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Relations
	User     User               `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Location AttendanceLocation `gorm:"foreignKey:LocationID" json:"location,omitempty"`
}

// TableName specifies the table name for Attendance model
func (Attendance) TableName() string {
	return "attendances"
}

// AttendanceResponse represents attendance data with relations
type AttendanceResponse struct {
	ID                   uint                `json:"id"`
	UserID               uint                `json:"user_id"`
	LocationID           uint                `json:"location_id"`
	CheckInTime          time.Time           `json:"check_in_time"`
	CheckOutTime         *time.Time          `json:"check_out_time"`
	CheckInLatitude      float64             `json:"check_in_latitude"`
	CheckInLongitude     float64             `json:"check_in_longitude"`
	CheckOutLatitude     *float64            `json:"check_out_latitude"`
	CheckOutLongitude    *float64            `json:"check_out_longitude"`
	DistanceFromLocation float64             `json:"distance_from_location"`
	Status               string              `json:"status"`
	Notes                string              `json:"notes"`
	PhotoURL             string              `json:"photo_url"`
	WorkDuration         *string             `json:"work_duration,omitempty"` // calculated field
	User                 *UserResponse       `json:"user,omitempty"`
	Location             *LocationResponse   `json:"location,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

// ToResponse converts Attendance to AttendanceResponse
func (a *Attendance) ToResponse() AttendanceResponse {
	response := AttendanceResponse{
		ID:                   a.ID,
		UserID:               a.UserID,
		LocationID:           a.LocationID,
		CheckInTime:          a.CheckInTime,
		CheckOutTime:         a.CheckOutTime,
		CheckInLatitude:      a.CheckInLatitude,
		CheckInLongitude:     a.CheckInLongitude,
		CheckOutLatitude:     a.CheckOutLatitude,
		CheckOutLongitude:    a.CheckOutLongitude,
		DistanceFromLocation: a.DistanceFromLocation,
		Status:               a.Status,
		Notes:                a.Notes,
		PhotoURL:             a.PhotoURL,
		CreatedAt:            a.CreatedAt,
		UpdatedAt:            a.UpdatedAt,
	}

	// Calculate work duration if checked out
	if a.CheckOutTime != nil {
		duration := a.CheckOutTime.Sub(a.CheckInTime)
		durationStr := formatDuration(duration)
		response.WorkDuration = &durationStr
	}

	// Add user info if loaded
	if a.User.ID != 0 {
		userResp := a.User.ToResponse()
		response.User = &userResp
	}

	// Add location info if loaded
	if a.Location.ID != 0 {
		locResp := a.Location.ToResponse()
		response.Location = &locResp
	}

	return response
}

// formatDuration formats duration to human-readable string
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return time.Duration(hours*int(time.Hour) + minutes*int(time.Minute)).String()
}
