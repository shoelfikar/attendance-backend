package controller

import (
	"net/http"
	"strconv"

	"github.com/attendance/backend/internal/service"
	"github.com/attendance/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type ScheduleController struct {
	scheduleService *service.ScheduleService
}

func NewScheduleController(scheduleService *service.ScheduleService) *ScheduleController {
	return &ScheduleController{
		scheduleService: scheduleService,
	}
}

// CreateSchedule godoc
// @Summary Create new work schedule (Admin)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateScheduleRequest true "Create schedule request"
// @Success 201 {object} utils.Response
// @Router /api/v1/admin/schedules [post]
func (ctrl *ScheduleController) CreateSchedule(c *gin.Context) {
	var req service.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	schedule, err := ctrl.scheduleService.CreateSchedule(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create schedule", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Schedule created successfully", schedule.ToResponse())
}

// GetAllSchedules godoc
// @Summary Get all work schedules (Admin)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/schedules [get]
func (ctrl *ScheduleController) GetAllSchedules(c *gin.Context) {
	schedules, err := ctrl.scheduleService.GetAllSchedules()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get schedules", err.Error())
		return
	}

	// Convert to responses
	responses := make([]interface{}, len(schedules))
	for i, schedule := range schedules {
		responses[i] = schedule.ToResponse()
	}

	utils.SuccessResponse(c, http.StatusOK, "Schedules retrieved", responses)
}

// GetScheduleByID godoc
// @Summary Get schedule by ID (Admin)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Schedule ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/schedules/:id [get]
func (ctrl *ScheduleController) GetScheduleByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid schedule ID", err.Error())
		return
	}

	schedule, err := ctrl.scheduleService.GetScheduleByID(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Schedule not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Schedule retrieved", schedule.ToResponse())
}

// UpdateSchedule godoc
// @Summary Update work schedule (Admin)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Schedule ID"
// @Param request body service.UpdateScheduleRequest true "Update schedule request"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/schedules/:id [put]
func (ctrl *ScheduleController) UpdateSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid schedule ID", err.Error())
		return
	}

	var req service.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	schedule, err := ctrl.scheduleService.UpdateSchedule(uint(id), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update schedule", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Schedule updated successfully", schedule.ToResponse())
}

// DeleteSchedule godoc
// @Summary Delete work schedule (Admin)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Schedule ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/schedules/:id [delete]
func (ctrl *ScheduleController) DeleteSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid schedule ID", err.Error())
		return
	}

	if err := ctrl.scheduleService.DeleteSchedule(uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete schedule", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Schedule deleted successfully", nil)
}

// AssignSchedule godoc
// @Summary Assign schedule to user (Admin)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.AssignScheduleRequest true "Assign schedule request"
// @Success 201 {object} utils.Response
// @Router /api/v1/admin/schedules/assign [post]
func (ctrl *ScheduleController) AssignSchedule(c *gin.Context) {
	var req service.AssignScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	userSchedule, err := ctrl.scheduleService.AssignScheduleToUser(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to assign schedule", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Schedule assigned successfully", userSchedule.ToResponse())
}

// GetUserSchedules godoc
// @Summary Get user's assigned schedules (Admin)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param user_id query int true "User ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/schedules/user [get]
func (ctrl *ScheduleController) GetUserSchedules(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	userSchedules, err := ctrl.scheduleService.GetUserSchedules(uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user schedules", err.Error())
		return
	}

	// Convert to responses
	responses := make([]interface{}, len(userSchedules))
	for i, us := range userSchedules {
		responses[i] = us.ToResponse()
	}

	utils.SuccessResponse(c, http.StatusOK, "User schedules retrieved", responses)
}
