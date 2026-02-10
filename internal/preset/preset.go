package preset

import (
	"fmt"
	"math"
)

// FileInfo represents source file information (used for dynamic preset calculation)
type FileInfo struct {
	Width     int
	Height    int
	Framerate float64
}

// GetAllPresets returns all available SyncLauper presets
func GetAllPresets() []Preset {
	return []Preset{
		{
			Name:         "원본 설정 유지",
			Resolution:   "source",
			Framerate:    "source",
			Width:        0,
			Height:       0,
			Level:        "auto",
			FPS:          0,
			UseSourceFPS: true,
			UseSourceRes: true,
		},
		{
			Name:       "HEVC 4K|60p",
			Resolution: "4K",
			Framerate:  "60",
			Width:      3840,
			Height:     2160,
			Level:      "5.1",
			FPS:        60,
		},
		{
			Name:       "HEVC 4K|30p",
			Resolution: "4K",
			Framerate:  "30",
			Width:      3840,
			Height:     2160,
			Level:      "5.0",
			FPS:        30,
		},
		{
			Name:       "HEVC 4K|29.97p",
			Resolution: "4K",
			Framerate:  "29.97",
			Width:      3840,
			Height:     2160,
			Level:      "5.0",
			FPS:        29.97,
		},
		{
			Name:       "HEVC 4K|24p",
			Resolution: "4K",
			Framerate:  "24",
			Width:      3840,
			Height:     2160,
			Level:      "5.0",
			FPS:        24,
		},
		{
			Name:       "HEVC 4K|23.976p",
			Resolution: "4K",
			Framerate:  "23.976",
			Width:      3840,
			Height:     2160,
			Level:      "5.0",
			FPS:        23.976,
		},
		{
			Name:       "HEVC 1080p|60p",
			Resolution: "1080p",
			Framerate:  "60",
			Width:      1920,
			Height:     1080,
			Level:      "4.1",
			FPS:        60,
		},
		{
			Name:       "HEVC 1080p|30p",
			Resolution: "1080p",
			Framerate:  "30",
			Width:      1920,
			Height:     1080,
			Level:      "4.1",
			FPS:        30,
		},
		{
			Name:       "HEVC 1080p|29.97p",
			Resolution: "1080p",
			Framerate:  "29.97",
			Width:      1920,
			Height:     1080,
			Level:      "4.1",
			FPS:        29.97,
		},
		{
			Name:       "HEVC 1080p|24p",
			Resolution: "1080p",
			Framerate:  "24",
			Width:      1920,
			Height:     1080,
			Level:      "4.1",
			FPS:        24,
		},
		{
			Name:       "HEVC 1080p|23.976p",
			Resolution: "1080p",
			Framerate:  "23.976",
			Width:      1920,
			Height:     1080,
			Level:      "4.1",
			FPS:        23.976,
		},
	}
}

