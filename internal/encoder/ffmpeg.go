package encoder

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"syncLauperVideoConverter/internal/cmdutil"
)

// FFmpegConfig contains configuration for FFmpeg
type FFmpegConfig struct {
	ExecutablePath string // Path to ffmpeg executable
}

// DefaultFFmpegConfig returns the default configuration
func DefaultFFmpegConfig() FFmpegConfig {
	execName := "ffmpeg"
	if runtime.GOOS == "windows" {
		execName = "ffmpeg.exe"
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)

		// Check if ffmpeg is in the same directory as the executable
		localPath := filepath.Join(exeDir, execName)
		if _, err := os.Stat(localPath); err == nil {
			return FFmpegConfig{
				ExecutablePath: localPath,
			}
		}

		// On macOS, also check the directory containing the .app bundle
		// exePath is like: /path/to/App.app/Contents/MacOS/executable
		// We need to check: /path/to/ffmpeg
		if runtime.GOOS == "darwin" && strings.Contains(exePath, ".app/Contents/MacOS") {
			// Navigate up from Contents/MacOS to the .app's parent directory
			appBundleDir := exeDir // .../App.app/Contents/MacOS
			for i := 0; i < 3; i++ {
				appBundleDir = filepath.Dir(appBundleDir)
			}
			// appBundleDir is now the directory containing the .app
			bundlePath := filepath.Join(appBundleDir, execName)
			if _, err := os.Stat(bundlePath); err == nil {
				return FFmpegConfig{
					ExecutablePath: bundlePath,
				}
			}
		}
	}

	// Fall back to system PATH
	return FFmpegConfig{
		ExecutablePath: execName,
	}
}

// FFmpeg represents an FFmpeg CLI wrapper
type FFmpeg struct {
	config FFmpegConfig
}

// NewFFmpeg creates a new FFmpeg instance
func NewFFmpeg(config FFmpegConfig) *FFmpeg {
	return &FFmpeg{
		config: config,
	}
}

// CheckInstalled checks if FFmpeg is installed and accessible
func (f *FFmpeg) CheckInstalled() error {
	cmd := exec.Command(f.config.ExecutablePath, "-version")
	cmdutil.HideWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("FFmpeg not found: %v", err)
	}

	if !strings.Contains(string(output), "ffmpeg") {
		return fmt.Errorf("invalid FFmpeg output")
	}

	return nil
}

// GetVersion returns the FFmpeg version
func (f *FFmpeg) GetVersion() (string, error) {
	cmd := exec.Command(f.config.ExecutablePath, "-version")
	cmdutil.HideWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return strings.TrimSpace(string(output)), nil
}

// EncodeResult represents the result of an encoding operation
type EncodeResult struct {
	Success    bool   `json:"success"`
	OutputPath string `json:"outputPath"`
	Error      string `json:"error,omitempty"`
}

// ProgressCallback is called with progress updates during encoding
type ProgressCallback func(progress *EncodingProgress)

// Encode runs FFmpeg with the given arguments
func (f *FFmpeg) Encode(ctx context.Context, args []string, durationSecs float64, progressCb ProgressCallback) (*EncodeResult, error) {
	// Add -progress pipe:1 to get structured progress output on stdout
	// Add -y to overwrite output without asking
	fullArgs := make([]string, 0, len(args)+4)
	fullArgs = append(fullArgs, "-y")
	fullArgs = append(fullArgs, "-progress", "pipe:1")
	fullArgs = append(fullArgs, args...)

	cmd := exec.CommandContext(ctx, f.config.ExecutablePath, fullArgs...)
	cmdutil.HideWindow(cmd)

	// Capture stdout (progress) and stderr (logs)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start FFmpeg: %v", err)
	}

	// Drain stderr in goroutine to prevent pipe blocking
	var stderrOutput strings.Builder
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			stderrOutput.WriteString(scanner.Text())
			stderrOutput.WriteString("\n")
		}
	}()

	// Parse progress from stdout (key=value format)
	scanner := bufio.NewScanner(stdout)
	var lastSpeed string

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "out_time_us":
			// out_time_us is in microseconds
			if us, err := strconv.ParseInt(value, 10, 64); err == nil && durationSecs > 0 {
				currentSecs := float64(us) / 1_000_000.0
				percent := (currentSecs / durationSecs) * 100
				if percent > 100 {
					percent = 100
				}
				if percent < 0 {
					percent = 0
				}

				if progressCb != nil {
					progress := &EncodingProgress{
						Progress:    percent,
						Status:      StatusEncoding,
						PassNumber:  1,
						TotalPasses: 1,
						Speed:       lastSpeed,
					}
					// Calculate ETA
					if lastSpeed != "" {
						progress.ETA = calculateETA(percent, lastSpeed)
					}
					progressCb(progress)
				}
			}
		case "speed":
			lastSpeed = value
		}
	}

	// Wait for command to finish
	err = cmd.Wait()

	// Check for context cancellation
	if ctx.Err() == context.Canceled {
		return &EncodeResult{
			Success: false,
			Error:   "encoding cancelled",
		}, nil
	}

	if err != nil {
		errMsg := err.Error()
		// Include stderr for more details
		if stderrOutput.Len() > 0 {
			// Get last few lines of stderr for the error message
			lines := strings.Split(stderrOutput.String(), "\n")
			start := len(lines) - 5
			if start < 0 {
				start = 0
			}
			errMsg = strings.Join(lines[start:], "\n")
		}
		return &EncodeResult{
			Success: false,
			Error:   errMsg,
		}, nil
	}

	// Extract output path from args (-o is not used in ffmpeg, output is last arg)
	outputPath := ""
	if len(args) > 0 {
		outputPath = args[len(args)-1]
	}

	return &EncodeResult{
		Success:    true,
		OutputPath: outputPath,
	}, nil
}

// calculateETA estimates remaining time based on progress and speed
func calculateETA(percent float64, speedStr string) string {
	// speedStr is like "1.5x" or "0.8x"
	speedStr = strings.TrimSuffix(speedStr, "x")
	speed, err := strconv.ParseFloat(speedStr, 64)
	if err != nil || speed <= 0 || percent <= 0 {
		return ""
	}

	// Rough ETA calculation
	elapsed := percent / speed
	remaining := (100 - percent) / speed
	if remaining <= 0 {
		return "00:00:00"
	}
	_ = elapsed

	eta := time.Duration(remaining/100*3600) * time.Second
	hours := int(eta.Hours())
	minutes := int(eta.Minutes()) % 60
	seconds := int(eta.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
