# SyncLauper VideoConverter

[한국어](README.ko.md)

A batch video encoder for [SyncLauper](https://synclauper.studio42.kr) multi-screen synchronized video player. Easily converts video files to HEVC (H.265) codec.

## Features

- **Batch encoding** - Convert multiple video files at once
- **HEVC (H.265)** - High-efficiency encoding with customizable presets
- **Drag & drop** - Add files by dragging or clicking
- **Instant file analysis** - Native parsers for MP4/MOV/MKV/WebM/AVI (no ffprobe dependency for common formats)
- **Duration mismatch detection** - Warns if video durations differ (important for sync playback)
- **Upscale warning** - Alerts when resolution or framerate would be upscaled unnecessarily
- **Real-time progress** - Encoding progress with ETA and speed display
- **Preset system** - 4K/1080p at various framerates, or source-preserving mode

## Presets

| Preset | Resolution | Framerate |
|--------|-----------|-----------|
| Source | Original | Original |
| HEVC 4K | 3840x2160 | 60/30/29.97/24/23.976 fps |
| HEVC 1080p | 1920x1080 | 60/30/29.97/24/23.976 fps |

## Requirements

- **Windows 10/11** (WebView2 runtime, pre-installed on modern Windows)
- **FFmpeg** - Place `ffmpeg.exe` in the same folder as the application
  - [Download FFmpeg (essentials build recommended)](https://www.gyan.dev/ffmpeg/builds/)

## Installation

1. Download the latest release
2. Extract the zip file (right-click zip → Properties → Unblock before extracting)
3. Place `ffmpeg.exe` in the same folder as `syncLauperVideoConverter.exe`
4. Run `syncLauperVideoConverter.exe`

```
SyncLauperVideoConverter/
├── syncLauperVideoConverter.exe
└── ffmpeg.exe
```

## Building from Source

### Prerequisites

- [Go](https://go.dev/dl/) 1.21+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2
- [Node.js](https://nodejs.org/) 18+
- For Windows cross-compilation from Linux: `x86_64-w64-mingw32-gcc`

### Build

```bash
# Windows (native)
wails build

# Windows (cross-compile from WSL/Linux)
export CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc
wails build -platform windows/amd64
```

## Tech Stack

- **Backend**: Go + [Wails v2](https://wails.io/)
- **Frontend**: Svelte + TypeScript
- **Encoding**: FFmpeg (libx265)
- **File analysis**: Native Go parsers (MP4/MOV/M4V, MKV/WebM, AVI) + ffprobe fallback

## Project Structure

```
├── app.go                          # Main application logic
├── main.go                         # Entry point
├── wails.json                      # Wails configuration
├── internal/
│   ├── encoder/ffmpeg.go           # FFmpeg wrapper
│   ├── fileinfo/
│   │   ├── fileinfo.go             # File info dispatcher
│   │   ├── mp4parser.go            # Native MP4/MOV/M4V parser
│   │   ├── mkvparser.go            # Native MKV/WebM parser
│   │   └── aviparser.go            # Native AVI parser
│   ├── preset/preset.go            # Encoding presets
│   └── cmdutil/                    # Platform-specific utilities
└── frontend/src/
    ├── App.svelte                  # Main UI
    └── lib/
        ├── components/             # Svelte components
        ├── stores/                 # State management
        └── types/                  # TypeScript types
```

## License

Copyright (c) 2026 [studio42](https://studio42.kr). All rights reserved.

## Links

- [SyncLauper](https://synclauper.studio42.kr) - Multi-screen synchronized video player
- [studio42](https://studio42.kr)
