package main

import (
	"log"
	"video_transcoder/config"
	"video_transcoder/handler"
	"video_transcoder/models"
	"video_transcoder/worker"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	db := config.DatabaseConnection()
	
	// Auto-migrate database tables
	if err := db.AutoMigrate(&models.Video{}, &models.TranscodeJob{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	
	log.Println("Database connection established and tables migrated")

	// Initialize handler
	h := handler.New(db)
	
	// Initialize and start worker
	w := worker.NewWorker(db)
	w.Start()
	defer w.Stop()

	// Setup Gin router
	r := gin.Default()

	// Serve static files (for the web interface)
    // r.Static("/static", "./static")
    // r.StaticFile("/", "./static/upload.html")
	
	// Add CORS middleware for frontend integration
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Health check endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Video Transcoder API is running!",
			"version": "1.0.0",
			"status":  "healthy",
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Upload video for transcoding
		api.POST("/upload", h.UploadVideo)
		
		// Get video status and transcoding progress
		api.GET("/videos/:id/status", h.GetVideoStatus)
		
		// Download video (original or transcoded)
		api.GET("/videos/:id/download", h.DownloadVideo)
		
		// List all videos
		api.GET("/videos", h.ListVideos)
	}

	log.Println("Starting server on :8080")
	log.Println("API Endpoints:")
	log.Println("  POST /api/v1/upload - Upload video for transcoding")
	log.Println("  GET  /api/v1/videos/:id/status - Get video status")
	log.Println("  GET  /api/v1/videos/:id/download?quality=720p - Download video")
	log.Println("  GET  /api/v1/videos - List all videos")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}