package worker

import (
	"log"
	"time"
	"video_transcoder/models"
	"video_transcoder/services"

	"gorm.io/gorm"
)

type Worker struct {
	db               *gorm.DB
	transcodeService *services.TranscodeService
	running          bool
}

func NewWorker(db *gorm.DB) *Worker {
	return &Worker{
		db:               db,
		transcodeService: services.NewTranscodeService(db),
		running:          false,
	}
}

// Start begins the worker to process queued transcode jobs
func (w *Worker) Start() {
	if w.running {
		log.Println("Worker is already running")
		return
	}

	w.running = true
	log.Println("Starting transcode worker...")

	go w.processJobs()
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	log.Println("Stopping transcode worker...")
	w.running = false
}

// processJobs continuously processes queued transcode jobs
func (w *Worker) processJobs() {
	ticker := time.NewTicker(5 * time.Second) // Check for jobs every 5 seconds
	defer ticker.Stop()

	for w.running {
		<-ticker.C
		w.processQueuedJobs()
	}

	log.Println("Transcode worker stopped")
}

// processQueuedJobs finds and processes queued transcode jobs
func (w *Worker) processQueuedJobs() {
	var jobs []models.TranscodeJob

	// Find queued jobs
	if err := w.db.Where("status = ?", models.StatusQueued).
		Order("created_at ASC").
		Limit(5). // Process up to 5 jobs at a time
		Find(&jobs).Error; err != nil {
		log.Printf("Failed to fetch queued jobs: %v", err)
		return
	}

	if len(jobs) == 0 {
		return // No jobs to process
	}

	log.Printf("Found %d queued transcode jobs", len(jobs))

	for _, job := range jobs {
		if !w.running {
			break
		}
		w.processJob(job)
	}
}

// processJob processes a single transcode job
func (w *Worker) processJob(job models.TranscodeJob) {
	log.Printf("Processing transcode job ID: %d, Quality: %s", job.ID, job.Quality)

	// Update job status to processing
	if err := w.db.Model(&job).Update("status", models.StatusProcessing).Error; err != nil {
		log.Printf("Failed to update job status: %v", err)
		return
	}

	// Get quality preset
	preset, exists := services.QualityPresets[job.Quality]
	if !exists {
		log.Printf("Unknown quality preset: %s", job.Quality)
		w.db.Model(&job).Updates(map[string]interface{}{
			"status":        models.StatusFailed,
			"error_message": "Unknown quality preset",
		})
		return
	}

	// Perform the actual transcoding
	err := w.transcodeService.TranscodeToQuality(job.InputPath, job.OutputPath, preset, job.ID)
	if err != nil {
		log.Printf("Transcoding failed for job %d: %v", job.ID, err)
		w.db.Model(&job).Updates(map[string]interface{}{
			"status":        models.StatusFailed,
			"error_message": err.Error(),
		})
		
		// Update parent video status if all jobs failed
		w.updateVideoStatusIfNeeded(job.VideoID)
		return
	}

	// Update job status to completed
	if err := w.db.Model(&job).Updates(map[string]interface{}{
		"status":   models.StatusCompleted,
		"progress": 100,
	}).Error; err != nil {
		log.Printf("Failed to update job completion status: %v", err)
	}

	log.Printf("Transcode job %d completed successfully", job.ID)

	// Update parent video status
	w.updateVideoStatusIfNeeded(job.VideoID)
}

// updateVideoStatusIfNeeded checks if all jobs for a video are complete and updates video status
func (w *Worker) updateVideoStatusIfNeeded(videoID string) {
	var totalJobs, completedJobs, failedJobs int64

	// Count job statuses
	w.db.Model(&models.TranscodeJob{}).Where("video_id = ?", videoID).Count(&totalJobs)
	w.db.Model(&models.TranscodeJob{}).Where("video_id = ? AND status = ?", videoID, models.StatusCompleted).Count(&completedJobs)
	w.db.Model(&models.TranscodeJob{}).Where("video_id = ? AND status = ?", videoID, models.StatusFailed).Count(&failedJobs)

	var newStatus models.TranscodeStatus
	var transcodedPaths models.StringMap = make(models.StringMap)

	if completedJobs == totalJobs {
		// All jobs completed successfully
		newStatus = models.StatusCompleted
		
		// Get all completed job paths
		var completedJobsList []models.TranscodeJob
		w.db.Where("video_id = ? AND status = ?", videoID, models.StatusCompleted).Find(&completedJobsList)
		
		for _, job := range completedJobsList {
			transcodedPaths[job.Quality] = job.OutputPath
		}
	} else if failedJobs == totalJobs {
		// All jobs failed
		newStatus = models.StatusFailed
	} else if completedJobs > 0 {
		// Some jobs completed
		newStatus = models.StatusCompleted // Partial success
		
		// Get completed job paths
		var completedJobsList []models.TranscodeJob
		w.db.Where("video_id = ? AND status = ?", videoID, models.StatusCompleted).Find(&completedJobsList)
		
		for _, job := range completedJobsList {
			transcodedPaths[job.Quality] = job.OutputPath
		}
	} else {
		// Jobs still processing
		return
	}

	// Update video status
	updates := map[string]interface{}{
		"status": newStatus,
	}
	
	if len(transcodedPaths) > 0 {
		updates["transcoded_paths"] = transcodedPaths
	}

	if err := w.db.Model(&models.Video{}).Where("id = ?", videoID).Updates(updates).Error; err != nil {
		log.Printf("Failed to update video status: %v", err)
	}

	log.Printf("Updated video %s status to %s", videoID, newStatus)
}
