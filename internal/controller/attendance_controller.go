package controller

import (
	"net/http"
	"strconv"

	"github.com/attendance/backend/internal/service"
	"github.com/attendance/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type AttendanceController struct {
	attendanceService *service.AttendanceService
}

func NewAttendanceController(attendanceService *service.AttendanceService) *AttendanceController {
	return &AttendanceController{
		attendanceService: attendanceService,
	}
}

// CheckIn godoc
// @Summary Check-in attendance
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CheckInRequest true "Check-in request"
// @Success 201 {object} utils.Response
// @Router /api/v1/attendance/check-in [post]
func (ctrl *AttendanceController) CheckIn(c *gin.Context) {
	var req service.CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	userID := c.GetUint("userID")
	attendance, err := ctrl.attendanceService.CheckIn(userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Check-in failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Check-in successful", attendance.ToResponse())
}

// CheckOut godoc
// @Summary Check-out attendance
// @Tags attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CheckOutRequest true "Check-out request"
// @Success 200 {object} utils.Response
// @Router /api/v1/attendance/check-out [post]
func (ctrl *AttendanceController) CheckOut(c *gin.Context) {
	var req service.CheckOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err.Error())
		return
	}

	userID := c.GetUint("userID")
	attendance, err := ctrl.attendanceService.CheckOut(userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Check-out failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Check-out successful", attendance.ToResponse())
}

// GetTodayAttendance godoc
// @Summary Get today's attendance
// @Tags attendance
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/attendance/today [get]
func (ctrl *AttendanceController) GetTodayAttendance(c *gin.Context) {
	userID := c.GetUint("userID")
	attendance, err := ctrl.attendanceService.GetTodayAttendance(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "No attendance found for today", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Today's attendance retrieved", attendance.ToResponse())
}

// GetAttendanceStatus godoc
// @Summary Get current attendance status
// @Tags attendance
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/attendance/status [get]
func (ctrl *AttendanceController) GetAttendanceStatus(c *gin.Context) {
	userID := c.GetUint("userID")
	status, err := ctrl.attendanceService.GetAttendanceStatus(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get status", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Status retrieved", status)
}

// GetAttendanceHistory godoc
// @Summary Get attendance history
// @Tags attendance
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.Response
// @Router /api/v1/attendance/history [get]
func (ctrl *AttendanceController) GetAttendanceHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	userID := c.GetUint("userID")

	attendances, total, err := ctrl.attendanceService.GetUserAttendanceHistory(userID, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get history", err.Error())
		return
	}

	// Convert to responses
	responses := make([]interface{}, len(attendances))
	for i, att := range attendances {
		responses[i] = att.ToResponse()
	}

	utils.SuccessResponse(c, http.StatusOK, "History retrieved", gin.H{
		"data":       responses,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_page": (int(total) + limit - 1) / limit,
	})
}

// GetAllAttendances godoc
// @Summary Get all attendances (Admin)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param location_id query int false "Filter by location ID"
// @Param status query string false "Filter by status"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/attendances [get]
func (ctrl *AttendanceController) GetAllAttendances(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Build filters
	filters := make(map[string]interface{})
	if userID, err := strconv.ParseUint(c.Query("user_id"), 10, 32); err == nil {
		filters["user_id"] = uint(userID)
	}
	if locationID, err := strconv.ParseUint(c.Query("location_id"), 10, 32); err == nil {
		filters["location_id"] = uint(locationID)
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters["date_from"] = dateFrom
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		filters["date_to"] = dateTo
	}

	offset := (page - 1) * limit
	attendances, total, err := ctrl.attendanceService.GetAllAttendances(filters, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get attendances", err.Error())
		return
	}

	// Convert to responses
	responses := make([]interface{}, len(attendances))
	for i, att := range attendances {
		responses[i] = att.ToResponse()
	}

	utils.SuccessResponse(c, http.StatusOK, "Attendances retrieved", gin.H{
		"data":       responses,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_page": (int(total) + limit - 1) / limit,
	})
}
