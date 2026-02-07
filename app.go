package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"syncLauperVideoConverter/internal/encoder"
	"syncLauperVideoConverter/internal/fileinfo"
	"syncLauperVideoConverter/internal/preset"
)

// App struct
type App struct {
	ctx       context.Context
	encoder   *encoder.Encoder
	files     []*fileinfo.FileInfo
	outputDir string
	mu        sync.RWMutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		encoder: encoder.NewEncoder(),
		files:   make([]*fileinfo.FileInfo, 0),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Set up encoder callbacks
	a.encoder.SetProgressCallback(func(progress *encoder.EncodingProgress) {
		runtime.EventsEmit(a.ctx, "encoding:progress", progress)
	})

	a.encoder.SetCompleteCallback(func(result *encoder.EncodeResult, job *encoder.EncodingJob) {
		runtime.EventsEmit(a.ctx, "encoding:fileComplete", map[string]interface{}{
			"success":    result.Success,
			"outputPath": result.OutputPath,
			"filename":   job.FileInfo.Name,
		})
	})

	a.encoder.SetErrorCallback(func(err error, job *encoder.EncodingJob) {
		runtime.EventsEmit(a.ctx, "encoding:error", map[string]interface{}{
			"error":    err.Error(),
			"filename": job.FileInfo.Name,
		})
	})

	a.encoder.SetAllCompleteCallback(func(completed int, failed int) {
		runtime.EventsEmit(a.ctx, "encoding:allComplete", map[string]interface{}{
			"completed": completed,
			"failed":    failed,
		})
	})

	// Set default output directory to user's Videos folder
	homeDir, _ := os.UserHomeDir()
	a.outputDir = filepath.Join(homeDir, "Videos", "SyncLauper")
}

// GetPresets returns all available presets
func (a *App) GetPresets() []preset.Preset {
	return preset.GetAllPresets()
}

// AddFilesResult represents the result of adding files
type AddFilesResult struct {
	Added  []*fileinfo.FileInfo `json:"added"`
	Errors []string             `json:"errors"`
}

// AddFiles adds video files and returns basic info immediately.
// Full metadata is loaded separately via LoadNextMetadata calls.
func (a *App) AddFiles(paths []string) AddFilesResult {
	var validPaths []string
	var errors []string

	a.mu.RLock()
	existingPaths := make(map[string]bool)
	for _, f := range a.files {
		existingPaths[f.Path] = true
	}
	a.mu.RUnlock()

	for _, path := range paths {
		if existingPaths[path] {
			continue
		}
		if !fileinfo.IsSupportedFormat(path) {
			errors = append(errors, fmt.Sprintf("%s: 지원하지 않는 형식입니다", filepath.Base(path)))
			continue
		}
		validPaths = append(validPaths, path)
	}

	// Create basic file info instantly (no ffprobe, just os.Stat)
	var added []*fileinfo.FileInfo
	for _, path := range validPaths {
		stat, err := os.Stat(path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", filepath.Base(path), err))
			continue
		}
		info := &fileinfo.FileInfo{
			Path:     path,
			Name:     filepath.Base(path),
			FileSize: stat.Size(),
			Duration: "분석 중...",
		}
		added = append(added, info)
	}

	// Add to files list immediately
	a.mu.Lock()
	a.files = append(a.files, added...)
	a.mu.Unlock()

	return AddFilesResult{
		Added:  added,
		Errors: errors,
	}
}

// LoadAllMetadata runs ffprobe for all files without metadata in parallel.
// Returns the updated file list.
func (a *App) LoadAllMetadata() []*fileinfo.FileInfo {
	// Find files without metadata
	a.mu.RLock()
	var pending []*fileinfo.FileInfo
	for _, f := range a.files {
		if f.Codec == "" {
			pending = append(pending, f)
		}
	}
	a.mu.RUnlock()

	if len(pending) == 0 {
		return nil
	}

	// Process all in parallel (up to 4 concurrent ffprobe)
	type probeResult struct {
		target *fileinfo.FileInfo
		info   *fileinfo.FileInfo
		err    error
	}

	results := make(chan probeResult, len(pending))
	sem := make(chan struct{}, 4)

	for _, f := range pending {
		go func(target *fileinfo.FileInfo) {
			sem <- struct{}{}
			defer func() { <-sem }()
			info, err := fileinfo.GetFileInfo(target.Path)
			results <- probeResult{target: target, info: info, err: err}
		}(f)
	}

	// Collect results
	for i := 0; i < len(pending); i++ {
		r := <-results
		a.mu.Lock()
		if r.err != nil {
			r.target.Codec = "error"
			r.target.Duration = "분석 실패"
		} else {
			r.target.Width = r.info.Width
			r.target.Height = r.info.Height
			r.target.Duration = r.info.Duration
			r.target.DurationSeconds = r.info.DurationSeconds
			r.target.Framerate = r.info.Framerate
			r.target.Codec = r.info.Codec
			r.target.AudioCodec = r.info.AudioCodec
		}
		a.mu.Unlock()
	}

	// Check duration mismatch
	a.mu.RLock()
	allFiles := make([]*fileinfo.FileInfo, len(a.files))
	copy(allFiles, a.files)
	a.mu.RUnlock()

	if len(allFiles) > 1 {
		result := fileinfo.CheckDurationMismatch(allFiles, 1.0)
		if result.HasMismatch {
			runtime.EventsEmit(a.ctx, "duration:mismatch", result)
		}
	}

	return allFiles
}

