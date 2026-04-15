#!/usr/bin/env bash
set -e

INSTALL_DIR="/opt/syncLauperVideoConverter"
DESKTOP_FILE="/usr/share/applications/syncLauperVideoConverter.desktop"
SYMLINK="/usr/local/bin/syncLauperVideoConverter"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "$EUID" -ne 0 ]; then
  echo "root 권한이 필요합니다. sudo 로 다시 실행합니다..."
  exec sudo -E bash "$0" "$@"
fi

install_deps() {
  echo "==> 런타임 의존성 확인 (libwebkit2gtk-4.1, libgtk-3)"
  if ldconfig -p | grep -q 'libwebkit2gtk-4\.1\.so\.0'; then
    echo "    이미 설치되어 있습니다"
    return
  fi

  if command -v apt-get >/dev/null 2>&1; then
    echo "    apt 로 설치 중..."
    apt-get update
    apt-get install -y libwebkit2gtk-4.1-0 libgtk-3-0
  elif command -v dnf >/dev/null 2>&1; then
    echo "    dnf 로 설치 중..."
    dnf install -y webkit2gtk4.1 gtk3
  elif command -v pacman >/dev/null 2>&1; then
    echo "    pacman 으로 설치 중..."
    pacman -S --needed --noconfirm webkit2gtk-4.1 gtk3
  elif command -v zypper >/dev/null 2>&1; then
    echo "    zypper 로 설치 중..."
    zypper install -y libwebkit2gtk-4_1-0 libgtk-3-0
  else
    echo "    경고: 패키지 매니저를 찾을 수 없습니다. libwebkit2gtk-4.1 과 libgtk-3 를 수동 설치하세요."
  fi
}

install_deps

echo "==> $INSTALL_DIR 에 설치합니다"
mkdir -p "$INSTALL_DIR"
cp "$SCRIPT_DIR/syncLauperVideoConverter" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/ffmpeg" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/ffprobe" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/icon.png" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/LICENSE" "$INSTALL_DIR/" 2>/dev/null || true
cp "$SCRIPT_DIR/THIRD_PARTY_LICENSES.txt" "$INSTALL_DIR/" 2>/dev/null || true
cp "$SCRIPT_DIR/README.md" "$INSTALL_DIR/" 2>/dev/null || true
cp "$SCRIPT_DIR/uninstall.sh" "$INSTALL_DIR/"

chmod +x "$INSTALL_DIR/syncLauperVideoConverter"
chmod +x "$INSTALL_DIR/ffmpeg"
chmod +x "$INSTALL_DIR/ffprobe"
chmod +x "$INSTALL_DIR/uninstall.sh"

echo "==> Desktop entry 등록"
install -m 644 "$SCRIPT_DIR/syncLauperVideoConverter.desktop" "$DESKTOP_FILE"

echo "==> 심볼릭 링크 생성: $SYMLINK"
ln -sf "$INSTALL_DIR/syncLauperVideoConverter" "$SYMLINK"

if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database /usr/share/applications >/dev/null 2>&1 || true
fi

echo ""
echo "설치 완료!"
echo "  - 실행: 애플리케이션 메뉴에서 'SyncLauper VideoConverter' 또는 터미널에서 'syncLauperVideoConverter'"
echo "  - 제거: sudo $INSTALL_DIR/uninstall.sh"
