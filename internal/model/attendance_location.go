package model

import "time"

type AttendanceLocation struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Latitude    float64   `gorm:"not null;type:decimal(10,8)" json:"latitude"`
	Longitude   float64   `gorm:"not null;type:decimal(11,8)" json:"longitude"`
	Radius      int       `gorm:"default:10" json:"radius"` // in meters
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedBy   *uint     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName specifies the table name for AttendanceLocation model
func (AttendanceLocation) TableName() string {
	return "attendance_locations"
}

// LocationResponse represents location data with creator info
type LocationResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Radius      int       `json:"radius"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   *uint     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts AttendanceLocation to LocationResponse
func (l *AttendanceLocation) ToResponse() LocationResponse {
	return LocationResponse{
		ID:          l.ID,
		Name:        l.Name,
		Description: l.Description,
		Latitude:    l.Latitude,
		Longitude:   l.Longitude,
		Radius:      l.Radius,
		IsActive:    l.IsActive,
		CreatedBy:   l.CreatedBy,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}
