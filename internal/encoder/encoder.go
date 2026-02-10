package encoder

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"syncLauperVideoConverter/internal/fileinfo"
	"syncLauperVideoConverter/internal/preset"
)

// EncodingJob represents a single encoding job
type EncodingJob struct {
	ID         string             `json:"id"`
	InputPath  string             `json:"inputPath"`
	OutputPath string             `json:"outputPath"`
	Preset     *preset.Preset     `json:"preset"`
	FileInfo   *fileinfo.FileInfo `json:"fileInfo"`
	Status     string             `json:"status"`
	Progress   float64            `json:"progress"`
	Error      string             `json:"error,omitempty"`
}

// Encoder manages encoding jobs
type Encoder struct {
	ffmpeg          *FFmpeg
	jobs            []*EncodingJob
	currentJob      int
	isRunning       bool
	cancelFunc      context.CancelFunc
	cancelCtx       context.Context
	mu              sync.RWMutex
	progressCb      func(progress *EncodingProgress)
	completeCb      func(result *EncodeResult, job *EncodingJob)
	errorCb         func(err error, job *EncodingJob)
	allCompleteCb   func(completed int, failed int)
	selectedEncoder string // Selected encoder ID (e.g., "libx265", "hevc_nvenc")
}

// NewEncoder creates a new Encoder instance
func NewEncoder() *Encoder {
	return &Encoder{
		ffmpeg: NewFFmpeg(DefaultFFmpegConfig()),
		jobs:   make([]*EncodingJob, 0),
	}
}

// SetProgressCallback sets the callback for progress updates
func (e *Encoder) SetProgressCallback(cb func(progress *EncodingProgress)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.progressCb = cb
}

// SetCompleteCallback sets the callback for job completion
func (e *Encoder) SetCompleteCallback(cb func(result *EncodeResult, job *EncodingJob)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.completeCb = cb
}

// SetErrorCallback sets the callback for errors
func (e *Encoder) SetErrorCallback(cb func(err error, job *EncodingJob)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.errorCb = cb
}

// SetAllCompleteCallback sets the callback for when all jobs complete
func (e *Encoder) SetAllCompleteCallback(cb func(completed int, failed int)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.allCompleteCb = cb
}

// CheckFFmpeg checks if FFmpeg is available
func (e *Encoder) CheckFFmpeg() error {
	return e.ffmpeg.CheckInstalled()
}

// GetFFmpegVersion returns the FFmpeg version
func (e *Encoder) GetFFmpegVersion() (string, error) {
	return e.ffmpeg.GetVersion()
}

// GetAvailableEncoders returns all available HEVC encoders
func (e *Encoder) GetAvailableEncoders() []HWEncoder {
	return e.ffmpeg.GetAvailableHWEncoders()
}

// SetEncoder sets the encoder to use
func (e *Encoder) SetEncoder(encoderID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.selectedEncoder = encoderID
}

// GetSelectedEncoder returns the currently selected encoder
func (e *Encoder) GetSelectedEncoder() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.selectedEncoder == "" {
		return "libx265" // Default to software
	}
	return e.selectedEncoder
}

// AddJob adds a new encoding job to the queue
func (e *Encoder) AddJob(inputPath string, outputDir string, p *preset.Preset) (*EncodingJob, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get file info
	info, err := fileinfo.GetFileInfo(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	// Generate output path
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputPath := filepath.Join(outputDir, baseName+".mkv")

	job := &EncodingJob{
		ID:         fmt.Sprintf("job_%d", len(e.jobs)+1),
		InputPath:  inputPath,
		OutputPath: outputPath,
		Preset:     p,
		FileInfo:   info,
		Status:     StatusWaiting,
		Progress:   0,
	}

	e.jobs = append(e.jobs, job)
	return job, nil
}

// ClearJobs clears all jobs from the queue
func (e *Encoder) ClearJobs() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isRunning {
		return // Don't clear while running
	}

	e.jobs = make([]*EncodingJob, 0)
	e.currentJob = 0
}

// GetJobs returns all jobs
func (e *Encoder) GetJobs() []*EncodingJob {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Return a copy to prevent race conditions
	jobs := make([]*EncodingJob, len(e.jobs))
	copy(jobs, e.jobs)
	return jobs
}

