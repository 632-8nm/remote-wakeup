# WOL Web — 远程开机网页 (Wake-on-LAN)

通过网页按钮远程唤醒局域网内(或家中)任意一台支持 Wake-on-LAN 的电脑。
纯 Go 标准库实现,**零第三方运行时依赖**,编译为单个静态二进制,单文件部署。

## 链路

```
手机浏览器
   │
   ▼
Cloudflare Tunnel / 局域网 IP
   │
   ▼
本服务（WOL Web）  ── WOL Magic Packet ──►  目标电脑（Windows / 任意 WOL 设备）
```

- 登录(密码)后即可一键发送 WOL 魔术包唤醒目标电脑。
- 首页自动轮询目标电脑的 ping 状态,显示**在线 / 离线**。
- 可配合 [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) 将页面暴露到公网域名。

## 特性

- ✅ 登录保护(恒定时间密码比较,`crypto/subtle`)
- ✅ 会话签名 Cookie(无数据库、服务端零存储)
- ✅ 一键 WOL 唤醒 + 实时在线状态(ping)
- ✅ 单静态二进制,一个文件即可运行
- ✅ 零第三方依赖,纯 Go 标准库
- ✅ 前端页面经 `go:embed` 打包进二进制,无额外模板文件
- ✅ 配置通过 `.env` 或环境变量,**仓库不含任何真实凭据**

## 快速开始

### 方式一:直接运行二进制

```bash
# 下载 release 中的 wol-web(或自己编译):
git clone https://github.com/632-8nm/remote-wakeup.git
cd remote-wakeup
go build -o wol-web .          # 当前平台编译

# 交叉编译到 aarch64 板子:
# GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o wol-web .
```

配置(环境变量或 `.env`,二选一):

```bash
export WOL_MAC=AA:BB:CC:DD:EE:FF      # 目标电脑 MAC(必填)
export TARGET_IP=192.168.1.100        # 目标电脑 IP(必填,用于 ping 检测)
export ADMIN_PASSWORD=你的密码         # 登录密码(必填)
./wol-web                            # 监听 5000
```

### 方式二:一键部署为 systemd(扁平发布包) 

Release 提供扁平 zip,三个文件同级:`wol-web`、`.env.example`、`install.sh`。

```bash
unzip wol-web-vX.Y.Z-linux-arm64.zip
cd <解压目录>
cp .env.example .env        # 填写 WOL_MAC / TARGET_IP / ADMIN_PASSWORD
./install.sh                 # 自动注册 systemd 服务并启动
```

## 配置(环境变量)

| 变量 | 必填 | 默认 | 说明 |
|---|---|---|---|
| `WOL_MAC` | ✅ | — | 目标 MAC |
| `TARGET_IP` | ✅ | — | 目标 IP(ping 检测) |
| `ADMIN_PASSWORD` | ✅ | — | 登录密码 |
| `SECRET_KEY` | — | 随机 | 会话签名密钥 |
| `WOL_BROADCAST` | — | `255.255.255.255` | WOL 广播地址 |
| `WOL_PORT` | — | `9` | WOL 目标端口 |
| `SESSION_HOURS` | — | `1` | 登录有效期(小时,可小数) |
| `PORT` | — | `5000` | Web 监听端口 |

## 更新(已有部署的板子)

```bash
./update.sh     # 从 GitHub Release 拉最新二进制并重启服务
```

## systemd 服务

发布包与仓库根目录均有 `wol-web.service`(通用模板)。手动部署:

```bash
sudo cp wol-web.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now wol-web
```

> 凭据通过 `EnvironmentFile=.env` 注入,不写死在 unit 文件中。

### 修改密码

```bash
./change-password.sh    # 交互式更新 .env 中的密码
```

## Cloudflare Tunnel(可选公网接入)

在 Cloudflare Zero Trust Dashboard 建 Public Hostname,指向 `http://localhost:5000`;
或 `~/.cloudflared/config.yml`:

```yaml
ingress:
  - hostname: wol.example.com
    service: http://localhost:5000
  - service: http_status:404
```

可叠加 Cloudflare Access 策略做二次保护。

## 安全说明

- 认证用恒定时间比较,无时序侧信道。
- 会话为 HMAC 签名 Cookie,服务端零存储。
- 历史与代码中不含任何真实 MAC/IP/密码;请勿将 `.env` 提交进仓库。

## 目录结构

```
.
├── main.go / config.go / handlers.go / session.go / wol.go / templates.go / util.go
├── templates/            # 前端页面(go:embed)
├── deploy 相关脚本已扁平到根:install.sh / update.sh / wol-web.service
├── .env.example          # 配置模板
├── change-password.sh
└── wol-web.service       # systemd 模板
```

## License

[MIT](LICENSE)