// RemoveFile removes a file from the list
func (a *App) RemoveFile(path string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, f := range a.files {
		if f.Path == path {
			a.files = append(a.files[:i], a.files[i+1:]...)
			return true
		}
	}
	return false
}

// ClearFiles clears all files from the list
func (a *App) ClearFiles() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.files = make([]*fileinfo.FileInfo, 0)
}

// GetFiles returns all added files
func (a *App) GetFiles() []*fileinfo.FileInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Return a copy
	files := make([]*fileinfo.FileInfo, len(a.files))
	copy(files, a.files)
	return files
}

// CheckDurationMismatch checks if files have different durations
func (a *App) CheckDurationMismatch() fileinfo.DurationCheckResult {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return fileinfo.CheckDurationMismatch(a.files, 1.0)
}

// SelectOutputFolder opens a folder selection dialog
func (a *App) SelectOutputFolder() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "출력 폴더 선택",
		DefaultDirectory: a.outputDir,
	})
	if err != nil || dir == "" {
		return a.outputDir
	}

	a.mu.Lock()
	a.outputDir = dir
	a.mu.Unlock()

	return dir
}

// GetOutputFolder returns the current output folder
func (a *App) GetOutputFolder() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.outputDir
}

// SetOutputFolder sets the output folder
func (a *App) SetOutputFolder(path string) error {
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("폴더를 찾을 수 없습니다: %v", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("유효한 폴더가 아닙니다")
	}

	a.mu.Lock()
	a.outputDir = path
	a.mu.Unlock()

	return nil
}

// StartEncoding starts encoding all files with the selected preset
func (a *App) StartEncoding(presetName string) error {
	a.mu.RLock()
	files := a.files
	outputDir := a.outputDir
	a.mu.RUnlock()

	if len(files) == 0 {
		return fmt.Errorf("변환할 파일이 없습니다")
	}

	// Get preset
	p := preset.GetPresetByName(presetName)
	if p == nil {
		return fmt.Errorf("프리셋을 찾을 수 없습니다: %s", presetName)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("출력 폴더를 생성할 수 없습니다: %v", err)
	}

	// Clear previous jobs and add new ones
	a.encoder.ClearJobs()

	for _, file := range files {
		_, err := a.encoder.AddJob(file.Path, outputDir, p)
		if err != nil {
			runtime.EventsEmit(a.ctx, "encoding:error", map[string]interface{}{
				"error":    err.Error(),
				"filename": file.Name,
			})
		}
	}

	// Start encoding
	if err := a.encoder.Start(); err != nil {
		return err
	}

	runtime.EventsEmit(a.ctx, "encoding:started", map[string]interface{}{
		"totalFiles": len(files),
		"preset":     presetName,
	})

	return nil
}

// CancelEncoding cancels the current encoding
func (a *App) CancelEncoding() {
	a.encoder.Cancel()
	runtime.EventsEmit(a.ctx, "encoding:cancelled", nil)
}

// IsEncoding returns whether encoding is in progress
func (a *App) IsEncoding() bool {
	return a.encoder.IsRunning()
}

// CheckFFmpeg checks if FFmpeg is installed
func (a *App) CheckFFmpeg() error {
	return a.encoder.CheckFFmpeg()
}

// GetFFmpegVersion returns the FFmpeg version
func (a *App) GetFFmpegVersion() (string, error) {
	return a.encoder.GetFFmpegVersion()
}

// OpenFileDialog opens a file selection dialog
func (a *App) OpenFileDialog() ([]string, error) {
	files, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "비디오 파일 선택",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "비디오 파일",
				Pattern:     "*.mp4;*.mov;*.avi;*.mkv;*.webm;*.m4v;*.wmv;*.flv;*.mts;*.m2ts;*.ts",
			},
			{
				DisplayName: "모든 파일",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// OpenOutputFolder opens the output folder in the file explorer
func (a *App) OpenOutputFolder() error {
	a.mu.RLock()
	dir := a.outputDir
	a.mu.RUnlock()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	runtime.BrowserOpenURL(a.ctx, "file://"+dir)
	return nil
}

// GetAppInfo returns application information
func (a *App) GetAppInfo() map[string]string {
	version, _ := a.encoder.GetFFmpegVersion()
	return map[string]string{
		"appName":        "SyncLauper VideoConverter",
		"appVersion":     "1.0.0",
		"ffmpegVersion":  version,
	}
}
