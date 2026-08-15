#!/bin/bash
# 从 GitHub Release 拉取最新 wol-web-linux-arm64 并重启服务（板端更新）
#
# 用法：
#   REPORT=632-8nm/remote-wakeup ./update.sh
# 全局: 可用 GITHUB_TOKEN 或复用环境已有的 ORANGEPI_MONITOR_GITHUB_TOKEN
set -euo pipefail

REPO="${REPORT:-632-8nm/remote-wakeup}"
TOKEN="${GITHUB_TOKEN:-${ORANGEPI_MONITOR_GITHUB_TOKEN:-}}"

INSTALL_DIR="${INSTALL_DIR:-$HOME/project/remote-wakeup}"
BIN="$INSTALL_DIR/wol-web"
SERVICE="wol-web"

echo "==> 从 GitHub Release 拉取最新版 $REPO"
echo "    目标目录: $INSTALL_DIR"

# 1) 获取最新 release 的资产下载地址
AUTH=()
if [ -n "$TOKEN" ]; then
  AUTH=(-H "Authorization: Bearer $TOKEN")
fi

ASSET_URL="$(curl -fsSL "${AUTH[@]}" \
  "https://api.github.com/repos/$REPO/releases/latest" \
  | grep -oE '"browser_download_url": "[^"]+wol-web-linux-arm64[^"]*"' \
  | head -1 | sed 's/.*: "//; s/"$//')"

if [ -z "$ASSET_URL" ]; then
  echo "错误: 未在最新 release 找到 wol-web-linux-arm64 资产" >&2
  exit 1
fi
echo "    下载自: $ASSET_URL"

# 2) 备份当前二进制
if [ -x "$BIN" ]; then
  cp "$BIN" "$BIN.bak"
  echo "    已备份旧二进制 => $BIN.bak"
fi

# 3) 下载到临时文件并校验是有效 ELF
TMPBIN="$(mktemp)"
if ! curl -fsSL "${AUTH[@]}" -o "$TMPBIN" "$ASSET_URL"; then
  echo "下载失败" >&2
  rm -f "$TMPBIN"
  exit 1
fi
if ! file "$TMPBIN" | grep -q "ELF"; then
  echo "错误: 下载内容不是有效 ELF" >&2
  rm -f "$TMPBIN"
  exit 1
fi

# 4) 停止服务 → 替换 → 重启
sudo systemctl stop "$SERVICE" 2>/dev/null || true
install -m 0755 "$TMPBIN" "$BIN"
rm -f "$TMPBIN"
sudo systemctl daemon-reload
sudo systemctl start "$SERVICE"
sleep 1

if systemctl --quiet is-active "$SERVICE"; then
  echo "✓ 已更新并重启 $SERVICE"
  echo "  新版本 SHA256: $(sha256sum "$BIN" | awk '{print $1}')"
else
  echo "✗ 服务启动失败，回滚到备份"
  [ -f "$BIN.bak" ] && install -m 0755 "$BIN.bak" "$BIN"
  sudo systemctl daemon-reload
  sudo systemctl start "$SERVICE"
  exit 1
fi
