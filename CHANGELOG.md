# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-02-10

### Added

- **Hardware encoding support** - Auto-detect and select GPU encoders
  - macOS: Apple VideoToolbox (`hevc_videotoolbox`)
  - Windows: NVIDIA NVENC (`hevc_nvenc`), Intel QuickSync (`hevc_qsv`), AMD AMF (`hevc_amf`)
  - Linux: VAAPI (`hevc_vaapi`)
  - Automatic fallback to software encoding (`libx265`) when hardware unavailable
- **Encoder selection UI** - Dropdown to choose between available encoders
- **GitHub Actions workflow** - Automated builds for Windows and macOS (Intel + Apple Silicon)
- **FFmpeg bundling** - Release packages now include FFmpeg essentials

### Changed

- Updated README with macOS installation instructions
- Updated README with hardware encoding documentation
- Updated build instructions for all platforms

## [1.0.0] - 2025-02-09

### Added

- Initial release
- HEVC (H.265) batch encoding with FFmpeg
- Drag & drop file support
- Native file parsers for MP4/MOV/MKV/WebM/AVI
- Duration mismatch detection for sync playback
- Upscale warning alerts
- Real-time encoding progress with ETA
- Preset system (4K/1080p at various framerates)
- Source-preserving mode
