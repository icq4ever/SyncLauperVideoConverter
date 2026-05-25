package preset

import (
	"fmt"
	"math"
	"strings"
)

// FileInfo represents source file information (used for dynamic preset calculation)
type FileInfo struct {
	Width     int
	Height    int
	Framerate float64
	HasAudio  bool
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
	return p.ToFFmpegArgsWithEncoder(inputPath, outputPath, sourceInfo, "libx265", 0, 0, 0, 0)
}

// rotateFilter returns the ffmpeg filter expression for a given rotation in degrees,
// or "" if rotation is not applied.
func rotateFilter(degrees int) string {
	switch degrees {
	case 90:
		return "transpose=1"
	case 180:
		return "transpose=2,transpose=2"
	case 270:
		return "transpose=2"
	}
	return ""
}

// ToFFmpegArgsWithEncoder converts a preset to FFmpeg arguments with specified encoder
func (p *Preset) ToFFmpegArgsWithEncoder(inputPath, outputPath string, sourceInfo *FileInfo, encoderID string, quality int, blackIntroDuration int, blackOutroDuration int, rotation int) []string {
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

	hasAudio := sourceInfo != nil && sourceInfo.HasAudio

	if blackIntroDuration > 0 || blackOutroDuration > 0 {
		return p.buildArgsWithBlackPad(inputPath, outputPath, settings, encoderID, effectiveWidth, effectiveHeight, effectiveFPS, effectiveLevel, keyint, blackIntroDuration, blackOutroDuration, rotation, hasAudio)
	}

	return p.buildStandardArgs(inputPath, outputPath, settings, encoderID, effectiveWidth, effectiveHeight, effectiveFPS, effectiveLevel, keyint, rotation, hasAudio)
}

// buildStandardArgs builds FFmpeg args without black intro (original logic)
func (p *Preset) buildStandardArgs(inputPath, outputPath string, settings EncodingSettings, encoderID string, effectiveWidth, effectiveHeight int, effectiveFPS float64, effectiveLevel string, keyint int, rotation int, hasAudio bool) []string {
	args := []string{}

	// Add pre-input args for hardware encoders (must come before -i)
	args = append(args, getPreInputArgs(encoderID)...)

	args = append(args, "-i", inputPath)

	// If the source has no audio, inject a near-silent white noise track so the
	// output always carries an audio stream (keeps SyncLauper playback consistent).
	if !hasAudio {
		args = append(args,
			"-f", "lavfi", "-i", "anoisesrc=color=white:amplitude=0.001",
			"-map", "0:v", "-map", "1:a", "-shortest",
		)
	}

	// For 90/270 rotation the output dimensions are swapped — adjust the level
	// calculation to reflect the post-rotation frame size.
	encWidth, encHeight := effectiveWidth, effectiveHeight
	if rotation == 90 || rotation == 270 {
		encWidth, encHeight = effectiveHeight, effectiveWidth
	}

	// Add encoder-specific video codec options
	args = append(args, getEncoderArgs(encoderID, settings, effectiveLevel, keyint, encWidth, encHeight)...)

	// Build the video filter chain: decomb → scale → rotate
	filters := []string{}
	if settings.Decomb {
		filters = append(filters, "yadif=mode=0:parity=-1:deint=1")
	}
	if !p.UseSourceRes && p.Width > 0 && p.Height > 0 {
		filters = append(filters, fmt.Sprintf("scale=%d:%d", p.Width, p.Height))
	}
	if rf := rotateFilter(rotation); rf != "" {
		filters = append(filters, rf)
	}
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
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

	// Output format
	args = append(args, "-f", "matroska")
	args = append(args, outputPath)

	return args
}

