#!/usr/bin/env bash
set -euo pipefail
if [[ $# -ne 2 || "$2" != --stopped ]]; then
  echo '先停止原服务，然后执行：bash install-prebuilt.sh /原安装目录 --stopped'; exit 1
fi
[[ $(uname -s) == Linux && $(uname -m) == x86_64 ]] || { echo '本包仅适用于 Linux x86_64。'; exit 1; }
PACKAGE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
TARGET_DIR="$(cd "$1" && pwd -P)"
[[ "$TARGET_DIR" != / && -f "$TARGET_DIR/config.json" && -d "$TARGET_DIR/templates" ]] || { echo '目标不是有效的原安装目录。'; exit 1; }
case "$PACKAGE_DIR/" in "$TARGET_DIR/"*) echo '请将更新包解压到原安装目录外。'; exit 1;; esac
(cd "$PACKAGE_DIR"; sha256sum -c SHA256SUMS.txt >/dev/null)
for exe in /proc/[0-9]*/exe; do
  running="$(readlink "$exe" 2>/dev/null || true)"
  if [[ "$running" == "$TARGET_DIR/gophish" || "$running" == "$TARGET_DIR/gophish (deleted)" ]]; then
    echo '原程序仍在运行，请先停止服务。'; exit 1
  fi
done
while IFS= read -r file; do
  [[ -f "$PACKAGE_DIR/$file" ]] || { echo "缺少文件：$file"; exit 1; }
done < "$PACKAGE_DIR/ui-files.txt"
while IFS= read -r file; do
  if [[ -f "$TARGET_DIR/$file" ]] && ! cmp -s "$PACKAGE_DIR/$file" "$TARGET_DIR/$file"; then
    echo "已有迁移不同，请先人工核对：$file"; exit 1
  fi
done < "$PACKAGE_DIR/migrations.txt"
BACKUP_DIR="${TARGET_DIR}.backup-prebuilt-$(date +%Y%m%d-%H%M%S)-$$"
cp -a "$TARGET_DIR" "$BACKUP_DIR"
echo "完整目录备份：$BACKUP_DIR"
while IFS= read -r file; do
  mkdir -p "$TARGET_DIR/$(dirname "$file")"
  cp "$PACKAGE_DIR/$file" "$TARGET_DIR/$file"
done < "$PACKAGE_DIR/ui-files.txt"
while IFS= read -r file; do
  if [[ ! -f "$TARGET_DIR/$file" ]]; then
    mkdir -p "$TARGET_DIR/$(dirname "$file")"
    cp "$PACKAGE_DIR/$file" "$TARGET_DIR/$file"
  fi
done < "$PACKAGE_DIR/migrations.txt"
cp "$PACKAGE_DIR/gophish" "$TARGET_DIR/gophish.update-new"
chmod 755 "$TARGET_DIR/gophish.update-new"
mv -f "$TARGET_DIR/gophish.update-new" "$TARGET_DIR/gophish"
echo '升级完成。请进入原目录，用原配置和原方式启动一个实例。数据库结构迁移由程序启动时执行。'
