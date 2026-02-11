# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.1] - 2025-02-11

### Fixed

- **NVIDIA NVENC compatibility** - Removed B-frames option for compatibility with older NVIDIA GPUs (e.g., Quadro P1000)
- **Intel QuickSync compatibility** - Force `main` profile instead of `main10` for older Intel GPUs
- **AMD AMF compatibility** - Force `main` profile instead of `main10` for older AMD GPUs

## [1.1.0] - 2025-02-11

### Added

- **Hardware encoding support** - Auto-detect and select GPU encoders
  - macOS: Apple VideoToolbox (`hevc_videotoolbox`)
  - Windows: NVIDIA NVENC (`hevc_nvenc`), Intel QuickSync (`hevc_qsv`), AMD AMF (`hevc_amf`)
  - Linux: VAAPI (`hevc_vaapi`)
  - Automatic fallback to software encoding (`libx265`) when hardware unavailable
- **Encoder selection UI** - Dropdown to choose between available encoders
- **GitHub Actions workflow** - Automated builds for Windows and macOS (Intel + Apple Silicon)
- **FFmpeg bundling** - Release packages now include FFmpeg essentials
- **Encoding error details** - Show FFmpeg error messages in UI when encoding fails

### Fixed

- macOS `.app` bundle not finding ffmpeg/ffprobe in the same folder
- VideoToolbox (macOS) producing 0KB files due to unsupported `-profile:v` parameter
- FFmpeg error message hardcoded to `ffmpeg.exe` on macOS

### Changed

- Updated README with macOS installation instructions
- Updated README with hardware encoding documentation
- Updated build instructions for all platforms
- Clarified ffprobe is optional (only needed for WMV/FLV/TS fallback)
- Removed emoji icons from encoder dropdown
- Custom styled select dropdown (removed macOS native gradient)

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
