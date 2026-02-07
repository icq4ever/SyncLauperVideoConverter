package preset

// Preset represents a video encoding preset
type Preset struct {
	Name         string  `json:"name"`
	Resolution   string  `json:"resolution"`   // "4K", "1080p", or "source"
	Framerate    string  `json:"framerate"`    // "60", "30", "29.97", "24", "23.976", or "source"
	Width        int     `json:"width"`        // 0 = use source
	Height       int     `json:"height"`       // 0 = use source
	Level        string  `json:"level"`        // "5.1", "5.0", "4.1", "auto"
	FPS          float64 `json:"fps"`          // numeric framerate value
	UseSourceFPS bool    `json:"useSourceFps"` // true = use source framerate
	UseSourceRes bool    `json:"useSourceRes"` // true = use source resolution
}

// EncodingSettings contains the common encoding settings for all presets
type EncodingSettings struct {
	Encoder        string `json:"encoder"`        // "x265"
	EncoderPreset  string `json:"encoderPreset"`  // "fast"
	EncoderTune    string `json:"encoderTune"`    // "fastdecode"
	EncoderProfile string `json:"encoderProfile"` // "main"
	Quality        int    `json:"quality"`        // 22
	AudioEncoder   string `json:"audioEncoder"`   // "av_aac"
	AudioBitrate   int    `json:"audioBitrate"`   // 160
	AudioMixdown   string `json:"audioMixdown"`   // "stereo"
	Format         string `json:"format"`         // "av_mkv"
	MultiPass      bool   `json:"multiPass"`      // true
	TurboFirstPass bool   `json:"turboFirstPass"` // true
	Decomb         bool   `json:"decomb"`         // true
	CFR            bool   `json:"cfr"`            // true (constant framerate)
}

// DefaultSettings returns the default encoding settings for SyncLauper
func DefaultSettings() EncodingSettings {
	return EncodingSettings{
		Encoder:        "x265",
		EncoderPreset:  "fast",
		EncoderTune:    "fastdecode",
		EncoderProfile: "main",
		Quality:        22,
		AudioEncoder:   "av_aac",
		AudioBitrate:   160,
		AudioMixdown:   "stereo",
		Format:         "av_mkv",
		MultiPass:      false, // Disabled: causes HandBrakeCLI to hang
		TurboFirstPass: false,
		Decomb:         true,
		CFR:            true,
	}
}
