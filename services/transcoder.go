package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"video_transcoder/models"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gorm.io/gorm"
)

type TranscodeService struct {
	db *gorm.DB
}

func NewTranscodeService(db *gorm.DB) *TranscodeService {
	return &TranscodeService{db: db}
}

// Quality presets for transcoding
type QualityPreset struct {
	Name       string
	Resolution string
	Bitrate    string
	Codec      string
}

var QualityPresets = map[string]QualityPreset{
	"480p": {
		Name:       "480p",
		Resolution: "854x480",
		Bitrate:    "1000k",
		Codec:      "libx264",
	},
	"720p": {
		Name:       "720p",
		Resolution: "1280x720",
		Bitrate:    "2500k",
		Codec:      "libx264",
	},
	"1080p": {
		Name:       "1080p",
		Resolution: "1920x1080",
		Bitrate:    "5000k",
		Codec:      "libx264",
	},
}

// GetVideoInfo extracts video metadata using ffprobe
func (ts *TranscodeService) GetVideoInfo(inputPath string) (duration float64, fileSize int64, err error) {
	// Get file size
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get file info: %v", err)
	}
	fileSize = fileInfo.Size()

	// Get duration using ffprobe
	probe, err := ffmpeg.Probe(inputPath)
	if err != nil {
		return 0, fileSize, fmt.Errorf("failed to probe video: %v", err)
	}

	// Parse duration from probe output (simplified)
	// In a production app, you'd want to parse JSON output properly
	durationStr := ""
	lines := strings.Split(probe, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Duration:") {
			parts := strings.Split(line, "Duration: ")
			if len(parts) > 1 {
				durationStr = strings.Split(parts[1], ",")[0]
				break
			}
		}
	}

	if durationStr != "" {
		duration, err = parseDuration(durationStr)
		if err != nil {
			log.Printf("Failed to parse duration: %v", err)
		}
	}

	return duration, fileSize, nil
}

// parseDuration converts HH:MM:SS.mmm to seconds
func parseDuration(durationStr string) (float64, error) {
	parts := strings.Split(durationStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid duration format")
	}

	hours, _ := strconv.ParseFloat(parts[0], 64)
	minutes, _ := strconv.ParseFloat(parts[1], 64)
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	return hours*3600 + minutes*60 + seconds, nil
}

// TranscodeVideo processes a video file into multiple qualities
func (ts *TranscodeService) TranscodeVideo(videoID, inputPath string, qualities []string) error {
	// Update video status to processing
	if err := ts.db.Model(&models.Video{}).Where("id = ?", videoID).
		Update("status", models.StatusProcessing).Error; err != nil {
		return fmt.Errorf("failed to update video status: %v", err)
	}

	// Get video info
	duration, fileSize, err := ts.GetVideoInfo(inputPath)
	if err != nil {
		log.Printf("Failed to get video info: %v", err)
	}

	// Update video with metadata
	ts.db.Model(&models.Video{}).Where("id = ?", videoID).Updates(map[string]interface{}{
		"duration":  duration,
		"file_size": fileSize,
	})

	transcodedPaths := make(models.StringMap)
	allJobsSuccessful := true

	// Create output directory
	outputDir := filepath.Join("./videos", "transcoded", videoID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Process each quality
	for _, quality := range qualities {
		preset, exists := QualityPresets[quality]
		if !exists {
			log.Printf("Unknown quality preset: %s", quality)
			continue
		}

		// Create transcode job
		job := models.TranscodeJob{
			VideoID:    videoID,
			InputPath:  inputPath,
			OutputPath: filepath.Join(outputDir, fmt.Sprintf("%s_%s.mp4", videoID, quality)),
			Quality:    quality,
			Status:     models.StatusQueued,
		}

		if err := ts.db.Create(&job).Error; err != nil {
			log.Printf("Failed to create transcode job: %v", err)
			continue
		}

		// Update job status to processing
		ts.db.Model(&job).Update("status", models.StatusProcessing)

		// Perform transcoding
		err := ts.TranscodeToQuality(inputPath, job.OutputPath, preset, job.ID)
		if err != nil {
			log.Printf("Failed to transcode to %s: %v", quality, err)
			ts.db.Model(&job).Updates(map[string]interface{}{
				"status":        models.StatusFailed,
				"error_message": err.Error(),
			})
			allJobsSuccessful = false
		} else {
			ts.db.Model(&job).Updates(map[string]interface{}{
				"status":   models.StatusCompleted,
				"progress": 100,
			})
			transcodedPaths[quality] = job.OutputPath
		}
	}

	// Update video status
	var finalStatus models.TranscodeStatus
	if allJobsSuccessful {
		finalStatus = models.StatusCompleted
	} else if len(transcodedPaths) > 0 {
		finalStatus = models.StatusCompleted // Partial success
	} else {
		finalStatus = models.StatusFailed
	}

	return ts.db.Model(&models.Video{}).Where("id = ?", videoID).Updates(map[string]interface{}{
		"status":           finalStatus,
		"transcoded_paths": transcodedPaths,
	}).Error
}

// TranscodeToQuality performs the actual transcoding using ffmpeg
func (ts *TranscodeService) TranscodeToQuality(inputPath, outputPath string, preset QualityPreset, jobID uint) error {
	log.Printf("Starting transcode: %s -> %s (%s)", inputPath, outputPath, preset.Name)

	// Create ffmpeg stream
	stream := ffmpeg.Input(inputPath).
		Video().
		Filter("scale", ffmpeg.Args{preset.Resolution}).
		Output(outputPath, ffmpeg.KwArgs{
			"c:v":     preset.Codec,
			"b:v":     preset.Bitrate,
			"c:a":     "aac",
			"b:a":     "128k",
			"preset":  "fast",
			"movflags": "+faststart", // For web streaming
		}).
		OverWriteOutput()

	// Run the transcoding
	err := stream.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg transcoding failed: %v", err)
	}

	log.Printf("Transcode completed: %s", outputPath)
	return nil
}

// StartTranscodingJob starts transcoding for a video
func (ts *TranscodeService) StartTranscodingJob(videoID, inputPath string) {
	log.Printf("Starting transcoding job for video: %s", videoID)

	// Define default qualities to transcode
	qualities := []string{"480p", "720p"}

	// Check if input file has high enough resolution for 1080p
	// This is a simplified check - in production you'd probe the actual resolution
	qualities = append(qualities, "1080p")

	if err := ts.TranscodeVideo(videoID, inputPath, qualities); err != nil {
		log.Printf("Transcoding failed for video %s: %v", videoID, err)
		
		// Update video status to failed
		ts.db.Model(&models.Video{}).Where("id = ?", videoID).Updates(map[string]interface{}{
			"status":        models.StatusFailed,
			"error_message": err.Error(),
		})
	}
}