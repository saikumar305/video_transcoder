package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"time"
	"video_transcoder/controller"
	"video_transcoder/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	db         *gorm.DB
	controller *controller.TranscoderController
}

func New(db *gorm.DB) *Handler {
	return &Handler{
		db:         db,
		controller: controller.NewTranscoderController(db),
	}
}

type UploadResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type StatusResponse struct {
	ID              string                 `json:"id"`
	Status          models.TranscodeStatus `json:"status"`
	OriginalPath    string                 `json:"original_path"`
	TranscodedPaths models.StringMap       `json:"transcoded_paths"`
	Duration        float64                `json:"duration"`
	FileSize        int64                  `json:"file_size"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Jobs            []models.TranscodeJob  `json:"jobs,omitempty"`
}

func (h *Handler) UploadVideo(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
		return
	}

	// Validate file type (basic validation)
	if !isVideoFile(file.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Please upload a video file"})
		return
	}

	id := uuid.New().String()

	// Create upload directory if it doesn't exist
	uploadDir := "./videos/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Save original file
	originalPath := filepath.Join(uploadDir, id+"_"+file.Filename)
	if err := c.SaveUploadedFile(file, originalPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Create DB record
	video := models.Video{
		ID:               id,
		OriginalPath:     originalPath,
		OriginalFilename: file.Filename,
		Status:           models.StatusPending,
		TranscodedPaths:  make(models.StringMap),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := h.db.Create(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
		return
	}

	// Start transcoding in background
	h.controller.ProcessVideoTranscoding(id, originalPath)

	c.JSON(http.StatusOK, UploadResponse{
		ID:      id,
		Message: "Video uploaded successfully. Transcoding started.",
	})
}

func (h *Handler) GetVideoStatus(c *gin.Context) {
	videoID := c.Param("id")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID is required"})
		return
	}

	var video models.Video
	if err := h.db.Where("id = ?", videoID).First(&video).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve video"})
		return
	}

	// Get associated transcode jobs
	var jobs []models.TranscodeJob
	h.db.Where("video_id = ?", videoID).Find(&jobs)

	response := StatusResponse{
		ID:              video.ID,
		Status:          video.Status,
		OriginalPath:    video.OriginalPath,
		TranscodedPaths: video.TranscodedPaths,
		Duration:        video.Duration,
		FileSize:        video.FileSize,
		ErrorMessage:    video.ErrorMessage,
		CreatedAt:       video.CreatedAt,
		UpdatedAt:       video.UpdatedAt,
		Jobs:            jobs,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) DownloadVideo(c *gin.Context) {
	videoID := c.Param("id")
	quality := c.Query("quality") // e.g., ?quality=720p
	
	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID is required"})
		return
	}

	var video models.Video
	if err := h.db.Where("id = ?", videoID).First(&video).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve video"})
		return
	}

	var filePath string
	
	if quality == "" || quality == "original" {
		filePath = video.OriginalPath
	} else {
		if video.TranscodedPaths == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No transcoded versions available"})
			return
		}
		
		path, exists := video.TranscodedPaths[quality]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Requested quality not available"})
			return
		}
		filePath = path
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video file not found on disk"})
		return
	}

	// Serve the file
	filename := filepath.Base(filePath)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "video/mp4")
	c.File(filePath)
}

func (h *Handler) ListVideos(c *gin.Context) {
	var videos []models.Video
	
	// Get pagination parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	
	var pageInt, limitInt int = 1, 10
	// Parse pagination (simplified - you might want better validation)
	
	offset := (pageInt - 1) * limitInt
	
	if err := h.db.Offset(offset).Limit(limitInt).Order("created_at DESC").Find(&videos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve videos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"videos": videos,
		"page":   page,
		"limit":  limit,
	})
}

// Helper function to validate video file types
func isVideoFile(filename string) bool {
	ext := filepath.Ext(filename)
	videoExtensions := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm", ".m4v"}
	
	for _, validExt := range videoExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}
