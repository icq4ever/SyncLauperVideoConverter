package fileinfo

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"syncLauperVideoConverter/internal/cmdutil"
)

// getFFprobePath returns the path to ffprobe executable
func getFFprobePath() string {
	execName := "ffprobe"
	if runtime.GOOS == "windows" {
		execName = "ffprobe.exe"
	}

	// First, check if ffprobe is in the same directory as the executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		localPath := filepath.Join(exeDir, execName)
		if _, err := os.Stat(localPath); err == nil {
			return localPath
		}
	}

	// Fall back to system PATH
	return execName
}

// FileInfo represents video file information
type FileInfo struct {
	Path                string  `json:"path"`
	Name                string  `json:"name"`
	Width               int     `json:"width"`
	Height              int     `json:"height"`
	Duration            string  `json:"duration"`            // "HH:MM:SS" format
	DurationSeconds     float64 `json:"durationSeconds"`     // seconds (for duration comparison)
	Framerate           float64 `json:"framerate"`           // e.g., 29.97, 60.0
	Codec               string  `json:"codec"`               // e.g., "h264", "hevc"
	AudioCodec          string  `json:"audioCodec"`          // e.g., "aac", "ac3"
	FileSize            int64   `json:"fileSize"`            // bytes
	HasDurationMismatch bool    `json:"hasDurationMismatch"` // true if duration differs from other files
}

// DurationCheckResult represents the result of duration mismatch check
type DurationCheckResult struct {
	HasMismatch   bool                  `json:"hasMismatch"`
	BaseDuration  string                `json:"baseDuration"`
	Tolerance     float64               `json:"tolerance"` // in seconds
	MismatchFiles []DurationMismatchInfo `json:"mismatchFiles"`
}

// DurationMismatchInfo represents a file with duration mismatch
type DurationMismatchInfo struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Duration string `json:"duration"`
	Diff     string `json:"diff"` // e.g., "+5s", "-3s"
}

// ffprobeOutput represents the JSON output from ffprobe
type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	CodecType     string `json:"codec_type"`
	CodecName     string `json:"codec_name"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	RFrameRate    string `json:"r_frame_rate"`    // e.g., "30000/1001"
	AvgFrameRate  string `json:"avg_frame_rate"`  // e.g., "30000/1001"
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
	Size     string `json:"size"`
}

// GetFileInfo extracts video metadata. It tries native parsing first for
// MP4/MOV/M4V, MKV/WebM, and AVI files (instant, no external process).
// Falls back to ffprobe for other formats (WMV, FLV, TS, etc.).
func GetFileInfo(path string) (*FileInfo, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".mp4", ".mov", ".m4v":
		info, err := parseMP4(path)
		if err == nil {
			return info, nil
		}
	case ".mkv", ".webm":
		info, err := parseMKV(path)
		if err == nil {
			return info, nil
		}
	case ".avi":
		info, err := parseAVI(path)
		if err == nil {
			return info, nil
		}
	}

	// Fall back to ffprobe for unsupported native formats or parse errors
	return getFileInfoFFprobe(path)
}

// getFileInfoFFprobe extracts video metadata using ffprobe (external process)
func getFileInfoFFprobe(path string) (*FileInfo, error) {
	// Check if file exists
	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	// Run ffprobe with limited analysis for fast metadata extraction
	cmd := exec.Command(getFFprobePath(),
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		"-analyzeduration", "2000000",
		"-probesize", "2000000",
		path,
	)
	cmdutil.HideWindow(cmd)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %v", err)
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	info := &FileInfo{
		Path:     path,
		Name:     filepath.Base(path),
		FileSize: stat.Size(),
	}

	// Find video and audio streams
	for _, stream := range probe.Streams {
		if stream.CodecType == "video" && info.Codec == "" {
			info.Width = stream.Width
			info.Height = stream.Height
			info.Codec = stream.CodecName
			info.Framerate = parseFramerate(stream.RFrameRate)
			if info.Framerate == 0 {
				info.Framerate = parseFramerate(stream.AvgFrameRate)
			}
		}
		if stream.CodecType == "audio" && info.AudioCodec == "" {
			info.AudioCodec = stream.CodecName
		}
	}

	// Parse duration
	if probe.Format.Duration != "" {
		durationSec, _ := strconv.ParseFloat(probe.Format.Duration, 64)
		info.DurationSeconds = durationSec
		info.Duration = formatDuration(durationSec)
	}

	return info, nil
}

// parseFramerate parses a framerate string like "30000/1001" or "30"
func parseFramerate(s string) float64 {
	if s == "" || s == "0/0" {
		return 0
	}

	parts := strings.Split(s, "/")
	if len(parts) == 2 {
		num, _ := strconv.ParseFloat(parts[0], 64)
		den, _ := strconv.ParseFloat(parts[1], 64)
		if den != 0 {
			return num / den
		}
	}

	val, _ := strconv.ParseFloat(s, 64)
	return val
}

// formatDuration formats seconds to "HH:MM:SS" format
func formatDuration(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

// CheckDurationMismatch checks if files have different durations
// tolerance is in seconds (default 1 second)
func CheckDurationMismatch(files []*FileInfo, tolerance float64) DurationCheckResult {
	if tolerance <= 0 {
		tolerance = 1.0 // default 1 second tolerance
	}

	result := DurationCheckResult{
		HasMismatch:   false,
		Tolerance:     tolerance,
		MismatchFiles: []DurationMismatchInfo{},
	}

	if len(files) == 0 {
		return result
	}

	// Use first file as base
	baseDuration := files[0].DurationSeconds
	result.BaseDuration = files[0].Duration

	for _, file := range files {
		diff := file.DurationSeconds - baseDuration
		if math.Abs(diff) > tolerance {
			result.HasMismatch = true
			file.HasDurationMismatch = true

			diffStr := formatDiff(diff)
			result.MismatchFiles = append(result.MismatchFiles, DurationMismatchInfo{
				Path:     file.Path,
				Name:     file.Name,
				Duration: file.Duration,
				Diff:     diffStr,
			})
		}
	}

	return result
}

// formatDiff formats duration difference to a human-readable string
func formatDiff(seconds float64) string {
	if seconds >= 0 {
		return fmt.Sprintf("+%.1fs", seconds)
	}
	return fmt.Sprintf("%.1fs", seconds)
}

// GetFilesInfo extracts video metadata for multiple files
func GetFilesInfo(paths []string) ([]*FileInfo, []error) {
	var files []*FileInfo
	var errors []error

	for _, path := range paths {
		info, err := GetFileInfo(path)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %v", filepath.Base(path), err))
			continue
		}
		files = append(files, info)
	}

	return files, errors
}

// IsSupportedFormat checks if the file extension is supported
func IsSupportedFormat(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	supported := map[string]bool{
		".mp4":  true,
		".mov":  true,
		".avi":  true,
		".mkv":  true,
		".webm": true,
		".m4v":  true,
		".wmv":  true,
		".flv":  true,
		".mts":  true,
		".m2ts": true,
		".ts":   true,
	}
	return supported[ext]
}

// FormatFileSize formats file size to human-readable format
func FormatFileSize(bytes int64) string {
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
