package main

import (
	"log"

	"github.com/attendance/backend/internal/config"
	"github.com/attendance/backend/internal/controller"
	"github.com/attendance/backend/internal/middleware"
	"github.com/attendance/backend/internal/service"
	"github.com/attendance/backend/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Connect to database
	if err := database.Connect(cfg.Database.GetDSN()); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	log.Println("Database connected successfully")

	// Initialize services
	authService := service.NewAuthService(database.DB, cfg)
	userService := service.NewUserService(database.DB)
	locationService := service.NewLocationService(database.DB)
	attendanceService := service.NewAttendanceService(database.DB, locationService)
	scheduleService := service.NewScheduleService(database.DB)

	// Initialize controllers
	authController := controller.NewAuthController(authService)
	userController := controller.NewUserController(userService)
	locationController := controller.NewLocationController(locationService)
	attendanceController := controller.NewAttendanceController(attendanceService)
	scheduleController := controller.NewScheduleController(scheduleService)

	// Initialize Gin router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Attendance API is running",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/refresh-token", authController.RefreshToken)
			auth.POST("/logout", authController.Logout)

			// Protected auth routes
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthMiddleware(cfg))
			{
				authProtected.GET("/me", authController.GetMe)
			}
		}

		// Attendance routes (protected)
		attendance := v1.Group("/attendance")
		attendance.Use(middleware.AuthMiddleware(cfg))
		{
			attendance.GET("/locations", locationController.GetNearbyLocations)
			attendance.POST("/validate-location", locationController.ValidateLocation)
			attendance.POST("/check-in", attendanceController.CheckIn)
			attendance.POST("/check-out", attendanceController.CheckOut)
			attendance.GET("/today", attendanceController.GetTodayAttendance)
			attendance.GET("/status", attendanceController.GetAttendanceStatus)
			attendance.GET("/history", attendanceController.GetAttendanceHistory)
		}

		// Admin routes (protected + admin only)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg))
		admin.Use(middleware.AdminMiddleware())
		{
			// Profile management
			admin.GET("/profile", userController.GetMyProfile)
			admin.PUT("/profile", userController.UpdateMyProfile)
			admin.PUT("/profile/password", userController.UpdateMyPassword)

			// User management
			users := admin.Group("/users")
			{
				users.GET("", userController.GetAllUsers)
				users.GET("/stats", userController.GetUserStats)
				users.GET("/:id", userController.GetUserByID)
				users.POST("", userController.CreateUser)
				users.PUT("/:id", userController.UpdateUser)
				users.DELETE("/:id", userController.DeleteUser)
				users.PUT("/:id/password", userController.ChangeUserPassword)
			}

			// Location management
			locations := admin.Group("/locations")
			{
				locations.GET("", locationController.GetAllLocations)
				locations.GET("/:id", locationController.GetLocationByID)
				locations.POST("", locationController.CreateLocation)
				locations.PUT("/:id", locationController.UpdateLocation)
				locations.DELETE("/:id", locationController.DeleteLocation)
			}

			// Attendance management
			attendances := admin.Group("/attendances")
			{
				attendances.GET("", attendanceController.GetAllAttendances)
			}

			// Schedule management
			schedules := admin.Group("/schedules")
			{
				schedules.GET("", scheduleController.GetAllSchedules)
				schedules.GET("/:id", scheduleController.GetScheduleByID)
				schedules.POST("", scheduleController.CreateSchedule)
				schedules.PUT("/:id", scheduleController.UpdateSchedule)
				schedules.DELETE("/:id", scheduleController.DeleteSchedule)
				schedules.POST("/assign", scheduleController.AssignSchedule)
				schedules.GET("/user", scheduleController.GetUserSchedules)
			}
		}
	}

	// Start server
	port := ":" + cfg.Server.Port
	log.Printf("🚀 Server starting on port %s", cfg.Server.Port)
	log.Printf("📝 Environment: %s", cfg.Server.GinMode)
	log.Printf("💾 Database: %s", cfg.Database.DBName)

	if err := router.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
