package service

import (
	"errors"
	"time"

	"github.com/attendance/backend/internal/model"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ScheduleService struct {
	db *gorm.DB
}

func NewScheduleService(db *gorm.DB) *ScheduleService {
	return &ScheduleService{db: db}
}

// CreateScheduleRequest represents create schedule request
type CreateScheduleRequest struct {
	Name          string `json:"name" binding:"required"`
	CheckInStart  string `json:"check_in_start" binding:"required"`  // "08:00:00"
	CheckInEnd    string `json:"check_in_end" binding:"required"`    // "09:00:00"
	CheckOutStart string `json:"check_out_start" binding:"required"` // "17:00:00"
	WorkDays      []int  `json:"work_days" binding:"required"`       // [1,2,3,4,5]
}

// UpdateScheduleRequest represents update schedule request
type UpdateScheduleRequest struct {
	Name          string `json:"name"`
	CheckInStart  string `json:"check_in_start"`
	CheckInEnd    string `json:"check_in_end"`
	CheckOutStart string `json:"check_out_start"`
	WorkDays      []int  `json:"work_days"`
}

// AssignScheduleRequest represents assign schedule to user request
type AssignScheduleRequest struct {
	UserID        uint   `json:"user_id" binding:"required"`
	ScheduleID    uint   `json:"schedule_id" binding:"required"`
	LocationID    uint   `json:"location_id" binding:"required"`
	EffectiveFrom string `json:"effective_from" binding:"required"` // "2025-01-01"
	EffectiveTo   string `json:"effective_to"`                      // "2025-12-31" (optional)
}

// CreateSchedule creates a new work schedule
func (s *ScheduleService) CreateSchedule(req *CreateScheduleRequest) (*model.WorkSchedule, error) {
	// Convert []int to pq.Int64Array
	workDays := make(pq.Int64Array, len(req.WorkDays))
	for i, day := range req.WorkDays {
		workDays[i] = int64(day)
	}

	schedule := model.WorkSchedule{
		Name:          req.Name,
		CheckInStart:  req.CheckInStart,
		CheckInEnd:    req.CheckInEnd,
		CheckOutStart: req.CheckOutStart,
		WorkDays:      workDays,
	}

	if err := s.db.Create(&schedule).Error; err != nil {
		return nil, err
	}

	return &schedule, nil
}

// GetScheduleByID retrieves schedule by ID
func (s *ScheduleService) GetScheduleByID(id uint) (*model.WorkSchedule, error) {
	var schedule model.WorkSchedule
	if err := s.db.First(&schedule, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("schedule not found")
		}
		return nil, err
	}
	return &schedule, nil
}

// GetAllSchedules retrieves all work schedules
func (s *ScheduleService) GetAllSchedules() ([]model.WorkSchedule, error) {
	var schedules []model.WorkSchedule
	if err := s.db.Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

// UpdateSchedule updates schedule information
func (s *ScheduleService) UpdateSchedule(id uint, req *UpdateScheduleRequest) (*model.WorkSchedule, error) {
	schedule, err := s.GetScheduleByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		schedule.Name = req.Name
	}
	if req.CheckInStart != "" {
		schedule.CheckInStart = req.CheckInStart
	}
	if req.CheckInEnd != "" {
		schedule.CheckInEnd = req.CheckInEnd
	}
	if req.CheckOutStart != "" {
		schedule.CheckOutStart = req.CheckOutStart
	}
	if len(req.WorkDays) > 0 {
		workDays := make(pq.Int64Array, len(req.WorkDays))
		for i, day := range req.WorkDays {
			workDays[i] = int64(day)
		}
		schedule.WorkDays = workDays
	}

	if err := s.db.Save(&schedule).Error; err != nil {
		return nil, err
	}

	return schedule, nil
}

// DeleteSchedule deletes a work schedule
func (s *ScheduleService) DeleteSchedule(id uint) error {
	if _, err := s.GetScheduleByID(id); err != nil {
		return err
	}

	if err := s.db.Delete(&model.WorkSchedule{}, id).Error; err != nil {
		return err
	}

	return nil
}

// AssignScheduleToUser assigns a work schedule to a user
func (s *ScheduleService) AssignScheduleToUser(req *AssignScheduleRequest) (*model.UserSchedule, error) {
	// Validate schedule exists
	if _, err := s.GetScheduleByID(req.ScheduleID); err != nil {
		return nil, errors.New("schedule not found")
	}

	// Parse dates
	effectiveFrom, err := parseDate(req.EffectiveFrom)
	if err != nil {
		return nil, errors.New("invalid effective_from date format")
	}

	var effectiveTo *string
	if req.EffectiveTo != "" {
		effectiveTo = &req.EffectiveTo
	}

	userSchedule := model.UserSchedule{
		UserID:        req.UserID,
		ScheduleID:    req.ScheduleID,
		LocationID:    req.LocationID,
		EffectiveFrom: effectiveFrom,
	}

	if effectiveTo != nil {
		parsed, err := parseDate(*effectiveTo)
		if err != nil {
			return nil, errors.New("invalid effective_to date format")
		}
		userSchedule.EffectiveTo = &parsed
	}

	if err := s.db.Create(&userSchedule).Error; err != nil {
		return nil, err
	}

	// Load relations
	s.db.Preload("User").Preload("Schedule").Preload("Location").First(&userSchedule, userSchedule.ID)

	return &userSchedule, nil
}

// GetUserSchedules retrieves schedules assigned to a user
func (s *ScheduleService) GetUserSchedules(userID uint) ([]model.UserSchedule, error) {
	var userSchedules []model.UserSchedule
	if err := s.db.Preload("Schedule").Preload("Location").
		Where("user_id = ?", userID).
		Find(&userSchedules).Error; err != nil {
		return nil, err
	}
	return userSchedules, nil
}

// Helper function to parse date
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