// IsRunning returns whether encoding is in progress
func (e *Encoder) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isRunning
}

// Start starts encoding all jobs in the queue
func (e *Encoder) Start() error {
	e.mu.Lock()
	if e.isRunning {
		e.mu.Unlock()
		return fmt.Errorf("encoding already in progress")
	}

	if len(e.jobs) == 0 {
		e.mu.Unlock()
		return fmt.Errorf("no jobs in queue")
	}

	e.isRunning = true
	e.currentJob = 0
	e.cancelCtx, e.cancelFunc = context.WithCancel(context.Background())
	e.mu.Unlock()

	// Run in goroutine
	go e.processQueue()

	return nil
}

// Cancel cancels the current encoding
func (e *Encoder) Cancel() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cancelFunc != nil {
		e.cancelFunc()
	}
}

// processQueue processes all jobs in the queue
func (e *Encoder) processQueue() {
	for {
		e.mu.RLock()
		if e.currentJob >= len(e.jobs) {
			e.mu.RUnlock()
			break
		}

		job := e.jobs[e.currentJob]
		totalJobs := len(e.jobs)
		currentJobNum := e.currentJob + 1
		e.mu.RUnlock()

		// Update job status
		e.mu.Lock()
		job.Status = StatusEncoding
		e.mu.Unlock()

		// Build FFmpeg arguments with selected encoder
		sourceInfo := &preset.FileInfo{
			Width:     job.FileInfo.Width,
			Height:    job.FileInfo.Height,
			Framerate: job.FileInfo.Framerate,
		}
		encoderID := e.GetSelectedEncoder()
		args := job.Preset.ToFFmpegArgsWithEncoder(job.InputPath, job.OutputPath, sourceInfo, encoderID)

		// Progress callback wrapper
		progressWrapper := func(progress *EncodingProgress) {
			e.mu.Lock()
			job.Progress = progress.Progress
			e.mu.Unlock()

			// Add file and queue info
			progress.Filename = job.FileInfo.Name
			progress.CurrentFile = currentJobNum
			progress.TotalFiles = totalJobs

			if e.progressCb != nil {
				e.progressCb(progress)
			}
		}

		// Run encoding with duration for progress calculation
		result, err := e.ffmpeg.Encode(e.cancelCtx, args, job.FileInfo.DurationSeconds, progressWrapper)

		// Check for cancellation
		if e.cancelCtx.Err() == context.Canceled {
			e.mu.Lock()
			job.Status = StatusCancelled
			e.isRunning = false
			// Mark remaining jobs as cancelled
			for i := e.currentJob + 1; i < len(e.jobs); i++ {
				e.jobs[i].Status = StatusCancelled
			}
			e.mu.Unlock()
			return
		}

		// Handle result
		e.mu.Lock()
		if err != nil || !result.Success {
			job.Status = StatusError
			if err != nil {
				job.Error = err.Error()
			} else {
				job.Error = result.Error
			}

			if e.errorCb != nil {
				go e.errorCb(fmt.Errorf("%s", job.Error), job)
			}
		} else {
			job.Status = StatusCompleted
			job.Progress = 100

			if e.completeCb != nil {
				go e.completeCb(result, job)
			}
		}

		e.currentJob++
		e.mu.Unlock()
	}

	// All jobs completed
	e.mu.Lock()
	e.isRunning = false

	// Count completed and failed
	completed := 0
	failed := 0
	for _, j := range e.jobs {
		if j.Status == StatusCompleted {
			completed++
		} else if j.Status == StatusError {
			failed++
		}
	}

	if e.allCompleteCb != nil {
		go e.allCompleteCb(completed, failed)
	}
	e.mu.Unlock()
}

// GetCurrentProgress returns the current encoding progress
func (e *Encoder) GetCurrentProgress() *EncodingProgress {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.isRunning || e.currentJob >= len(e.jobs) {
		return nil
	}

	job := e.jobs[e.currentJob]
	return &EncodingProgress{
		Filename:    job.FileInfo.Name,
		Progress:    job.Progress,
		CurrentFile: e.currentJob + 1,
		TotalFiles:  len(e.jobs),
		Status:      job.Status,
	}
}
