# SyncLauper VideoConverter

[English](README.md)

[SyncLauper](https://synclauper.studio42.kr) 멀티스크린 동기화 영상 플레이어를 위한 배치 비디오 인코더입니다. 영상 파일을 HEVC (H.265) 코덱으로 손쉽게 변환할 수 있습니다.

## 주요 기능

- **일괄 인코딩** - 여러 영상 파일을 한 번에 변환
- **HEVC (H.265)** - 고효율 인코딩, 다양한 프리셋 제공
- **드래그 앤 드롭** - 파일을 드래그하거나 클릭하여 추가
- **즉시 파일 분석** - MP4/MOV/MKV/WebM/AVI를 Go 네이티브 파서로 즉시 분석 (ffprobe 불필요)
- **재생시간 불일치 감지** - 동영상 길이가 다를 때 경고 (동기화 재생 시 중요)
- **업스케일 경고** - 해상도나 프레임레이트가 불필요하게 업스케일될 때 알림
- **실시간 진행률** - 인코딩 진행률, 예상 시간, 속도 표시
- **프리셋 시스템** - 4K/1080p 다양한 프레임레이트 또는 원본 유지 모드

## 프리셋

| 프리셋 | 해상도 | 프레임레이트 |
|--------|--------|------------|
| 원본 설정 유지 | 원본 | 원본 |
| HEVC 4K | 3840x2160 | 60/30/29.97/24/23.976 fps |
| HEVC 1080p | 1920x1080 | 60/30/29.97/24/23.976 fps |

## 요구사항

- **Windows 10/11** (WebView2 런타임, 최신 Windows에 기본 설치됨)
- **FFmpeg** - `ffmpeg.exe`를 프로그램과 같은 폴더에 배치
  - [FFmpeg 다운로드 (essentials 빌드 권장)](https://www.gyan.dev/ffmpeg/builds/)

## 설치 방법

1. 최신 릴리스를 다운로드
2. zip 파일 압축 해제 (zip 우클릭 → 속성 → "차단 해제" 체크 후 압축 해제)
3. `ffmpeg.exe`를 `syncLauperVideoConverter.exe`와 같은 폴더에 배치
4. `syncLauperVideoConverter.exe` 실행

```
SyncLauperVideoConverter/
├── syncLauperVideoConverter.exe
└── ffmpeg.exe
```

## 소스에서 빌드하기

### 사전 요구사항

- [Go](https://go.dev/dl/) 1.21+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2
- [Node.js](https://nodejs.org/) 18+
- Linux/WSL에서 Windows 크로스 컴파일 시: `x86_64-w64-mingw32-gcc`

### 빌드

```bash
# Windows (네이티브)
wails build

# Windows (WSL/Linux에서 크로스 컴파일)
export CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc
wails build -platform windows/amd64
```

## 기술 스택

- **백엔드**: Go + [Wails v2](https://wails.io/)
- **프론트엔드**: Svelte + TypeScript
- **인코딩**: FFmpeg (libx265)
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
