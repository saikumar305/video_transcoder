package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type TranscodeStatus string

const (
    StatusPending     TranscodeStatus = "pending"
    StatusQueued      TranscodeStatus = "queued"
    StatusProcessing  TranscodeStatus = "processing"
    StatusCompleted   TranscodeStatus = "completed"
    StatusFailed      TranscodeStatus = "failed"
)

// Custom type for storing map as JSON in PostgreSQL
type StringMap map[string]string

func (sm StringMap) Value() (driver.Value, error) {
    return json.Marshal(sm)
}

func (sm *StringMap) Scan(value interface{}) error {
    if value == nil {
        *sm = make(StringMap)
        return nil
    }
    
    bytes, ok := value.([]byte)
    if !ok {
        return nil
    }
    
    return json.Unmarshal(bytes, sm)
}

type Video struct {
    ID              string         `gorm:"primaryKey" json:"id"`
    OriginalPath    string         `gorm:"not null" json:"original_path"`
    OriginalFilename string        `gorm:"not null" json:"original_filename"`
    FileSize        int64          `json:"file_size"`
    Duration        float64        `json:"duration"` // in seconds
    TranscodedPaths StringMap      `gorm:"type:jsonb" json:"transcoded_paths"` // e.g., {"720p": "/path/720p.mp4"}
    Status          TranscodeStatus `gorm:"not null" json:"status"`
    ErrorMessage    string         `json:"error_message,omitempty"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
}

func (Video) TableName() string {
    return "videos"
}

type TranscodeJob struct {
    ID               uint            `gorm:"primaryKey" json:"id"`
    VideoID          string          `json:"video_id" gorm:"not null;index"`
    Video            Video           `gorm:"foreignKey:VideoID" json:"video,omitempty"`
    InputPath        string          `json:"input_path"`
    OutputPath       string          `json:"output_path"`
    Quality          string          `json:"quality"` // 720p, 480p, 1080p
    Status           TranscodeStatus `json:"status" gorm:"type:text"`
    Progress         int             `json:"progress"` // 0-100
    ErrorMessage     string          `json:"error_message,omitempty"`
    CreatedAt        time.Time       `json:"created_at"`
    UpdatedAt        time.Time       `json:"updated_at"`
    DeletedAt        gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (TranscodeJob) TableName() string {
    return "transcode_jobs"
}