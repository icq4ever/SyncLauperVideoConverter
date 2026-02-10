# 변경 이력

이 프로젝트의 주요 변경 사항을 기록합니다.

## [1.1.0] - 2025-02-10

### 추가

- **하드웨어 인코딩 지원** - GPU 인코더 자동 감지 및 선택
  - macOS: Apple VideoToolbox (`hevc_videotoolbox`)
  - Windows: NVIDIA NVENC (`hevc_nvenc`), Intel QuickSync (`hevc_qsv`), AMD AMF (`hevc_amf`)
  - Linux: VAAPI (`hevc_vaapi`)
  - 하드웨어 인코더 미지원 시 소프트웨어 인코딩(`libx265`) 자동 폴백
- **인코더 선택 UI** - 사용 가능한 인코더 중 선택 가능한 드롭다운 추가
- **GitHub Actions 워크플로우** - Windows 및 macOS (Intel + Apple Silicon) 자동 빌드
- **FFmpeg 번들링** - 릴리스 패키지에 FFmpeg essentials 포함

### 변경

- README에 macOS 설치 방법 추가
- README에 하드웨어 인코딩 관련 문서 추가
- 모든 플랫폼 빌드 방법 업데이트

## [1.0.0] - 2025-02-09

### 추가

- 최초 릴리스
- FFmpeg 기반 HEVC (H.265) 일괄 인코딩
- 드래그 앤 드롭 파일 추가 지원
- MP4/MOV/MKV/WebM/AVI 네이티브 파서
- 동기화 재생을 위한 재생시간 불일치 감지
- 업스케일 경고 알림
- 실시간 인코딩 진행률 및 예상 시간 표시
- 프리셋 시스템 (4K/1080p, 다양한 프레임레이트)
- 원본 설정 유지 모드
