# WOL Web — 远程开机网页 (Wake-on-LAN)

通过网页按钮远程唤醒局域网内(或家中)任意一台支持 Wake-on-LAN 的电脑。
纯 Go 标准库实现,**零第三方运行时依赖**,编译为单个静态二进制,单文件部署。

> 📦 **部署架构与从零搭建流程见 [DEPLOYMENT.md](DEPLOYMENT.md)**（CI/CD、Cloudflare 隧道、/opt 部署、密钥配置）

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

- 登录(密码)后一键发送 WOL 魔术包唤醒目标电脑;首页自动轮询目标 ping 状态,显示**在线 / 离线**。
- 主要使用方式是通过 [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) 暴露到公网域名,手机浏览器即可访问。

## 特性

- ✅ 登录保护(恒定时间密码比较,`crypto/subtle`)
- ✅ 会话签名 Cookie(无数据库、服务端零存储)
- ✅ 一键 WOL 唤醒 + 实时在线状态(ping)
- ✅ 单静态二进制,一个文件即可运行;零第三方依赖,纯 Go 标准库
- ✅ 前端页面经 `go:embed` 打包进二进制,无额外模板文件
- ✅ 配置通过 `.env` 或环境变量,**仓库不含任何真实凭据**

## Cloudflare Tunnel(公网接入)

通过 Cloudflare Tunnel 把本服务暴露到公网域名,无需公网 IP / 端口转发。
步骤如下:

1. **前提**:域名已托管在 Cloudflare;登录 [Cloudflare Zero Trust](https://one.dash.cloudflare.com)。
2. **板子装 cloudflared**:官方脚本装 `cloudflared`(arm64 版),`cloudflared --version` 验证。
3. **创建隧道**:Zero Trust → **Networks → Tunnels** → Create a tunnel → 选 Cloudflared → 记下 token。
4. **接入隧道**:`cloudflared tunnel run --token <TOKEN>`(先前台测试,确认 `Registered tunnel connection`)。
5. **配域名**:Tunnel 页 → Public Hostname → Add;Service 类型 `HTTP`、URL `localhost:5000`。
6. **常驻**:`sudo cloudflared service install <TOKEN>` + `systemctl enable --now cloudflared`。
7. **验证**:浏览器打开 `https://你的域名`,应能登录并看到在线状态。
8. **(可选)** Cloudflare Access 中为域名加一条 Access 策略,做二次访问保护。

## 快速开始

本项目按"启动方式"分两种,配置项见下节(两方式所需配置相同)。

### 方式一:从源码构建

```bash
git clone https://github.com/632-8nm/remote-wakeup.git
cd remote-wakeup
go build -o wol-web ./cmd/wol-web        # 当前平台编译
# 交叉编译到 aarch64 板子:
# GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o wol-web ./cmd/wol-web
./wol-web                                # 配置见下节，监听 5000
```

### 方式二:下载预编译二进制

从 GitHub Release 下载预编译好的 zip(内含 `wol-web`、`.env.example`、`dev.sh`):

```bash
# 下载 release 中的 wol-web-vX.Y.Z-linux-arm64.zip 并解压
unzip wol-web-vX.Y.Z-linux-arm64.zip
cd <解压目录>
cp .env.example .env      # 编辑 .env 填写配置(见下节)
./dev.sh                  # 注册 systemd 服务并启动
```

> `dev.sh` 自适应:目录中有源码(go.mod)时先构建再部署;只有预编译二进制时直接部署。

## 配置(环境变量)

必填三项,其余可选。可用环境变量注入,或写入 `.env`(方式二用 `.env`)。

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
# 源码用户：更新代码后重新构建并重启
./dev.sh

# release 用户：下载新 release zip 解压，覆盖 wol-web 后用 dev.sh 重启
./dev.sh
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

直接编辑 `.env` 中的 `ADMIN_PASSWORD`,然后重启服务生效:

```bash
# 编辑 .env 中的 ADMIN_PASSWORD=新密码，然后：
sudo systemctl restart wol-web
```

## 安全说明

- 认证用恒定时间比较,无时序侧信道。
- 会话为 HMAC 签名 Cookie,服务端零存储。
- 历史与代码中不含任何真实 MAC/IP/密码;请勿将 `.env` 提交进仓库。

## 目录结构

```
.
├── cmd/wol-web/          # 程序入口
├── internal/
│   ├── config/           # 配置加载与校验
│   ├── web/              # HTTP 服务（路由/handlers/会话/安全/模板渲染）
│   │   └── templates/    # 前端页面(go:embed)
│   └── wol/              # WOL 魔术包
├── dev.sh                # 部署脚本（自适应：源码构建或预编译二进制）
├── .env.example          # 配置模板
└── wol-web.service       # systemd 模板
```

## License

[MIT](LICENSE)
