#!/bin/bash
# WOL Web 一键部署脚本（面向所有使用者，不含任何真实凭据）
#
# 用法：三个文件同级（wol-web、.env.example、install.sh）放在同一目录后：
#   ./install.sh
#
# 脚本会：
#   1) 找到同目录的 wol-web 二进制
#   2) 若没有 .env，则从 .env.example 生成并引导填写
#   3) 安装为 systemd 服务（需 sudo）
#   4) 启动并设置开机自启
set -euo pipefail

# 脚本所在目录（扁平结构：二进制/配置/脚本同级）
SRC_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN="$SRC_DIR/wol-web"
ENV_FILE="$SRC_DIR/.env"
SVC_NAME="wol-web"
RUN_USER="${SUDO_USER:-$USER}"

echo "==> WOL Web 安装脚本"
echo "    来源目录: $SRC_DIR"

# 1) 二进制存在性检查
if [ ! -x "$BIN" ]; then
  echo "错误: 未找到可执行文件 $BIN" >&2
  echo "请把 release 里的 wol-web 与 install.sh 放同一目录" >&2
  exit 1
fi

# 2) .env 配置（缺失则从模板生成）
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

# 3) 生成 systemd 单元（用 .env 绝对路径，替换为实际用户）
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

# 4) 端口显示
PORT="$(grep -E '^PORT=' "$ENV_FILE" | cut -d= -f2- | head -1)"
[ -z "$PORT" ] && PORT=5000

sudo systemctl daemon-reload
sudo systemctl enable --now "$SVC_NAME"
sleep 1
if systemctl --quiet is-active "$SVC_NAME"; then
  echo "✓ 服务已启动（http://<本机IP>:$PORT）"
else
  echo "✗ 服务启动失败，日志: "
  sudo journalctl -u "$SVC_NAME" -n 20 --no-pager
  exit 1
fi

echo "==> 完成"