// buildArgsWithBlackPad builds FFmpeg args with black intro and/or outro padding
func (p *Preset) buildArgsWithBlackPad(inputPath, outputPath string, settings EncodingSettings, encoderID string, effectiveWidth, effectiveHeight int, effectiveFPS float64, effectiveLevel string, keyint int, introDuration, outroDuration int, rotation int, hasAudio bool) []string {
	fpsStr := fmt.Sprintf("%.3f", effectiveFPS)

	// Output (post-rotation) dimensions used for black pad frames and encoder level.
	padWidth, padHeight := effectiveWidth, effectiveHeight
	if rotation == 90 || rotation == 270 {
		padWidth, padHeight = effectiveHeight, effectiveWidth
	}

	// Add pre-input args for hardware encoders (must come before -i)
	args := getPreInputArgs(encoderID)

	// Build inputs in order: [intro v], [intro a], [src], [outro v], [outro a]
	inputIdx := 0
	introVIdx, introAIdx := -1, -1
	outroVIdx, outroAIdx := -1, -1

	if introDuration > 0 {
		args = append(args,
			"-f", "lavfi", "-i", fmt.Sprintf("color=black:s=%dx%d:d=%d:r=%s", padWidth, padHeight, introDuration, fpsStr),
		)
		introVIdx = inputIdx
		inputIdx++
		args = append(args,
			"-f", "lavfi", "-t", fmt.Sprintf("%d", introDuration), "-i", "anullsrc=r=48000:cl=stereo",
		)
		introAIdx = inputIdx
		inputIdx++
	}

	args = append(args, "-i", inputPath)
	srcIdx := inputIdx
	inputIdx++

	// If source has no audio stream, add a near-silent white noise track and use
	// it in place of the missing source audio during concat.
	srcAudioIdx := srcIdx
	if !hasAudio {
		args = append(args,
			"-f", "lavfi", "-i", "anoisesrc=color=white:amplitude=0.001:sample_rate=48000",
		)
		srcAudioIdx = inputIdx
		inputIdx++
	}

	if outroDuration > 0 {
		args = append(args,
			"-f", "lavfi", "-i", fmt.Sprintf("color=black:s=%dx%d:d=%d:r=%s", padWidth, padHeight, outroDuration, fpsStr),
		)
		outroVIdx = inputIdx
		inputIdx++
		args = append(args,
			"-f", "lavfi", "-t", fmt.Sprintf("%d", outroDuration), "-i", "anullsrc=r=48000:cl=stereo",
		)
		outroAIdx = inputIdx
		inputIdx++
	}

	// Build per-source video filter chain: decomb → scale → rotate
	srcFilters := []string{}
	if settings.Decomb {
		srcFilters = append(srcFilters, "yadif=mode=0:parity=-1:deint=1")
	}
	if !p.UseSourceRes && p.Width > 0 && p.Height > 0 {
		srcFilters = append(srcFilters, fmt.Sprintf("scale=%d:%d", p.Width, p.Height))
	}
	if rf := rotateFilter(rotation); rf != "" {
		srcFilters = append(srcFilters, rf)
	}

	srcVideoLabel := fmt.Sprintf("[%d:v]", srcIdx)
	videoFilter := ""
	if len(srcFilters) > 0 {
		videoFilter = fmt.Sprintf("[%d:v]%s[srcv];", srcIdx, strings.Join(srcFilters, ","))
		srcVideoLabel = "[srcv]"
	}

	// Build concat segments
	parts := ""
	n := 0
	if introDuration > 0 {
		parts += fmt.Sprintf("[%d:v][%d:a]", introVIdx, introAIdx)
		n++
	}
	parts += fmt.Sprintf("%s[%d:a]", srcVideoLabel, srcAudioIdx)
	n++
	if outroDuration > 0 {
		parts += fmt.Sprintf("[%d:v][%d:a]", outroVIdx, outroAIdx)
		n++
	}

	filterComplex := fmt.Sprintf("%s%sconcat=n=%d:v=1:a=1[v][a]", videoFilter, parts, n)
	args = append(args, "-filter_complex", filterComplex)
	args = append(args, "-map", "[v]", "-map", "[a]")

	// Add encoder-specific video codec options (use post-rotation dimensions)
	args = append(args, getEncoderArgs(encoderID, settings, effectiveLevel, keyint, padWidth, padHeight)...)

	// Add framerate if not using source
	if !p.UseSourceFPS && p.FPS > 0 {
		args = append(args, "-r", fpsStr)
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

	// Output format
	args = append(args, "-f", "matroska")
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

// getPreInputArgs returns FFmpeg arguments that must appear before -i for hardware encoders
func getPreInputArgs(encoderID string) []string {
	switch encoderID {
	case "hevc_qsv":
		return []string{"-init_hw_device", "qsv=hw", "-filter_hw_device", "hw"}
	case "hevc_vaapi":
		return []string{"-init_hw_device", "vaapi=hw:/dev/dri/renderD128", "-filter_hw_device", "hw"}
	default:
		return nil
	}
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
		// Note: B-frames removed for compatibility with older NVIDIA GPUs (e.g., Quadro P1000)
		return []string{
			"-c:v", "hevc_nvenc",
			"-rc", "vbr",
			"-cq", fmt.Sprintf("%d", settings.Quality),
			"-preset", mapNvencPreset(settings.EncoderPreset),
			"-profile:v", settings.EncoderProfile,
			"-level:v", level,
			"-g", fmt.Sprintf("%d", keyint),
		}

	case "hevc_qsv":
		// Intel QuickSync
		// Use CQP mode for maximum compatibility (ICQ/global_quality not supported on all GPUs)
		// Use low_power mode for newer Intel GPUs (Iris Xe etc.) that only support VDENC path
		profile := settings.EncoderProfile
		if profile == "main10" {
			profile = "main" // Fallback for compatibility
		}
		return []string{
			"-c:v", "hevc_qsv",
			"-low_power", "1",
			"-rc:v", "CQP",
			"-qp", fmt.Sprintf("%d", settings.Quality),
			"-preset", mapQsvPreset(settings.EncoderPreset),
			"-profile:v", profile,
			"-g", fmt.Sprintf("%d", keyint),
		}

	case "hevc_amf":
		// AMD AMF
		// Note: Older AMD GPUs may not support main10 profile
		profile := settings.EncoderProfile
		if profile == "main10" {
			profile = "main" // Fallback for compatibility
		}
		return []string{
			"-c:v", "hevc_amf",
			"-rc", "cqp",
			"-qp_i", fmt.Sprintf("%d", settings.Quality),
			"-qp_p", fmt.Sprintf("%d", settings.Quality),
			"-quality", mapAmfQuality(settings.EncoderPreset),
			"-profile:v", profile,
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
