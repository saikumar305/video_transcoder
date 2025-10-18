package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dirName, err)
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filename string) (int64, error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// CleanupFile removes a file if it exists
func CleanupFile(filename string) error {
	if FileExists(filename) {
		return os.Remove(filename)
	}
	return nil
}

// GetVideoExtension returns the extension of a video file
func GetVideoExtension(filename string) string {
	return filepath.Ext(filename)
}

// IsValidVideoExtension checks if the file extension is a valid video format
func IsValidVideoExtension(filename string) bool {
	ext := filepath.Ext(filename)
	validExtensions := map[string]bool{
		".mp4":  true,
		".avi":  true,
		".mov":  true,
		".mkv":  true,
		".wmv":  true,
		".flv":  true,
		".webm": true,
		".m4v":  true,
		".3gp":  true,
		".mpg":  true,
		".mpeg": true,
	}
	return validExtensions[ext]
}

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration converts seconds to HH:MM:SS format
func FormatDuration(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}