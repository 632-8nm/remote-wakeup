#!/bin/bash
# 修改 WOL Web 管理员密码。
# 密码保存在项目目录的 .env 文件（ADMIN_PASSWORD 字段）。
# 用法： ./change-password.sh
# 若以 systemd 部署，请补充 systemctl restart wol-web 的调用。
set -e

cd "$(dirname "$0")"

if [ ! -f .env ]; then
  echo "未找到 .env，请先 cp .env.example .env 并配置。" >&2
  exit 1
fi

read -s -p "新密码: " NEW_PW
echo
if [ -z "$NEW_PW" ]; then
  echo "密码不能为空" >&2
  exit 1
fi

if grep -q '^ADMIN_PASSWORD=' .env; then
  sed -i "s|^ADMIN_PASSWORD=.*|ADMIN_PASSWORD=$NEW_PW|" .env
else
  echo "ADMIN_PASSWORD=$NEW_PW" >> .env
fi
echo "已更新 .env 中的 ADMIN_PASSWORD。"

# 如由 systemd 管理，可取消下面两行注释以自动重启：
# SYSTEMD_SERVICE="${WOL_SERVICE:-wol-web}"
# systemctl restart "$SYSTEMD_SERVICE" && echo "服务 $SYSTEMD_SERVICE 已重启。"
