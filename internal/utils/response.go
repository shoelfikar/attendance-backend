package utils

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SuccessResponse sends success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends error response
func ErrorResponse(c *gin.Context, statusCode int, message string, err interface{}) {
	c.JSON(statusCode, Response{
		Status:  "error",
		Message: message,
		Error:   err,
	})
}

// ValidationErrorResponse sends validation error response
func ValidationErrorResponse(c *gin.Context, errors interface{}) {
	c.JSON(400, Response{
		Status:  "error",
		Message: "Validation failed",
		Error:   errors,
	})
}
