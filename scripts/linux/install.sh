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
