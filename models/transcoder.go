package models

import (
	"time"

	"gorm.io/gorm"
)


type TranscodeStatus string

const (
    StatusQueued   TranscodeStatus = "queued"
    StatusRunning  TranscodeStatus = "running"
    StatusSuccess  TranscodeStatus = "success"
    StatusFailed   TranscodeStatus = "failed"
)


type Video struct {
	ID        uint   `gorm:"primaryKey"`
	FileName  string
	FilePath  string
	FileSize  int64  // Size in bytes
	Format    string
	Duration  int    // Duration in seconds
	CreatedAt int64  // Unix timestamp
}

type TranscodeJob struct {
    ID               uint   `gorm:"primaryKey" json:"id"`
	VideoID          uint   `json:"video_id" gorm:"not null"`
    OriginalFilename string `json:"original_filename"`
    InputPath        string `json:"input_path"`
    OutputPath       string `json:"output_path"`
    FileSize         int64  `json:"file_size"` // Size in bytes
    Preset           string `json:"preset"` // e.g., "mp4_720p_h264" - flexible for future
    Status           TranscodeStatus `json:"status" gorm:"type:text"`
    ErrorMessage     string `json:"error_message,omitempty"`
    CreatedAt        time.Time
    UpdatedAt        time.Time
    DeletedAt        gorm.DeletedAt `gorm:"index"`
}