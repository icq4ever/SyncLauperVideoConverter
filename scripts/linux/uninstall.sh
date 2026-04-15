#!/usr/bin/env bash
set -e

INSTALL_DIR="/opt/syncLauperVideoConverter"
DESKTOP_FILE="/usr/share/applications/syncLauperVideoConverter.desktop"
SYMLINK="/usr/local/bin/syncLauperVideoConverter"

if [ "$EUID" -ne 0 ]; then
  echo "root 권한이 필요합니다. sudo 로 다시 실행합니다..."
  exec sudo -E bash "$0" "$@"
fi

echo "==> 제거 중..."
rm -f "$DESKTOP_FILE"
rm -f "$SYMLINK"
rm -rf "$INSTALL_DIR"

if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database /usr/share/applications >/dev/null 2>&1 || true
fi

echo "제거 완료."
