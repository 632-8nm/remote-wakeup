# 部署架构与搭建流程（remote-wakeup）

> 本文档记录本项目从零搭建的完整流程与当前生产架构，供维护者参考。
> 敏感信息一律使用占位符，不包含真实 token / 密码。

## 1. 整体架构

```
开发者本地 (Windows)
   │  写代码 → go vet / go test（本地自测）
   │  git commit + push
   ▼
GitHub (remote-wakeup 仓库)
   │  触发 GitHub Actions（workflow: deploy.yml）
   ▼
CI/CD（GitHub 云端 runner，ubuntu-latest）
   │  ① go vet ./... + go test ./...（质量门禁）
   │  ② 交叉编译 arm64 二进制（CGO_ENABLED=0 GOOS=linux GOARCH=arm64）
   │  ③ 通过 Cloudflare Tunnel (ssh.<your-domain>) SSH 到板子
   ▼
板子（Orange Pi Zero 3）
   │  停服务 → scp 二进制 → 重启
   ▼
/opt/wol-web/wol-web  （systemd: wol-web，监听 :5000）
```

## 2. 生产部署位置

| 项 | 值 |
|---|---|
| 部署目录 | `/opt/wol-web/` |
| 二进制 | `/opt/wol-web/wol-web` |
| 配置 | `/opt/wol-web/.env`（权限 600） |
| systemd 服务 | `wol-web.service` |
| 监听端口 | `5000` |
| 公网入口 | `remote-wakeup.<your-domain>`（Cloudflare 隧道 → http://localhost:5000） |
| CI 部署入口 | `ssh.<your-domain>`（Cloudflare 隧道 → ssh://localhost:22） |

## 3. 板子侧组件

| 组件 | 说明 | 状态 |
|---|---|---|
| `cloudflared` | Cloudflare 隧道客户端（token 模式） | systemd 服务，常驻 |
| `/opt/wol-web/` | 生产部署目录（root 属主，二进制 750 / .env 600） | 由 CI 更新 |
| `wol-web.service` | systemd 单元，指向 `/opt/wol-web/wol-web` | 常驻 |

## 4. Cloudflare 侧配置

| 项 | 配置 | 用途 |
|---|---|---|
| 域名 | `<your-domain>` | 托管在 Cloudflare |
| 隧道 | token 模式（tunnel run --token） | 板子主动连 Cloudflare |
| Public Hostname | `remote-wakeup.<your-domain>` → `http://localhost:5000` | 公网访问 Web |
| Public Hostname | `ssh.<your-domain>` → `ssh://localhost:22` | CI/CD 部署 SSH 入口 |
| Service Token | 在 Access → Service Auth 创建 | CI 认证（格式 ID:SECRET） |
| Access 策略 | `ci-deploy`（Service Auth + token） | 放行 CI 的 cloudflared 连接 |

## 5. GitHub 侧配置

| 项 | 值 | 说明 |
|---|---|---|
| 仓库 | `<your-org>/remote-wakeup` | — |
| Workflow | `.github/workflows/deploy.yml` | push main 触发 |
| Workflow | `.github/workflows/release.yml` | 打 v* tag 触发，发布 Release |
| Secret: `BOARD_SSH_KEY` | 板子 `~/.ssh/deploy` 私钥 | 云端 SSH 登录板子 |
| Secret: `CLOUDFLARED_TOKEN` | `ClientID:ClientSecret` | cloudflared access 认证 |

## 6. 从零搭建步骤

### 6.1 板子准备
```bash
# 安装 cloudflared
curl -L --output /usr/local/bin/cloudflared \
  https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64
chmod +x /usr/local/bin/cloudflared

# 配置免密 sudo（仅 systemctl/journalctl/tee，供 CI 使用）
sudo tee /etc/sudoers.d/orangepi-systemd <<'EOF'
orangepi ALL=(ALL) NOPASSWD: /usr/bin/systemctl, /bin/systemctl, /usr/bin/journalctl, /usr/bin/tee
EOF
sudo chmod 440 /etc/sudoers.d/orangepi-systemd

# 生成 CI 部署密钥
ssh-keygen -t ed25519 -N "" -f ~/.ssh/deploy
cat ~/.ssh/deploy.pub >> ~/.ssh/authorized_keys

# 创建生产部署目录
sudo mkdir -p /opt/wol-web
sudo chown orangepi:orangepi /opt/wol-web

# 配置 systemd 服务
sudo tee /etc/systemd/system/wol-web.service <<'EOF'
[Unit]
Description=WOL Web - Remote Wake-on-LAN (Go)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=orangepi
WorkingDirectory=/opt/wol-web
ExecStart=/opt/wol-web/wol-web
EnvironmentFile=/opt/wol-web/.env
Restart=always
RestartSec=5
UMask=0022

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl daemon-reload
sudo systemctl enable --now wol-web
```

### 6.2 Cloudflare 配置
1. 域名托管在 Cloudflare（`<your-domain>`）
2. 板子安装 cloudflared，用 token 接入隧道
3. 加 Public Hostname：
   - `remote-wakeup.<your-domain>` → `http://localhost:5000`
   - `ssh.<your-domain>` → `ssh://localhost:22`
4. Access → Service Auth 创建 Service Token（记下 Client ID/Secret）
5. Access → 为 `ssh.<your-domain>` 配置 `ci-deploy` 策略（Service Auth）

### 6.3 GitHub 配置
1. 仓库加两个 Secret：
   - `BOARD_SSH_KEY` = 板子 `~/.ssh/deploy` 私钥全文
   - `CLOUDFLARED_TOKEN` = `ClientID:ClientSecret`
2. 推送 `.github/workflows/deploy.yml`（push main 自动部署）
3. 推送 `.github/workflows/release.yml`（打 tag 自动构建 Release）

### 6.4 验证
```bash
# 在板子测试隧道 SSH（模拟 CI）
cloudflared access tcp --hostname ssh.<your-domain> --url localhost:2222 \
  --service-token-id <ID> --service-token-secret <SECRET> &
ssh -i ~/.ssh/deploy -p 2222 orangepi@127.0.0.1 'echo OK'
```
推一次代码到 main，观察 GitHub Actions 的 Deploy 是否 success，板子服务是否更新。

## 7. 日常维护

```bash
# 查看服务
systemctl status wol-web
# 查看日志
journalctl -u wol-web -f
# 手动重启
sudo systemctl restart wol-web
# 改配置（.env）
sudo nano /opt/wol-web/.env && sudo systemctl restart wol-web
```

## 8. 回滚

CI 部署的是云端编译的固定版本二进制。回滚方式：
- 用 `git revert` 回退代码后 push（触发重新部署旧版）
- 或手动替换 `/opt/wol-web/wol-web` 为上一版二进制并重启
