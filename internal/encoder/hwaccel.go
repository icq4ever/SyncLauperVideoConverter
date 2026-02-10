package encoder

import (
	"os/exec"
	"runtime"
	"strings"

	"syncLauperVideoConverter/internal/cmdutil"
)

// HWEncoder represents a hardware encoder
type HWEncoder struct {
	ID          string `json:"id"`          // FFmpeg encoder name (e.g., "hevc_videotoolbox")
	Name        string `json:"name"`        // Display name (e.g., "Apple VideoToolbox")
	Description string `json:"description"` // Short description
	Available   bool   `json:"available"`   // Whether the encoder is available
	Priority    int    `json:"priority"`    // Higher = preferred
}

// GetAvailableHWEncoders returns all available HEVC hardware encoders
func (f *FFmpeg) GetAvailableHWEncoders() []HWEncoder {
	// Define all known HEVC hardware encoders
	allEncoders := getKnownHWEncoders()

	// Get list of encoders from FFmpeg
	availableEncoders := f.getFFmpegEncoders()

	// Check which encoders are available
	var result []HWEncoder
	for _, enc := range allEncoders {
		enc.Available = availableEncoders[enc.ID]
		if enc.Available {
			result = append(result, enc)
		}
	}

	// Always add software encoder as fallback
	result = append(result, HWEncoder{
		ID:          "libx265",
		Name:        "Software (x265)",
		Description: "CPU 인코딩 - 느리지만 호환성 최고",
		Available:   true,
		Priority:    0,
	})

	return result
}

// getKnownHWEncoders returns all known HEVC hardware encoders for the current platform
func getKnownHWEncoders() []HWEncoder {
	var encoders []HWEncoder

	switch runtime.GOOS {
	case "darwin":
		encoders = []HWEncoder{
			{
				ID:          "hevc_videotoolbox",
				Name:        "Apple VideoToolbox",
				Description: "macOS 하드웨어 인코딩 - 빠르고 효율적",
				Priority:    100,
			},
		}
	case "windows":
		encoders = []HWEncoder{
			{
				ID:          "hevc_nvenc",
				Name:        "NVIDIA NVENC",
				Description: "NVIDIA GPU 인코딩 - 매우 빠름",
				Priority:    100,
			},
			{
				ID:          "hevc_qsv",
				Name:        "Intel QuickSync",
				Description: "Intel 내장 GPU 인코딩",
				Priority:    90,
			},
			{
				ID:          "hevc_amf",
				Name:        "AMD AMF",
				Description: "AMD GPU 인코딩",
				Priority:    80,
			},
		}
	case "linux":
		encoders = []HWEncoder{
			{
				ID:          "hevc_nvenc",
				Name:        "NVIDIA NVENC",
				Description: "NVIDIA GPU 인코딩 - 매우 빠름",
				Priority:    100,
			},
			{
				ID:          "hevc_vaapi",
				Name:        "VAAPI",
				Description: "Linux VA-API 인코딩 (Intel/AMD)",
				Priority:    90,
			},
		}
	}

	return encoders
}

// getFFmpegEncoders returns a map of available encoders from FFmpeg
func (f *FFmpeg) getFFmpegEncoders() map[string]bool {
	result := make(map[string]bool)

	cmd := exec.Command(f.config.ExecutablePath, "-encoders", "-hide_banner")
	cmdutil.HideWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return result
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Encoder lines start with type flags like "V..... hevc_nvenc"
		if len(line) > 7 && line[0] == 'V' {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				encoderName := parts[1]
				// Only track HEVC encoders
				if strings.Contains(encoderName, "hevc") || strings.Contains(encoderName, "265") {
					result[encoderName] = true
				}
			}
		}
	}

	return result
}

// GetBestEncoder returns the best available encoder (highest priority)
func (f *FFmpeg) GetBestEncoder() HWEncoder {
	encoders := f.GetAvailableHWEncoders()
	if len(encoders) == 0 {
		return HWEncoder{
			ID:        "libx265",
			Name:      "Software (x265)",
			Available: true,
			Priority:  0,
		}
	}

	best := encoders[0]
	for _, enc := range encoders {
		if enc.Priority > best.Priority {
			best = enc
		}
	}
	return best
}
