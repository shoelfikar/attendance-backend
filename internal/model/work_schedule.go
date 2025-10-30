package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type WorkSchedule struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	Name           string        `gorm:"not null" json:"name"`
	CheckInStart   string        `gorm:"not null;type:time" json:"check_in_start"`   // e.g., "08:00:00"
	CheckInEnd     string        `gorm:"not null;type:time" json:"check_in_end"`     // e.g., "09:00:00"
	CheckOutStart  string        `gorm:"not null;type:time" json:"check_out_start"`  // e.g., "17:00:00"
	WorkDays       pq.Int64Array `gorm:"type:integer[]" json:"work_days"`            // [1,2,3,4,5] for Mon-Fri
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// TableName specifies the table name for WorkSchedule model
func (WorkSchedule) TableName() string {
	return "work_schedules"
}

// ScheduleResponse represents work schedule data
type ScheduleResponse struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	CheckInStart  string    `json:"check_in_start"`
	CheckInEnd    string    `json:"check_in_end"`
	CheckOutStart string    `json:"check_out_start"`
	WorkDays      []int     `json:"work_days"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToResponse converts WorkSchedule to ScheduleResponse
func (w *WorkSchedule) ToResponse() ScheduleResponse {
	workDays := make([]int, len(w.WorkDays))
	for i, day := range w.WorkDays {
		workDays[i] = int(day)
	}

	return ScheduleResponse{
		ID:            w.ID,
		Name:          w.Name,
		CheckInStart:  w.CheckInStart,
		CheckInEnd:    w.CheckInEnd,
		CheckOutStart: w.CheckOutStart,
		WorkDays:      workDays,
		CreatedAt:     w.CreatedAt,
		UpdatedAt:     w.UpdatedAt,
	}
}

type UserSchedule struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"not null" json:"user_id"`
	ScheduleID    uint       `gorm:"not null" json:"schedule_id"`
	LocationID    uint       `gorm:"not null" json:"location_id"`
	EffectiveFrom time.Time  `gorm:"not null;type:date" json:"effective_from"`
	EffectiveTo   *time.Time `gorm:"type:date" json:"effective_to"`
	CreatedAt     time.Time  `json:"created_at"`

	// Relations
	User     User               `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Schedule WorkSchedule       `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
	Location AttendanceLocation `gorm:"foreignKey:LocationID" json:"location,omitempty"`
}

// TableName specifies the table name for UserSchedule model
func (UserSchedule) TableName() string {
	return "user_schedules"
}

// UserScheduleResponse represents user schedule data with relations
type UserScheduleResponse struct {
	ID            uint              `json:"id"`
	UserID        uint              `json:"user_id"`
	ScheduleID    uint              `json:"schedule_id"`
	LocationID    uint              `json:"location_id"`
	EffectiveFrom time.Time         `json:"effective_from"`
	EffectiveTo   *time.Time        `json:"effective_to"`
	User          *UserResponse     `json:"user,omitempty"`
	Schedule      *ScheduleResponse `json:"schedule,omitempty"`
	Location      *LocationResponse `json:"location,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

// ToResponse converts UserSchedule to UserScheduleResponse
func (us *UserSchedule) ToResponse() UserScheduleResponse {
	response := UserScheduleResponse{
		ID:            us.ID,
		UserID:        us.UserID,
		ScheduleID:    us.ScheduleID,
		LocationID:    us.LocationID,
		EffectiveFrom: us.EffectiveFrom,
		EffectiveTo:   us.EffectiveTo,
		CreatedAt:     us.CreatedAt,
	}

	// Add user info if loaded
	if us.User.ID != 0 {
		userResp := us.User.ToResponse()
		response.User = &userResp
	}

	// Add schedule info if loaded
	if us.Schedule.ID != 0 {
		scheduleResp := us.Schedule.ToResponse()
		response.Schedule = &scheduleResp
	}

	// Add location info if loaded
	if us.Location.ID != 0 {
		locResp := us.Location.ToResponse()
		response.Location = &locResp
	}

	return response
}

// Value implements driver.Valuer for JSON marshaling
func (w WorkSchedule) Value() (driver.Value, error) {
	return json.Marshal(w)
}

// Scan implements sql.Scanner for JSON unmarshaling
func (w *WorkSchedule) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, &w)
}
