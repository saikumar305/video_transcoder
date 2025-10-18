package controller

import (
	"log"
	"video_transcoder/services"

	"gorm.io/gorm"
)

type TranscoderController struct {
	db               *gorm.DB
	transcodeService *services.TranscodeService
}

func NewTranscoderController(db *gorm.DB) *TranscoderController {
	return &TranscoderController{
		db:               db,
		transcodeService: services.NewTranscodeService(db),
	}
}

// ProcessVideoTranscoding handles the transcoding process for uploaded videos
func (tc *TranscoderController) ProcessVideoTranscoding(videoID, inputPath string) {
	log.Printf("Processing video transcoding for ID: %s", videoID)
	
	// Start transcoding in a goroutine to avoid blocking
	go tc.transcodeService.StartTranscodingJob(videoID, inputPath)
}

// GetTranscodingStatus retrieves the current status of a video transcoding job
func (tc *TranscoderController) GetTranscodingStatus(videoID string) (interface{}, error) {
	// This will be implemented to return job status
	return tc.transcodeService, nil
}