// GetPresetByName returns a preset by its name
func GetPresetByName(name string) *Preset {
	for _, p := range GetAllPresets() {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// DetermineLevel determines the appropriate H.265 level based on resolution and framerate
func DetermineLevel(width, height int, fps float64) string {
	is4K := width > 1920 || height > 1080
	isHighFPS := fps > 30

	if is4K && isHighFPS {
		return "5.1"
	} else if is4K {
		return "5.0"
	}
	return "4.1"
}

// ToFFmpegArgs converts a preset to FFmpeg arguments
func (p *Preset) ToFFmpegArgs(inputPath, outputPath string, sourceInfo *FileInfo) []string {
	return p.ToFFmpegArgsWithEncoder(inputPath, outputPath, sourceInfo, "libx265", 0)
}

// ToFFmpegArgsWithEncoder converts a preset to FFmpeg arguments with specified encoder
func (p *Preset) ToFFmpegArgsWithEncoder(inputPath, outputPath string, sourceInfo *FileInfo, encoderID string, quality int) []string {
	settings := DefaultSettings()
	if quality > 0 {
		settings.Quality = quality
	}

	// Determine effective values for source-based preset
	effectiveWidth := p.Width
	effectiveHeight := p.Height
	effectiveFPS := p.FPS
	effectiveLevel := p.Level

	if p.UseSourceRes && sourceInfo != nil {
		effectiveWidth = sourceInfo.Width
		effectiveHeight = sourceInfo.Height
	}

	if p.UseSourceFPS && sourceInfo != nil {
		effectiveFPS = sourceInfo.Framerate
	}

	// Calculate level for "auto" or source-based preset
	if effectiveLevel == "auto" && sourceInfo != nil {
		effectiveLevel = DetermineLevel(effectiveWidth, effectiveHeight, effectiveFPS)
	}

	// Build keyint for GOP settings
	keyint := int(math.Round(effectiveFPS))
	if keyint <= 0 {
		keyint = 30 // fallback
	}

	args := []string{
		"-i", inputPath,
	}

	// Add encoder-specific video codec options
	args = append(args, getEncoderArgs(encoderID, settings, effectiveLevel, keyint, effectiveWidth, effectiveHeight)...)

	// Add resolution if not using source
	if !p.UseSourceRes && p.Width > 0 && p.Height > 0 {
		args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", p.Width, p.Height))
	}

	// Add framerate if not using source
	if !p.UseSourceFPS && p.FPS > 0 {
		args = append(args, "-r", fmt.Sprintf("%.3f", p.FPS))
	}

	// CFR mode
	if settings.CFR {
		args = append(args, "-vsync", "cfr")
	}

	// Audio settings
	args = append(args,
		"-c:a", "aac",
		"-b:a", fmt.Sprintf("%dk", settings.AudioBitrate),
		"-ac", "2",
	)

	// Deinterlace filter (equivalent to HandBrake's decomb)
	if settings.Decomb {
		// Use yadif in auto mode: only deinterlace if input is interlaced
		if !p.UseSourceRes && p.Width > 0 {
			// Already have a -vf, need to prepend yadif
			for i, arg := range args {
				if arg == "-vf" && i+1 < len(args) {
					args[i+1] = "yadif=mode=0:parity=-1:deint=1," + args[i+1]
					break
				}
			}
		} else {
			args = append(args, "-vf", "yadif=mode=0:parity=-1:deint=1")
		}
	}

	// Output format
	args = append(args, "-f", "matroska")

	// Output path (must be last)
	args = append(args, outputPath)

	return args
}

// GetPresetInfo returns a human-readable description of the preset
func (p *Preset) GetPresetInfo() string {
	if p.UseSourceRes && p.UseSourceFPS {
		return "원본 해상도 및 프레임레이트 유지, HEVC 인코딩"
	}
	return fmt.Sprintf("%s @ %sfps, HEVC 인코딩", p.Resolution, p.Framerate)
}

// getEncoderArgs returns encoder-specific FFmpeg arguments
func getEncoderArgs(encoderID string, settings EncodingSettings, level string, keyint int, width int, height int) []string {
	switch encoderID {
	case "hevc_videotoolbox":
		// Apple VideoToolbox (macOS)
		// Calculate bitrate based on resolution to match libx265 CRF quality
		// Base: ~3 Mbps for 1080p, scales linearly with pixel count
		pixels := width * height
		if pixels == 0 {
			pixels = 1920 * 1080 // default to 1080p
		}
		baseBitrateK := float64(pixels) / float64(1920*1080) * 3000
		// Quality adjustment: each CRF point changes bitrate by ~12%
		qualityFactor := math.Pow(1.12, float64(22-settings.Quality))
		bitrateK := int(baseBitrateK * qualityFactor)
		if bitrateK < 500 {
			bitrateK = 500
		}
		if bitrateK > 50000 {
			bitrateK = 50000
		}
		return []string{
			"-c:v", "hevc_videotoolbox",
			"-b:v", fmt.Sprintf("%dk", bitrateK),
			"-tag:v", "hvc1",
			"-allow_sw", "1",
		}

	case "hevc_nvenc":
		// NVIDIA NVENC
		// CQ mode with quality value (0-51, lower is better)
		return []string{
			"-c:v", "hevc_nvenc",
			"-rc", "vbr",
			"-cq", fmt.Sprintf("%d", settings.Quality),
			"-preset", mapNvencPreset(settings.EncoderPreset),
			"-profile:v", settings.EncoderProfile,
			"-level:v", level,
			"-g", fmt.Sprintf("%d", keyint),
			"-bf", "3",
		}

	case "hevc_qsv":
		// Intel QuickSync
		return []string{
			"-c:v", "hevc_qsv",
			"-global_quality", fmt.Sprintf("%d", settings.Quality),
			"-preset", mapQsvPreset(settings.EncoderPreset),
			"-profile:v", settings.EncoderProfile,
			"-level:v", level,
			"-g", fmt.Sprintf("%d", keyint),
		}

	case "hevc_amf":
		// AMD AMF
		return []string{
			"-c:v", "hevc_amf",
			"-rc", "cqp",
			"-qp_i", fmt.Sprintf("%d", settings.Quality),
			"-qp_p", fmt.Sprintf("%d", settings.Quality),
			"-quality", mapAmfQuality(settings.EncoderPreset),
			"-profile:v", settings.EncoderProfile,
			"-level:v", level,
			"-gops_per_idr", "1",
		}

	case "hevc_vaapi":
		// Linux VAAPI
		return []string{
			"-c:v", "hevc_vaapi",
			"-qp", fmt.Sprintf("%d", settings.Quality),
			"-profile:v", settings.EncoderProfile,
			"-level:v", level,
			"-g", fmt.Sprintf("%d", keyint),
		}

	default:
		// libx265 (software)
		x265Params := fmt.Sprintf(
			"keyint=%d:min-keyint=%d:open-gop=0:scenecut=0:repeat-headers=1:ref=4:bframes=3:hrd=1",
			keyint, keyint,
		)
		return []string{
			"-c:v", "libx265",
			"-crf", fmt.Sprintf("%d", settings.Quality),
			"-preset", settings.EncoderPreset,
			"-tune", settings.EncoderTune,
			"-profile:v", settings.EncoderProfile,
			"-level:v", level,
			"-x265-params", x265Params,
		}
	}
}

// mapNvencPreset maps x265 preset names to NVENC preset names
func mapNvencPreset(preset string) string {
	switch preset {
	case "ultrafast", "superfast", "veryfast":
		return "p1"
	case "faster", "fast":
		return "p4"
	case "medium":
		return "p5"
	case "slow":
		return "p6"
	case "slower", "veryslow":
		return "p7"
	default:
		return "p4"
	}
}

// mapQsvPreset maps x265 preset names to QSV preset names
func mapQsvPreset(preset string) string {
	switch preset {
	case "ultrafast", "superfast", "veryfast":
		return "veryfast"
	case "faster", "fast":
		return "fast"
	case "medium":
		return "medium"
	case "slow", "slower", "veryslow":
		return "slow"
	default:
		return "fast"
	}
}

// mapAmfQuality maps x265 preset to AMF quality
func mapAmfQuality(preset string) string {
	switch preset {
	case "ultrafast", "superfast", "veryfast", "faster", "fast":
		return "speed"
	case "medium":
		return "balanced"
	case "slow", "slower", "veryslow":
		return "quality"
	default:
		return "balanced"
	}
}
