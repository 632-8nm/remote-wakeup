#!/bin/bash
# WOL Web 部署脚本（统一脚本）
#
# 自适应两种用法：
#   A) 源码构建：目录里有 go.mod 和 go 工具链 → 先本地构建再部署
#   B) 预编译二进制：目录里只有 wol-web（release 解压）→ 跳过构建，直接部署
#
# 每次运行都会：
#   1) （若源码场景）本地 go build
#   2) 若缺少 .env，则从 .env.example 生成并引导填写
#   3) 注册/更新 systemd 服务（需 sudo）
#   4) 重启服务并验证
set -euo pipefail

SRC_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SRC_DIR"

BIN="$SRC_DIR/wol-web"
ENV_FILE="$SRC_DIR/.env"
SVC_NAME="wol-web"
RUN_USER="${SUDO_USER:-$USER}"

echo "==> WOL Web dev.sh（部署脚本）"
echo "    目录: $SRC_DIR"

# 1) .env 配置（缺失则从模板生成）
if [ ! -f "$ENV_FILE" ]; then
  if [ -f "$SRC_DIR/.env.example" ]; then
    cp "$SRC_DIR/.env.example" "$ENV_FILE"
    chmod 600 "$ENV_FILE"
    echo "已从 .env.example 生成 $ENV_FILE，请编辑后重跑本脚本。"
    echo "  必填: WOL_MAC / TARGET_IP / ADMIN_PASSWORD"
    exit 1
  else
    echo "错误: 缺少 .env.example，无法生成配置" >&2
    exit 1
  fi
fi

# 校验必填项是否仍是占位符
for k in WOL_MAC TARGET_IP ADMIN_PASSWORD; do
  val="$(grep -E "^$k=" "$ENV_FILE" | head -1 | cut -d= -f2-)"
  if [ -z "$val" ] || [[ "$val" == change-me* ]] || [[ "$val" == AA:BB:* ]]; then
    echo "请在 $ENV_FILE 中设置 $k（当前仍是占位符/空值）" >&2
    exit 1
  fi
done

# 2) 自适应构建：源码场景（有 go.mod + go）则构建；否则要求已有二进制
if [ -f "$SRC_DIR/go.mod" ] && command -v go >/dev/null 2>&1; then
  echo "==> 检测到源码（go.mod），本地构建 ..."
  CGO_ENABLED=0 go build -trimpath -o "$BIN" ./cmd/wol-web
  echo "    ✓ 构建完成: $BIN"
else
  echo "==> 未检测到源码构建环境，使用已有二进制"
  if [ ! -x "$BIN" ]; then
    echo "错误: 未找到可执行文件 $BIN（release 解压目录里应有 wol-web）" >&2
    exit 1
  fi
  echo "    使用: $BIN"
fi

# 3) 生成/更新 systemd 单元
SVC_FILE="/etc/systemd/system/$SVC_NAME.service"
UNIT="[Unit]
Description=WOL Web - Remote Wake-on-LAN (Go)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$RUN_USER
WorkingDirectory=$SRC_DIR
ExecStart=$BIN
EnvironmentFile=$ENV_FILE
Restart=always
RestartSec=5
UMask=0022

[Install]
WantedBy=multi-user.target
"
echo "$UNIT" | sudo tee "$SVC_FILE" >/dev/null
echo "已写入 $SVC_FILE"
sudo systemctl daemon-reload

# 4) 重启（已存在则重启，否则 enable --now）
if systemctl --quiet list-unit-files "$SVC_NAME.service" 2>/dev/null; then
  sudo systemctl restart "$SVC_NAME"
else
  sudo systemctl enable --now "$SVC_NAME"
fi
sleep 1

# 5) 验证
if systemctl --quiet is-active "$SVC_NAME"; then
  PORT="$(grep -E '^PORT=' "$ENV_FILE" | cut -d= -f2- | head -1)"
  [ -z "$PORT" ] && PORT=5000
  echo "✓ 服务已启动（http://<本机IP>:$PORT）"
else
  echo "✗ 服务启动失败，日志: "
  sudo journalctl -u "$SVC_NAME" -n 20 --no-pager
  exit 1
fi

echo "==> 完成"
