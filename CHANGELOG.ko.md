# 변경 이력

이 프로젝트의 주요 변경 사항을 기록합니다.

## [1.1.3] - 2026-02-28

### 수정

- **Intel QuickSync (QSV) 인코딩 실패 수정** - Intel Iris Xe 및 최신 GPU에서 "Could not open encoder before EOF" 에러 해결
  - ICQ 모드(`-global_quality`)에서 CQP 모드(`-rc:v CQP -qp`)로 변경하여 호환성 향상
  - VDENC 전용 GPU (Iris Xe, 11세대 이상)를 위한 `-low_power 1` 플래그 추가
  - `main10` 미지원 시 `main` 프로파일 자동 폴백
- **인코딩 에러가 무음으로 사라지는 문제** - 인코딩 실패 시 진행바가 즉시 숨겨져 에러를 확인할 수 없던 문제 수정
  - 인코딩 완료 후 결과 요약 UI 추가 (성공/일부 실패/실패)
  - 실패한 파일을 확인하고 직접 닫을 수 있도록 개선

### 추가

- **인코더 런타임 검증** - 하드웨어 인코더를 목록에 표시하기 전 1프레임 테스트 인코딩으로 실제 작동 여부 확인
- **Linux QuickSync/VAAPI 지원** - Linux에서 `hevc_qsv` 및 `hevc_vaapi` 인코더 감지 추가
- **프리릴리스 워크플로우** - 버전 태그에 `-`가 포함된 경우(예: `v1.1.3-rc.1`) GitHub에서 자동으로 프리릴리스로 표시

## [1.1.2] - 2026-02-25

### 추가

- **앞 공백(블랙 인트로) 옵션** - 인코딩 시 영상 앞에 1~3초 블랙(무음) 구간 삽입 기능
  - 인코딩 시작 버튼 좌측에 체크박스 + 초 선택 드롭다운 배치
  - 기본값: 1초, 활성화
- **프리셋/인코더 설명 정렬** - 설명 행이 위쪽 셀렉터 열과 좌우 정렬되도록 개선

## [1.1.1] - 2025-02-11

### 수정

- **NVIDIA NVENC 호환성** - 구형 NVIDIA GPU(예: Quadro P1000) 호환을 위해 B-frames 옵션 제거
- **Intel QuickSync 호환성** - 구형 Intel GPU 호환을 위해 `main10` 대신 `main` 프로파일 강제 사용
- **AMD AMF 호환성** - 구형 AMD GPU 호환을 위해 `main10` 대신 `main` 프로파일 강제 사용

## [1.1.0] - 2025-02-11

### 추가

- **하드웨어 인코딩 지원** - GPU 인코더 자동 감지 및 선택
  - macOS: Apple VideoToolbox (`hevc_videotoolbox`)
  - Windows: NVIDIA NVENC (`hevc_nvenc`), Intel QuickSync (`hevc_qsv`), AMD AMF (`hevc_amf`)
  - Linux: VAAPI (`hevc_vaapi`)
  - 하드웨어 인코더 미지원 시 소프트웨어 인코딩(`libx265`) 자동 폴백
- **인코더 선택 UI** - 사용 가능한 인코더 중 선택 가능한 드롭다운 추가
- **GitHub Actions 워크플로우** - Windows 및 macOS (Intel + Apple Silicon) 자동 빌드
- **FFmpeg 번들링** - 릴리스 패키지에 FFmpeg essentials 포함
- **인코딩 에러 상세 표시** - 인코딩 실패 시 FFmpeg 에러 메시지를 UI에 표시

### 수정

- macOS `.app` 번들에서 같은 폴더의 ffmpeg/ffprobe를 찾지 못하는 문제
- VideoToolbox (macOS)에서 `-profile:v` 파라미터로 인해 0KB 파일이 생성되는 문제
- macOS에서 FFmpeg 에러 메시지가 `ffmpeg.exe`로 표시되는 문제

### 변경

- README에 macOS 설치 방법 추가
- README에 하드웨어 인코딩 관련 문서 추가
- 모든 플랫폼 빌드 방법 업데이트
- ffprobe는 선택사항임을 명시 (WMV/FLV/TS 폴백용으로만 필요)
- 인코더 드롭다운에서 이모지 아이콘 제거
- macOS 네이티브 그라데이션 대신 커스텀 드롭다운 스타일 적용

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
