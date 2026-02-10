# SyncLauper VideoConverter

[English](README.md)

[SyncLauper](https://synclauper.studio42.kr) 멀티스크린 동기화 영상 플레이어를 위한 배치 비디오 인코더입니다. 영상 파일을 HEVC (H.265) 코덱으로 손쉽게 변환할 수 있습니다.

## 주요 기능

- **일괄 인코딩** - 여러 영상 파일을 한 번에 변환
- **HEVC (H.265)** - 고효율 인코딩, 다양한 프리셋 제공
- **하드웨어 가속** - GPU 인코딩 자동 감지 및 선택
- **드래그 앤 드롭** - 파일을 드래그하거나 클릭하여 추가
- **즉시 파일 분석** - MP4/MOV/MKV/WebM/AVI를 Go 네이티브 파서로 즉시 분석 (ffprobe 불필요)
- **재생시간 불일치 감지** - 동영상 길이가 다를 때 경고 (동기화 재생 시 중요)
- **업스케일 경고** - 해상도나 프레임레이트가 불필요하게 업스케일될 때 알림
- **실시간 진행률** - 인코딩 진행률, 예상 시간, 속도 표시
- **프리셋 시스템** - 4K/1080p 다양한 프레임레이트 또는 원본 유지 모드

## 하드웨어 인코딩 지원

사용 가능한 하드웨어 인코더를 자동으로 감지하여 빠른 인코딩을 지원합니다.

| 플랫폼 | 인코더 | 설명 |
|--------|--------|------|
| macOS | Apple VideoToolbox | Mac 내장 하드웨어 가속 |
| Windows | NVIDIA NVENC | NVIDIA GPU |
| Windows | Intel QuickSync | Intel 내장 GPU |
| Windows | AMD AMF | AMD GPU |
| Linux | VAAPI | VA-API (Intel/AMD) |
| 모든 플랫폼 | x265 (소프트웨어) | CPU 인코딩 (폴백) |

## 프리셋

| 프리셋 | 해상도 | 프레임레이트 |
|--------|--------|------------|
| 원본 설정 유지 | 원본 | 원본 |
| HEVC 4K | 3840x2160 | 60/30/29.97/24/23.976 fps |
| HEVC 1080p | 1920x1080 | 60/30/29.97/24/23.976 fps |

## 요구사항

### Windows
- **Windows 10/11** (WebView2 런타임, 최신 Windows에 기본 설치됨)
- **FFmpeg** - 릴리스 패키지에 포함됨

### macOS
- **macOS 11 Big Sur** 이상
- **FFmpeg** - 릴리스 패키지에 포함됨

## 설치 방법

### Windows

1. [Releases](https://github.com/icq4ever/SyncLauperVideoConverter/releases)에서 `*-windows-amd64.zip` 다운로드
2. zip 파일 압축 해제 (zip 우클릭 → 속성 → "차단 해제" 체크 후 압축 해제)
3. `syncLauperVideoConverter.exe` 실행

```
SyncLauperVideoConverter/
├── syncLauperVideoConverter.exe
├── ffmpeg.exe
└── ffprobe.exe (선택사항)
```

> **참고**: ffprobe는 WMV/FLV/TS 등 일부 형식의 폴백 분석용입니다. MP4/MOV/MKV/WebM/AVI는 네이티브 파서를 사용하므로 ffprobe 없이도 작동합니다.

### macOS

1. [Releases](https://github.com/icq4ever/SyncLauperVideoConverter/releases)에서 다운로드:
   - Intel Mac: `*-macos-intel.zip`
   - Apple Silicon (M1/M2/M3): `*-macos-arm64.zip`
2. zip 파일 압축 해제
3. `SyncLauperVideoConverter.app` 실행
   - 최초 실행 시: 우클릭 → 열기 (또는 시스템 설정 → 보안에서 허용)

```
SyncLauperVideoConverter/
├── SyncLauperVideoConverter.app
├── ffmpeg
└── ffprobe (선택사항)
```

## 소스에서 빌드하기

### 사전 요구사항

- [Go](https://go.dev/dl/) 1.21+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2
- [Node.js](https://nodejs.org/) 18+
- Linux/WSL에서 Windows 크로스 컴파일 시: `x86_64-w64-mingw32-gcc`

### Wails CLI 설치

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 빌드

```bash
# 현재 플랫폼용 빌드
wails build

# macOS (Intel)
wails build -platform darwin/amd64

# macOS (Apple Silicon)
wails build -platform darwin/arm64

# Windows
wails build -platform windows/amd64

# Windows (WSL/Linux에서 크로스 컴파일)
export CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc
wails build -platform windows/amd64
```

### 개발 모드

```bash
wails dev
```

### 빌드 결과물

- **Windows**: `build/bin/syncLauperVideoConverter.exe`
- **macOS**: `build/bin/syncLauperVideoConverter.app`

> **참고**: 배포 시 `ffmpeg`는 필수이며, `ffprobe`는 선택사항입니다 (일부 비표준 형식 지원용).

## 기술 스택

- **백엔드**: Go + [Wails v2](https://wails.io/)
- **프론트엔드**: Svelte + TypeScript
- **인코딩**: FFmpeg (하드웨어 가속 + libx265 폴백)
- **파일 분석**: Go 네이티브 파서 (MP4/MOV/M4V, MKV/WebM, AVI) + ffprobe 폴백

## 프로젝트 구조

```
├── app.go                          # 메인 앱 로직
├── main.go                         # 진입점
├── wails.json                      # Wails 설정
├── internal/
│   ├── encoder/ffmpeg.go           # FFmpeg 래퍼
│   ├── fileinfo/
│   │   ├── fileinfo.go             # 파일 분석 디스패처
│   │   ├── mp4parser.go            # MP4/MOV/M4V 네이티브 파서
│   │   ├── mkvparser.go            # MKV/WebM 네이티브 파서
│   │   └── aviparser.go            # AVI 네이티브 파서
│   ├── preset/preset.go            # 인코딩 프리셋
│   └── cmdutil/                    # 플랫폼별 유틸리티
└── frontend/src/
    ├── App.svelte                  # 메인 UI
    └── lib/
        ├── components/             # Svelte 컴포넌트
        ├── stores/                 # 상태 관리
        └── types/                  # TypeScript 타입
```

## 라이선스

이 프로젝트는 [MIT 라이선스](LICENSE)로 배포됩니다.

## 링크

- [SyncLauper](https://synclauper.studio42.kr) - 멀티스크린 동기화 영상 플레이어
- [studio42](https://studio42.kr)
